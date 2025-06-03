package misc

import "encoding/json"

func Copier(instance any, data any) error {

	// marshal any data to JSON
	b, err := json.Marshal(data)

	// unmarshal data to a given instance
	if err == nil {
		err = json.Unmarshal(b, instance)
	}

	return err

}
