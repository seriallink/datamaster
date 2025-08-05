package misc

// In checks whether a given value exists within a list of values.
//
// Parameters:
//   - value: the target value to search for.
//   - list: a variadic list of values to search within.
//
// Returns:
//   - bool: true if the value is found in the list; false otherwise.
func In(value any, list ...any) bool {
	for _, element := range list {
		if value == element {
			return true
		}
	}
	return false
}

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

// Ternary returns one of two values based on a boolean condition.
//
// This function mimics the ternary (conditional) operator found in other languages,
// returning `x` if the condition is true, or `y` otherwise. It accepts any type
// as input and returns an interface{} (`any`), so the caller is responsible for type assertion if needed.
//
// Parameters:
//   - condition: boolean expression to evaluate.
//   - x: value returned if the condition is true.
//   - y: value returned if the condition is false.
//
// Returns:
//   - any: `x` if condition is true; otherwise `y`.
func Ternary(condition bool, x, y any) any {
	if condition {
		return x
	}
	return y
}
