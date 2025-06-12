package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/seriallink/datamaster/app/bronze"
	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func Process(ctx context.Context, cfg aws.Config, objectKey string) error {

	item := core.ProcessingControl{}
	err := core.GetUnmarshalledDynamoDBItem(cfg, ctx, objectKey, &item)
	if err != nil {
		return fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	if item.Status == enum.ProcessSuccess.String() || item.Status == enum.ProcessRunning.String() {
		log.Printf("Skipping object %s with status: %s", item.ObjectKey, item.Status)
		return nil
	}

	if err = item.Start(ctx, cfg); err != nil {
		return fmt.Errorf("failed to mark processing start: %w", err)
	}

	bucket, err := (&core.Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "StageBucketName")
	if err != nil {
		return fmt.Errorf("failed to get StageBucketName: %w", err)
	}

	object, err := core.DownloadS3Object(ctx, cfg, bucket, item.ObjectKey)
	if err != nil {
		return fmt.Errorf("failed to download gzip object: %w", item.Finish(ctx, cfg, err))
	}
	defer object.Body.Close()

	reader, err := gzip.NewReader(object.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", item.Finish(ctx, cfg, err))
	}
	defer reader.Close()

	model, err := bronze.LoadModel(fmt.Sprintf("%s.%s", item.SchemaName, item.TableName))
	if err != nil {
		return fmt.Errorf("failed to load model: %w", item.Finish(ctx, cfg, err))
	}

	buf, err := dispatchStreamProcessing(item.FileFormat, reader, model)
	if err != nil {
		return fmt.Errorf("failed to process file format %s: %w", item.FileFormat, item.Finish(ctx, cfg, err))
	}

	bucket, err = (&core.Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "DataLakeBucketName")
	if err != nil {
		return fmt.Errorf("failed to get bucket name from stack output: %w", item.Finish(ctx, cfg, err))
	}

	err = core.UploadDataToS3(cfg, ctx, bucket, item.DestinationKey(), buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to upload data to S3: %w", item.Finish(ctx, cfg, err))
	}

	log.Printf("Successfully processed object %s and uploaded to %s/%s\n", objectKey, bucket, item.DestinationKey())

	return item.Finish(ctx, cfg, nil)

}

func dispatchStreamProcessing(format string, r io.Reader, model bronze.Model) (*bytes.Buffer, error) {
	switch format {
	case enum.FileFormatCsv.String():
		return core.StreamCsvToParquet(r, model)
	case enum.FileFormatJson.String():
		return core.StreamJsonToParquet(r, model)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", format)
	}
}

func main() {

	objectKey := os.Getenv("OBJECT_KEY")
	if objectKey == "" {
		log.Fatal("Missing OBJECT_KEY environment variable")
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	if err = Process(ctx, cfg, objectKey); err != nil {
		log.Fatalf("processing failed: %v", err)
	}

}
