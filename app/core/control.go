package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ProcessingControl struct {
	ObjectKey     string  `dynamodbav:"object_key"`
	ParentKey     *string `dynamodbav:"parent_key,omitempty"`
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

func NewProcessingControl(layer, key string) (*ProcessingControl, error) {

	parts := strings.Split(key, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid key format: %s", key)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	item := &ProcessingControl{
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

func (m *ProcessingControl) DestinationKey(targetLayer string) string {

	parts := strings.Split(m.ObjectKey, "/")
	if len(parts) < 3 {
		panic(fmt.Errorf("invalid object key: %s", m.ObjectKey))
	}

	filename := parts[2]
	filename = strings.TrimSuffix(filename, ".json.gz")
	filename = strings.TrimSuffix(filename, ".csv.gz")
	filename = strings.TrimSuffix(filename, ".gz")
	filename = strings.TrimSuffix(filename, ".parquet")

	key := filepath.Join(targetLayer, parts[1], filename+".parquet")
	return key

}

func (m *ProcessingControl) Put(ctx context.Context, cfg aws.Config) error {

	client := dynamodb.NewFromConfig(cfg)

	tableName, err := (&Stack{Name: misc.StackNameProcessing}).GetStackOutput(cfg, "ProcessingControlTableName")
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

func (m *ProcessingControl) RegisterNextLayerControl(ctx context.Context, cfg aws.Config, data []byte) (*ProcessingControl, error) {

	if m.ObjectKey == "" {
		return nil, fmt.Errorf("invalid parent ProcessingControl: %v", m)
	}

	item, err := NewProcessingControl(misc.LayerSilver, m.DestinationKey(misc.LayerBronze))
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(data)

	item.ParentKey = aws.String(m.ObjectKey)
	item.RecordCount = m.RecordCount
	item.FileFormat = "parquet"
	item.ComputeTarget = "glue"
	item.FileSize = int64(len(data))
	item.Checksum = hex.EncodeToString(hash[:])

	if err = item.Put(ctx, cfg); err != nil {
		return nil, err
	}

	return item, nil

}
