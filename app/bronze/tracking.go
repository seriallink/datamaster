package bronze

import "github.com/google/uuid"

func init() {
	Register(&Tracking{})
}

type Tracking struct {
	TrackingID     string `json:"tracking_id"     parquet:"tracking_id,optional,snappy"`
	DeliveryID     string `json:"delivery_id"     parquet:"delivery_id,optional,snappy"`
	Notes          string `json:"notes"           parquet:"notes,optional,snappy"`
	TrackingStatus string `json:"tracking_status" parquet:"tracking_status,optional,snappy,dict"`
	EventTimestamp string `json:"event_timestamp" parquet:"event_timestamp,optional,snappy"`
	CreatedAt      string `json:"created_at"      parquet:"created_at,optional,snappy"`
	UpdatedAt      string `json:"updated_at"      parquet:"updated_at,optional,snappy"`
	Operation      string `json:"operation"       parquet:"operation,optional,snappy,dict"`
}

func (m *Tracking) TableName() string {
	return "dm_bronze.tracking"
}

func (m *Tracking) TableId() uuid.UUID {
	return uuid.MustParse("2f4d5a52-f698-4c7c-b337-1d41a51481f5")
}
