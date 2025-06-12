package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Client     *s3.Client
	dynamoClient *dynamodb.Client
	tableName    = "dm-processing-control"
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	s3Client = s3.NewFromConfig(cfg)
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, event events.S3Event) error {

	for _, record := range event.Records {

		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		parts := strings.Split(key, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid key format: %s", key)
		}

		// Download the .gz file from S3
		resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("failed to get S3 object: %v", err)
		}
		defer resp.Body.Close()

		// Decompress the gzip content
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gz.Close()

		// Buffer the stream for dual read (record count and checksum)
		var buf bytes.Buffer
		tee := io.TeeReader(gz, &buf)

		now := time.Now().UTC().Format(time.RFC3339)

		item := core.ProcessingControl{
			ObjectKey:    key,
			SchemaName:   misc.NameWithDefaultPrefix(misc.LayerBronze, '_'),
			TableName:    parts[1],
			FileSize:     record.S3.Object.Size,
			Status:       enum.ProcessPending.String(),
			AttemptCount: 0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		// Count the number of records based on file format
		if strings.Contains(key, "dm-batch-process") {
			item.RecordCount, err = countCsvLines(tee)
			item.FileFormat = enum.FileFormatCsv.String()
		} else {
			item.RecordCount, err = countJsonObjects(tee)
			item.FileFormat = enum.FileFormatJson.String()
		}
		if err != nil {
			return fmt.Errorf("failed to count records: %v", err)
		}

		// Compute SHA256 checksum from buffered data
		item.Checksum, err = computeChecksum(bytes.NewReader(buf.Bytes()))
		if err != nil {
			return fmt.Errorf("failed to compute checksum: %v", err)
		}

		// Recommend compute target based on record count
		item.ComputeTarget = recommendCompute(item.RecordCount)

		itemMap, err := attributevalue.MarshalMap(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %v", err)
		}

		// Only insert if the object_key does not already exist
		_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName:           aws.String(tableName),
			Item:                itemMap,
			ConditionExpression: aws.String("attribute_not_exists(object_key)"),
		})
		if err != nil {
			var cfe *types.ConditionalCheckFailedException
			if errors.As(err, &cfe) {
				fmt.Printf("Item already exists, skipping: %s\n", key)
				return nil
			}
			return fmt.Errorf("failed to put item: %v", err)
		}
	}

	return nil

}

func computeChecksum(r io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func countCsvLines(r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)
	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count - 1, nil // remove header
}

func countJsonObjects(r io.Reader) (int, error) {
	count := 0
	dec := json.NewDecoder(r)
	for dec.More() {
		var obj map[string]interface{}
		if err := dec.Decode(&obj); err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

func recommendCompute(count int) string {
	switch {
	case count <= 100_000:
		return "lambda"
	default:
		return "ecs"
	}
}

func main() {
	lambda.Start(handler)
}
