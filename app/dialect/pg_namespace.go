package dialect

import "github.com/jackc/pgtype"

type PgNamespace struct {
	Oid      pgtype.OID          `json:"oid"      gorm:"column:oid;type:oid;primary_key"`
	NspName  pgtype.Name         `json:"nspname"  gorm:"column:nspname;type:name"`
	NspOwner pgtype.OID          `json:"nspowner" gorm:"column:nspowner;type:oid"`
	NspAcl   pgtype.ACLItemArray `json:"nspacl"   gorm:"column:nspacl;type:_aclitem"`
}

func (m *PgNamespace) TableName() string {
	return "pg_catalog.pg_namespace"
}
