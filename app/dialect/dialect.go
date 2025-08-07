// Package dialect defines internal representations of PostgreSQL system catalogs
// used to extract schema metadata for automatic generation of Glue tables.
//
// It maps low-level PostgreSQL entities such as tables, columns, types, constraints,
// and inheritance structures using strongly typed Go structs based on pg_catalog.
//
// These models are used by the Data Master CLI to infer schema definitions,
// resolve data types, and build a complete metastore representation of the source database.
//
// All entities are structured to support querying via GORM and cross-referencing
// between related catalog tables (e.g., pg_class → pg_attribute → pg_type).
package dialect
