package model

import (
	"encoding/json"
)

func init() {
	Register(&Beer{})
}

type Beer struct {
	BeerID    json.Number  `json:"beer_id"    gorm:"column:beer_id;primaryKey"`
	BreweryID json.Number  `json:"brewery_id" gorm:"column:brewery_id"`
	BeerName  string       `json:"beer_name"  gorm:"column:beer_name"`
	BeerStyle string       `json:"beer_style" gorm:"column:beer_style"`
	BeerABV   *json.Number `json:"beer_abv"   gorm:"column:beer_abv"`
}

func (Beer) TableName() string {
	return "dm_core.beer"
}
