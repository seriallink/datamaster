package cmd

import (
	"github.com/abiosoft/ishell"
	"github.com/seriallink/datamaster/app/misc"
)

// Welcome prints the welcome banner and message for the CLI.
func Welcome(shell *ishell.Shell) {
	shell.Println(misc.Yellow("    ____        __           __  ___           __            "))
	shell.Println(misc.Yellow("   / __ \\____ _/ /_____ _   /  |/  /___ ______/ /____  _____"))
	shell.Println(misc.Yellow("  / / / / __ `/ __/ __ `/  / /|_/ / __ `/ ___/ __/ _ \\/ ___/"))
	shell.Println(misc.Yellow(" / /_/ / /_/ / /_/ /_/ /  / /  / / /_/ (__  ) /_/  __/ /     "))
	shell.Println(misc.Yellow("/_____/\\__,_/\\__/\\__,_/  /_/  /_/\\__,_/____/\\__/\\___/_/"))
	shell.Println()
	shell.Println("Welcome to Data Master CLI! Type 'help' to see available commands.")
}
