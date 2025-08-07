package dialect

import "github.com/jackc/pgtype"

// PgAttribute represents a column definition in the PostgreSQL system catalog (pg_attribute).
//
// Each instance corresponds to a single attribute (column) in a table or view,
// and includes metadata such as name, data type, nullability, default value, storage settings,
// and additional behavioral flags.
//
// This struct is used during catalog introspection to generate table schemas
// for the Glue Catalog and support type resolution during data ingestion.
//
// Most fields are mapped directly from pg_attribute, with types compatible
// with the pgtype package and relationships resolved via GORM.
type PgAttribute struct {
	AttRelId      pgtype.OID          `json:"attrelid"          gorm:"column:attrelid;type:oid"`
	AttName       pgtype.Name         `json:"attname"           gorm:"column:attname;type:name"`
	AttTypId      pgtype.OID          `json:"atttypid"          gorm:"column:atttypid;type:oid"`
	AttStatTarget pgtype.Int4         `json:"attstattarget"     gorm:"column:attstattarget;type:int4"`
	AttLen        pgtype.Int2         `json:"attlen"            gorm:"column:attlen;type:int2"`
	AttNum        pgtype.Int2         `json:"attnum"            gorm:"column:attnum;type:int2"`
	AttNdims      pgtype.Int4         `json:"attndims"          gorm:"column:attndims;type:int4"`
	AttCacheOff   pgtype.Int4         `json:"attcacheoff"       gorm:"column:attcacheoff;type:int4"`
	AttTypMod     pgtype.Int4         `json:"atttypmod"         gorm:"column:atttypmod;type:int4"`
	AttByVal      pgtype.Bool         `json:"attbyval"          gorm:"column:attbyval;type:bool"`
	AttStorage    pgtype.BPChar       `json:"attstorage"        gorm:"column:attstorage;type:char"`
	AttAlign      pgtype.BPChar       `json:"attalign"          gorm:"column:attalign;type:char"`
	AttNotNull    pgtype.Bool         `json:"attnotnull"        gorm:"column:attnotnull;type:bool"`
	AttHasDef     pgtype.Bool         `json:"atthasdef"         gorm:"column:atthasdef;type:bool"`
	AttHasMissing pgtype.Bool         `json:"atthasmissing"     gorm:"column:atthasmissing;type:bool"`
	AttIdentity   pgtype.BPChar       `json:"attidentity"       gorm:"column:attidentity;type:char"`
	AttGenerated  pgtype.BPChar       `json:"attgenerated"      gorm:"column:attgenerated;type:char"`
	AttIsDropped  pgtype.Bool         `json:"attisdropped"      gorm:"column:attisdropped;type:bool"`
	AttIsLocal    pgtype.Bool         `json:"attislocal"        gorm:"column:attislocal;type:bool"`
	AttInhCount   pgtype.Int4         `json:"attinhcount"       gorm:"column:attinhcount;type:int4"`
	AttCollation  pgtype.OID          `json:"attcollation"      gorm:"column:attcollation;type:oid"`
	AttAcl        pgtype.ACLItemArray `json:"attacl"            gorm:"column:attacl;type:_aclitem"`
	AttOptions    pgtype.TextArray    `json:"attoptions"        gorm:"column:attoptions;type:_text"`
	AttFdwOptions pgtype.TextArray    `json:"attfdwoptions"     gorm:"column:attfdwoptions;type:_text"`
	AttMissingVal pgtype.TextArray    `json:"attmissingval"     gorm:"column:attmissingval;type:anyarray"`
	PgType        *PgType             `json:"pg_type,omitempty" gorm:"ForeignKey:AttTypId;References:Oid"`
}

func (m *PgAttribute) TableName() string {
	return "pg_catalog.pg_attribute"
}
