package main

import (
	"embed"
	"fmt"

	"github.com/seriallink/datamaster/cli/cmd"
	"github.com/seriallink/datamaster/cli/core"
	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
)

//go:embed cli/help/*.tmpl
var helps embed.FS

//go:embed infra/templates/*.yml
var stacks embed.FS

//go:embed database/migrations/*.sql
var scripts embed.FS

//go:embed artifacts/*.zip
var artifacts embed.FS

func main() {

	defer func() {
		if err := core.CloseConnection(); err != nil {
			fmt.Println(misc.Red("Error closing db connection: %v\n", err))
		}
	}()

	shell := ishell.New()

	shell.Println(misc.Yellow("    ____        __           __  ___           __            "))
	shell.Println(misc.Yellow("   / __ \\____ _/ /_____ _   /  |/  /___ ______/ /____  _____"))
	shell.Println(misc.Yellow("  / / / / __ `/ __/ __ `/  / /|_/ / __ `/ ___/ __/ _ \\/ ___/"))
	shell.Println(misc.Yellow(" / /_/ / /_/ / /_/ /_/ /  / /  / / /_/ (__  ) /_/  __/ /     "))
	shell.Println(misc.Yellow("/_____/\\__,_/\\__/\\__,_/  /_/  /_/\\__,_/____/\\__/\\___/_/"))

	shell.Println("\nWelcome to Data Master CLI! Enter 'help' for a list of commands.")

	shell.AddCmd(cmd.AuthCmd())
	shell.AddCmd(cmd.WhoAmICmd())
	shell.AddCmd(cmd.StacksCmd(stacks))
	shell.AddCmd(cmd.DeployCmd(stacks))
	shell.AddCmd(cmd.MigrationCmd(scripts))
	shell.AddCmd(cmd.CatalogCmd())
	shell.AddCmd(cmd.PipeCmd())
	shell.AddCmd(cmd.ArtifactsCmd(artifacts))
	shell.AddCmd(cmd.LambdaCmd(artifacts))
	shell.AddCmd(cmd.StreamCmd())
	shell.AddCmd(cmd.InspectCmd())
	shell.AddCmd(cmd.ExitCmd(shell))
	shell.AddCmd(cmd.ClearCmd(shell))
	shell.AddCmd(cmd.Help(helps))

	shell.Run()

}
