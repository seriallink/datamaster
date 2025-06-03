package cmd

import (
	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/misc"

	"github.com/abiosoft/ishell"
)

// WrapperFunc defines the signature for wrapped command handlers.
type WrapperFunc func(c *ishell.Context)

// WithAuth wraps a command handler with AWS auth validation logic.
func WithAuth(f WrapperFunc) func(c *ishell.Context) {
	return func(c *ishell.Context) {
		cfg := core.GetAWSConfig()
		if cfg.Credentials == nil {
			c.Println(misc.Red("No AWS credentials found. Please authenticate first.\n"))
			return
		}
		f(c)
	}
}
