package tests

import (
	"testing"

	"github.com/seriallink/datamaster/app/core"
	"github.com/stretchr/testify/require"
)

func TestEnsureECSServiceLinkedRole(t *testing.T) {
	MustPersistTestConfig()
	cfg := core.GetAWSConfig()
	err := core.EnsureECSServiceLinkedRole(cfg)
	require.NoError(t, err)
}
