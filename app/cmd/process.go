package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"
)

// ProcessCmd triggers the processing pipeline for a specific layer (e.g. silver, gold)
// with optional filtering by table names.
func ProcessCmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "process",
		Help: "Trigger processing pipeline for a given layer",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("process", flag.ContinueOnError)
			layer := fs.String("layer", "", "Processing layer (e.g. silver, gold)")
			tables := fs.String("tables", "", "Comma-separated list of table names to process")
			if !ParseShellFlags(c, fs) {
				return
			}

			if *layer == "" {
				c.Println(misc.Red("Flag --layer is required."))
				return
			}

			var tableList []string
			if *tables != "" {
				tableList = strings.Split(*tables, ",")
			}

			if len(tableList) > 0 {
				c.Println(misc.Blue(fmt.Sprintf("You are about to trigger processing for %d table(s) in layer '%s':", len(tableList), *layer)))
				for _, tbl := range tableList {
					c.Println("-", tbl)
				}
			} else {
				c.Println(misc.Blue(fmt.Sprintf("You are about to trigger processing for all tables in layer: %s", *layer)))
			}

			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Processing cancelled.\n"))
				return
			}

			if err := core.TriggerProcessing(core.GetAWSConfig(), *layer, tableList); err != nil {
				c.Println(misc.Red(fmt.Sprintf("Error triggering processing: %v", err)))
				return
			}

			c.Println(misc.Green("Processing started successfully.\n"))
		}),
	}
}
