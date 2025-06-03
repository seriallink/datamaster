package core

import (
	"context"
	"fmt"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func GetUnmarshalledDynamoDBItem(cfg aws.Config, ctx context.Context, key string, md any) error {

	var (
		err       error
		tableName string
		output    *dynamodb.GetItemOutput
	)

	client := dynamodb.NewFromConfig(cfg)

	tableName, err = (&Stack{Name: misc.StackNameProcessing}).GetStackOutput(cfg, "ProcessingControlTableName")
	if err != nil {
		return fmt.Errorf("failed to get ProcessingControlTableName: %w", err)
	}

	output, err = client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"object_key": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if output.Item == nil {
		return fmt.Errorf("object_key not found: %s", key)
	}

	err = attributevalue.UnmarshalMap(output.Item, md)
	if err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return nil

}

func PutDynamoDBItem(cfg aws.Config, ctx context.Context, item *ProcessingControl) error {

	client := dynamodb.NewFromConfig(cfg)

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("dm-processing-control"),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	return nil

}
