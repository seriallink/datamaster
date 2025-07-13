package misc

import (
	"fmt"
)

// NameWithDefaultPrefix prepends the default project prefix to a name using the given separator.
//
// Parameters:
//   - name: string - The name to be prefixed.
//   - separator: rune - The character used to separate the prefix and the name.
//
// Returns:
//   - string: A string in the format "<prefix><separator><name>".
func NameWithDefaultPrefix(name string, separator rune) string {
	return fmt.Sprintf("%s%c%s", DefaultProjectPrefix, separator, name)
}
