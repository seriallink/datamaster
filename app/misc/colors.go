package misc

import "github.com/fatih/color"

// Yellow prints formatted text in yellow.
// Usage: fmt.Println(Yellow("your message"))
var Yellow = color.New(color.FgYellow).SprintfFunc()

// Red prints formatted text in red.
// Usage: fmt.Println(Red("your message"))
var Red = color.New(color.FgRed).SprintfFunc()

// Green prints formatted text in green.
// Usage: fmt.Println(Green("your message"))
var Green = color.New(color.FgGreen).SprintfFunc()

// Blue prints formatted text in blue.
// Usage: fmt.Println(Blue("your message"))
var Blue = color.New(color.FgBlue).SprintfFunc()
