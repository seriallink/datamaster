package bronze

import "github.com/google/uuid"

func init() {
	Register(&Category{})
}

type Category struct {
	CategoryID          string `json:"category_id"          parquet:"category_id,optional,snappy"`
	CategoryName        string `json:"category_name"        parquet:"category_name,optional,snappy"`
	CategoryDescription string `json:"category_description" parquet:"category_description,optional,snappy"`
	CreatedAt           string `json:"created_at"           parquet:"created_at,optional,snappy"`
	UpdatedAt           string `json:"updated_at"           parquet:"updated_at,optional,snappy"`
	Operation           string `json:"operation"            parquet:"operation,optional,snappy,dict"`
}

func (m *Category) TableName() string {
	return "dm_bronze.category"
}

func (m *Category) TableId() uuid.UUID {
	return uuid.MustParse("e595191a-1d76-4ca5-a5b1-c88eaa604447")
}
