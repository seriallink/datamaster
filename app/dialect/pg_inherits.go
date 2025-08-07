package dialect

import "github.com/jackc/pgtype"

// PgInherits represents table inheritance relationships in PostgreSQL,
// as defined in the pg_inherits system catalog.
//
// Each row indicates that a child table inherits from a parent table,
// preserving column structure and optionally constraints.
// This model is used to resolve hierarchical schemas and detect inherited tables
// during catalog introspection.
type PgInherits struct {
	InhRelid         pgtype.OID  `json:"inhrelid"         gorm:"column:inhrelid;type:oid"`
	InhParent        pgtype.OID  `json:"inhparent"        gorm:"column:inhparent;type:oid"`
	InhSeqNo         pgtype.Int4 `json:"inhseqno"         gorm:"column:inhseqno;type:int4"`
	InhDetachPending pgtype.Bool `json:"inhdetachpending" gorm:"column:inhdetachpending;type:bool"`
}

func (m *PgInherits) TableName() string {
	return "pg_catalog.pg_inherits"
}
