package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

func PipeCmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "pipe",
		Help: "Register dynamic pipe(s) to process table updates",
		Func: WithAuth(func(c *ishell.Context) {
			fs := flag.NewFlagSet("pipe", flag.ContinueOnError)
			table := fs.String("table", "", "Name of the table to register (optional)")

			if !ParseShellFlags(c, fs) {
				return
			}

			if *table != "" {
				c.Println(misc.Blue(fmt.Sprintf("You are about to register a dynamic pipe for table: %s", *table)))
			} else {
				c.Println(misc.Blue("You are about to register dynamic pipes for ALL tables"))
			}

			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Pipe registration cancelled.\n"))
				return
			}

			var err error
			if *table != "" {
				err = core.RegisterPipeForTable(*table)
			} else {
				err = core.RegisterPipeForAllTables()
			}

			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Error: %v", err)))
				return
			}

			c.Println(misc.Green("Pipe(s) registered successfully.\n"))
		}),
	}
}
