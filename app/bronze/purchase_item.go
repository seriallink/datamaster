package bronze

import "github.com/google/uuid"

func init() {
	Register(&PurchaseItem{})
}

type PurchaseItem struct {
	PurchaseID string `json:"purchase_id" parquet:"purchase_id,optional,snappy"`
	ProductID  string `json:"product_id"  parquet:"product_id,optional,snappy"`
	Quantity   string `json:"quantity"    parquet:"quantity,optional,snappy"`
	UnitPrice  string `json:"unit_price"  parquet:"unit_price,optional,snappy"`
	TotalPrice string `json:"total_price" parquet:"total_price,optional,snappy"`
	CreatedAt  string `json:"created_at"  parquet:"created_at,optional,snappy"`
	UpdatedAt  string `json:"updated_at"  parquet:"updated_at,optional,snappy"`
	Operation  string `json:"operation"   parquet:"operation,optional,snappy,dict"`
}

func (m *PurchaseItem) TableName() string {
	return "dm_bronze.purchase_item"
}

func (m *PurchaseItem) TableId() uuid.UUID {
	return uuid.MustParse("b04b185a-ff56-4541-b9e7-4bd0151095b2")
}
