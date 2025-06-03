package core

import (
	"fmt"
	"reflect"

	"github.com/seriallink/datamaster/cli/misc"
)

func UnmarshalRecords(model any, data []map[string]any) (any, error) {

	// expected return type
	rt := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(model)).Type())

	// init slice with the correct type
	slice := reflect.New(rt).Interface()

	err := misc.Copier(slice, data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal records: %w", err)
	}

	return slice, nil

}
