package cmd

import (
	"embed"
	"strings"

	"github.com/seriallink/datamaster/cli/core"

	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
)

func DeployCmd(templates embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "deploy",
		Help: "Deploy all infrastructure stacks",
		Func: func(c *ishell.Context) {
			c.Println(Blue("You are about to deploy all stacks."))
			c.Println("This will deploy the following stacks:")
			c.Println("- network")
			c.Println("- s3-buckets")
			c.Println("- glue-jobs")
			c.Println("- lakeformation")
			c.Println("- redshift")
			c.Print("Type 'go' to continue: ")

			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(Red("Deployment cancelled.\n"))
				return
			}

			stacks := []struct {
				name     string
				template string
			}{
				{"network", "network.yml"},
				{"s3-buckets", "s3-buckets.yml"},
				{"glue-jobs", "glue-jobs.yml"},
				{"lakeformation", "lakeformation.yml"},
				{"redshift", "redshift.yml"},
			}

			for _, stack := range stacks {

				templateContent, err := templates.ReadFile("infra/templates/" + stack.template)
				if err != nil {
					c.Println(Red("Error reading template %s: %v\n", stack.template, err))
					return
				}

				if err = core.DeployStack(stack.name, templateContent); err != nil {
					c.Println(color.HiRedString("Error deploying stack %s: %v\n", stack.name, err))
					return
				}

			}

			c.Println(color.HiGreenString("All stacks deployed successfully!\n"))

		},
	}
}
