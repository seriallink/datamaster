package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// BenchmarkCmd returns an interactive CLI command that runs the ingestion benchmark
// using one of the supported execution environments: ECS (Go) or Glue (PySpark).
//
// Usage:
//
//	benchmark --run <ecs|glue>
//
// Parameters:
//
//	--run: Required flag specifying the implementation to execute.
//	  - ecs: Executes the benchmark using a Go binary running on ECS.
//	  - glue: Executes the benchmark using a PySpark script on AWS Glue.
//
// Example:
//
//	benchmark --run ecs
//
// The command prompts for confirmation before launching the benchmark. After execution,
// it prints the result or any error encountered.
func BenchmarkCmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "benchmark",
		Help: "Run benchmark using ECS (Go) or Glue (PySpark)",
		Func: WithAuth(func(c *ishell.Context) {
			fs := flag.NewFlagSet("benchmark", flag.ContinueOnError)
			run := fs.String("run", "", "Required: implementation to benchmark (ecs, glue)")

			if !ParseShellFlags(c, fs) {
				return
			}

			valid := map[string]bool{"ecs": true, "glue": true}
			if !valid[*run] {
				c.Println(misc.Red("You must specify --run ecs or glue"))
				return
			}

			c.Println(misc.Blue(fmt.Sprintf("You are about to run the benchmark for: %s", *run)))
			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Benchmark cancelled.\n"))
				return
			}

			var err error
			switch *run {
			case "ecs":
				err = core.RunEcsBenchmark()
			case "glue":
				err = core.RunGlueBenchmark()
			}

			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Benchmark failed: %v", err)))
				return
			}

			c.Println(misc.Green("Benchmark completed successfully.\n"))
		}),
	}
}
