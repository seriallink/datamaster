package core

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/cli/misc"

	_ "github.com/lib/pq"
)

func RunMigration(fs embed.FS, script string) error {

	var (
		err             error
		sqlContent      []byte
		securityOutputs map[string]string
		db              *sql.DB
	)

	if script == "" {
		script = misc.DefaultScript
	}

	if sqlContent, err = fs.ReadFile(script); err != nil {
		return fmt.Errorf("failed to read script %s: %w", script, err)
	}

	if securityOutputs, err = GetStackOutputs(&Stack{Name: misc.StackNameSecurity}); err != nil {
		return fmt.Errorf("failed to get outputs from security stack: %w", err)
	}

	if db, err = ConnectAuroraFromSecret(securityOutputs["SecretArn"]); err != nil {
		return fmt.Errorf("failed to connect to Aurora: %w", err)
	}
	defer db.Close()

	if _, err = db.Exec(strings.TrimPrefix(string(sqlContent), "\uFEFF")); err != nil {
		return fmt.Errorf("failed to execute script %s: %w", script, err)
	}

	return nil

}
