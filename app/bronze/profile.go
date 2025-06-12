package bronze

import "errors"

func init() {
	Register(&Profile{})
}

type Profile struct {
	ProfileID   int64  `json:"profile_id"   parquet:"profile_id,optional,snappy"`
	ProfileName string `json:"profile_name" parquet:"profile_name,optional,snappy"`
	Email       string `json:"email"        parquet:"email,optional,snappy"`
	State       string `json:"state"        parquet:"state,optional,snappy"`
	Operation   string `json:"operation"    parquet:"operation,optional,snappy,dict"`
}

func (m *Profile) TableName() string {
	return "dm_bronze.profile"
}

func (m *Profile) FromCSV(line []string, idx map[string]int) (any, error) {
	return nil, errors.New("FromCSV not implemented for Profile model")
}
