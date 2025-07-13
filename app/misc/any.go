package misc

// NotIn checks if a value is not present in the provided list.
//
// Parameters:
//   - value: any - The value to check.
//   - list: ...any - A variadic list of values to compare against.
//
// Returns:
//   - bool: true if value is not in the list; false otherwise.
func NotIn(value any, list ...any) bool {
	for _, element := range list {
		if value == element {
			return false
		}
	}
	return true
}
