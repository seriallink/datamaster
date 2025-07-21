package core

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/misc"

	_ "github.com/lib/pq"
)

// RunAllMigrations executes all predefined SQL migration scripts in order.
//
// It runs the core, view, and mart scripts sequentially using the embedded filesystem.
//
// Parameters:
//   - fs: Embedded filesystem containing the SQL migration scripts.
//
// Returns:
//   - error: An error if any of the migration scripts fail to execute.
func RunAllMigrations(fs embed.FS) error {

	scripts := []string{
		misc.MigrationCoreScript,
		misc.MigrationViewScript,
		misc.MigrationMartScript,
	}

	for _, script := range scripts {
		fmt.Println(misc.Blue("Running %s script...", script))
		if err := RunMigration(fs, script); err != nil {
			return err
		}
	}

	return nil

}

// RunMigration executes an SQL script embedded via embed.FS against the Aurora PostgreSQL database.
//
// Parameters:
//   - fs: embedded file system containing SQL scripts.
//   - script: path to the SQL script within the embedded FS.
//
// Behavior:
//   - Reads the script file from the embedded FS.
//   - Opens a connection to Aurora using GetConnection.
//   - Executes the SQL content (removing optional BOM prefix if present).
//
// Returns:
//   - error: if reading the script, connecting to the DB, or executing the SQL fails.
func RunMigration(fs embed.FS, script string) error {

	var (
		err        error
		sqlContent []byte
		sqlDbConn  *sql.DB
	)

	if sqlContent, err = fs.ReadFile(script); err != nil {
		return fmt.Errorf("failed to read script %s: %w", script, err)
	}

	if dbConn, err = GetConnection(); err != nil {
		return fmt.Errorf("failed to connect to Aurora: %w", err)
	}

	if sqlDbConn, err = dbConn.DB(); err != nil {
		return fmt.Errorf("failed to get native DB connection: %w", err)
	}

	if _, err = sqlDbConn.Exec(strings.TrimPrefix(string(sqlContent), "\uFEFF")); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Printf("Script %s has already been executed. Skipping...\n", script)
			return nil
		}
		return fmt.Errorf("failed to execute script %s: %w", script, err)
	}

	if script == misc.MigrationCoreScript {
		err = ensureCDCStarted()
		if err != nil {
			return fmt.Errorf("failed to start CDC: %w", err)
		}
	}

	return nil

}
