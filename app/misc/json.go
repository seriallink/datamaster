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
