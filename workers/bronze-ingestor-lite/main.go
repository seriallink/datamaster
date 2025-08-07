// Lightweight Lambda function responsible for transforming a raw JSON file into a Parquet dataset
// and storing it in the bronze layer of the data lake.
//
// Triggered by a Step Function or another orchestrator, it performs the following steps:
//   - Retrieves processing metadata from DynamoDB
//   - Downloads the raw file from S3
//   - Applies PII masking using AWS Comprehend
//   - Maps the payload to a typed schema
//   - Writes the result to Parquet format
//   - Uploads the output to the bronze path in the data lake
//   - Registers the next processing step (silver) in DynamoDB
//
// This function is designed for small to medium workloads and serves as
// the event-driven ingestion entry point for the bronze layer.
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/seriallink/datamaster/app/bronze"
	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

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

	err = core.QueryByObjectKey(cfg, ctx, objectKey, &item)
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

	if item.Status == enum.ProcessError.String() && item.AttemptCount >= core.GetMaxAttempts() {
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

	data, err = core.MaskPIIData(cfg, ctx, data)
	if err != nil {
		return fmt.Errorf("failed to mask PII data: %w", item.Finish(ctx, cfg, err))
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

func main() {
	lambda.Start(handler)
	return
}
