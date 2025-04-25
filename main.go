package main

import (
	"embed"
	"github.com/abiosoft/ishell"
	"github.com/seriallink/datamaster/cli/cmd"
)

//go:embed infra/templates/*.yml
var templates embed.FS

func main() {

	shell := ishell.New()

	shell.Println(cmd.Yellow("    ____        __           __  ___           __            "))
	shell.Println(cmd.Yellow("   / __ \\____ _/ /_____ _   /  |/  /___ ______/ /____  _____"))
	shell.Println(cmd.Yellow("  / / / / __ `/ __/ __ `/  / /|_/ / __ `/ ___/ __/ _ \\/ ___/"))
	shell.Println(cmd.Yellow(" / /_/ / /_/ / /_/ /_/ /  / /  / / /_/ (__  ) /_/  __/ /     "))
	shell.Println(cmd.Yellow("/_____/\\__,_/\\__/\\__,_/  /_/  /_/\\__,_/____/\\__/\\___/_/"))

	shell.Println("\nWelcome to DataMaster CLI! Enter 'help' for a list of commands.")

	shell.AddCmd(cmd.AuthCmd())
	shell.AddCmd(cmd.WhoAmICmd())
	shell.AddCmd(cmd.DeployCmd(templates))

	shell.Run()

}
