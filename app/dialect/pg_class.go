package dialect

import "github.com/jackc/pgtype"

// PgClass represents a relation (table, view, index, etc.) in the PostgreSQL system catalog (pg_class).
//
// Each instance contains metadata about a database object, including its name,
// storage parameters, owner, persistence settings, and structural flags.
// This model is used to extract schema definitions during catalog introspection
// and to generate Glue table metadata in the Data Master pipeline.
//
// It includes foreign key associations to attributes and constraints,
// enabling traversal of a table's full logical structure.
type PgClass struct {
	Oid                 pgtype.OID          `json:"oid"                      gorm:"column:oid;type:oid;primary_key"`
	RelName             pgtype.Name         `json:"relname"                  gorm:"column:relname;type:name"`
	RelNamespace        pgtype.OID          `json:"relnamespace"             gorm:"column:relnamespace;type:oid"`
	RelType             pgtype.OID          `json:"reltype"                  gorm:"column:reltype;type:oid"`
	RelOfType           pgtype.OID          `json:"reloftype"                gorm:"column:reloftype;type:oid"`
	RelOwner            pgtype.OID          `json:"relowner"                 gorm:"column:relowner;type:oid"`
	RelAm               pgtype.OID          `json:"relam"                    gorm:"column:relam;type:oid"`
	RelFileNode         pgtype.OID          `json:"relfilenode"              gorm:"column:relfilenode;type:oid"`
	RelTablespace       pgtype.OID          `json:"reltablespace"            gorm:"column:reltablespace;type:oid"`
	RelPages            pgtype.Int4         `json:"relpages"                 gorm:"column:relpages;type:int4"`
	RelTuples           pgtype.Float4       `json:"reltuples"                gorm:"column:reltuples;type:float4"`
	RelAllVisible       pgtype.Int4         `json:"relallvisible"            gorm:"column:relallvisible;type:int4"`
	RelToastRelId       pgtype.OID          `json:"reltoastrelid"            gorm:"column:reltoastrelid;type:oid"`
	RelHasIndex         pgtype.Bool         `json:"relhasindex"              gorm:"column:relhasindex;type:bool"`
	RelIsShared         pgtype.Bool         `json:"relisshared"              gorm:"column:relisshared;type:bool"`
	RelPersistence      pgtype.BPChar       `json:"relpersistence"           gorm:"column:relpersistence;type:char"`
	RelKind             pgtype.BPChar       `json:"relkind"                  gorm:"column:relkind;type:char"`
	RelNatts            pgtype.Int2         `json:"relnatts"                 gorm:"column:relnatts;type:int2"`
	RelChecks           pgtype.Int2         `json:"relchecks"                gorm:"column:relchecks;type:int2"`
	RelHasRules         pgtype.Bool         `json:"relhasrules"              gorm:"column:relhasrules;type:bool"`
	RelHasTriggers      pgtype.Bool         `json:"relhastriggers"           gorm:"column:relhastriggers;type:bool"`
	RelHasSubclass      pgtype.Bool         `json:"relhassubclass"           gorm:"column:relhassubclass;type:bool"`
	RelRowSecurity      pgtype.Bool         `json:"relrowsecurity"           gorm:"column:relrowsecurity;type:bool"`
	RelForceRowSecurity pgtype.Bool         `json:"relforcerowsecurity"      gorm:"column:relforcerowsecurity;type:bool"`
	RelIsPopulated      pgtype.Bool         `json:"relispopulated"           gorm:"column:relispopulated;type:bool"`
	RelReplIdent        pgtype.BPChar       `json:"relreplident"             gorm:"column:relreplident;type:char"`
	RelIsPartition      pgtype.Bool         `json:"relispartition"           gorm:"column:relispartition;type:bool"`
	RelRewrite          pgtype.OID          `json:"relrewrite"               gorm:"column:relrewrite;type:oid"`
	RelFrozenXid        pgtype.XID          `json:"relfrozenxid"             gorm:"column:relfrozenxid;type:xid"`
	RelMinmXid          pgtype.XID          `json:"relminmxid"               gorm:"column:relminmxid;type:xid"`
	RelAcl              pgtype.ACLItemArray `json:"relacl"                   gorm:"column:relacl;type:_aclitem"`
	RelOptions          pgtype.TextArray    `json:"reloptions"               gorm:"column:reloptions;type:_text"`
	RelPartBound        pgtype.Text         `json:"relpartbound"             gorm:"column:relpartbound;type:pg_node_tree"`
	PgAttributes        []PgAttribute       `json:"pg_attributes,omitempty"  gorm:"ForeignKey:AttRelId;References:Oid"`
	PgConstraints       []PgConstraint      `json:"pg_constraints,omitempty" gorm:"ForeignKey:ConRelId;References:Oid"`
	//PgIndexes           []PgIndex      `json:"pg_indexes,omitempty"     gorm:"ForeignKey:IndexRelId;References:Oid"`
	//PgDepends           []PgDepend     `json:"pg_depends,omitempty"     gorm:"ForeignKey:RefObjId;References:Oid"`
}

func (m *PgClass) TableName() string {
	return "pg_catalog.pg_class"
}
