package dialect

import "github.com/jackc/pgtype"

// PgDepend represents a dependency relationship between PostgreSQL objects,
// as defined in the pg_depend system catalog.
//
// It is used to track how objects such as constraints, indexes, and columns
// depend on one another, which is critical for understanding inheritance,
// cascading behavior, and internal linkage within the catalog.
//
// This model is used to resolve indirect relationships during schema introspection,
// such as linking constraints to indexes or types to attributes.
type PgDepend struct {
	ClassId       pgtype.OID     `json:"classid,omitempty"        gorm:"column:classid;type:oid"`
	ObjId         pgtype.OID     `json:"objid,omitempty"          gorm:"column:objid;type:oid"`
	ObjSubId      pgtype.OID     `json:"objsubid,omitempty"       gorm:"column:objsubid;type:oid"`
	RefClassId    pgtype.OID     `json:"refclassid,omitempty"     gorm:"column:refclassid;type:oid"`
	RefObjId      pgtype.OID     `json:"refobjid,omitempty"       gorm:"column:refobjid;type:oid"`
	RefObjSubId   pgtype.Int4    `json:"refobjsubid,omitempty"    gorm:"column:refobjsubid;type:int4"`
	DepType       pgtype.BPChar  `json:"deptype,omitempty"        gorm:"column:deptype;type:char"`
	PgConstraint  *PgConstraint  `json:"pg_constraint,omitempty"  gorm:"ForeignKey:Oid;References:ObjId"`
	PgClasses     []PgClass      `json:"pg_classes,omitempty"     gorm:"ForeignKey:Oid;References:ObjId"`
	PgConstraints []PgConstraint `json:"pg_constraints,omitempty" gorm:"ForeignKey:Oid;References:ObjId"`
}

func (m *PgDepend) TableName() string {
	return "pg_catalog.pg_depend"
}
