package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/seriallink/datamaster/app/bronze"
	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

const defaultMaxAttempts = 3

type Event struct {
	ObjectKey string `json:"object_key"`
}

func Process(ctx context.Context, cfg aws.Config, objectKey string) error {

	var (
		err     error
		item    core.ProcessingControl
		data    []map[string]any
		model   bronze.Model
		records any
		bucket  string
	)

	err = core.GetUnmarshalledDynamoDBItem(cfg, ctx, objectKey, &item)
	if err != nil {
		if strings.Contains(err.Error(), "object_key not found") {
			log.Printf("Object %s not found in DynamoDB, skipping processing", objectKey)
			return nil
		}
		return fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	if item.Status == enum.ProcessSuccess.String() || item.Status == enum.ProcessRunning.String() {
		log.Printf("Skipping object %s with status: %s", item.ObjectKey, item.Status)
		return nil
	}

	if item.Status == enum.ProcessError.String() && item.AttemptCount >= getMaxAttempts() {
		log.Printf("Skipping object %s: exceeded max retries (%d)", item.ObjectKey, item.AttemptCount)
		return nil
	}

	err = item.Start(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to start processing item: %w", err)
	}

	data, err = core.LoadRawS3Data(cfg, ctx, &item)
	if err != nil {
		return fmt.Errorf("failed to load raw data from S3: %w", item.Finish(ctx, cfg, err))
	}

	model, err = bronze.LoadModel(fmt.Sprintf("%s.%s", item.SchemaName, item.TableName))
	if err != nil {
		return fmt.Errorf("failed to load model: %w", item.Finish(ctx, cfg, err))
	}

	records, err = core.UnmarshalRecords(model, data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal records: %w", item.Finish(ctx, cfg, err))
	}

	buf := &bytes.Buffer{}
	err = core.WriteParquet(records, buf)
	if err != nil {
		return fmt.Errorf("failed to write Parquet: %w", item.Finish(ctx, cfg, err))
	}

	bucket, err = (&core.Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "DataLakeBucketName")
	if err != nil {
		return fmt.Errorf("failed to get bucket name from stack output: %w", item.Finish(ctx, cfg, err))
	}

	err = core.UploadDataToS3(cfg, ctx, bucket, item.DestinationKey(misc.LayerBronze), buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to upload data to S3: %w", item.Finish(ctx, cfg, err))
	}

	_, err = item.RegisterNextLayerControl(ctx, cfg, buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to register next layer control: %w", item.Finish(ctx, cfg, err))
	}

	log.Printf("Successfully processed object %s and uploaded to %s/%s\n", objectKey, bucket, item.DestinationKey(misc.LayerBronze))

	return item.Finish(ctx, cfg, nil)

}

func handler(ctx context.Context, event Event) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	return Process(ctx, cfg, event.ObjectKey)
}

func getMaxAttempts() int {
	val := os.Getenv("MAX_ATTEMPTS")
	if val == "" {
		return defaultMaxAttempts
	}
	if num, err := strconv.Atoi(val); err == nil && num > 0 {
		return num
	}
	return defaultMaxAttempts
}

func main() {
	lambda.Start(handler)
	return
}
