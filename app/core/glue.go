package core

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/glue/types"
)

func FetchGlueSchema(cfg aws.Config, ctx context.Context, database, table string) ([]types.Column, error) {

	client := glue.NewFromConfig(cfg)

	resp, err := client.GetTable(ctx, &glue.GetTableInput{
		DatabaseName: aws.String(NameWithPrefix(database)),
		Name:         aws.String(table),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get table[%s.%s] from Glue: %w", database, table, err)
	}

	if resp.Table == nil || resp.Table.StorageDescriptor == nil || resp.Table.StorageDescriptor.Columns == nil {
		return nil, fmt.Errorf("invalid table metadata from Glue")
	}

	return resp.Table.StorageDescriptor.Columns, nil

}
