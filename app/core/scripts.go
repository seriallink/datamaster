package core

import (
	"context"
	"embed"
	"fmt"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// UploadPythonScripts uploads the python scripts from the embedded filesystem to the S3 artifacts bucket.
//
// Parameters:
//   - cfg: the AWS configuration used to access S3 and resolve the bucket name.
//   - scripts: the embedded filesystem containing the python scripts files.
//
// Returns:
//   - error: an error if any of the assets fail to be read or uploaded.
func UploadPythonScripts(cfg aws.Config, scripts embed.FS, bucket string, files ...string) error {

	for _, file := range files {

		data, err := scripts.ReadFile(fmt.Sprintf("scripts/%s", file))
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		key := path.Join("scripts", path.Base(file))

		err = UploadDataToS3(cfg, context.TODO(), bucket, key, data)
		if err != nil {
			return fmt.Errorf("failed to upload %s to S3: %w", file, err)
		}

		fmt.Printf("Uploaded %s to s3://%s/%s\n", file, bucket, key)

	}

	return nil

}
