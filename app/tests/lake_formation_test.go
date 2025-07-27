package tests

import (
	"testing"

	"github.com/seriallink/datamaster/app/core"

	"github.com/stretchr/testify/require"
)

func TestGrantDataLocationAccess(t *testing.T) {
	MustPersistTestConfig()
	cfg := core.GetAWSConfig()
	err := core.GrantDataLocationAccess(cfg)
	require.NoError(t, err)
}
