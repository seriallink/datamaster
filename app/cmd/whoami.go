package cmd

import (
	"context"
	"fmt"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// WhoAmICmd returns an interactive shell command that displays the current AWS identity.
// It uses the AWS STS GetCallerIdentity API to retrieve and print the UserId, Account, and ARN.
// AWS authentication is required and handled by the WithAuth middleware.
func WhoAmICmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "whoami",
		Help: "Display current AWS identity",
		Func: WithAuth(func(c *ishell.Context) {

			identity, err := core.GetCallerIdentity(context.TODO(), core.GetAWSConfig())
			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Error retrieving AWS identity: %v\n", err)))
				return
			}

			c.Println(misc.Green("Current AWS Identity:\n  UserId: %s\n  Account: %s\n  ARN: %s\n", *identity.UserId, *identity.Account, *identity.Arn))

		}),
	}
}
