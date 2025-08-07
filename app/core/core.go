// Package core implements all core functionalities shared across the Data Master project,
// serving as the backbone for both the CLI and data processing components.
//
// It provides reusable integrations with AWS services (S3, DynamoDB, Glue, DMS, Lake Formation),
// handles processing control logic, data catalog synchronization, infrastructure operations,
// benchmarking support, and orchestration utilities.
//
// All orchestration logic, resource abstraction, and operational primitives used by the
// CLI commands and ECS/Lambda workers are defined in this package, enabling consistency,
// automation, and decoupling across the Medallion data pipeline.
package core
