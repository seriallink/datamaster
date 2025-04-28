package cmd

import (
	"context"

	"github.com/seriallink/datamaster/cli/core"
	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
)

func WhoAmICmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "whoami",
		Help: "Display current AWS identity",
		Func: func(c *ishell.Context) {

			cfg := core.GetAWSConfig()
			if cfg.Credentials == nil {
				c.Println(misc.Red("No AWS credentials found. Please authenticate first.\n"))
				return
			}

			identity, err := core.ValidateAWSCredentials(context.TODO(), core.GetAWSConfig())
			if err != nil {
				c.Println(misc.Red("Error retrieving AWS identity: %v\n", err))
				return
			}

			c.Println(misc.Green("Current AWS Identity:\n  UserId: %s\n  Account: %s\n  ARN: %s\n", *identity.UserId, *identity.Account, *identity.Arn))

		},
	}
}
