package core

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var awsConfig aws.Config

func GetAWSConfig() aws.Config {
	return awsConfig
}

func PersistAWSConfig(profile, accessKey, secretKey, region string) (err error) {
	awsConfig, err = LoadAWSConfig(profile, accessKey, secretKey, region)
	return
}

func LoadAWSConfig(profile, accessKey, secretKey, region string) (aws.Config, error) {

	if region == "" {
		region = "us-east-1"
	}

	if profile != "" {
		return config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithSharedConfigProfile(profile),
		)
	}

	if accessKey != "" && secretKey != "" {
		return config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
			),
		)
	}

	return config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))

}

func ValidateAWSCredentials(ctx context.Context, cfg aws.Config) (*sts.GetCallerIdentityOutput, error) {
	client := sts.NewFromConfig(cfg)
	return client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
}
