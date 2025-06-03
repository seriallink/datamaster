package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func ConfigureS3Notification(ctx context.Context, s3Client *s3.Client, lambdaClient *lambda.Client) error {

	var (
		err         error
		stageBucket string
		lambdaArn   string
	)

	cfg := GetAWSConfig()

	stageBucket, err = (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "StageBucketName")
	if err != nil {
		return fmt.Errorf("failed to get StageBucketName: %w", err)
	}

	lambdaArn, err = (&Stack{Name: misc.StackNameFunctions}).GetStackOutput(cfg, "ProcessingControllerFunctionArn")
	if err != nil {
		return fmt.Errorf("failed to get ProcessingControllerFunctionArn: %w", err)
	}

	// Grant permission for S3 to invoke the Lambda
	_, err = lambdaClient.AddPermission(ctx, &lambda.AddPermissionInput{
		FunctionName: aws.String(lambdaArn),
		StatementId:  aws.String("allow-s3-invoke"),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("s3.amazonaws.com"),
		SourceArn:    aws.String("arn:aws:s3:::" + stageBucket),
	})
	if err != nil {
		if strings.Contains(err.Error(), "ResourceConflictException") {
			fmt.Println(misc.Yellow("Lambda permission already exists. Skipping."))
		} else {
			return fmt.Errorf("failed to add lambda invoke permission: %w", err)
		}
	} else {
		fmt.Println(misc.Green("Lambda invoke permission added."))
	}

	// Configure S3 Notification
	_, err = s3Client.PutBucketNotificationConfiguration(ctx, &s3.PutBucketNotificationConfigurationInput{
		Bucket: aws.String(stageBucket),
		NotificationConfiguration: &types.NotificationConfiguration{
			LambdaFunctionConfigurations: []types.LambdaFunctionConfiguration{
				{
					Id:                aws.String("raw-upload-notification"),
					LambdaFunctionArn: aws.String(lambdaArn),
					Events: []types.Event{
						types.EventS3ObjectCreatedPut,
						types.EventS3ObjectCreatedPost,
						types.EventS3ObjectCreatedCopy,
						types.EventS3ObjectCreatedCompleteMultipartUpload,
					},
					Filter: &types.NotificationConfigurationFilter{
						Key: &types.S3KeyFilter{
							FilterRules: []types.FilterRule{
								{Name: types.FilterRuleNamePrefix, Value: aws.String("raw/")},
								{Name: types.FilterRuleNameSuffix, Value: aws.String(".gz")},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to set bucket notification for bucket %s and lambda %s: %w", stageBucket, lambdaArn, err)
	}

	fmt.Println("S3 notification successfully configured.")
	return nil

}

func LoadRawS3Data(cfg aws.Config, ctx context.Context, objectKey string) ([]map[string]any, error) {

	var (
		err    error
		bucket string
		object *s3.GetObjectOutput
		reader *gzip.Reader
		data   []map[string]any
	)

	client := s3.NewFromConfig(cfg)

	bucket, err = (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "StageBucketName")
	if err != nil {
		return nil, fmt.Errorf("failed to get StageBucketName: %w", err)
	}

	object, err = client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from s3: %w", err)
	}
	defer object.Body.Close()

	reader, err = gzip.NewReader(object.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	dec := json.NewDecoder(reader)
	for dec.More() {
		var record map[string]any
		if err = dec.Decode(&record); err != nil {
			return nil, fmt.Errorf("failed to decode payload: %w", err)
		}
		data = append(data, record)
	}

	return data, nil

}

func UploadDataToS3(cfg aws.Config, ctx context.Context, bucket, key string, data []byte) error {

	_, err := s3.NewFromConfig(cfg).PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to upload Parquet to S3: %w", err)
	}

	return nil

}
