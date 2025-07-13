package mock

import (
	"github.com/alifiroozi80/duckdb"
	"gorm.io/gorm"
)

// OpenDuckDB opens a new in-memory DuckDB connection and wraps it with GORM.
//
// Returns:
//   - *gorm.DB: the GORM-wrapped DuckDB connection.
//   - error: any error encountered while opening the connection.
func OpenDuckDB() (*gorm.DB, error) {
	return gorm.Open(duckdb.Open(""), &gorm.Config{})
}
