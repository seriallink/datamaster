package core

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// singleton instance of AWS configuration
var awsConfig aws.Config

// GetAWSConfig returns the currently persisted AWS configuration.
// If the configuration has not been initialized, it returns an empty aws.Config
// and any AWS SDK client using it will likely fail unless initialized via PersistAWSConfig.
func GetAWSConfig() aws.Config {
	return awsConfig
}

// PersistAWSConfig loads and persists the AWS configuration globally,
// based on the provided profile or static credentials.
// If both profile and static credentials are provided, the profile takes precedence.
//
// Parameters:
//   - profile: name of the AWS named profile to load.
//   - accessKey: AWS access key ID (used if profile is empty).
//   - secretKey: AWS secret access key (used if profile is empty).
//   - region: AWS region to use (default is "us-east-1" if empty).
//
// Returns:
//   - error: if loading the configuration fails.
func PersistAWSConfig(profile, accessKey, secretKey, region string) (err error) {
	awsConfig, err = LoadAWSConfig(profile, accessKey, secretKey, region)
	return
}

// LoadAWSConfig loads an AWS configuration based on the input parameters.
// It supports named profiles, static credentials, or defaults.
//
// Parameters:
//   - profile: AWS profile name (overrides other methods if set).
//   - accessKey: AWS access key ID (used only if profile is empty).
//   - secretKey: AWS secret access key (used only if profile is empty).
//   - region: AWS region (defaults to "us-east-1" if empty).
//
// Returns:
//   - aws.Config: the loaded AWS configuration.
//   - error: if loading the configuration fails.
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

// GetCallerIdentity returns the identity of the current AWS caller,
// including the account, ARN, and user ID.
//
// Parameters:
//   - ctx: context for cancellation and timeout.
//   - cfg: an initialized AWS configuration.
//
// Returns:
//   - *sts.GetCallerIdentityOutput: information about the caller identity.
//   - error: if the STS call fails.
func GetCallerIdentity(ctx context.Context, cfg aws.Config) (*sts.GetCallerIdentityOutput, error) {
	client := sts.NewFromConfig(cfg)
	return client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
}
