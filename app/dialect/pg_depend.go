package dialect

import "github.com/jackc/pgtype"

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
