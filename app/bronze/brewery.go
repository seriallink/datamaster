package bronze

import "errors"

func init() {
	Register(&Brewery{})
}

// Brewery represents the schema of the brewery dataset in the bronze layer.
//
// This model is used for Parquet serialization and includes the minimum
// information necessary to identify and track breweries in the pipeline.
//
// Fields:
//   - BreweryId: Unique identifier for the brewery.
//   - BreweryName: Name of the brewery.
//   - Operation: Type of change (insert, update, delete) for CDC tracking.
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
