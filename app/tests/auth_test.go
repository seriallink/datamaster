package tests

import (
	"context"
	"os"
	"testing"

	"github.com/seriallink/datamaster/app/core"

	"github.com/stretchr/testify/assert"
)

func TestAuthWithKeys(t *testing.T) {

	accessKey := os.Getenv("AWS_ACCESS_KEY")
	secretKey := os.Getenv("AWS_SECRET_KEY")
	region := os.Getenv("AWS_REGION")

	assert.NotEmpty(t, accessKey, "AWS_ACCESS_KEY env var is required")
	assert.NotEmpty(t, secretKey, "AWS_SECRET_KEY env var is required")
	assert.NotEmpty(t, region, "AWS_REGION is required")

	config, err := core.LoadAWSConfig("", accessKey, secretKey, region)
	assert.NoError(t, err)

	identity, err := core.GetCallerIdentity(context.TODO(), config)
	assert.NoError(t, err)
	assert.NotNil(t, identity)

}

func TestAuthWithDefaultProfile(t *testing.T) {

	region := os.Getenv("AWS_REGION")
	assert.NotEmpty(t, region, "AWS_REGION is required for default profile auth test")

	config, err := core.LoadAWSConfig("", "", "", region)
	assert.NoError(t, err)

	identity, err := core.GetCallerIdentity(context.TODO(), config)
	assert.NoError(t, err)
	assert.NotNil(t, identity)
	assert.NotEmpty(t, *identity.Account)

}

func TestAuthWithNamedProfile(t *testing.T) {

	profile := os.Getenv("AWS_PROFILE")
	region := os.Getenv("AWS_REGION")

	assert.NotEmpty(t, profile, "AWS_PROFILE env var is required")
	assert.NotEmpty(t, region, "AWS_REGION is required")

	config, err := core.LoadAWSConfig(profile, "", "", region)
	assert.NoError(t, err)

	identity, err := core.GetCallerIdentity(context.TODO(), config)
	assert.NoError(t, err)
	assert.NotNil(t, identity)
	assert.NotEmpty(t, *identity.Account)

}
