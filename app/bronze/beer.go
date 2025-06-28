package bronze

import "errors"

func init() {
	Register(&Beer{})
}

type Beer struct {
	BeerID    int64   `json:"beer_id"    parquet:"beer_id"`
	BreweryID int64   `json:"brewery_id" parquet:"brewery_id"`
	BeerName  string  `json:"beer_name"  parquet:"beer_name"`
	BeerStyle string  `json:"beer_style" parquet:"beer_style,dict"`
	BeerAbv   float64 `json:"beer_abv"   parquet:"beer_abv,optional"`
	Operation string  `json:"operation"  parquet:"operation,dict"`
}

func (m *Beer) TableName() string {
	return "dm_bronze.beer"
}

func (m *Beer) FromCSV(line []string, idx map[string]int) (any, error) {
	return nil, errors.New("FromCSV not implemented for Beer model")
}
