// Package cmd defines all CLI commands available in the Data Master project.
//
// It includes entrypoints for provisioning infrastructure, catalog synchronization,
// dataset seeding, benchmarking, processing orchestration, and dashboard generation.
//
// Each command is modular and self-contained, and the CLI is implemented using Cobra,
// allowing flexible composition, argument parsing, and extensibility.
//
// This package is the user-facing interface for executing and automating operations
// across all stages of the Medallion Architecture (raw, bronze, silver, gold).
package cmd
