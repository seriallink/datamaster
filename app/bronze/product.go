package bronze

import "github.com/google/uuid"

func init() {
	Register(&Product{})
}

type Product struct {
	ProductID    string `json:"product_id"    parquet:"product_id,optional,snappy"`
	CategoryID   string `json:"category_id"   parquet:"category_id,optional,snappy"`
	ProductName  string `json:"product_name"  parquet:"product_name,optional,snappy"`
	CurrentStock string `json:"current_stock" parquet:"current_stock,optional,snappy"`
	UnitPrice    string `json:"unit_price"    parquet:"unit_price,optional,snappy"`
	CreatedAt    string `json:"created_at"    parquet:"created_at,optional,snappy"`
	UpdatedAt    string `json:"updated_at"    parquet:"updated_at,optional,snappy"`
	Operation    string `json:"operation"     parquet:"operation,optional,snappy,dict"`
}

func (m *Product) TableName() string {
	return "dm_bronze.product"
}

func (m *Product) TableId() uuid.UUID {
	return uuid.MustParse("cc5ea8c8-95d5-445f-95f4-f522abc5897e")
}
