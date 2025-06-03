package cmd

import (
	"embed"
	"flag"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func DeployCmd(templates, artifacts embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "deploy",
		Help: "Deploy infrastructure",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
			stackName := fs.String("stack", "", "Stack name to deploy")
			params := fs.String("params", "", "Comma-separated list of key=value pairs")
			if !ParseShellFlags(c, fs) {
				return
			}

			// Parse --params into []types.Parameter
			var parsedParams []types.Parameter
			if *params != "" {
				for _, pair := range strings.Split(*params, ",") {
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) == 2 {
						parsedParams = append(parsedParams, types.Parameter{
							ParameterKey:   aws.String(kv[0]),
							ParameterValue: aws.String(kv[1]),
						})
					}
				}
			}

			// Deploy a specific stack
			if *stackName != "" {
				stack := core.Stack{
					Name:   *stackName,
					Params: parsedParams,
				}

				if _, err := stack.GetTemplateBody(templates); err != nil {
					c.Println(misc.Red(fmt.Sprintf("Stack '%s' does not exist.", stack.Name)))
					c.Println("Run the command 'stacks' to list available stack names.")
					return
				}

				c.Println(misc.Blue("You are about to deploy stack:"), stack.FullStackName())
				c.Print("Type 'go' to continue: ")
				if strings.ToLower(c.ReadLine()) != "go" {
					c.Println(misc.Red("Deployment cancelled.\n"))
					return
				}

				if err := core.DeployStack(&stack, templates, artifacts); err != nil {
					c.Println(misc.Red(fmt.Sprintf("Deployment failed: %v", err)))
					return
				}

				c.Println(misc.Green("Stack '%s' deployed successfully!\n", stack.FullStackName()))
				return
			}

			// Deploy all stacks
			c.Println(misc.Blue("You are about to deploy all stacks."))
			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Deployment cancelled.\n"))
				return
			}

			if err := core.DeployAllStacks(templates, artifacts); err != nil {
				c.Println(misc.Red(fmt.Sprintf("Deployment failed: %v", err)))
				return
			}

			c.Println(misc.Green("All stacks deployed successfully!\n"))
		}),
	}
}
