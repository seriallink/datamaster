package cmd

import "github.com/abiosoft/ishell"

func ClearCmd(shell *ishell.Shell) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "clear",
		Help: "Clear the screen",
		Func: func(c *ishell.Context) {
			_ = shell.ClearScreen()
		},
	}
}
