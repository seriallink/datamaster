package model

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/google/uuid"
)

var models sync.Map

type Model interface {
	TableId() uuid.UUID
	TableName() string
}

func Register(m Model) {

	// validate model
	if m == nil {
		panic("nil model")
	}

	// check if the model was not registered previously
	if _, exists := models.Load(m.TableId()); exists {
		panic("register called twice for model " + m.TableName())
	}

	// register model
	models.Store(m.TableId(), m)

}

func LoadModel(modelID uuid.UUID) (Model, error) {

	// load model
	md, ok := models.Load(modelID)

	// the model was not loaded
	if !ok {
		return nil, fmt.Errorf("model not registered: %s", modelID.String())
	}

	// model not found
	return md.(Model), nil

}

func LoadInstance(modelID uuid.UUID) (Model, error) {

	// load model
	md, err := LoadModel(modelID)

	// the model could not be loaded
	if err != nil {
		return nil, err
	}

	// return a new instance of the model
	return reflect.New(reflect.TypeOf(md).Elem()).Interface().(Model), nil

}
