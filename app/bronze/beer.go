package bronze

import "errors"

func init() {
	Register(&Beer{})
}

type Beer struct {
	BeerID    int64   `json:"beer_id"    parquet:"beer_id,optional,snappy"`
	BreweryID int64   `json:"brewery_id" parquet:"brewery_id,optional,snappy"`
	BeerName  string  `json:"beer_name"  parquet:"beer_name,optional,snappy"`
	BeerStyle string  `json:"beer_style" parquet:"beer_style,optional,snappy"`
	BeerAbv   float64 `json:"beer_abv"   parquet:"beer_abv,optional,snappy,dict"`
	Operation string  `json:"operation"  parquet:"operation,optional,snappy,dict"`
}

func (m *Beer) TableName() string {
	return "dm_bronze.beer"
}

func (m *Beer) FromCSV(line []string, idx map[string]int) (any, error) {
	return nil, errors.New("FromCSV not implemented for Beer model")
}
