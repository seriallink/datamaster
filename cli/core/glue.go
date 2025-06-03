package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
)

func FetchGlueSchema(cfg aws.Config, ctx context.Context, database, table string) (*arrow.Schema, error) {

	client := glue.NewFromConfig(cfg)

	resp, err := client.GetTable(ctx, &glue.GetTableInput{
		DatabaseName: &database,
		Name:         &table,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get table from Glue: %w", err)
	}

	if resp.Table == nil || resp.Table.StorageDescriptor == nil || resp.Table.StorageDescriptor.Columns == nil {
		return nil, fmt.Errorf("invalid table metadata from Glue")
	}

	var fields []arrow.Field
	for _, col := range resp.Table.StorageDescriptor.Columns {
		colType := "string"
		if col.Type != nil {
			colType = *col.Type
		}
		dt, err := glueTypeToArrow(colType)
		if err != nil {
			return nil, fmt.Errorf("failed to map type %s: %w", colType, err)
		}
		fields = append(fields, arrow.Field{Name: *col.Name, Type: dt, Nullable: true})
	}

	return arrow.NewSchema(fields, nil), nil

}

func glueTypeToArrow(glueType string) (arrow.DataType, error) {
	switch strings.ToLower(glueType) {
	case "string", "uuid":
		return arrow.BinaryTypes.String, nil
	case "int", "integer":
		return arrow.PrimitiveTypes.Int32, nil
	case "bigint":
		return arrow.PrimitiveTypes.Int64, nil
	case "float":
		return arrow.PrimitiveTypes.Float32, nil
	case "double":
		return arrow.PrimitiveTypes.Float64, nil
	case "boolean":
		return arrow.FixedWidthTypes.Boolean, nil
	case "timestamp", "timestamptz":
		return arrow.FixedWidthTypes.Timestamp_us, nil
	case "date":
		return arrow.FixedWidthTypes.Date32, nil
	case "binary":
		return arrow.BinaryTypes.Binary, nil
	default:
		return nil, fmt.Errorf("unsupported Glue type: %s", glueType)
	}
}
