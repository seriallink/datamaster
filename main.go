package main

import (
	"embed"
	"fmt"

	"github.com/seriallink/datamaster/app/cmd"
	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

//go:embed artifacts/*
var artifacts embed.FS

//go:embed etl/main.py etl/bundle.zip
var assets embed.FS

//go:embed dashboards/*.json
var dashboards embed.FS

//go:embed app/help/*.tmpl
var helps embed.FS

//go:embed database/migrations/*.sql
var scripts embed.FS

//go:embed stacks/*.yml
var stacks embed.FS

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
	shell.AddCmd(cmd.DeployCmd(stacks, artifacts, assets))
	shell.AddCmd(cmd.MigrationCmd(scripts))
	shell.AddCmd(cmd.CatalogCmd())
	shell.AddCmd(cmd.ArtifactsCmd(artifacts))
	shell.AddCmd(cmd.LambdaCmd(artifacts))
	shell.AddCmd(cmd.SeedCmd())
	shell.AddCmd(cmd.StreamCmd())
	shell.AddCmd(cmd.ProcessCmd())
	shell.AddCmd(cmd.GrafanaCmd(dashboards))
	shell.AddCmd(cmd.ExitCmd(shell))
	shell.AddCmd(cmd.ClearCmd(shell))
	shell.AddCmd(cmd.Help(helps))

	shell.Run()

}
