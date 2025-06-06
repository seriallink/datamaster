package cmd

import (
	"embed"
	"flag"
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"
)

func MigrationCmd(migrations embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "migration",
		Help: "Run database migration scripts",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("migration", flag.ContinueOnError)
			script := fs.String("script", "", "Optional script name to run")
			if !ParseShellFlags(c, fs) {
				return
			}

			// Run specific script
			if *script != "" {
				c.Println(misc.Blue("You are about to run a specific migration script:"))
				c.Println("  →", *script)
				c.Print("Type 'go' to continue: ")
				if strings.ToLower(c.ReadLine()) != "go" {
					c.Println(misc.Red("Migration cancelled.\n"))
					return
				}

				if err := core.RunMigration(migrations, *script); err != nil {
					c.Println(misc.Red(fmt.Sprintf("Migration failed: %v", err)))
					return
				}

				c.Println(misc.Green("Script '%s' executed successfully!\n", *script))
				return
			}

			// Run all scripts
			c.Println(misc.Blue("You are about to run all migration scripts in order."))
			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Migration cancelled.\n"))
				return
			}

			if err := core.RunAllMigrations(migrations); err != nil {
				c.Println(misc.Red(fmt.Sprintf("Migration failed: %v", err)))
				return
			}

			c.Println(misc.Green("All migration scripts executed successfully!\n"))
		}),
	}
}
