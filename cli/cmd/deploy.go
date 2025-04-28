package cmd

import (
	"embed"
	"strings"

	"github.com/seriallink/datamaster/cli/core"
	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
)

func DeployCmd(templates embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "deploy",
		Help: "Deploy all infrastructure stacks",
		Func: func(c *ishell.Context) {
			c.Println(misc.Blue("You are about to deploy all stacks."))
			c.Println("This will deploy the following stacks:")
			c.Println("- Network")
			c.Println("- Security")
			c.Println("- Aurora PostgreSQL")
			c.Print("Type 'go' to continue: ")

			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Deployment cancelled.\n"))
				return
			}

			err := core.DeployAllStacks(c, templates)
			if err != nil {
				c.Println(misc.Red("Deployment failed: %v", err))
				return
			}

			c.Println(color.HiGreenString("All stacks deployed successfully!\n"))
		},
	}
}
