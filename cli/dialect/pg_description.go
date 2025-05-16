package dialect

import "github.com/jackc/pgtype"

type PgDescription struct {
	ObjOid      pgtype.OID  `json:"objoid,omitempty"      gorm:"column:objoid;type:oid"`
	ClassOid    pgtype.OID  `json:"classoid,omitempty"    gorm:"column:classoid;type:oid"`
	ObjSubId    pgtype.Int4 `json:"objsubid,omitempty"    gorm:"column:objsubid;type:int4"`
	Description pgtype.Text `json:"description,omitempty" gorm:"column:description;type:text"`
}

func (m *PgDescription) TableName() string {
	return "pg_catalog.pg_description"
}
