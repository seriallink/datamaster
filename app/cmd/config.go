package cmd

import (
	"github.com/abiosoft/readline"
	"github.com/seriallink/datamaster/app/misc"
)

// CustomReadlineConfig returns a readline.Config with specific settings
// for interactive shells using ishell.
//
// It disables Meta key sequences that commonly interfere with normal input,
// especially on buggy terminals that send ESC-prefixed characters by default.
//
// Blocked runes include:
//   - MetaBackward (Alt+b): moves backward by word
//   - MetaForward (Alt+f): moves forward by word
//   - MetaDelete (Alt+d): deletes forward word
//   - MetaBackspace (Alt+Backspace): deletes backward word
//   - MetaTranspose (Alt+t): transposes characters
//
// Returns:
//   - *readline.Config: customized configuration for readline shell input.
func CustomReadlineConfig() *readline.Config {
	return &readline.Config{
		Prompt: ">>> ",
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if misc.In(r,
				readline.MetaBackward,
				readline.MetaForward,
				readline.MetaDelete,
				readline.MetaBackspace,
				readline.MetaTranspose) {
				return r, false
			}
			return r, true
		},
	}
}
