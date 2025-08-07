package model

import (
	"encoding/json"
)

func init() {
	Register(&Beer{})
}

// Beer represents a record from the "beer" table in the raw dataset.
//
// Each field corresponds to a column in the source data and is annotated
// for both JSON unmarshaling and GORM ORM mapping. The `Beer` struct is
// typically used for ingestion into the bronze layer of the data lake.
//
// Fields:
//   - BeerID: Unique identifier of the beer (primary key).
//   - BreweryID: Foreign key referencing the brewery that produced the beer.
//   - BeerName: Name of the beer.
//   - BeerStyle: Style or category of the beer (e.g., IPA, Lager).
//   - BeerABV: Alcohol by volume (ABV) percentage of the beer, optional.
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
