// Data Master Project is a modular and fully automated data platform,
// built on a serverless, event-driven architecture using AWS services.
//
// It implements the Medallion Architecture (raw, bronze, silver, gold),
// with high-performance processing in Go and PySpark across ingestion,
// transformation, and analytics layers.
//
// The project includes:
//   - A Go-based CLI for provisioning, cataloging, and orchestration
//   - Lambda functions for lightweight, event-driven processing
//   - ECS workers for scalable ingestion of large datasets
//   - EMR Spark jobs for complex ETL workflows
//   - Integrated benchmarking tools to compare processing performance
//
// Governance, cost optimization, and observability are built-in,
// making the platform suitable for both experimentation and production workloads.
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

//go:embed dashboards/*.json
var dashboards embed.FS

//go:embed app/help/*.tmpl
var helps embed.FS

//go:embed database/migrations/*.sql
var migrations embed.FS

//go:embed scripts/*.py scripts/bundle.zip
var scripts embed.FS

//go:embed stacks/*.yml
var stacks embed.FS

func main() {

	defer func() {
		if err := core.CloseConnection(); err != nil {
			fmt.Println(misc.Red("Error closing db connection: %v\n", err))
		}
	}()

	shell := ishell.NewWithConfig(cmd.CustomReadlineConfig())

	cmd.Welcome(shell)

	shell.AddCmd(cmd.AuthCmd())
	shell.AddCmd(cmd.WhoAmICmd())
	shell.AddCmd(cmd.StacksCmd(stacks))
	shell.AddCmd(cmd.DeployCmd(stacks, artifacts, scripts))
	shell.AddCmd(cmd.MigrationCmd(migrations))
	shell.AddCmd(cmd.CatalogCmd())
	shell.AddCmd(cmd.ArtifactsCmd(artifacts))
	shell.AddCmd(cmd.LambdaCmd(artifacts))
	shell.AddCmd(cmd.ECRCmd(artifacts))
	shell.AddCmd(cmd.SeedCmd())
	shell.AddCmd(cmd.ProcessCmd())
	shell.AddCmd(cmd.GrafanaCmd(dashboards))
	shell.AddCmd(cmd.StreamCmd())
	shell.AddCmd(cmd.BenchmarkCmd())
	shell.AddCmd(cmd.ExitCmd(shell))
	shell.AddCmd(cmd.ClearCmd(shell))
	shell.AddCmd(cmd.Help(helps))

	shell.Run()

}
