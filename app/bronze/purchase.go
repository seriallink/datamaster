package bronze

import "github.com/google/uuid"

func init() {
	Register(&Purchase{})
}

type Purchase struct {
	PurchaseID     string `json:"purchase_id"     parquet:"purchase_id,optional,snappy"`
	CustomerID     string `json:"customer_id"     parquet:"customer_id,optional,snappy"`
	PurchaseStatus string `json:"purchase_status" parquet:"purchase_status,optional,snappy,dict"`
	TotalValue     string `json:"total_value"     parquet:"total_value,optional,snappy"`
	PurchaseDate   string `json:"purchase_date"   parquet:"purchase_date,optional,snappy"`
	CreatedAt      string `json:"created_at"      parquet:"created_at,optional,snappy"`
	UpdatedAt      string `json:"updated_at"      parquet:"updated_at,optional,snappy"`
	Operation      string `json:"operation"       parquet:"operation,optional,snappy,dict"`
}

func (m *Purchase) TableName() string {
	return "dm_bronze.purchase"
}

func (m *Purchase) TableId() uuid.UUID {
	return uuid.MustParse("90456ba3-b9c5-4031-ae5f-60b68e22d901")
}
