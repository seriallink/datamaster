package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// CatalogCmd returns an interactive shell command that creates or updates AWS Glue catalog tables.
// The user can optionally specify a data lake layer (--layer) and a comma-separated list of table names (--tables).
// If no flags are provided, all tables in all layers (bronze, silver, gold) will be processed.
func CatalogCmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "catalog",
		Help: "Create or update catalog tables",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("catalog", flag.ContinueOnError)
			layer := fs.String("layer", "", "Catalog layer (bronze, silver, gold)")
			tables := fs.String("tables", "", "Comma-separated list of table names to include")
			if !ParseShellFlags(c, fs) {
				return
			}

			if *layer != "" && misc.NotIn(*layer, misc.LayerBronze, misc.LayerSilver, misc.LayerGold) {
				c.Println(misc.Red(fmt.Sprintf("Invalid layer: '%s'. Valid options are: bronze, silver, gold.", *layer)))
				return
			}

			var tableList []string
			if *tables != "" {
				tableList = strings.Split(*tables, ",")
			}

			if *layer == "" && len(tableList) > 0 {
				c.Println(misc.Red("Flag --layer is required when using --tables."))
				return
			}

			if *layer != "" && len(tableList) > 0 {
				c.Println(misc.Blue(fmt.Sprintf("You are about to create or update %d table(s) in layer '%s':", len(tableList), *layer)))
				for _, tbl := range tableList {
					c.Println("-", tbl)
				}
			} else if *layer != "" {
				c.Println(misc.Blue(fmt.Sprintf("You are about to sync all tables in layer: %s", *layer)))
			} else {
				c.Println(misc.Blue("You are about to sync all databases:"))
				c.Println("- bronze")
				c.Println("- silver")
				c.Println("- gold")
			}

			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Catalog creation cancelled.\n"))
				return
			}

			db, err := core.GetConnection()
			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Error connecting to database: %v", err)))
			}

			if *layer != "" {
				if err := core.SyncCatalogFromDatabaseSchema(db, *layer, tableList...); err != nil {
					c.Println(misc.Red(fmt.Sprintf("Error: %v", err)))
					return
				}
			} else {
				for _, layerType := range []string{misc.LayerBronze, misc.LayerSilver, misc.LayerGold} {
					if err := core.SyncCatalogFromDatabaseSchema(db, layerType); err != nil {
						c.Println(misc.Red(fmt.Sprintf("Error: %v", err)))
						return
					}
				}
			}

			c.Println(misc.Green("Catalog creation completed successfully.\n"))

		}),
	}
}
