package mock

import (
	"gorm.io/gorm"
)

// SeedFakePgCatalogWithSchema inserts simulated pg_catalog metadata for testing purposes,
// using the provided schema name (e.g., "dm_core").
func SeedFakePgCatalogWithSchema(db *gorm.DB, schema string) error {

	stmts := []string{
		`CREATE TABLE "pg_catalog.pg_namespace" (
			oid INTEGER,
			nspname TEXT
		);`,

		`CREATE TABLE "pg_catalog.pg_class" (
			oid INTEGER,
			relname TEXT,
			relnamespace INTEGER,
			relkind TEXT
		);`,

		`CREATE TABLE "pg_catalog.pg_attribute" (
			attrelid INTEGER,
			attname TEXT,
			attnum INTEGER,
			atttypid INTEGER
		);`,

		`CREATE TABLE "pg_catalog.pg_type" (
			oid INTEGER,
			typname TEXT
		);`,

		`INSERT INTO "pg_catalog.pg_class" (oid, relname, relnamespace, relkind)
		 VALUES (200, 'brewery', 100, 'r');`,

		`INSERT INTO "pg_catalog.pg_attribute" (attrelid, attname, attnum, atttypid)
		 VALUES 
			(200, 'id', 1, 300),
			(200, 'name', 2, 301),
			(200, 'city', 3, 301),
			(200, 'state', 4, 301);`,

		`INSERT INTO "pg_catalog.pg_type" (oid, typname)
		 VALUES 
			(300, 'uuid'),
			(301, 'varchar');`,
	}

	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	return nil

}
