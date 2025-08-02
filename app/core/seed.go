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

// SeedThemAll seeds all predefined datasets sequentially.
//
// This function iterates over a hardcoded list of dataset names and calls
// `SeedFile` for each one. If any dataset fails to seed, the process is aborted.
//
// Returns:
//   - error: An error if any seeding operation fails; otherwise, nil.
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

// SeedFile downloads a compressed CSV dataset from GitHub and seeds it.
//
// Parameters:
//   - name: Name of the dataset (e.g., "brewery", "beer").
//
// Returns:
//   - error: An error if the download, read, or seeding fails; otherwise, nil.
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

// SeedFromReader seeds a dataset into Aurora or uploads it to S3, depending on the table name.
//
// It checks for duplicates by querying the ProcessingControl table using the schema ("dm_bronze")
// and table name. If an entry already exists, the operation is skipped.
//
// Parameters:
//   - data: Compressed CSV content in bytes (gzip).
//   - name: Dataset name. Must be one of: "brewery", "beer", "profile", or "review".
//
// Returns:
//   - error: An error if decompression, duplicate check, or seeding/upload fails.
func SeedFromReader(data []byte, name string) error {

	cfg := GetAWSConfig()
	dynamo := dynamodb.NewFromConfig(cfg)

	outputs, err := (&Stack{Name: misc.StackNameControl}).GetStackOutputs(cfg)
	if err != nil {
		return fmt.Errorf("failed to get Processing stack outputs: %w", err)
	}

	out, err := dynamo.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(outputs["ProcessingControlTableName"]),
		IndexName:              aws.String(outputs["SchemaTableIndexName"]),
		KeyConditionExpression: aws.String("schema_name = :s AND table_name = :t"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": &types.AttributeValueMemberS{Value: "dm_bronze"},
			":t": &types.AttributeValueMemberS{Value: name},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("failed to query DynamoDB: %w", err)
	}
	if len(out.Items) > 0 {
		fmt.Printf(misc.Yellow("Seed already exists for schema='dm_bronze' and table='%s', skipping.\n", name))
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

// LoadCsvToAurora reads CSV data from a reader, parses it into a slice of records,
// and inserts the data into the corresponding Aurora PostgreSQL table.
//
// Parameters:
//   - r: Reader containing CSV data (typically a gzip-decompressed CSV).
//   - name: Name of the table to insert the data into (e.g., "brewery", "beer").
//
// Returns:
//   - error: If any step fails (e.g., DB connection, model loading, CSV parsing, or insertion), an error is returned.
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

// UploadBatchFile uploads a batch data file to the raw ingestion path in S3.
//
// The uploaded object is placed under the "raw/{tableName}/" prefix with a
// unique name based on timestamp and UUID.
//
// Parameters:
//   - reader: io.Reader providing the compressed file content (e.g., gzip).
//   - tableName: Name of the logical table associated with the batch file.
//
// Returns:
//   - error: If the upload fails due to stack resolution, read error, or S3 operation.
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

// insertInChunks inserts a slice of records into the database in batches.
//
// This function splits the input slice into chunks of size `chunkSize` and performs
// batched inserts using GORM's `Create` method to improve performance and reduce
// memory usage on large datasets.
//
// Parameters:
//   - db: *gorm.DB - Active GORM database connection.
//   - slice: any - Pointer to a slice of records to be inserted.
//   - chunkSize: int - Number of records per batch.
//
// Returns:
//   - error: If the input is not a slice or any insert operation fails.
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

// computeChecksum calculates the SHA-256 checksum of the given data.
//
// Parameters:
//   - data: []byte - The input byte slice to hash.
//
// Returns:
//   - string: Hexadecimal-encoded SHA-256 hash of the input.
func computeChecksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
