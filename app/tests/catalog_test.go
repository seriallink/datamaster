package tests

import (
	"testing"

	"github.com/seriallink/datamaster/app/core"

	"github.com/stretchr/testify/require"
)

func TestSyncCatalogFromAurora(t *testing.T) {
	MustPersistTestConfig()
	db, err := core.GetConnection()
	require.NoError(t, err)
	err = core.SyncCatalogFromDatabaseSchema(db, "bronze", "brewery")
	require.NoError(t, err)
}
