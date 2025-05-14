package misc

import "strings"

func Trim(s string) string {
	return strings.TrimSpace(Dedent(s))
}

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
