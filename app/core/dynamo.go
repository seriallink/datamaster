package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// QueryByObjectKey retrieves a single item from the ProcessingControl table using the object_key GSI.
// It queries the appropriate DynamoDB index and unmarshals the result into the provided destination.
//
// Parameters:
//   - cfg: the AWS configuration used to connect to DynamoDB.
//   - ctx: the context for the AWS request.
//   - objectKey: the S3 object key to query.
//   - md: a pointer to a struct where the result will be unmarshaled.
//
// Returns:
//   - error: an error if the query fails, no record is found, or unmarshaling fails.
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

// PutDynamoDBItem stores a ProcessingControl item in the DynamoDB table "dm-processing-control".
// It marshals the item into a DynamoDB attribute map and performs a PutItem operation.
//
// Parameters:
//   - cfg: the AWS configuration used to connect to DynamoDB.
//   - ctx: the context for the AWS request.
//   - item: the ProcessingControl struct to persist.
//
// Returns:
//   - error: an error if the item fails to be marshaled or inserted into DynamoDB.
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

// FetchPendingTables returns a list of unique table names that have pending items
// for the given schema name in the dm-processing-control table.
//
// Parameters:
//   - cfg: AWS configuration used to connect to DynamoDB.
//   - schema: like "dm_silver" or "dm_gold"
//
// Returns:
//   - []string: list of table names
//   - error: in case of failure
func FetchPendingTables(cfg aws.Config, schema string) ([]string, error) {

	client := dynamodb.NewFromConfig(cfg)

	output, err := (&Stack{Name: misc.StackNameControl}).GetStackOutputs(cfg)
	if err != nil {
		return nil, err
	}

	exprAttrNames := map[string]string{
		"#schema": "schema_name",
		"#status": "status",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":s": &types.AttributeValueMemberS{Value: schema},
		":p": &types.AttributeValueMemberS{Value: enum.ProcessPending.String()},
	}

	out, err := client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(output["ProcessingControlTableName"]),
		IndexName:                 aws.String(output["SchemaStatusIndexName"]),
		KeyConditionExpression:    aws.String("#schema = :s AND #status = :p"),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
	}

	tableSet := make(map[string]bool)
	for _, item := range out.Items {
		if attr, ok := item["table_name"]; ok {
			if sAttr, ok := attr.(*types.AttributeValueMemberS); ok {
				table := strings.TrimSpace(sAttr.Value)
				if table != "" {
					tableSet[table] = true
				}
			}
		}
	}

	var result []string
	for table := range tableSet {
		result = append(result, table)
	}

	return result, nil

}
