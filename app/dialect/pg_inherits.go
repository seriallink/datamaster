package dialect

import "github.com/jackc/pgtype"

type PgInherits struct {
	InhRelid         pgtype.OID  `json:"inhrelid"         gorm:"column:inhrelid;type:oid"`
	InhParent        pgtype.OID  `json:"inhparent"        gorm:"column:inhparent;type:oid"`
	InhSeqNo         pgtype.Int4 `json:"inhseqno"         gorm:"column:inhseqno;type:int4"`
	InhDetachPending pgtype.Bool `json:"inhdetachpending" gorm:"column:inhdetachpending;type:bool"`
}

func (m *PgInherits) TableName() string {
	return "pg_catalog.pg_inherits"
}
