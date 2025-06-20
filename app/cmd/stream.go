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
			scenario := fs.String("mode", "profile", "Simulation scenario to run (e.g., profile, review)")
			rps := fs.Int("rps", 10, "Records per second per stream")
			duration := fs.Duration("duration", 30*time.Second, "Total duration of the simulation")
			parallel := fs.Int("parallel", 1, "Number of goroutines per scenario")
			quality := fs.String("quality", "high", "Data quality level: high, medium, low")

			if !ParseShellFlags(c, fs) {
				return
			}

			opts := core.SimOptions{
				RPS:      *rps,
				Duration: *duration,
				Quality:  *quality,
				Parallel: *parallel,
			}

			if scenarioFn, ok := core.SupportedStreams[*scenario]; ok {
				c.Println(misc.Blue(fmt.Sprintf("Running scenario: %s", *scenario)))
				for i := 0; i < opts.Parallel; i++ {
					go scenarioFn(opts)
				}
			} else {
				c.Println(misc.Red(fmt.Sprintf("Invalid scenario: %s", *scenario)))
				return
			}

			time.Sleep(opts.Duration + 2*time.Second)
			c.Println(misc.Green("Streaming finished.\n"))

		}),
	}
}
