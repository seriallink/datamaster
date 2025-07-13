package cmd

import "github.com/abiosoft/ishell"

// ClearCmd returns an interactive shell command that clears the screen output.
// It invokes the ClearScreen method from the provided shell instance.
func ClearCmd(shell *ishell.Shell) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "clear",
		Help: "Clear the screen",
		Func: func(c *ishell.Context) {
			_ = shell.ClearScreen()
		},
	}
}
