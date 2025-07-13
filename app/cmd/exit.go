package cmd

import "github.com/abiosoft/ishell"

// ExitCmd returns an interactive shell command that terminates the program.
// When executed, it prints a goodbye message and stops the shell session.
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
