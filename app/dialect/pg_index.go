package dialect

import (
	"github.com/jackc/pgtype"
)

// PgIndex represents an index definition in the PostgreSQL system catalog (pg_index).
//
// It describes structural properties of an index, including uniqueness,
// clustering, validity, replication identity, and expression-based logic.
// It also stores references to indexed columns, collation, access methods,
// and predicate conditions (for partial indexes).
//
// This model is used to reconstruct index metadata during schema introspection
// and to understand constraint enforcement or performance optimization strategies.
type PgIndex struct {
	IndexRelId     pgtype.OID    `json:"indexrelid"              gorm:"column:indexrelid;type:oid;primary_key"`
	IndRelId       pgtype.OID    `json:"indrelid"                gorm:"column:indrelid;type:oid"`
	IndNAtts       pgtype.Int2   `json:"indnatts"                gorm:"column:indnatts;type:int2"`
	IndNKeyAtts    pgtype.Int2   `json:"indnkeyatts"             gorm:"column:indnkeyatts;type:int2"`
	IndIsUnique    pgtype.Bool   `json:"indisunique"             gorm:"column:indisunique;type:bool"`
	IndIsPrimary   pgtype.Bool   `json:"indisprimary"            gorm:"column:indisprimary;type:bool"`
	IndIsExclusion pgtype.Bool   `json:"indisexclusion"          gorm:"column:indisexclusion;type:bool"`
	IndImmediate   pgtype.Bool   `json:"indimmediate"            gorm:"column:indimmediate;type:bool"`
	IndIsClustered pgtype.Bool   `json:"indisclustered"          gorm:"column:indisclustered;type:bool"`
	IndIsValid     pgtype.Bool   `json:"indisvalid"              gorm:"column:indisvalid;type:bool"`
	IndCheckxmin   pgtype.Bool   `json:"indcheckxmin"            gorm:"column:indcheckxmin;type:bool"`
	IndIsReady     pgtype.Bool   `json:"indisready"              gorm:"column:indisready;type:bool"`
	IndIsLive      pgtype.Bool   `json:"indislive"               gorm:"column:indislive;type:bool"`
	IndIsReplIdent pgtype.Bool   `json:"indisreplident"          gorm:"column:indisreplident;type:bool"`
	IndKey         pgtype.Vec2   `json:"indkey"                  gorm:"column:indkey;type:int2vector"`
	IndCollation   pgtype.Vec2   `json:"indcollation"            gorm:"column:indcollation;type:oidvector"`
	IndClass       pgtype.Vec2   `json:"indclass"                gorm:"column:indclass;type:oidvector"`
	IndOption      pgtype.Vec2   `json:"indoption"               gorm:"column:indoption;type:int2vector"`
	IndExprs       pgtype.Text   `json:"indexprs"                gorm:"column:indexprs;type:pg_node_tree"`
	IndPred        pgtype.Text   `json:"indpred"                 gorm:"column:indpred;type:pg_node_tree"`
	PgAttributes   []PgAttribute `json:"pg_attributes,omitempty" gorm:"-"`
}

func (m *PgIndex) TableName() string {
	return "pg_catalog.pg_index"
}

//func (m *PgIndex) AfterFind(db *gorm.DB) error {
//
//	// set model for searching
//	criteria := &PgAttribute{
//		AttRelId: m.IndexRelId,
//	}
//
//	return db.Where(criteria).Where("attnum IN (?)", m.IndKey).Find(&m.PgAttributes).Error
//
//}
