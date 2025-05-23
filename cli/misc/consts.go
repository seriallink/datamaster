package misc

// Defaults used throughout the CLI.
const (
	DefaultProjectPrefix = "dm"
	TemplatesPath        = "stacks"
	ArtifactsPath        = "artifacts"
	TemplateExtension    = ".yml"
	DefaultScript        = "database/migrations/data-master.sql"
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
	StackNameStreaming     = "streaming"
	StackNameBatch         = "batch"
	StackNameProcessing    = "processing"
	StackNameConsumption   = "consumption"
	StackNameObservability = "observability"
	StackNameCosts         = "costs"
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
