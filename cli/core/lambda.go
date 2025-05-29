package core

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// DeployLambdaFromArtifact uploads the specified Lambda ZIP artifact to S3 and deploys the function.
// If the function does not exist, it is created. If it exists, its code and configuration are updated.
//
// Parameters:
//   - artifacts: an embedded filesystem containing the ZIP artifacts.
//   - name: the name of the Lambda function (must match the artifact filename without extension).
//   - memory: the memory size in MB to assign to the Lambda.
//   - timeout: the timeout in seconds to assign to the Lambda.
//
// Returns:
//   - error: if any step in the deployment process fails.
func DeployLambdaFromArtifact(artifacts embed.FS, name string, memory, timeout int) error {

	var (
		err    error
		role   string
		bucket string
	)

	ctx := context.TODO()
	client := lambda.NewFromConfig(GetAWSConfig())

	// Upload the artifact for the specific function
	if err = UploadArtifacts(artifacts, name); err != nil {
		return fmt.Errorf("failed to upload artifact: %w", err)
	}

	role, err = (&Stack{Name: misc.StackNameRoles}).GetStackOutput("LambdaExecutionRoleArn")
	if err != nil {
		return fmt.Errorf("failed to get LambdaExecutionRoleArn: %w", err)
	}

	bucket, err = (&Stack{Name: misc.StackNameStorage}).GetStackOutput("ArtifactsBucketName")
	if err != nil {
		return fmt.Errorf("failed to get ArtifactsBucketName: %w", err)
	}

	// Ensures the function name follows the project naming convention using the default prefix
	funcName := fmt.Sprintf("%s-%s", misc.DefaultProjectPrefix, name)

	// Check if function exists
	_, err = client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(funcName),
	})

	if err != nil && strings.Contains(err.Error(), "ResourceNotFoundException") {
		fmt.Printf("Creating Lambda function: %s\n", funcName)
		_, err = client.CreateFunction(ctx, &lambda.CreateFunctionInput{
			FunctionName: aws.String(funcName),
			Role:         aws.String(role),
			Runtime:      types.RuntimeProvidedal2023,
			Handler:      aws.String("bootstrap"),
			Description:  aws.String(lambdaDescription(name)),
			Code: &types.FunctionCode{
				S3Bucket: aws.String(bucket),
				S3Key:    aws.String(name + ".zip"),
			},
			MemorySize: aws.Int32(int32(memory)),
			Timeout:    aws.Int32(int32(timeout)),
			Architectures: []types.Architecture{
				types.ArchitectureX8664,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create lambda: %w", err)
		}
	} else if err == nil {
		fmt.Printf("Updating Lambda function: %s\n", name)

		_, err = client.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
			FunctionName: aws.String(funcName),
			S3Bucket:     aws.String(bucket),
			S3Key:        aws.String(name + ".zip"),
		})
		if err != nil {
			return fmt.Errorf("failed to update lambda code: %w", err)
		}

		// Wait until the function is ready for config update
		if err = waitForLambdaReady(ctx, client, funcName, 10); err != nil {
			return fmt.Errorf("lambda not ready for config update: %w", err)
		}

		_, err = client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(funcName),
			Description:  aws.String(lambdaDescription(name)),
			MemorySize:   aws.Int32(int32(memory)),
			Timeout:      aws.Int32(int32(timeout)),
		})
		if err != nil {
			return fmt.Errorf("failed to update lambda config: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check lambda: %w", err)
	}

	return nil

}

// UploadArtifacts uploads one or more embedded Lambda ZIP artifacts to the configured S3 bucket.
//
// Parameters:
//   - artifacts: an embedded filesystem containing the ZIP artifacts.
//   - functions: optional list of function names to upload (without ".zip"); if empty, all ZIPs are uploaded.
//
// Returns:
//   - error: if reading embedded files or uploading to S3 fails.
func UploadArtifacts(artifacts embed.FS, functions ...string) error {

	var (
		err    error
		bucket string
		files  []fs.DirEntry
	)

	client := s3.NewFromConfig(GetAWSConfig())

	bucket, err = (&Stack{Name: misc.StackNameStorage}).GetStackOutput("ArtifactsBucketName")
	if err != nil {
		return fmt.Errorf("ArtifactsBucketName not found in stack outputs")
	}

	files, err = artifacts.ReadDir(misc.ArtifactsPath)
	if err != nil {
		return fmt.Errorf("failed to read artifacts directory: %w", err)
	}

	// Create a map for a quick lookup if specific functions are passed
	only := map[string]bool{}
	if len(functions) > 0 {
		for _, f := range functions {
			only[f+".zip"] = true
		}
	}

	for _, file := range files {

		var content []byte

		name := file.Name()

		if file.IsDir() || !strings.HasSuffix(name, ".zip") {
			continue
		}

		// If specific functions were provided, skip others
		if len(only) > 0 && !only[name] {
			continue
		}

		embeddedPath := path.Join(misc.ArtifactsPath, name)

		content, err = artifacts.ReadFile(embeddedPath)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", embeddedPath, err)
		}

		fmt.Printf("Uploading embedded %s to s3://%s/%s...\n", name, bucket, name)

		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(name),
			Body:   bytes.NewReader(content),
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", name, err)
		}
	}

	return nil

}

// lambdaDescription returns a human-readable description for a given Lambda function name.
// Used to populate the Description field when deploying functions.
//
// Parameters:
//   - name: the logical name of the Lambda function (without prefix).
//
// Returns:
//   - string: a description to be used in AWS Lambda metadata.
func lambdaDescription(name string) string {
	switch name {
	case "firehose-processor":
		return "Sets the Firehose prefix dynamically based on the table-name field"
	case "validate-contract":
		return "Validates business contract rules before ingestion"
	default:
		return fmt.Sprintf("Lambda function: %s", name)
	}
}

// waitForLambdaReady polls the AWS Lambda function configuration until it reaches a stable state.
//
// Parameters:
//   - ctx: context for cancellation and timeout control.
//   - client: Lambda client used to fetch function configuration.
//   - name: the name of the Lambda function.
//   - maxAttempts: the maximum number of polling attempts.
//
// Returns:
//   - nil if the function reaches a 'Successful' update state.
//   - error if the function does not stabilize within the allowed attempts or if an API call fails.
func waitForLambdaReady(ctx context.Context, client *lambda.Client, name string, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		resp, err := client.GetFunctionConfiguration(ctx, &lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(name),
		})
		if err != nil {
			return err
		}

		if resp.LastUpdateStatus == types.LastUpdateStatusSuccessful {
			return nil
		}

		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout waiting for Lambda function to stabilize")
}
