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
	config, err := core.LoadAWSConfig("", accessKey, secretKey, region)
	assert.NoError(t, err)

	identity, err := core.GetCallerIdentity(context.TODO(), config)
	assert.NoError(t, err)
	assert.NotNil(t, identity)

}
