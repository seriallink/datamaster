// Lambda function responsible for initializing the bronze ingestion pipeline
// by registering metadata about new raw files in the ProcessingControl table.
//
// Triggered by S3 Event Notifications, it performs the following actions:
//   - Decompresses the uploaded .csv.gz or .json.gz file
//   - Counts the number of records (CSV or line-delimited JSON)
//   - Computes a SHA-256 checksum of the uncompressed content
//   - Estimates the file's compute target (Lambda or ECS)
//   - Stores a new ProcessingControl entry in DynamoDB, including record count,
//     format, checksum, file size, and destination layer
//
// This function ensures that all incoming files are tracked, validated, and
// prepared for downstream ingestion into the bronze layer.
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/enum"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

var cfg aws.Config

func init() {
	cfg, _ = config.LoadDefaultConfig(context.TODO())
}

func handler(ctx context.Context, event events.S3Event) error {

	for _, record := range event.Records {

		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		object, err := core.DownloadS3Object(ctx, cfg, bucket, key)
		if err != nil {
			return fmt.Errorf("failed to get S3 object: %v", err)
		}
		defer object.Body.Close()

		gz, err := gzip.NewReader(object.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gz.Close()

		// Buffer the stream for dual read (record count and checksum)
		var buf bytes.Buffer
		tee := io.TeeReader(gz, &buf)

		item, err := core.NewProcessingControl(misc.LayerBronze, key)
		if err != nil {
			return fmt.Errorf("failed to init processing control item: %v", err)
		}

		if strings.Contains(key, "dm-batch-process") {
			item.RecordCount, err = countCsvLines(tee)
			item.FileFormat = enum.FileFormatCsv.String()
		} else {
			item.RecordCount, err = countJsonObjects(tee)
			item.FileFormat = enum.FileFormatJson.String()
		}
		if err != nil {
			return fmt.Errorf("failed to count records: %v", err)
		}

		item.Checksum, err = computeChecksum(bytes.NewReader(buf.Bytes()))
		if err != nil {
			return fmt.Errorf("failed to compute checksum: %v", err)
		}

		item.FileSize = record.S3.Object.Size
		item.ComputeTarget = recommendCompute(item.RecordCount)

		err = item.Put(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to put item in DynamoDB: %v", err)
		}

	}

	return nil

}

func computeChecksum(r io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func countCsvLines(r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)
	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count - 1, nil // remove header
}

func countJsonObjects(r io.Reader) (int, error) {
	count := 0
	dec := json.NewDecoder(r)
	for dec.More() {
		var obj map[string]interface{}
		if err := dec.Decode(&obj); err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

func recommendCompute(count int) string {
	switch {
	case count <= 100_000:
		return enum.ComputeTargetLambda.String()
	default:
		return enum.ComputeTargetECS.String()
	}
}

func main() {
	lambda.Start(handler)
}
