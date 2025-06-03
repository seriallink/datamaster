package dialect

import (
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

type PgConstraint struct {
	Oid           pgtype.OID       `json:"oid"                     gorm:"column:oid;type:oid;primary_key"`
	ConName       pgtype.Name      `json:"conname"                 gorm:"column:conname;type:name"`
	ConNamespace  pgtype.OID       `json:"connamespace"            gorm:"column:connamespace;type:oid"`
	ConType       pgtype.BPChar    `json:"contype"                 gorm:"column:contype;type:char"`
	ConDeferrable pgtype.Bool      `json:"condeferrable"           gorm:"column:condeferrable;type:bool"`
	ConDeferred   pgtype.Bool      `json:"condeferred"             gorm:"column:condeferred;type:bool"`
	ConValidated  pgtype.Bool      `json:"convalidated"            gorm:"column:convalidated;type:bool"`
	ConRelId      pgtype.OID       `json:"conrelid"                gorm:"column:conrelid;type:oid"`
	ConTypId      pgtype.OID       `json:"contypid"                gorm:"column:contypid;type:oid"`
	ConIndId      pgtype.OID       `json:"conindid"                gorm:"column:conindid;type:oid"`
	ConParentId   pgtype.OID       `json:"conparentid"             gorm:"column:conparentid;type:oid"`
	ConFrelId     pgtype.OID       `json:"confrelid"               gorm:"column:confrelid;type:oid"`
	ConFUpdType   pgtype.BPChar    `json:"confupdtype"             gorm:"column:confupdtype;type:char"`
	ConFDelType   pgtype.BPChar    `json:"confdeltype"             gorm:"column:confdeltype;type:char"`
	ConFMatchType pgtype.BPChar    `json:"confmatchtype"           gorm:"column:confmatchtype;type:char"`
	ConIsLocal    pgtype.Bool      `json:"conislocal"              gorm:"column:conislocal;type:bool"`
	ConInhCount   pgtype.Int4      `json:"coninhcount"             gorm:"column:coninhcount;type:int4"`
	ConNoInherit  pgtype.Bool      `json:"connoinherit"            gorm:"column:connoinherit;type:bool"`
	ConKey        pgtype.Int2Array `json:"conkey"                  gorm:"column:conkey;type:_int2"`
	ConFKey       pgtype.Int2Array `json:"confkey"                 gorm:"column:confkey;type:_int2"`
	ConPfeqop     pgtype.Int8Array `json:"conpfeqop"               gorm:"column:conpfeqop;type:_oid"`
	ConPpeqop     pgtype.Int8Array `json:"conppeqop"               gorm:"column:conppeqop;type:_oid"`
	ConFfeqop     pgtype.Int8Array `json:"conffeqop"               gorm:"column:conffeqop;type:_oid"`
	ConExclop     pgtype.Int8Array `json:"conexclop"               gorm:"column:conexclop;type:_oid"`
	ConBin        pgtype.Text      `json:"conbin"                  gorm:"column:conbin;type:pg_node_tree"`
	PgClass       *PgClass         `json:"pg_class,omitempty"      gorm:"ForeignKey:ConRelId;References:Oid"`
	PgAttributes  []PgAttribute    `json:"pg_attributes,omitempty" gorm:"-"`
	PgDepends     []PgDepend       `json:"pg_depends,omitempty"    gorm:"ForeignKey:RefObjId;References:ConIndId"`
}

func (m *PgConstraint) TableName() string {
	return "pg_catalog.pg_constraint"
}

func (m *PgConstraint) AfterFind(db *gorm.DB) error {

	// set model for searching
	criteria := &PgAttribute{
		AttRelId: m.ConRelId,
	}

	return db.Where(criteria).Where("attnum IN (?)", m.ConKey.Elements).Find(&m.PgAttributes).Error

}
