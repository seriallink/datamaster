package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/seriallink/datamaster/cli/core"
	"github.com/seriallink/datamaster/cli/enum"
	"github.com/seriallink/datamaster/cli/misc"

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

		count, err := countRecords(bucket, key)
		if err != nil {
			return fmt.Errorf("failed to count records: %v", err)
		}

		now := time.Now().UTC().Format(time.RFC3339)

		item := core.ProcessingControl{
			ObjectKey:     key,
			Schema:        misc.NameWithDefaultPrefix(misc.LayerBronze, '_'),
			Table:         parts[1],
			RecordCount:   count,
			FileSize:      record.S3.Object.Size,
			Status:        enum.ProcessPending.String(),
			AttemptCount:  0,
			ComputeTarget: recommendCompute(count),
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		itemMap, err := attributevalue.MarshalMap(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %v", err)
		}

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

func countRecords(bucket, key string) (int, error) {

	resp, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return 0, err
	}
	defer gz.Close()

	count := 0
	dec := json.NewDecoder(gz)

	for dec.More() {
		var obj map[string]interface{}
		if err = dec.Decode(&obj); err != nil {
			return 0, err
		}
		count++
	}

	return count, nil

}

func recommendCompute(count int) string {
	switch {
	case count <= 10000:
		return "lambda"
	default:
		return "ecs"
	}
}

func main() {
	lambda.Start(handler)
}
