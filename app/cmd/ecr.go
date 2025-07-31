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

// ECRCmd returns an interactive shell command that publishes one or all Docker images to ECR from embedded .tar artifacts.
// The user can specify an image name with --name, or publish all available .tar artifacts if not provided.
func ECRCmd(artifacts embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "ecr",
		Help: "Publish one or all Docker images to ECR from embedded artifacts",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("ecr", flag.ContinueOnError)
			name := fs.String("name", "", "Docker image name (optional)")

			if !ParseShellFlags(c, fs) {
				return
			}

			var images []string

			if *name == "" {
				files, err := artifacts.ReadDir(misc.ArtifactsPath)
				if err != nil {
					c.Println(misc.Red(fmt.Sprintf("Failed to read embedded artifacts: %v", err)))
					return
				}

				for _, f := range files {
					if !f.IsDir() && strings.HasSuffix(f.Name(), ".tar") {
						imageName := strings.TrimSuffix(f.Name(), ".tar")
						images = append(images, imageName)
					}
				}

				if len(images) == 0 {
					c.Println(misc.Red("No embedded Docker image artifacts (.tar) found.\n"))
					return
				}

				c.Println(misc.Blue(fmt.Sprintf(
					"You are about to publish ALL embedded Docker images (%d) to ECR",
					len(images),
				)))
			} else {
				c.Println(misc.Blue(fmt.Sprintf(
					"You are about to publish Docker image '%s' to ECR",
					*name,
				)))
				images = []string{*name}
			}

			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Docker image publication cancelled.\n"))
				return
			}

			_, err := core.PublishDockerImages(core.GetAWSConfig(), artifacts, images...)
			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Error publishing dockers images: %v", err)))
				return
			}

			c.Println(misc.Green("Docker image publication completed successfully.\n"))

		}),
	}
}
