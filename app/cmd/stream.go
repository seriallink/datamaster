package cmd

import (
	"flag"
	"fmt"
	"time"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

func StreamCmd() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "stream",
		Help: "Simulate streaming data into Aurora tables",
		Func: WithAuth(func(c *ishell.Context) {
			fs := flag.NewFlagSet("stream", flag.ContinueOnError)
			scenario := fs.String("scenario", "all", "Simulation scenario to run (e.g., customer, product, purchase, delivery)")
			rps := fs.Int("rps", 10, "Records per second per stream")
			duration := fs.Duration("duration", 30*time.Second, "Total duration of the simulation")
			parallel := fs.Int("parallel", 1, "Number of goroutines per scenario")
			quality := fs.String("quality", "high", "Data quality level: high, medium, low")
			eventType := fs.String("event-type", "insert", "Event type: insert, update, delete or mix")

			if !ParseShellFlags(c, fs) {
				return
			}

			opts := core.SimOptions{
				RPS:       *rps,
				Duration:  *duration,
				Quality:   *quality,
				EventType: *eventType,
				Parallel:  *parallel,
			}

			if *scenario == "all" {
				c.Println(misc.Yellow("Streaming all supported scenarios..."))

				for name, scenarioFn := range core.SupportedScenarios {
					c.Println(misc.Blue(fmt.Sprintf("Running scenario: %s", name)))
					for i := 0; i < opts.Parallel; i++ {
						go scenarioFn(opts)
					}
				}
			} else {
				if scenarioFn, ok := core.SupportedScenarios[*scenario]; ok {
					c.Println(misc.Blue(fmt.Sprintf("Running scenario: %s", *scenario)))
					for i := 0; i < opts.Parallel; i++ {
						go scenarioFn(opts)
					}
				} else {
					c.Println(misc.Red(fmt.Sprintf("Invalid scenario: %s", *scenario)))
					return
				}
			}

			time.Sleep(opts.Duration + 2*time.Second)
			c.Println(misc.Green("Streaming finished.\n"))
		}),
	}
}
