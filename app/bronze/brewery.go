package bronze

import "errors"

func init() {
	Register(&Brewery{})
}

type Brewery struct {
	BreweryId   int64  `json:"brewery_id"   parquet:"brewery_id"`
	BreweryName string `json:"brewery_name" parquet:"brewery_name"`
	Operation   string `json:"operation"    parquet:"operation,dict"`
}

func (m *Brewery) TableName() string {
	return "dm_bronze.brewery"
}

func (m *Brewery) FromCSV(line []string, idx map[string]int) (any, error) {
	return nil, errors.New("FromCSV not implemented for Brewery model")
}
