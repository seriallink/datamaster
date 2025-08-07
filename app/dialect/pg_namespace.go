package dialect

import "github.com/jackc/pgtype"

// PgNamespace represents a schema in PostgreSQL, as defined in the pg_namespace system catalog.
//
// Each record identifies a logical namespace used to group database objects
// such as tables, types, and functions.
//
// This model is used to resolve fully qualified names and manage object scoping
// during schema extraction and Glue Catalog generation.
type PgNamespace struct {
	Oid      pgtype.OID          `json:"oid"      gorm:"column:oid;type:oid;primary_key"`
	NspName  pgtype.Name         `json:"nspname"  gorm:"column:nspname;type:name"`
	NspOwner pgtype.OID          `json:"nspowner" gorm:"column:nspowner;type:oid"`
	NspAcl   pgtype.ACLItemArray `json:"nspacl"   gorm:"column:nspacl;type:_aclitem"`
}

func (m *PgNamespace) TableName() string {
	return "pg_catalog.pg_namespace"
}
