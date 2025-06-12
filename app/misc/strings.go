package misc

import "strings"

// Trim removes leading/trailing whitespace from the input string,
// after first removing any common indentation using Dedent.
//
// It is useful for formatting multi-line string literals where indentation
// is used for code readability but should not be preserved.
//
// Example:
//
//	input := `
//	    SELECT *
//	    FROM users
//	`
//	cleaned := Trim(input)
//
// Returns:
//   - string: trimmed and dedented version of the input.
func Trim(s string) string {
	return strings.TrimSpace(Dedent(s))
}

// Dedent removes common leading whitespace from every line of the input string.
// This is typically used for multi-line strings defined inline in code.
//
// It finds the smallest indentation (spaces or tabs) among all non-empty lines
// and removes that indentation from every line.
//
// Example:
//
//	input := `
//	    line 1
//	    line 2
//	`
//	output := Dedent(input)
//
// Returns:
//   - string: dedented version of the input.
func Dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1
	for _, line := range lines {
		if trimmed := strings.TrimLeft(line, " \t"); trimmed != "" {
			indent := len(line) - len(trimmed)
			if minIndent == -1 || indent < minIndent {
				minIndent = indent
			}
		}
	}
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}
	return strings.Join(lines, "\n")
}

func ToPascalCase(input string) string {
	parts := strings.Split(input, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
