package model

import (
	"encoding/json"
)

func init() {
	Register(&Brewery{})
}

// Brewery represents a record from the "brewery" table in the raw dataset.
//
// This struct is used to map and store brewery metadata, including unique identifiers
// and names, during the ingestion process into the bronze layer.
//
// Fields:
//   - BreweryID: Unique identifier of the brewery (primary key).
//   - BreweryName: Name of the brewery.
type Brewery struct {
	BreweryID   json.Number `json:"brewery_id"   gorm:"column:brewery_id;primaryKey"`
	BreweryName string      `json:"brewery_name" gorm:"column:brewery_name"`
}

func (Brewery) TableName() string {
	return "dm_core.brewery"
}
