package main

import (
	"embed"

	"github.com/seriallink/datamaster/cli/cmd"
	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
)

//go:embed infra/templates/*.yml
var templates embed.FS

func main() {

	shell := ishell.New()

	shell.Println(misc.Yellow("    ____        __           __  ___           __            "))
	shell.Println(misc.Yellow("   / __ \\____ _/ /_____ _   /  |/  /___ ______/ /____  _____"))
	shell.Println(misc.Yellow("  / / / / __ `/ __/ __ `/  / /|_/ / __ `/ ___/ __/ _ \\/ ___/"))
	shell.Println(misc.Yellow(" / /_/ / /_/ / /_/ /_/ /  / /  / / /_/ (__  ) /_/  __/ /     "))
	shell.Println(misc.Yellow("/_____/\\__,_/\\__/\\__,_/  /_/  /_/\\__,_/____/\\__/\\___/_/"))

	shell.Println("\nWelcome to Data Master CLI! Enter 'help' for a list of commands.")

	shell.AddCmd(cmd.AuthCmd())
	shell.AddCmd(cmd.WhoAmICmd())
	shell.AddCmd(cmd.DeployCmd(templates))

	shell.Run()

}
