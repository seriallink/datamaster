package cmd

import (
	"context"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

const (
	Option1 = "1"
	Option2 = "2"
)

// AuthCmd returns an interactive shell command that handles AWS authentication.
// The user is prompted to authenticate either using a named profile or direct access/secret keys.
// After selecting the authentication method and region, the credentials are saved locally
// and tested using the STS GetCallerIdentity call. On success, the authenticated identity is displayed.
func AuthCmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "auth",
		Help: "Authenticate with AWS",
		Func: func(c *ishell.Context) {

			var (
				profileName string
				accessKey   string
				secretKey   string
				region      string
			)

		loop:
			for {
				c.Println(misc.Blue("Select authentication method:"))
				c.Println("1. Use Profile")
				c.Println("2. Use Access/Secret Keys")
				c.Print("Enter option: ")
				choice := c.ReadLine()

				switch choice {
				case Option1:
					c.Print("Enter profile name (leave blank for default): ")
					profileName = c.ReadLine()
					break loop

				case Option2:
					c.Print("Enter access key: ")
					accessKey = c.ReadLine()
					c.Print("Enter secret key: ")
					secretKey = c.ReadLine()
					break loop

				default:
					c.Println(misc.Red("Invalid option. Choose 1 or 2."))
					continue

				}
			}

			c.Print("Enter AWS region (default: us-east-1): ")
			region = c.ReadLine()

			err := core.PersistAWSConfig(profileName, accessKey, secretKey, region)
			if err != nil {
				c.Println(misc.Red("Error saving AWS credentials: %v", err))
				return
			}

			identity, err := core.GetCallerIdentity(context.TODO(), core.GetAWSConfig())
			if err != nil {
				c.Println(misc.Red("Error testing AWS credentials: %v", err))
				return
			}

			c.Println(misc.Green("Authenticated as:\n  UserId: %s\n  Account: %s\n  ARN: %s\n", *identity.UserId, *identity.Account, *identity.Arn))
			return

		},
	}
}
