package tests

import (
	"embed"
	"testing"

	"github.com/seriallink/datamaster/app/core"

	"github.com/stretchr/testify/require"
)

//go:embed migrations/test.sql
var migrations embed.FS

func TestRunMigration(t *testing.T) {
	MustPersistTestConfig()
	err := core.RunMigration(migrations, "migrations/test.sql")
	require.NoError(t, err)
}
