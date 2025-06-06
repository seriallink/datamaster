package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
)

func SetIcebergMetadata(ctx context.Context, layerType, tableName string, columns []string) error {

	cfg := GetAWSConfig()
	client := athena.NewFromConfig(cfg)

	dbName := NameWithPrefix(layerType)
	location, err := GetStorageLocation(layerType, tableName)
	if err != nil {
		return err
	}

	var nullFields []string
	for _, col := range columns {
		colParts := strings.Split(col, " ")
		if len(colParts) != 2 {
			return fmt.Errorf("invalid column definition: %s", col)
		}
		colName := colParts[0]
		colType := colParts[1]
		nullFields = append(nullFields, fmt.Sprintf("CAST(NULL AS %s) AS %s", colType, colName))
	}

	query := fmt.Sprintf(`
		CREATE TABLE %s.%s
		WITH (
			format = 'ICEBERG',
			location = '%s'
		)
		AS SELECT %s WHERE false
	`, dbName, tableName, location, strings.Join(nullFields, ", "))

	bucket, err := (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "ObserverBucketName")
	if err != nil {
		return fmt.Errorf("failed to get storage bucket name: %w", err)
	}

	workgroup, err := (&Stack{Name: misc.StackNameConsumption}).GetStackOutput(cfg, "AthenaWorkGroupName")
	if err != nil {
		return fmt.Errorf("failed to get Athena workgroup name: %w", err)
	}

	outputLocation := fmt.Sprintf("s3://%s/athena/", bucket)

	startOutput, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
		QueryString: aws.String(query),
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: aws.String(outputLocation),
		},
		WorkGroup: aws.String(workgroup),
	})
	if err != nil {
		return fmt.Errorf("failed to start query execution: %w", err)
	}

	queryID := *startOutput.QueryExecutionId
	for {
		time.Sleep(2 * time.Second)

		statusOutput, err := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
			QueryExecutionId: aws.String(queryID),
		})
		if err != nil {
			return fmt.Errorf("failed to get query status: %w", err)
		}

		state := statusOutput.QueryExecution.Status.State
		if state == types.QueryExecutionStateSucceeded {
			fmt.Println(misc.Green("Iceberg metadata initialized with snapshot for %s.%s", dbName, tableName))
			return nil
		}
		if state == types.QueryExecutionStateFailed || state == types.QueryExecutionStateCancelled {
			return fmt.Errorf("query failed: %s", aws.ToString(statusOutput.QueryExecution.Status.StateChangeReason))
		}
	}

}

//func SetIcebergMetadata(ctx context.Context, layerType, tableName string, columns []string) error {
//
//	cfg := GetAWSConfig()
//	client := athena.NewFromConfig(cfg)
//
//	dbName := NameWithPrefix(layerType)
//	location, err := GetStorageLocation(layerType, tableName)
//	if err != nil {
//		return err
//	}
//
//	bucket, err := (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "ObserverBucketName")
//	if err != nil {
//		return fmt.Errorf("failed to get storage bucket name: %w", err)
//	}
//
//	workgroup, err := (&Stack{Name: misc.StackNameConsumption}).GetStackOutput(cfg, "AthenaWorkGroupName")
//	if err != nil {
//		return fmt.Errorf("failed to get Athena workgroup name: %w", err)
//	}
//
//	columnDefs := strings.Join(columns, ", ")
//	query := fmt.Sprintf(`
//		CREATE TABLE IF NOT EXISTS %s.%s (
//			%s
//		)
//		LOCATION '%s'
//		TBLPROPERTIES (
//			'table_type'='ICEBERG',
//			'format'='parquet'
//		)
//	`, dbName, tableName, columnDefs, location)
//
//	outputS3 := bucket + "/athena"
//
//	startOutput, err := client.StartQueryExecution(ctx, &athena.StartQueryExecutionInput{
//		QueryString: aws.String(query),
//		ResultConfiguration: &types.ResultConfiguration{
//			OutputLocation: aws.String(outputS3),
//		},
//		WorkGroup: aws.String(workgroup),
//	})
//	if err != nil {
//		return fmt.Errorf("failed to start query execution: %w", err)
//	}
//
//	// Poll until completion
//	queryID := *startOutput.QueryExecutionId
//	for {
//		time.Sleep(2 * time.Second)
//
//		statusOutput, err := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
//			QueryExecutionId: aws.String(queryID),
//		})
//		if err != nil {
//			return fmt.Errorf("failed to get query status: %w", err)
//		}
//
//		state := statusOutput.QueryExecution.Status.State
//		if state == types.QueryExecutionStateSucceeded {
//			fmt.Println(misc.Green("Iceberg metadata initialized for %s.%s", dbName, tableName))
//			return nil
//		}
//		if state == types.QueryExecutionStateFailed || state == types.QueryExecutionStateCancelled {
//			return fmt.Errorf("query failed: %s", aws.ToString(statusOutput.QueryExecution.Status.StateChangeReason))
//		}
//	}
//
//}
