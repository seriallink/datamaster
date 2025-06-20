package cmd

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

func ArtifactsCmd(artifacts embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "artifacts",
		Help: "List embedded Lambda artifacts.",
		Func: func(c *ishell.Context) {

			entries, err := artifacts.ReadDir(misc.ArtifactsPath)
			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Failed to read embedded artifacts: %v", err)))
				return
			}

			var names []string
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".zip") {
					name := strings.TrimSuffix(entry.Name(), ".zip")
					names = append(names, name)
				}
			}

			sort.Strings(names)

			if len(names) == 0 {
				c.Println(misc.Red("No embedded artifacts found."))
				return
			}

			c.Println("Available Lambda artifacts:")
			for _, name := range names {
				c.Printf("  %s\n", name)
			}

		},
	}
}
