package dialect

import "github.com/jackc/pgtype"

// PgDescription represents a user-defined comment or description attached to a database object,
// as stored in the pg_description system catalog.
//
// It can be associated with tables, columns, types, and other entities,
// and is used to enrich schema metadata with human-readable annotations.
// These descriptions can be extracted and propagated to the Glue Catalog or data documentation layers.
type PgDescription struct {
	ObjOid      pgtype.OID  `json:"objoid,omitempty"      gorm:"column:objoid;type:oid"`
	ClassOid    pgtype.OID  `json:"classoid,omitempty"    gorm:"column:classoid;type:oid"`
	ObjSubId    pgtype.Int4 `json:"objsubid,omitempty"    gorm:"column:objsubid;type:int4"`
	Description pgtype.Text `json:"description,omitempty" gorm:"column:description;type:text"`
}

func (m *PgDescription) TableName() string {
	return "pg_catalog.pg_description"
}
