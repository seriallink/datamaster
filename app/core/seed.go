package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/seriallink/datamaster/app/misc"
	"github.com/seriallink/datamaster/app/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func SeedThemAll() error {
	datasets := []string{"brewery", "beer", "profile", "review"}

	for _, name := range datasets {
		fmt.Printf("→ Seeding '%s'...\n", name)
		if err := SeedFile(name); err != nil {
			return fmt.Errorf("seeding aborted due to error on '%s': %w", name, err)
		}
	}

	return nil
}

func SeedFile(name string) error {

	url := fmt.Sprintf("https://raw.githubusercontent.com/seriallink/beer-datasets/main/%s.csv.gz", name)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status %s from GitHub", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return SeedFromReader(body, name)

}

func SeedFromReader(data []byte, name string) error {

	cfg := GetAWSConfig()
	dynamo := dynamodb.NewFromConfig(cfg)
	checksum := computeChecksum(data)

	outputs, err := (&Stack{Name: misc.StackNameControl}).GetStackOutputs(cfg)
	if err != nil {
		return fmt.Errorf("failed to get Processing stack outputs: %w", err)
	}

	// Check for duplication before proceeding
	out, err := dynamo.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(outputs["ProcessingControlTableName"]),
		IndexName:              aws.String(outputs["TableChecksumIndexName"]),
		KeyConditionExpression: aws.String("table_name = :t AND checksum = :c"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":t": &types.AttributeValueMemberS{Value: name},
			":c": &types.AttributeValueMemberS{Value: checksum},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("failed to query DynamoDB: %w", err)
	}
	if len(out.Items) > 0 {
		fmt.Printf("✔ Seed already exists for table='%s' and checksum='%s', skipping.\n", name, checksum)
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(name)) {
	case "brewery", "beer", "profile":
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to decompress: %w", err)
		}
		defer gr.Close()
		return LoadCsvToAurora(gr, name)

	case "review":
		return UploadBatchFile(bytes.NewReader(data), name)

	default:
		return fmt.Errorf("unsupported dataset: %s", name)
	}

}

func LoadCsvToAurora(r io.Reader, name string) error {

	db, err := GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	fullTableName := fmt.Sprintf("%s.%s", misc.NameWithDefaultPrefix(misc.SchemaCore, '_'), name)
	pointer, err := model.LoadModel(fullTableName)
	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}

	data, err := CsvToMap(r, pointer)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	slice, err := UnmarshalRecords(pointer, data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal records: %w", err)
	}

	if err = insertInChunks(db, slice, 1000); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	return nil

}

func UploadBatchFile(reader io.Reader, tableName string) error {

	cfg := GetAWSConfig()
	client := s3.NewFromConfig(cfg)

	bucket, err := (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "StageBucketName")
	if err != nil {
		return fmt.Errorf("failed to get StageBucketName from stack: %w", err)
	}

	id := uuid.New().String()
	timestamp := time.Now().UTC().Format("2006-01-02-15-04-05")
	key := fmt.Sprintf("raw/%s/dm-batch-process-1-%s-%s.gz", tableName, timestamp, id)

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read content body: %w", err)
	}

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(body),
		ContentLength: aws.Int64(int64(len(body))),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	fmt.Printf("Uploaded batch file to s3://%s/%s\n", bucket, key)
	return nil

}

func insertInChunks(db *gorm.DB, slice any, chunkSize int) error {
	v := reflect.ValueOf(slice)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("insertInChunks: expected slice, got %s", v.Kind())
	}
	for i := 0; i < v.Len(); i += chunkSize {
		end := i + chunkSize
		if end > v.Len() {
			end = v.Len()
		}
		chunk := v.Slice(i, end).Interface()
		if err := db.Create(chunk).Error; err != nil {
			return fmt.Errorf("insert chunk [%d:%d] failed: %w", i, end, err)
		}
	}
	return nil
}

func computeChecksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
