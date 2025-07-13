package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// SeedCmd returns an interactive shell command that seeds datasets into Aurora or S3.
// The user can optionally specify a dataset file with --file to seed a single dataset,
// or omit it to seed all datasets from the default source (GitHub).
func SeedCmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "seed",
		Help: "Seed datasets into Aurora or S3",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("seed", flag.ContinueOnError)
			file := fs.String("file", "", "Optional dataset name to seed")
			if !ParseShellFlags(c, fs) {
				return
			}

			if *file != "" {
				c.Println(misc.Blue(fmt.Sprintf("You are about to seed a specific dataset: %s", *file)))
			} else {
				c.Println(misc.Blue("You are about to seed all available datasets from GitHub."))
			}
			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Seeding cancelled.\n"))
				return
			}

			var err error
			if *file != "" {
				err = core.SeedFile(*file)
			} else {
				err = core.SeedThemAll()
			}

			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Seeding failed: %v", err)))
				return
			}

			c.Println(misc.Green("Seeding completed successfully.\n"))

		}),
	}

}
