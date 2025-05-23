package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
)

// ParseShellFlags parses shell flags and returns true if parsing was successful.
// It also prints errors in a consistent format using the misc.Red wrapper.
func ParseShellFlags(c *ishell.Context, fs *flag.FlagSet) bool {

	fs.SetOutput(io.Discard)

	if err := fs.Parse(c.Args); err != nil {
		c.Println(misc.Red(fmt.Sprintf("Invalid arguments: %v", err)))
		return false
	}

	if fs.NArg() > 0 {
		c.Println(misc.Red(fmt.Sprintf("Unexpected argument: %s", fs.Arg(0))))
		return false
	}

	return true

}
