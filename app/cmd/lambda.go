package cmd

import (
	"embed"
	"flag"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// LambdaCmd returns an interactive shell command that deploys one or all Lambda functions from embedded artifacts.
// The user can specify a function name with --name, or deploy all available .zip artifacts if not provided.
// Optional flags include --memory (in MB) and --timeout (in seconds).
func LambdaCmd(artifacts embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "lambda",
		Help: "Deploy one or all Lambda functions from embedded artifacts",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("lambda", flag.ContinueOnError)
			name := fs.String("name", "", "Lambda function name (optional)")
			memory := fs.Int("memory", 128, "Memory size (MB)")
			timeout := fs.Int("timeout", 60, "Timeout (seconds)")

			if !ParseShellFlags(c, fs) {
				return
			}

			var functions []string

			if *name == "" {
				c.Println(misc.Blue(fmt.Sprintf(
					"You are about to deploy ALL embedded Lambda functions with memory: %dMB, timeout: %ds",
					*memory, *timeout,
				)))

				files, err := artifacts.ReadDir(misc.ArtifactsPath)
				if err != nil {
					c.Println(misc.Red(fmt.Sprintf("Failed to read embedded artifacts: %v", err)))
					return
				}

				for _, f := range files {
					if !f.IsDir() && strings.HasSuffix(f.Name(), ".zip") {
						functionName := strings.TrimSuffix(f.Name(), ".zip")
						functions = append(functions, functionName)
					}
				}

				if len(functions) == 0 {
					c.Println(misc.Red("No embedded Lambda artifacts found.\n"))
					return
				}

			} else {
				c.Println(misc.Blue(fmt.Sprintf(
					"You are about to deploy Lambda '%s' with memory: %dMB, timeout: %ds",
					*name, *memory, *timeout,
				)))
				functions = []string{*name}
			}

			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Lambda deployment cancelled.\n"))
				return
			}

			for _, fn := range functions {
				err := core.DeployLambdaFromArtifact(artifacts, fn, *memory, *timeout)
				if err != nil {
					c.Println(misc.Red(fmt.Sprintf("Error deploying %s: %v", fn, err)))
					return
				}
			}

			c.Println(misc.Green("Lambda deployment completed successfully.\n"))

		}),
	}
}
