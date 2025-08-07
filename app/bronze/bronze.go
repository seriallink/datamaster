// Package bronze defines the strongly typed data models used for the bronze layer
// of the Data Master pipeline.
//
// Each model represents a specific raw dataset (e.g., beer, brewery, profile, review)
// and is mapped to a corresponding Parquet schema via struct tags.
//
// The package also provides a dynamic model registry that allows runtime loading
// and instantiation of models based on the fully qualified table name.
//
// All models implement the Model interface, which supports schema-aware
// conversions from CSV and JSON inputs into typed Go structs.
package bronze

import (
	"fmt"
	"sync"
)

var models sync.Map

type Model interface {
	TableName() string
	FromCSV([]string, map[string]int) (any, error)
}

// Register stores a model implementation in the global registry for later access.
//
// It ensures that each model is registered only once, using the fully qualified table name
// (e.g., "schema.table") as the key. If a model is nil or has already been registered,
// the function panics to prevent silent misconfiguration.
//
// Parameters:
//   - m: A model implementing the Model interface. Must not be nil and must have a unique TableName.
//
// Usage:
//
//	Call this during application startup to make the model available for schema introspection,
//	dynamic mapping, and downstream processing logic.
func Register(m Model) {

	// validate model
	if m == nil {
		panic("nil model")
	}

	// check if the model was not registered previously
	if _, exists := models.Load(m.TableName()); exists {
		panic("register called twice for model " + m.TableName())
	}

	// register model
	models.Store(m.TableName(), m)

}

// LoadModel retrieves a registered model by its fully qualified table name.
//
// It looks up the model from the global registry using the format "schema.table".
// If the model is not found, it returns an error indicating that the model
// has not been registered.
//
// Parameters:
//   - fullyQualifiedTableName: The unique identifier for the model in the format "schema.table".
//
// Returns:
//   - Model: The registered model instance.
//   - error: An error if the model is not found in the registry.
func LoadModel(fullyQualifiedTableName string) (Model, error) {

	// load model
	md, ok := models.Load(fullyQualifiedTableName)

	// the model was not loaded
	if !ok {
		return nil, fmt.Errorf("model not registered: %s", fullyQualifiedTableName)
	}

	// model not found
	return md.(Model), nil

}
