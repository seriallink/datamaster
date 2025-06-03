package bronze

import "github.com/google/uuid"

func init() {
	Register(&Customer{})
}

type Customer struct {
	CustomerID    string `json:"customer_id"    parquet:"name=customer_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	CustomerName  string `json:"customer_name"  parquet:"name=customer_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	CustomerEmail string `json:"customer_email" parquet:"name=customer_email, type=BYTE_ARRAY, convertedtype=UTF8"`
	CustomerPhone string `json:"customer_phone" parquet:"name=customer_phone, type=BYTE_ARRAY, convertedtype=UTF8"`
	Gender        string `json:"gender"         parquet:"name=gender, type=BYTE_ARRAY, convertedtype=UTF8"`
	CreatedAt     string `json:"created_at"     parquet:"name=created_at, type=BYTE_ARRAY, convertedtype=UTF8"`
	UpdatedAt     string `json:"updated_at"     parquet:"name=updated_at, type=BYTE_ARRAY, convertedtype=UTF8"`
	Operation     string `json:"operation"      parquet:"name=operation, type=BYTE_ARRAY, convertedtype=UTF8"`
}

func (m *Customer) TableName() string {
	return "dm_bronze.customer"
}

func (m *Customer) TableId() uuid.UUID {
	return uuid.MustParse("def5ae04-92ce-40ce-8430-d28a04604324")
}
