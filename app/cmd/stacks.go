package cmd

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// StacksCmd returns an interactive shell command that lists all available infrastructure stacks.
// It reads the embedded templates directory and extracts stack names based on the configured template extension.
func StacksCmd(templates embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "stacks",
		Help: "List available stacks.",
		Func: func(c *ishell.Context) {

			entries, err := templates.ReadDir(misc.TemplatesPath)
			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Failed to read stack templates: %v", err)))
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
