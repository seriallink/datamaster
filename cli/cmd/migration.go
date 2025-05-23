package cmd

import (
	"embed"
	"flag"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/cli/core"
	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
)

func MigrationCmd(migrations embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "migration",
		Help: "Run database migration scripts",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("migration", flag.ContinueOnError)
			script := fs.String("script", "", "Optional script name to run (e.g., xyz.sql)")
			if !ParseShellFlags(c, fs) {
				return
			}

			c.Println(misc.Blue("You are about to run a migration script on Aurora."))
			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Migration cancelled."))
				return
			}

			if err := core.RunMigration(migrations, *script); err != nil {
				c.Println(misc.Red(fmt.Sprintf("Migration failed: %v", err)))
				return
			}

			c.Println(misc.Green("Migration completed successfully."))
		}),
	}
}
