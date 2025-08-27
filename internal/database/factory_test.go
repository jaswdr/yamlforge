package database

import (
	"testing"

	"github.com/yamlforge/yamlforge/internal/parser"
)

func TestNewDatabase_SQLite(t *testing.T) {
	config := &parser.DatabaseConfig{
		Type: "sqlite",
		Path: "./test.db",
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	_, ok := db.(*SQLiteDB)
	if !ok {
		t.Errorf("Expected SQLiteDB instance, got: %T", db)
	}
}

func TestNewDatabase_UnsupportedType(t *testing.T) {
	config := &parser.DatabaseConfig{
		Type: "redis",
		Path: "./test.db",
	}

	_, err := NewDatabase(config)
	if err == nil {
		t.Fatal("Expected error for unsupported database type")
	}

	expected := "unsupported database type: redis"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got: %s", expected, err.Error())
	}
}

func TestNewDatabase_PostgreSQL(t *testing.T) {
	config := &parser.DatabaseConfig{
		Type:       "postgresql",
		Connection: "postgres://user:pass@localhost/dbname",
	}

	_, err := NewDatabase(config)
	if err == nil {
		t.Fatal("Expected error for unsupported PostgreSQL")
	}
}

func TestNewDatabase_MySQL(t *testing.T) {
	config := &parser.DatabaseConfig{
		Type:       "mysql",
		Connection: "user:pass@tcp(localhost:3306)/dbname",
	}

	_, err := NewDatabase(config)
	if err == nil {
		t.Fatal("Expected error for unsupported MySQL")
	}
}