package database

import (
	"fmt"

	"github.com/yamlforge/yamlforge/internal/parser"
)

func NewDatabase(config *parser.DatabaseConfig) (Database, error) {
	switch parser.DatabaseType(config.Type) {
	case parser.DatabaseSQLite:
		return NewSQLite(config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

