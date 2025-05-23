package cmd

import (
	"flag"
	"fmt"

	"github.com/seriallink/datamaster/cli/core"
	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
)

func InspectCmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "inspect",
		Help: "Show recent records from the Kinesis stream",
		Func: WithAuth(func(c *ishell.Context) {
			fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
			table := fs.String("table", "", "Table name (optional)")

			if !ParseShellFlags(c, fs) {
				return
			}

			if err := core.InspectStream(*table); err != nil {
				c.Println(misc.Red(fmt.Sprintf("Error: %v", err)))
			}
		}),
	}
}
