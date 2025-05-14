package cmd

import (
	"sort"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/seriallink/datamaster/cli/misc"
)

var helpDetails map[string]string

func init() {

	helpDetails = make(map[string]string)

	helpDetails["auth"] = misc.Trim(`

		Authenticate with AWS.

		You can authenticate using:
		  - A named AWS profile
		  - An access key and secret

		The CLI will guide you through the process interactively.

	`)

	helpDetails["whoami"] = misc.Trim(`

			Show the AWS identity used in the current session.

			Displays:
			  - AWS Account ID
			  - IAM Role/User
			  - Region

		`)

	helpDetails["deploy"] = misc.Trim(`

			Deploy infrastructure.

			Usage:
			  deploy
			      Deploy all infrastructure stacks in the correct order.

			  deploy --stack <name> [--params <k1=v1,...>]
			      Deploy a specific stack and optionally pass parameters.

			Examples:
			  deploy --stack network
			  deploy --stack storage --params "Environment=prod,EnableLogs=true"

			Note:
			  To see a list of available stack names, run: stacks

		`)

	helpDetails["migration"] = misc.Trim(`

			Run Aurora database migrations.

			Usage:
			  migration
			      Executes the default migration script (data-master.sql) against Aurora.

			  migration --script <filename>
			      Executes a specific migration script embedded in the binary.

			Examples:
			  migration
			  migration --script database/data-master.sql

			Note:
			  To see a list of available scripts, run: scripts

		`)

}

func Help() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "help",
		Help: "Show help for commands. Use 'help <command>' for details.",
		Func: func(c *ishell.Context) {
			cmds := c.Cmds()

			// Sort commands by name
			sort.Slice(cmds, func(i, j int) bool {
				return cmds[i].Name < cmds[j].Name
			})

			if len(c.Args) == 0 {
				c.Println("Available commands:")
				for _, cmd := range cmds {
					short := strings.SplitN(cmd.Help, ".", 2)[0]
					c.Printf("  %-12s %s\n", cmd.Name, short)
				}
				c.Println("\nUse 'help <command>' for more details.")
				return
			}

			cmdName := c.Args[0]
			for _, cmd := range cmds {
				if cmd.Name == cmdName {
					if detailed, ok := helpDetails[cmdName]; ok {
						c.Printf("Command: %s\n\n%s\n", cmd.Name, detailed)
					} else {
						c.Printf("Command: %s\n\n%s\n", cmd.Name, cmd.Help)
					}
					return
				}
			}

			c.Println(misc.Red("Unknown command: "), cmdName)
		},
	}
}
