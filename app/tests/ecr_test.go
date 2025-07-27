package tests

import (
	"testing"

	"github.com/seriallink/datamaster/app/core"
	"github.com/stretchr/testify/require"
)

func TestLoginToECRWithSDK(t *testing.T) {
	MustPersistTestConfig()
	cfg := core.GetAWSConfig()
	err := core.LoginToECRWithSDK(cfg)
	require.NoError(t, err)
}
