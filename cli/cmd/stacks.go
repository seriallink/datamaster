package cmd

import (
	"embed"
	"sort"
	"strings"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
)

func StacksCmd(templates embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "stacks",
		Help: "List available stacks.",
		Func: func(c *ishell.Context) {
			entries, err := templates.ReadDir(misc.TemplatesPath)
			if err != nil {
				c.Println(misc.Red("Failed to read stack templates: %v", err))
				return
			}

			var stacks []string
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), misc.TemplateExtension) {
					name := strings.TrimSuffix(entry.Name(), misc.TemplateExtension)
					stacks = append(stacks, name)
				}
			}

			sort.Strings(stacks)

			c.Println("Available stacks:")
			for _, name := range stacks {
				c.Printf("  %s\n", name)
			}
		},
	}
}
