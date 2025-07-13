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
