package core

import (
	"context"
	"fmt"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
)

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
