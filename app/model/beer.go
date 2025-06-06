package model

type Beer struct {
	BeerID    int64   `gorm:"column:beer_id;primaryKey"`
	BreweryID int64   `gorm:"column:brewery_id"`
	BeerName  string  `gorm:"column:beer_name"`
	BeerStyle string  `gorm:"column:beer_style"`
	BeerABV   float64 `gorm:"column:beer_abv"`
}

func (Beer) TableName() string {
	return "dm_core.beer"
}
