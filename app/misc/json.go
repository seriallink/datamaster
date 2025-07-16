package misc

import "encoding/json"

// Copier copies data from a source into a target instance by marshaling to JSON and unmarshaling.
//
// Parameters:
//   - instance: any - The target instance to populate (must be a pointer).
//   - data: any - The source data to copy from.
//
// Returns:
//   - error: an error if the copy operation fails, or nil if successful.
func Copier(instance any, data any) error {

	// marshal any data to JSON
	b, err := json.Marshal(data)

	// unmarshal data to a given instance
	if err == nil {
		err = json.Unmarshal(b, instance)
	}

	return err

}

// Stringify converts any Go data structure to its string representation.
//
// If the input is already a string, it returns the value directly. Otherwise, it
// marshals the data to a JSON-formatted string.
//
// Parameters:
//   - data: interface{} - The value to stringify.
//
// Returns:
//   - string: The string or JSON-encoded representation of the input.
//
// Panics:
//   - If the data cannot be marshaled to JSON, the function will panic.
func Stringify(data interface{}) string {

	// data is already a string
	if str, ok := data.(string); ok {
		return str
	}

	var bytes []byte

	bytes, err := json.Marshal(data)

	if err != nil {
		panic(err)
	}

	return string(bytes)

}
