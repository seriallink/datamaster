package cmd

import "github.com/abiosoft/ishell"

func ExitCmd(shell *ishell.Shell) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "exit",
		Help: "Exit the program",
		Func: func(c *ishell.Context) {
			c.Println("Bye!")
			shell.Stop()
		},
	}
}
