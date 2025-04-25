package cmd

import (
	"context"
	"github.com/abiosoft/ishell"
	"github.com/seriallink/datamaster/cli/core"
)

func WhoAmICmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "whoami",
		Help: "Display current AWS identity",
		Func: func(c *ishell.Context) {

			cfg := core.GetAWSConfig()
			if cfg.Credentials == nil {
				c.Println(Red("No AWS credentials found. Please authenticate first.\n"))
				return
			}

			identity, err := core.ValidateAWSCredentials(context.TODO(), core.GetAWSConfig())
			if err != nil {
				c.Println(Red("Error retrieving AWS identity: %v\n", err))
				return
			}

			c.Println(Green("Current AWS Identity:\n  UserId: %s\n  Account: %s\n  ARN: %s\n", *identity.UserId, *identity.Account, *identity.Arn))

		},
	}
}
