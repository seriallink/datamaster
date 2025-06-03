package misc

import (
	"fmt"
)

func NameWithDefaultPrefix(name string, separator rune) string {
	return fmt.Sprintf("%s%c%s", DefaultProjectPrefix, separator, name)
}
