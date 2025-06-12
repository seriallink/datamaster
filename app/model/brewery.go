package model

import (
	"encoding/json"
)

func init() {
	Register(&Brewery{})
}

type Brewery struct {
	BreweryID   json.Number `json:"brewery_id"   gorm:"column:brewery_id;primaryKey"`
	BreweryName string      `json:"brewery_name" gorm:"column:brewery_name"`
}

func (Brewery) TableName() string {
	return "dm_core.brewery"
}
