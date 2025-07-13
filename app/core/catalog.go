package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/dialect"
	"github.com/seriallink/datamaster/app/misc"

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

// SchemaToLayer maps a PostgreSQL schema name to its corresponding
// logical lakehouse layer name ("bronze", "silver", or "gold").
//
// Parameters:
//   - schema: the physical PostgreSQL schema name (e.g., "dm_core", "dm_view", "dm_mart").
//
// Returns:
//   - string: the associated logical lakehouse layer name.
//
// Panics:
//   - if the provided schema name is not recognized.
func SchemaToLayer(schema string) string {
	switch schema {
	case misc.SchemaCore:
		return misc.LayerBronze
	case misc.SchemaView:
		return misc.LayerSilver
	case misc.SchemaMart:
		return misc.LayerGold
	default:
		panic(fmt.Sprintf("Invalid schema: %s", schema))
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
//   - db: GORM database connection to the Aurora PostgreSQL instance.
//   - schema: name of the schema to inspect.
//   - tables: optional list of specific table names.
//
// Returns:
//   - []dialect.PgClass: table metadata with column details.
//   - error: if the query fails.
func LoadAuroraTablesWithColumns(db *gorm.DB, schema string, tables ...string) (classes []dialect.PgClass, err error) {

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
// into a slice of AWS Glue Column definitions compatible with Glue Catalog.
//
// Behavior:
//   - For the "bronze" layer, all columns are forced to type "string",
//     and an additional column named "operation" is injected at the end.
//   - For "silver" and "gold", types are inferred from PostgreSQL definitions.
//
// Parameters:
//   - layerType: lakehouse layer ("bronze", "silver", or "gold").
//   - class: pointer to PgClass containing attributes.
//   - db: GORM database connection for resolving PostgreSQL types.
//
// Returns:
//   - []types.Column: list of Glue-compatible column definitions.
func ConvertPgAttributesToGlueColumns(layerType string, class *dialect.PgClass, db *gorm.DB) (columns []types.Column) {

	for _, attr := range class.PgAttributes {

		// Force all fields to string type for the bronze layer
		columns = append(columns, types.Column{
			Name: aws.String(attr.AttName.String),
			//Type: aws.String(misc.TernaryStr(layerType == misc.LayerBronze, "string", CastPgType(attr, *attr.PgType, db))),
			Type: aws.String(CastPgType(attr, *attr.PgType, db)),
		})

	}

	// Inject 'operation' column explicitly for bronze
	if layerType == misc.LayerBronze {
		columns = append(columns, types.Column{
			Name: aws.String("operation"),
			Type: aws.String("string"),
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
		def = fmt.Sprintf(def, "double")

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
func SyncCatalogFromDatabaseSchema(db *gorm.DB, layerType string, tableList ...string) error {

	schemaName := NameWithPrefix(LayerToSchema(layerType))

	tables, err := LoadAuroraTablesWithColumns(db, schemaName, tableList...)
	if err != nil {
		return err
	}

	for _, table := range tables {
		if err = SyncGlueTable(layerType, table.RelName.String, ConvertPgAttributesToGlueColumns(layerType, &table, db)); err != nil {
			return fmt.Errorf("failed to create or update table %s: %w", table.RelName.String, err)
		}
	}

	return nil

}

// SyncGlueTable creates or updates a Glue table in the AWS Glue Catalog,
// using the appropriate format based on the lakehouse layer.
//
// For the "bronze" layer, it creates a standard Hive table using Parquet,
// compatible with Athena and Glue without Iceberg dependencies.
//
// For the "silver" and "gold" layers, it creates a native Iceberg table,
// enabling support for ACID operations, versioning, and time travel.
//
// Parameters:
//   - layerType: the lakehouse layer ("bronze", "silver", or "gold").
//   - tableName: the name of the table to create or update.
//   - columns: a slice of Glue-compatible column definitions.
//
// Returns:
//   - error: if creation or update fails.
func SyncGlueTable(layerType, tableName string, columns []types.Column) error {

	client := glue.NewFromConfig(GetAWSConfig())
	dbName := NameWithPrefix(layerType)
	location, _ := GetStorageLocation(layerType, tableName)

	_, err := client.GetTable(context.TODO(), &glue.GetTableInput{
		DatabaseName: aws.String(dbName),
		Name:         aws.String(tableName),
	})
	tableExists := err == nil

	if err != nil && !strings.Contains(err.Error(), "EntityNotFoundException") {
		return err
	}

	// Set format-specific configs
	var (
		inputFormat  string
		outputFormat string
		serdeLib     string
		tableType    string
		parameters   map[string]string
	)

	if layerType == misc.LayerBronze {

		// For bronze layer, create a standard Hive table with Parquet format
		tableType = "EXTERNAL_TABLE"
		inputFormat = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
		outputFormat = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"
		serdeLib = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"
		parameters = map[string]string{
			"classification": "parquet",
			"EXTERNAL":       "TRUE",
		}

	} else {

		// For silver and gold layers, create an Iceberg table
		tableType = "ICEBERG"
		inputFormat = "org.apache.iceberg.mr.hive.HiveIcebergInputFormat"
		outputFormat = "org.apache.iceberg.mr.hive.HiveIcebergOutputFormat"
		serdeLib = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"
		parameters = map[string]string{
			"classification":       "iceberg",
			"table_type":           "ICEBERG",
			"optimize_small_files": "true",
			"EXTERNAL":             "TRUE",
		}

		//var columnDefs []string
		//for _, col := range columns {
		//	columnDefs = append(columnDefs, fmt.Sprintf("%s %s", *col.Name, *col.Type))
		//}
		//if err = SetIcebergMetadata(context.TODO(), layerType, tableName, columnDefs); err != nil {
		//	return err
		//}

	}

	tableInput := &types.TableInput{
		Name:       aws.String(tableName),
		TableType:  aws.String(tableType),
		Parameters: parameters,
		StorageDescriptor: &types.StorageDescriptor{
			Columns:      columns,
			Location:     aws.String(location),
			InputFormat:  aws.String(inputFormat),
			OutputFormat: aws.String(outputFormat),
			SerdeInfo: &types.SerDeInfo{
				SerializationLibrary: aws.String(serdeLib),
				Parameters:           map[string]string{"serialization.format": "1"},
			},
		},
	}

	if tableExists {
		_, err = client.UpdateTable(context.TODO(), &glue.UpdateTableInput{
			DatabaseName: aws.String(dbName),
			TableInput:   tableInput,
		})
		if err != nil {
			return fmt.Errorf("failed to update table %s: %w", tableName, err)
		}
		fmt.Println(misc.Green("TableName %s updated successfully in database %s", tableName, dbName))
	} else {
		_, err = client.CreateTable(context.TODO(), &glue.CreateTableInput{
			DatabaseName: aws.String(dbName),
			TableInput:   tableInput,
		})
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
		fmt.Println(misc.Green("TableName %s created successfully in database %s", tableName, dbName))
	}

	return nil

}
