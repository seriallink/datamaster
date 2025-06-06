package model

type Brewery struct {
	BreweryID   int64  `gorm:"column:brewery_id;primaryKey"`
	BreweryName string `gorm:"column:brewery_name"`
}

func (Brewery) TableName() string {
	return "dm_core.brewery"
}
