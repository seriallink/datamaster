package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const defaultMaxAttempts = 3

type ProcessingControl struct {
	ControlID     string  `dynamodbav:"control_id"`
	ObjectKey     string  `dynamodbav:"object_key"`
	SchemaName    string  `dynamodbav:"schema_name"`
	TableName     string  `dynamodbav:"table_name"`
	RecordCount   int     `dynamodbav:"record_count"`
	FileFormat    string  `dynamodbav:"file_format"`
	FileSize      int64   `dynamodbav:"file_size"`
	Checksum      string  `dynamodbav:"checksum"`
	Status        string  `dynamodbav:"status"`
	AttemptCount  int     `dynamodbav:"attempt_count"`
	ComputeTarget string  `dynamodbav:"compute_target"`
	CreatedAt     string  `dynamodbav:"created_at"`
	UpdatedAt     string  `dynamodbav:"updated_at"`
	StartedAt     *string `dynamodbav:"started_at,omitempty"`
	FinishedAt    *string `dynamodbav:"finished_at,omitempty"`
	Duration      *int64  `dynamodbav:"duration,omitempty"`
	ErrorMessage  *string `dynamodbav:"error_message,omitempty"`
}

// NewProcessingControl creates a new ProcessingControl instance based on the given data lake layer and S3 object key.
//
// Parameters:
//   - layer: the logical lakehouse layer name (e.g., "bronze", "silver", "gold").
//   - key: the S3 object key in the format "<prefix>/<table>/<filename>".
//
// Returns:
//   - *ProcessingControl: the initialized control record for tracking processing status.
//   - error: an error if the key format is invalid.
func NewProcessingControl(layer, key string) (*ProcessingControl, error) {

	parts := strings.Split(key, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid key format: %s", key)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	item := &ProcessingControl{
		ControlID:    uuid.NewString(),
		ObjectKey:    key,
		SchemaName:   misc.NameWithDefaultPrefix(layer, '_'),
		TableName:    parts[1],
		Status:       enum.ProcessPending.String(),
		AttemptCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return item, nil

}

// DestinationKey returns the destination S3 key for the processed file in the specified lakehouse layer.
//
// If the target layer is "raw", the key will preserve the flat structure and original `.gz` extension.
// For all other layers (e.g., "bronze", "silver", "gold"), the key will follow a Hive-style partitioning
// layout based on the `created_at` timestamp, and the file will be renamed with a `.parquet` extension.
//
// Parameters:
//   - targetLayer: the destination layer name ("raw", "bronze", "silver", or "gold").
//
// Returns:
//   - string: the destination S3 object key, formatted according to the layer's storage strategy.
func (m *ProcessingControl) DestinationKey(targetLayer string) string {

	parts := strings.Split(m.ObjectKey, "/")
	if len(parts) < 3 {
		panic(fmt.Errorf("invalid object key: %s", m.ObjectKey))
	}

	table := parts[1]
	filename := parts[2]
	filename = strings.TrimSuffix(filename, ".json.gz")
	filename = strings.TrimSuffix(filename, ".csv.gz")
	filename = strings.TrimSuffix(filename, ".gz")
	filename = strings.TrimSuffix(filename, ".parquet")

	if targetLayer == misc.LayerRaw {
		return filepath.Join(targetLayer, table, filename+".gz")
	}

	t, err := time.Parse(time.RFC3339, m.CreatedAt)
	if err != nil {
		panic(fmt.Errorf("invalid created_at format: %s", m.CreatedAt))
	}

	key := filepath.Join(
		targetLayer,
		table,
		fmt.Sprintf("year=%04d", t.Year()),
		fmt.Sprintf("month=%02d", int(t.Month())),
		fmt.Sprintf("day=%02d", t.Day()),
		filename+".parquet",
	)

	return key

}

// Put stores the ProcessingControl item in the DynamoDB table if it does not already exist.
// It uses a conditional expression to prevent overwriting existing items.
//
// Parameters:
//   - ctx: the context for the AWS request.
//   - cfg: the AWS configuration used to create the DynamoDB client.
//
// Returns:
//   - error: an error if the operation fails (except when the item already exists).
func (m *ProcessingControl) Put(ctx context.Context, cfg aws.Config) error {

	client := dynamodb.NewFromConfig(cfg)

	tableName, err := (&Stack{Name: misc.StackNameControl}).GetStackOutput(cfg, "ProcessingControlTableName")
	if err != nil {
		return fmt.Errorf("failed to resolve table name from stack output: %w", err)
	}

	itemMap, err := attributevalue.MarshalMap(m)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(tableName),
		Item:                itemMap,
		ConditionExpression: aws.String("attribute_not_exists(object_key)"),
	})
	if err != nil {
		var cfe *types.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			fmt.Printf("Item already exists, skipping: %s\n", m.ObjectKey)
			return nil
		}
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil

}

// Start marks the beginning of the processing execution for the current control item.
// It updates the status to "running", increments the attempt count, resets the duration and error state,
// and sets the start timestamp.
//
// Parameters:
//   - ctx: the context for the AWS request.
//   - cfg: the AWS configuration used to connect to DynamoDB.
//
// Returns:
//   - error: an error if the item fails to be persisted in DynamoDB.
func (m *ProcessingControl) Start(ctx context.Context, cfg aws.Config) error {

	now := time.Now().UTC().Format(time.RFC3339)

	m.Status = enum.ProcessRunning.String()
	m.AttemptCount++
	m.StartedAt = &now
	m.FinishedAt = nil
	m.Duration = aws.Int64(0)
	m.ErrorMessage = nil

	return PutDynamoDBItem(cfg, ctx, m)

}

// Finish finalizes the processing execution by updating the control item with completion metadata.
// It sets the finish timestamp, calculates the duration (if start time is available),
// and updates the status to "success" or "error" based on the presence of a failure.
//
// Parameters:
//   - ctx: the context for the AWS request.
//   - cfg: the AWS configuration used to connect to DynamoDB.
//   - failure: an error indicating if the execution failed; nil means success.
//
// Returns:
//   - error: the original failure if present, or an error saving the updated item to DynamoDB.
func (m *ProcessingControl) Finish(ctx context.Context, cfg aws.Config, failure error) error {

	finish := time.Now().UTC()
	finishStr := finish.Format(time.RFC3339)

	m.FinishedAt = &finishStr

	if m.StartedAt != nil {
		if start, parseErr := time.Parse(time.RFC3339, *m.StartedAt); parseErr == nil {
			m.Duration = aws.Int64(finish.Sub(start).Milliseconds())
		}
	}

	if failure != nil {
		m.Status = enum.ProcessError.String()
		m.ErrorMessage = aws.String(failure.Error())
	} else {
		m.Status = enum.ProcessSuccess.String()
		m.ErrorMessage = nil
	}

	if err := PutDynamoDBItem(cfg, ctx, m); err != nil {
		return err
	}

	return failure

}

// RegisterNextLayerControl creates and stores a new ProcessingControl entry for the next lakehouse layer.
// It derives the destination key from the current control, calculates metadata (e.g., file size, checksum),
// and sets attributes such as format, compute target, and record count.
//
// Parameters:
//   - ctx: the context for the AWS request.
//   - cfg: the AWS configuration used to connect to DynamoDB.
//   - data: the raw byte content of the transformed file.
//
// Returns:
//   - *ProcessingControl: the newly created control item for the next layer.
//   - error: an error if the control creation or persistence fails.
func (m *ProcessingControl) RegisterNextLayerControl(ctx context.Context, cfg aws.Config, data []byte) (*ProcessingControl, error) {

	if m.ObjectKey == "" {
		return nil, fmt.Errorf("invalid parent ProcessingControl: %v", m)
	}

	item, err := NewProcessingControl(misc.LayerSilver, m.DestinationKey(misc.LayerBronze))
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(data)

	item.RecordCount = m.RecordCount
	item.FileFormat = enum.FileFormatParquet.String()
	item.ComputeTarget = enum.ComputeTargetEMR.String()
	item.FileSize = int64(len(data))
	item.Checksum = hex.EncodeToString(hash[:])

	if err = item.Put(ctx, cfg); err != nil {
		return nil, err
	}

	return item, nil

}

// Delete removes the current ProcessingControl item from the DynamoDB table.
//
// It uses the control_id field as the primary key for the DeleteItem operation.
// The table name is resolved dynamically from the stack output "ProcessingControlTableName".
//
// Parameters:
//   - ctx: the context for the AWS request.
//   - cfg: the AWS configuration used to connect to DynamoDB.
//
// Returns:
//   - error: an error if the table name cannot be resolved or the item fails to be deleted.
func (m *ProcessingControl) Delete(ctx context.Context, cfg aws.Config) error {
	client := dynamodb.NewFromConfig(cfg)

	tableName, err := (&Stack{Name: misc.StackNameControl}).GetStackOutput(cfg, "ProcessingControlTableName")
	if err != nil {
		return fmt.Errorf("failed to get ProcessingControlTableName: %w", err)
	}

	_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"control_id": &types.AttributeValueMemberS{Value: m.ControlID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// GetMaxAttempts returns the maximum number of processing attempts allowed.
// It reads the MAX_ATTEMPTS environment variable and falls back to a default value
// if the variable is unset, invalid, or less than or equal to zero.
//
// Returns:
//   - int: the configured or default maximum number of attempts.
func GetMaxAttempts() int {
	val := os.Getenv("MAX_ATTEMPTS")
	if val == "" {
		return defaultMaxAttempts
	}
	if num, err := strconv.Atoi(val); err == nil && num > 0 {
		return num
	}
	return defaultMaxAttempts
}
