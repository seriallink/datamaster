package tests

import (
	"testing"

	"github.com/seriallink/datamaster/app/core"
	"github.com/stretchr/testify/require"
)

func TestStartReplication(t *testing.T) {
	MustPersistTestConfig()
	err := core.StartReplication()
	require.NoError(t, err)
}
