package dialect

import "github.com/jackc/pgtype"

type PgEnum struct {
	Oid           pgtype.OID    `json:"oid"           gorm:"column:oid;type:oid;primary_key"`
	EnumTypId     pgtype.OID    `json:"enumtypid"     gorm:"column:enumtypid;type:oid"`
	EnumSortOrder pgtype.Float4 `json:"enumsortorder" gorm:"column:enumsortorder;type:float4"`
	EnumLabel     pgtype.Name   `json:"enumlabel"     gorm:"column:enumlabel;type:name"`
}

func (m *PgEnum) TableName() string {
	return "pg_catalog.pg_enum"
}
