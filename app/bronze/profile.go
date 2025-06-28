package bronze

import "errors"

func init() {
	Register(&Profile{})
}

type Profile struct {
	ProfileID   int64  `json:"profile_id"   parquet:"profile_id"`
	ProfileName string `json:"profile_name" parquet:"profile_name"`
	Email       string `json:"email"        parquet:"email"`
	State       string `json:"state"        parquet:"state"`
	Operation   string `json:"operation"    parquet:"operation,dict"`
}

func (m *Profile) TableName() string {
	return "dm_bronze.profile"
}

func (m *Profile) FromCSV(line []string, idx map[string]int) (any, error) {
	return nil, errors.New("FromCSV not implemented for Profile model")
}
