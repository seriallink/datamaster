package cmd

import (
	"embed"
	"flag"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// GrafanaCmd returns an interactive shell command that creates or updates Grafana dashboards.
// The user can optionally specify a single dashboard name with the --dashboard flag.
// If no dashboard is specified, all available dashboards embedded in the filesystem will be pushed.
func GrafanaCmd(dashboards embed.FS) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "grafana",
		Help: "Create or update dashboards in Grafana",
		Func: WithAuth(func(c *ishell.Context) {

			fs := flag.NewFlagSet("grafana", flag.ContinueOnError)
			dashboard := fs.String("dashboard", "", "Optional dashboard name to create (e.g. analytics, logs, costs)")
			if !ParseShellFlags(c, fs) {
				return
			}

			if *dashboard != "" {
				c.Println(misc.Blue(fmt.Sprintf("You are about to create the Grafana dashboard: %s", *dashboard)))
			} else {
				c.Println(misc.Blue("You are about to create all available Grafana dashboards."))
			}
			c.Print("Type 'go' to continue: ")
			if strings.ToLower(c.ReadLine()) != "go" {
				c.Println(misc.Red("Dashboard creation cancelled.\\n"))
				return
			}

			var err error
			if *dashboard == "" {
				err = core.PushAllDashboards(dashboards)
			} else {
				err = core.PushDashboard(*dashboard, dashboards)
			}

			if err != nil {
				c.Println(misc.Red(fmt.Sprintf("Failed to create dashboard %s: %v", *dashboard, err)))
				return
			}

		}),
	}

}
