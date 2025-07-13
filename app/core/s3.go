package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ConfigureS3Notification sets up an S3 event notification to trigger a Lambda function
// when `.gz` files are uploaded to the `raw/` prefix of the staging bucket.
//
// It first grants invoke permission to the S3 service on the specified Lambda, and then
// configures the notification rules on the bucket. If the permission already exists, it skips that step.
//
// Parameters:
//   - ctx: The context for the AWS SDK requests.
//   - s3Client: A configured S3 client used to set the bucket notification.
//   - lambdaClient: A configured Lambda client used to add invoke permissions.
//
// Returns:
//   - error: Any error encountered while configuring the notification or permission. Returns nil if successful.
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

// UploadDataToS3 uploads a byte slice to a specified S3 bucket and key.
//
// It creates an S3 client from the provided AWS config and sends the data using `PutObject`.
//
// Parameters:
//   - cfg: The AWS configuration used to create the S3 client.
//   - ctx: The context for the upload operation.
//   - bucket: The name of the S3 bucket.
//   - key: The key (path) where the object will be stored.
//   - data: The byte slice to be uploaded.
//
// Returns:
//   - error: Any error encountered during the upload. Returns nil if the upload is successful.
func UploadDataToS3(cfg aws.Config, ctx context.Context, bucket, key string, data []byte) error {

	_, err := s3.NewFromConfig(cfg).PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object to S3: %w", err)
	}

	return nil

}

// DownloadS3Object retrieves an object from an S3 bucket using the given bucket and key.
//
// It uses the provided AWS configuration to create an S3 client and performs a `GetObject` call.
//
// Parameters:
//   - ctx: Context for the request.
//   - cfg: AWS configuration used to create the S3 client.
//   - bucket: Name of the S3 bucket.
//   - key: Key (path) of the object to retrieve.
//
// Returns:
//   - *s3.GetObjectOutput: The S3 object output, including metadata and the object's body.
//   - error: Any error encountered during the download operation. Returns nil if successful.
func DownloadS3Object(ctx context.Context, cfg aws.Config, bucket, key string) (*s3.GetObjectOutput, error) {
	client := s3.NewFromConfig(cfg)
	object, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return object, nil
}

// LoadRawS3Data downloads and parses a raw data file from S3 based on the metadata in ProcessingControl.
//
// The file is expected to be compressed with gzip and stored in the `StageBucket`. It supports CSV and JSON formats.
//
// Parameters:
//   - cfg: AWS configuration used to access S3.
//   - ctx: Context for the request.
//   - item: ProcessingControl containing metadata like ObjectKey and FileFormat.
//
// Returns:
//   - []map[string]any: A slice of parsed records, where each record is a key-value map.
//   - error: Any error encountered during the process (S3 access, decompression, or parsing).
func LoadRawS3Data(cfg aws.Config, ctx context.Context, item *ProcessingControl) ([]map[string]any, error) {

	bucket, err := (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "StageBucketName")
	if err != nil {
		return nil, fmt.Errorf("failed to get StageBucketName: %w", err)
	}

	object, err := DownloadS3Object(ctx, cfg, bucket, item.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get object from s3: %w", err)
	}
	defer object.Body.Close()

	reader, err := gzip.NewReader(object.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	switch item.FileFormat {
	case enum.FileFormatCsv.String():
		return parseCSV(reader)
	case enum.FileFormatJson.String():
		return parseJSON(reader)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", item.FileFormat)
	}

}

// parseCSV reads and parses a CSV input stream into a slice of maps.
//
// Each row in the CSV is converted to a map where the keys are column headers
// and the values are the corresponding cell values (as strings).
//
// Parameters:
//   - r: io.Reader providing the CSV data.
//
// Returns:
//   - []map[string]any: Slice of maps representing the parsed CSV records.
//   - error: Any error encountered while reading or parsing the CSV.
func parseCSV(r io.Reader) ([]map[string]any, error) {

	var data []map[string]any

	csvReader := csv.NewReader(r)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}
		record := make(map[string]any)
		for i, col := range line {
			record[headers[i]] = col
		}
		data = append(data, record)
	}

	return data, nil

}

// parseJSON reads and parses a stream of JSON objects into a slice of maps.
//
// The input is expected to be a JSON stream where each line or object represents
// a single record (e.g., NDJSON format or array of objects).
//
// Parameters:
//   - r: io.Reader providing the JSON data.
//
// Returns:
//   - []map[string]any: Slice of maps representing the decoded JSON records.
//   - error: Any error encountered while decoding the JSON.
func parseJSON(r io.Reader) ([]map[string]any, error) {

	var data []map[string]any

	dec := json.NewDecoder(r)
	for dec.More() {

		var record map[string]any
		if err := dec.Decode(&record); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}

		row := make(map[string]any, len(record))
		for k, v := range record {
			row[k] = v
		}

		data = append(data, row)

	}

	return data, nil

}
