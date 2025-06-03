package cmd

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

func Help(helps embed.FS) *ishell.Cmd {
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
			filename := fmt.Sprintf("app/help/%s.tmpl", cmdName)

			content, err := helps.ReadFile(filename)
			if err == nil {
				c.Printf("Command: %s\n\n%s\n", cmdName, string(content))
				return
			}

			// fallback to built-in command help
			for _, cmd := range cmds {
				if cmd.Name == cmdName {
					c.Printf("Command: %s\n\n%s\n", cmd.Name, cmd.Help)
					return
				}
			}

			c.Println(misc.Red("Unknown command: "), cmdName)
		},
	}
}
