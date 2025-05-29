package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/cli/misc"

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

	stageBucket, err = (&Stack{Name: misc.StackNameStorage}).GetStackOutput("StageBucketName")
	if err != nil {
		return fmt.Errorf("failed to get StageBucketName: %w", err)
	}

	lambdaArn, err = (&Stack{Name: misc.StackNameFunctions}).GetStackOutput("ProcessingControllerFunctionArn")
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
