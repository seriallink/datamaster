package dialect

import "github.com/jackc/pgtype"

// PgType represents a data type definition in PostgreSQL, as defined in the pg_type system catalog.
//
// It includes metadata for scalar types, composite types, enum types, arrays, and domains.
// Each entry provides details about storage, alignment, input/output functions,
// category, and constraints associated with the type.
//
// This model is used to resolve column types during catalog introspection
// and to generate accurate type mappings for Glue table creation and schema evolution.
type PgType struct {
	Oid            pgtype.OID          `json:"oid"                     gorm:"column:oid;type:oid;primary_key"`
	TypName        pgtype.Name         `json:"typname"                 gorm:"column:typname;type:name"`
	TypNamespace   pgtype.OID          `json:"typnamespace"            gorm:"column:typnamespace;type:oid"`
	TypOwner       pgtype.OID          `json:"typowner"                gorm:"column:typowner;type:oid"`
	TypLen         pgtype.Int2         `json:"typlen"                  gorm:"column:typlen;type:int2"`
	TypByVal       pgtype.Bool         `json:"typbyval"                gorm:"column:typbyval;type:bool"`
	TypType        pgtype.BPChar       `json:"typtype"                 gorm:"column:typtype;type:char"`
	TypCategory    pgtype.BPChar       `json:"typcategory"             gorm:"column:typcategory;type:char"`
	TypIsPreferred pgtype.Bool         `json:"typispreferred"          gorm:"column:typispreferred;type:bool"`
	TypIsDefined   pgtype.Bool         `json:"typisdefined"            gorm:"column:typisdefined;type:bool"`
	TypDelim       pgtype.BPChar       `json:"typdelim"                gorm:"column:typdelim;type:char"`
	TypRelId       pgtype.OID          `json:"typrelid"                gorm:"column:typrelid;type:oid"`
	TypSubscript   pgtype.Text         `json:"typsubscript"            gorm:"column:typsubscript;type:regproc"`
	TypElem        pgtype.OID          `json:"typelem"                 gorm:"column:typelem;type:oid"`
	TypArray       pgtype.OID          `json:"typarray"                gorm:"column:typarray;type:oid"`
	TypInput       pgtype.Text         `json:"typinput"                gorm:"column:typinput;type:regproc"`
	TypOutput      pgtype.Text         `json:"typoutput"               gorm:"column:typoutput;type:regproc"`
	TypReceive     pgtype.Text         `json:"typreceive"              gorm:"column:typreceive;type:regproc"`
	TypSend        pgtype.Text         `json:"typsend"                 gorm:"column:typsend;type:regproc"`
	TypModIn       pgtype.Text         `json:"typmodin"                gorm:"column:typmodin;type:regproc"`
	TypModOut      pgtype.Text         `json:"typmodout"               gorm:"column:typmodout;type:regproc"`
	TypAnalyze     pgtype.Text         `json:"typanalyze"              gorm:"column:typanalyze;type:regproc"`
	TypAlign       pgtype.BPChar       `json:"typalign"                gorm:"column:typalign;type:char"`
	TypStorage     pgtype.BPChar       `json:"typstorage"              gorm:"column:typstorage;type:char"`
	TypNotNull     pgtype.Bool         `json:"typnotnull"              gorm:"column:typnotnull;type:bool"`
	TypBaseType    pgtype.OID          `json:"typbasetype"             gorm:"column:typbasetype;type:oid"`
	TypTypMod      pgtype.Int4         `json:"typtypmod"               gorm:"column:typtypmod;type:int4"`
	TypNdims       pgtype.Int4         `json:"typndims"                gorm:"column:typndims;type:int4"`
	TypCollation   pgtype.OID          `json:"typcollation"            gorm:"column:typcollation;type:oid"`
	TypDefaultBin  pgtype.Text         `json:"typdefaultbin"           gorm:"column:typdefaultbin;type:pg_node_tree"`
	TypDefault     pgtype.Text         `json:"typdefault"              gorm:"column:typdefault;type:text"`
	TypAcl         pgtype.ACLItemArray `json:"typacl"                  gorm:"column:typacl;type:_aclitem"`
	PgAttributes   []PgAttribute       `json:"pg_attributes,omitempty" gorm:"ForeignKey:AttRelId;References:TypRelId"`
	PgEnums        []PgEnum            `json:"pg_enums,omitempty"      gorm:"ForeignKey:EnumTypId;References:Oid"`
}

func (m *PgType) TableName() string {
	return "pg_catalog.pg_type"
}
