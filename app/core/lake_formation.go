package core

import (
	"context"
	"fmt"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
)

// GrantDataLocationAccess grants Lake Formation data location access to the Glue service role.
//
// It retrieves the data lake bucket name and the Glue service role ARN from CloudFormation stack outputs,
// then issues a GrantPermissions call to Lake Formation, allowing Glue to access the specified S3 bucket.
//
// Parameters:
//   - cfg: AWS configuration used to create the Lake Formation client and access stack outputs.
//
// Returns:
//   - error: An error if retrieving outputs or granting permissions fails.
func GrantDataLocationAccess(cfg aws.Config) error {

	client := lakeformation.NewFromConfig(cfg)

	bucket, err := (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "DataLakeBucketName")
	if err != nil {
		return err
	}

	identifier, err := (&Stack{Name: misc.StackNameRoles}).GetStackOutput(cfg, "GlueServiceRoleArn")
	if err != nil {
		return err
	}

	input := &lakeformation.GrantPermissionsInput{
		Principal: &types.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(identifier),
		},
		Resource: &types.Resource{
			DataLocation: &types.DataLocationResource{
				ResourceArn: aws.String(fmt.Sprintf("arn:aws:s3:::%s", bucket)),
			},
		},
		Permissions: []types.Permission{
			types.PermissionDataLocationAccess,
		},
	}

	_, err = client.GrantPermissions(context.Background(), input)
	if err != nil {
		return err
	}

	fmt.Println(misc.Green("Successfully granted Data Location access to Lake Formation."))
	return nil

}

// CreateLfTag creates an AWS Lake Formation tag with the key "classification" and value "PII".
//
// This tag can be used to classify sensitive data, such as fields containing Personally Identifiable Information (PII),
// for access control and governance purposes.
//
// Parameters:
//   - cfg: aws.Config - The AWS SDK configuration used to initialize the Lake Formation client.
//
// Returns:
//   - error: Any error encountered during tag creation, or nil if the operation was successful.
func CreateLfTag(cfg aws.Config) error {
	client := lakeformation.NewFromConfig(cfg)
	_, err := client.CreateLFTag(context.TODO(), &lakeformation.CreateLFTagInput{
		TagKey:    aws.String("classification"),
		TagValues: []string{"PII"},
	})
	return err
}

// TagColumnAsPII applies the Lake Formation tag "classification=PII" to a specific column of a table.
//
// This is useful for marking columns that contain Personally Identifiable Information (PII) so that
// access control and data governance rules can be applied accordingly.
//
// Parameters:
//   - cfg: aws.Config - The AWS SDK configuration used to initialize the Lake Formation client.
//   - schema: string - The name of the AWS Glue database (schema).
//   - table: string - The name of the table within the database.
//   - column: string - The name of the column to tag.
//
// Returns:
//   - error: Any error encountered while tagging the column, or nil if successful.
func TagColumnAsPII(cfg aws.Config, schema, table, column string) error {
	client := lakeformation.NewFromConfig(cfg)
	_, err := client.AddLFTagsToResource(context.TODO(), &lakeformation.AddLFTagsToResourceInput{
		Resource: &types.Resource{
			TableWithColumns: &types.TableWithColumnsResource{
				DatabaseName: aws.String(schema),
				Name:         aws.String(table),
				ColumnNames:  []string{column},
			},
		},
		LFTags: []types.LFTagPair{
			{
				TagKey:    aws.String("classification"),
				TagValues: []string{"PII"},
			},
		},
	})
	return err
}
