package core

import (
	"context"
	"fmt"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func QueryByObjectKey(cfg aws.Config, ctx context.Context, objectKey string, md any) error {

	var (
		err       error
		tableName string
		indexName string
		output    *dynamodb.QueryOutput
	)

	client := dynamodb.NewFromConfig(cfg)

	tableName, err = (&Stack{Name: misc.StackNameControl}).GetStackOutput(cfg, "ProcessingControlTableName")
	if err != nil {
		return fmt.Errorf("failed to get ProcessingControlTableName: %w", err)
	}

	indexName, err = (&Stack{Name: misc.StackNameControl}).GetStackOutput(cfg, "ObjectKeyIndexName")
	if err != nil {
		return fmt.Errorf("failed to get ObjectKeyIndexName: %w", err)
	}

	output, err = client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String("object_key = :v"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v": &types.AttributeValueMemberS{Value: objectKey},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("failed to query item by object_key: %w", err)
	}
	if len(output.Items) == 0 {
		return fmt.Errorf("no record found for object_key: %s", objectKey)
	}

	err = attributevalue.UnmarshalMap(output.Items[0], md)
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
