package bronze

import "github.com/google/uuid"

func init() {
	Register(&Delivery{})
}

type Delivery struct {
	DeliveryID            string `json:"delivery_id"             parquet:"delivery_id,optional,snappy"`
	PurchaseID            string `json:"purchase_id"             parquet:"purchase_id,optional,snappy"`
	CourierID             string `json:"courier_id"              parquet:"courier_id,optional,snappy"`
	DeliveryStatus        string `json:"delivery_status"         parquet:"delivery_status,optional,snappy,dict"`
	DeliveryEstimatedDate string `json:"delivery_estimated_date" parquet:"delivery_estimated_date,optional,snappy"`
	DeliveryDate          string `json:"delivery_date"           parquet:"delivery_date,optional,snappy"`
	CreatedAt             string `json:"created_at"              parquet:"created_at,optional,snappy"`
	UpdatedAt             string `json:"updated_at"              parquet:"updated_at,optional,snappy"`
	Operation             string `json:"operation"               parquet:"operation,optional,snappy,dict"`
}

func (m *Delivery) TableName() string {
	return "dm_bronze.delivery"
}

func (m *Delivery) TableId() uuid.UUID {
	return uuid.MustParse("6892583b-c67c-4716-b038-d7ec283e765e")
}
