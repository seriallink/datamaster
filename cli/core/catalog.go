package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/cli/dialect"
	"github.com/seriallink/datamaster/cli/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

// NameWithPrefix applies the default project prefix (e.g., "dm") to a schema, database, or layer name.
func NameWithPrefix(name string) string {
	return misc.NameWithDefaultPrefix(strings.ToLower(name), '_')
}

// LayerToSchema maps a lakehouse layer name (bronze, silver, or gold)
// to its corresponding PostgreSQL schema name.
//
// Parameters:
//   - layerType: the logical lakehouse layer ("bronze", "silver", or "gold").
//
// Returns:
//   - string: the associated PostgreSQL schema name.
//
// Panics:
//   - if the provided layerType is not recognized.
func LayerToSchema(layerType string) string {
	switch layerType {
	case misc.LayerBronze:
		return misc.SchemaCore
	case misc.LayerSilver:
		return misc.SchemaView
	case misc.LayerGold:
		return misc.SchemaMart
	default:
		panic(fmt.Sprintf("Invalid layer: %s", layerType))
	}
}

// GetStorageLocation returns the S3 path for storing a table's data,
// based on its schema and corresponding lakehouse layer.
//
// The location follows the structure:
//
//	s3://<DataLakeBucketName>/<layerType>/<tableName>/
//
// Parameters:
//   - schemaName: PostgreSQL schema name (e.g., "core", "view", or "mart").
//   - tableName: name of the table.
//
// Returns:
//   - string: fully qualified S3 path.
//   - error: if fetching stack outputs or mapping schema fails.
func GetStorageLocation(layerType, tableName string) (string, error) {

	stack := &Stack{Name: misc.StackNameStorage}

	bucketName, err := stack.GetStackOutput(GetAWSConfig(), "DataLakeBucketName")
	if err != nil {
		return "", fmt.Errorf("DataLakeBucketName not found in stack outputs")
	}

	return fmt.Sprintf("s3://%s/%s/%s/", bucketName, layerType, tableName), nil

}

// LoadAuroraTablesWithColumns queries PostgreSQL system catalogs to load user-defined tables,
// their columns, and data types from a given schema.
//
// Optionally filters by table name. Returns a slice of PgClass structs populated with PgAttributes.
//
// Parameters:
//   - schema: name of the schema to inspect.
//   - tables: optional list of specific table names.
//
// Returns:
//   - []dialect.PgClass: table metadata with column details.
//   - error: if the query fails.
func LoadAuroraTablesWithColumns(schema string, tables ...string) (classes []dialect.PgClass, err error) {

	var db *gorm.DB

	if db, err = GetConnection(); err != nil {
		return nil, err
	}

	// filter schema from pg_namespace
	namespace := &dialect.PgNamespace{
		NspName: pgtype.Name{
			String: schema,
			Status: pgtype.Present,
		},
	}

	if err = db.Where(namespace).Take(namespace).Error; err != nil {
		return
	}

	// filter all tables from pg_class by namespace (schema)
	criteria := &dialect.PgClass{
		RelNamespace: namespace.Oid,
		RelKind: pgtype.BPChar{
			String: misc.RelKindTable,
			Status: pgtype.Present,
		},
	}

	// mount lazy loading to get tables and columns
	db = db.
		// get column data type
		Preload("PgAttributes.PgType").
		// get columns (ignore system's and dropped columns with negative numbers)
		Preload("PgAttributes", "attnum > 0 AND attname not like '........pg.dropped.%'").
		// filter by schema and table name (optional)
		Where(&criteria).
		// only tables, partitions, and views
		Where("relkind IN (?)", []string{misc.RelKindTable, misc.RelKindPartition, misc.RelKindView})

	// filter by table name if provided
	if len(tables) > 0 {
		db = db.Where("relname IN (?)", tables)
	}

	// execute query
	if err = db.Order("relname").Find(&classes).Error; err != nil {
		return
	}

	return

}

// ConvertPgAttributesToGlueColumns converts a loaded PgClass (PostgreSQL table definition)
// into a slice of AWS Glue Column definitions compatible with Glue Catalog and Iceberg.
//
// Parameters:
//   - class: pointer to PgClass containing attributes.
//
// Returns:
//   - []types.Column: Glue column definitions.
func ConvertPgAttributesToGlueColumns(class *dialect.PgClass, db *gorm.DB) (columns []types.Column) {
	for _, attr := range class.PgAttributes {
		columnName := attr.AttName.String
		glueType := CastPgType(attr, *attr.PgType, db)
		columns = append(columns, types.Column{
			Name: aws.String(columnName),
			Type: aws.String(glueType),
		})
	}
	return
}

// CastPgType converts a PostgreSQL type (from pg_type and pg_attribute)
// into a logical representation used for Glue, Iceberg, or other downstream systems.
//
// It handles scalar types, arrays, enums, user-defined types, and falls back based on category.
//
// Parameters:
//   - pgattribute: the pg_attribute entry for the column.
//   - pgtype: the pg_type definition for the column's data type.
//   - db: GORM DB instance used for recursive type resolution (e.g., array elements).
//
// Returns:
//   - string: logical type (e.g., "string", "int", "timestamp", "map", "array").
//
// Panics:
//   - if a type cannot be mapped or an array element's type fails to load.
func CastPgType(pgattribute dialect.PgAttribute, pgtype dialect.PgType, db *gorm.DB) (def string) {

	// init def
	def = "%s"

	// check if type is defined as an array
	if pgattribute.PgType.TypCategory.String == "A" {
		def = "array"
	}

	switch pgtype.TypName.String {

	case "bytea":
		def = "binary"

	case "bool":
		def = fmt.Sprintf(def, "boolean")

	case "bpchar", "char", "name", "text", "varchar", "inet":
		def = fmt.Sprintf(def, "string")

	case "int2", "int4", "int8":
		def = fmt.Sprintf(def, "int")

	case "float4", "float8", "numeric":
		def = fmt.Sprintf(def, "decimal")

	case "date", "time", "timestamp", "timestamptz":
		def = fmt.Sprintf(def, "timestamp")

	case "interval":
		def = fmt.Sprintf(def, "string")

	case "uuid":
		def = fmt.Sprintf(def, "string")

	case "json", "jsonb":
		def = fmt.Sprintf(def, "map")

	default:

		switch pgtype.TypCategory.String {

		case "A": // array type

			// set criteria
			pgelement := dialect.PgType{
				Oid: pgtype.TypElem,
			}

			// get element type
			err := db.Take(&pgelement).Error
			if err != nil {
				panic(err)
			}

			// recursively cast type to array
			return CastPgType(pgattribute, pgelement, db)

		case "C": // composite type
			def = fmt.Sprintf(def, "string")

		case "D": // date/time types
			def = fmt.Sprintf(def, "timestamp")

		case "E": // enum type
			def = fmt.Sprintf(def, "string")

		case "U": // user-defined type (should be mapped to an object)
			def = fmt.Sprintf(def, "map")

		case "S": // string type
			def = fmt.Sprintf(def, "string")

		default: // not mapped
			panic(fmt.Sprintf("type could not be casted %s", pgtype.TypName.String))

		}

	}

	return

}

// SyncCatalogFromDatabaseSchema inspects a lakehouse layer (bronze, silver, or gold),
// loads metadata from the corresponding Aurora PostgreSQL schema,
// and ensures each table is registered as an Iceberg table in the AWS Glue Catalog.
//
// Parameters:
//   - layerType: logical layer name ("bronze", "silver", or "gold").
//   - tableList: optional list of table names to filter.
//
// Returns:
//   - error: if any step of the synchronization process fails.
func SyncCatalogFromDatabaseSchema(layerType string, tableList ...string) error {

	// open a database connection here to avoid multiple and unnecessary connections
	db, err := GetConnection()
	if err != nil {
		return err
	}

	schemaName := NameWithPrefix(LayerToSchema(layerType))

	tables, err := LoadAuroraTablesWithColumns(schemaName, tableList...)
	if err != nil {
		return err
	}

	for _, table := range tables {
		if err = SyncGlueTable(layerType, table.RelName.String, ConvertPgAttributesToGlueColumns(&table, db)); err != nil {
			return fmt.Errorf("failed to create or update table %s: %w", table.RelName.String, err)
		}
	}

	return nil

}

// SyncGlueTable creates or updates a Glue table in Iceberg format using the provided column definitions.
//
// Parameters:
//   - schemaName: logical layer name (bronze, silver, gold).
//   - tableName: name of the table.
//   - columns: list of Glue-compatible column definitions.
//
// Returns:
//   - error: if creation or update fails.
func SyncGlueTable(layerType, tableName string, columns []types.Column) (err error) {

	client := glue.NewFromConfig(GetAWSConfig())
	dbName := NameWithPrefix(layerType)
	location, _ := GetStorageLocation(layerType, tableName)

	_, err = client.GetTable(context.TODO(), &glue.GetTableInput{
		DatabaseName: aws.String(dbName),
		Name:         aws.String(tableName),
	})

	tableInput := &types.TableInput{
		Name: aws.String(tableName),
		StorageDescriptor: &types.StorageDescriptor{
			Columns: columns,
			// physical path in S3 that contains Iceberg metadata and data files
			Location: aws.String(location),
			// required to make the table readable via Athena/Glue
			// point to Iceberg-compatible Hive input/output formats
			InputFormat:  aws.String("org.apache.iceberg.mr.hive.HiveIcebergInputFormat"),
			OutputFormat: aws.String("org.apache.iceberg.mr.hive.HiveIcebergOutputFormat"),
			// technically unused by Iceberg, but required by Glue API,
			// a generic SerDe is passed as placeholder
			SerdeInfo: &types.SerDeInfo{
				SerializationLibrary: aws.String("org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"),
				Parameters:           map[string]string{"serialization.format": "1"},
			},
		},
		// declares the table as a native Iceberg dataset
		TableType: aws.String("ICEBERG"),
		Parameters: map[string]string{
			// used by AWS Glue for metadata discovery and consistency
			"classification": "iceberg",
			// must match TableType for compatibility with Athena and Glue jobs
			"table_type": "ICEBERG",
			// enables compaction-awareness for future rewrites
			"optimize_small_files": "true",
			// indicates the table data lives outside the Glue-managed warehouse (in S3)
			"EXTERNAL": "TRUE",
		},
	}

	if err != nil && !strings.Contains(err.Error(), "EntityNotFoundException") {
		return err
	}

	// table exists, update it
	if err == nil {
		_, err = client.UpdateTable(context.TODO(), &glue.UpdateTableInput{
			DatabaseName: aws.String(dbName),
			TableInput:   tableInput,
		})
		if err != nil {
			return fmt.Errorf("failed to update table %s: %w", tableName, err)
		}

		fmt.Println(misc.Green("Table %s updated successfully in database %s", tableName, dbName))
		return nil
	}

	// table does not exist, create it
	input := &glue.CreateTableInput{
		DatabaseName: aws.String(dbName),
		TableInput:   tableInput,
	}

	if _, err = client.CreateTable(context.TODO(), input); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	fmt.Println(misc.Green("Table %s created successfully in database %s", tableName, dbName))
	return nil

}
