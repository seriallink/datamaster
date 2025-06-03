package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type ProcessingControl struct {
	ObjectKey     string  `dynamodbav:"object_key"`
	Schema        string  `dynamodbav:"schema"`
	Table         string  `dynamodbav:"table"`
	RecordCount   int     `dynamodbav:"record_count"`
	FileSize      int64   `dynamodbav:"file_size"`
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

func (m *ProcessingControl) TableName() string {
	return "dm-processing-control"
}

func (m *ProcessingControl) DestinationKey() string {

	parts := strings.Split(m.ObjectKey, "/")
	if len(parts) < 3 {
		panic(fmt.Errorf("invalid object key: %s", m.ObjectKey))
	}

	key := filepath.Join(misc.LayerBronze, parts[1], parts[2])
	key = strings.TrimSuffix(key, ".json.gz") + ".parquet"

	return key

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
