package dialect

import "github.com/jackc/pgtype"

// PgEnum represents a single value of an enumerated type in PostgreSQL,
// as defined in the pg_enum system catalog.
//
// Each entry links to a specific enum type and defines its label and sort order.
// This model is used to resolve and document enum-based column types
// when generating schema definitions for downstream systems like the Glue Catalog.
type PgEnum struct {
	Oid           pgtype.OID    `json:"oid"           gorm:"column:oid;type:oid;primary_key"`
	EnumTypId     pgtype.OID    `json:"enumtypid"     gorm:"column:enumtypid;type:oid"`
	EnumSortOrder pgtype.Float4 `json:"enumsortorder" gorm:"column:enumsortorder;type:float4"`
	EnumLabel     pgtype.Name   `json:"enumlabel"     gorm:"column:enumlabel;type:name"`
}

func (m *PgEnum) TableName() string {
	return "pg_catalog.pg_enum"
}
