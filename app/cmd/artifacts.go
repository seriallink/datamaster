package cmd

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// ArtifactsCmd returns an interactive shell command that lists all embedded artifacts (.zip for Lambda, .tar for Docker).
// It reads the embedded filesystem under the predefined artifact path and prints the artifact names with their type.
func ArtifactsCmd(artifacts embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "artifacts",
		Help: "List embedded Lambda and Docker artifacts.",
		Func: func(c *ishell.Context) {

			entries, err := artifacts.ReadDir(misc.ArtifactsPath)
			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Failed to read embedded artifacts: %v", err)))
				return
			}

			var artifactsList []string

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				switch {
				case strings.HasSuffix(entry.Name(), ".zip"):
					name := strings.TrimSuffix(entry.Name(), ".zip")
					artifactsList = append(artifactsList, fmt.Sprintf("%s (lambda)", name))

				case strings.HasSuffix(entry.Name(), ".tar"):
					name := strings.TrimSuffix(entry.Name(), ".tar")
					artifactsList = append(artifactsList, fmt.Sprintf("%s (docker)", name))
				}
			}

			sort.Strings(artifactsList)

			if len(artifactsList) == 0 {
				c.Println(misc.Red("No embedded artifacts found."))
				return
			}

			c.Println("Available artifacts:")
			for _, item := range artifactsList {
				c.Printf("  %s\n", item)
			}
		},
	}
}
