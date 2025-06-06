package bronze

import "github.com/google/uuid"

func init() {
	Register(&Customer{})
}

type Customer struct {
	CustomerID    string `json:"customer_id"    parquet:"customer_id,optional,snappy"`
	CustomerName  string `json:"customer_name"  parquet:"customer_name,optional,snappy"`
	CustomerEmail string `json:"customer_email" parquet:"customer_email,optional,snappy"`
	CustomerPhone string `json:"customer_phone" parquet:"customer_phone,optional,snappy"`
	Gender        string `json:"gender"         parquet:"gender,optional,snappy,dict"`
	CreatedAt     string `json:"created_at"     parquet:"created_at,optional,snappy"`
	UpdatedAt     string `json:"updated_at"     parquet:"updated_at,optional,snappy"`
	Operation     string `json:"operation"      parquet:"operation,optional,snappy,dict"`
}

func (m *Customer) TableName() string {
	return "dm_bronze.customer"
}

func (m *Customer) TableId() uuid.UUID {
	return uuid.MustParse("def5ae04-92ce-40ce-8430-d28a04604324")
}
