package core

import (
	"context"
	"embed"
	"fmt"
	"path"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// UploadEtlAssets uploads the ETL assets (e.g., main.py and bundle.zip) from the embedded filesystem to the S3 artifacts bucket.
//
// Parameters:
//   - cfg: the AWS configuration used to access S3 and resolve the bucket name.
//   - assets: the embedded filesystem containing the ETL asset files.
//
// Returns:
//   - error: an error if any of the assets fail to be read or uploaded.
func UploadEtlAssets(cfg aws.Config, assets embed.FS) error {

	files := []string{
		"etl/main.py",
		"etl/bundle.zip",
	}

	bucket, err := (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "ArtifactsBucketName")
	if err != nil {
		return fmt.Errorf("failed to get ArtifactsBucketName: %w", err)
	}

	for _, file := range files {

		data, err := assets.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		key := path.Join("etl", path.Base(file))

		err = UploadDataToS3(cfg, context.TODO(), bucket, key, data)
		if err != nil {
			return fmt.Errorf("failed to upload %s to S3: %w", file, err)
		}

		fmt.Printf("Uploaded %s to s3://%s/%s\n", file, bucket, key)

	}

	return nil

}
