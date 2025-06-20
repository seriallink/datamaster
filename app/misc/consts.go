package misc

// Defaults used throughout the CLI.
const (
	DefaultProjectPrefix = "dm"
	TemplatesPath        = "stacks"
	ArtifactsPath        = "artifacts"
	TemplateExtension    = ".yml"
)

// Logical names of all stacks managed by the CLI.
const (
	StackNameNetwork       = "network"
	StackNameRoles         = "roles"
	StackNameSecurity      = "security"
	StackNameDatabase      = "database"
	StackNameStorage       = "storage"
	StackNameCatalog       = "catalog"
	StackNameGovernance    = "governance"
	StackNameConsumption   = "consumption"
	StackNameControl       = "control"
	StackNameFunctions     = "functions"
	StackNameStreaming     = "streaming"
	StackNameIngestion     = "ingestion"
	StackNameProcessing    = "processing"
	StackNameAnalytics     = "analytics"
	StackNameObservability = "observability"
	StackNameCosts         = "costs"
)

const (
	MigrationCoreScript = "database/migrations/001_create_dm_core.sql"
	MigrationViewScript = "database/migrations/002_create_dm_view.sql"
	MigrationMartScript = "database/migrations/003_create_dm_mart.sql"
)

const (
	RelKindTable     = "r"
	RelKindPartition = "p"
	RelKindView      = "v"
)

// Logical names of all schemas managed by the CLI.
const (
	SchemaCore = "core"
	SchemaView = "view"
	SchemaMart = "mart"
)

// Logical names of all layers managed by the CLI.
const (
	LayerBronze = "bronze"
	LayerSilver = "silver"
	LayerGold   = "gold"
)
