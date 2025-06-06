package bronze

import "github.com/google/uuid"

func init() {
	Register(&Courier{})
}

type Courier struct {
	CourierID           string `json:"courier_id"            parquet:"courier_id,optional,snappy"`
	CourierName         string `json:"courier_name"          parquet:"courier_name,optional,snappy"`
	ContactName         string `json:"contact_name"          parquet:"contact_name,optional,snappy"`
	ContactEmail        string `json:"contact_email"         parquet:"contact_email,optional,snappy"`
	ContactPhone        string `json:"contact_phone"         parquet:"contact_phone,optional,snappy"`
	CostPerKM           string `json:"cost_per_km"           parquet:"cost_per_km,optional,snappy"`
	AverageDeliveryTime string `json:"average_delivery_time" parquet:"average_delivery_time,optional,snappy"`
	ReliabilityScore    int16  `json:"reliability_score"     parquet:"reliability_score,optional,snappy"`
	CreatedAt           string `json:"created_at"            parquet:"created_at,optional,snappy"`
	UpdatedAt           string `json:"updated_at"            parquet:"updated_at,optional,snappy"`
	Operation           string `json:"operation"             parquet:"operation,optional,snappy,dict"`
}

func (m *Courier) TableName() string {
	return "dm_bronze.courier"
}

func (m *Courier) TableId() uuid.UUID {
	return uuid.MustParse("ec024b0a-0c03-4d61-87b2-eec12b317abb")
}
