package database

import (
	"testing"

	"github.com/yamlforge/yamlforge/internal/parser"
)

func TestNew(t *testing.T) {
	config := &parser.DatabaseConfig{
		Type: "sqlite",
		Path: "./test.db",
	}

	db, err := New(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if db == nil {
		t.Fatal("Expected database instance, got nil")
	}
}

func TestDB_GetDriverName(t *testing.T) {
	db := &DB{dbType: parser.DatabaseSQLite}
	
	driver := db.getDriverName()
	if driver != "sqlite3" {
		t.Errorf("Expected 'sqlite3', got: %s", driver)
	}

	db.dbType = parser.DatabaseType("unknown")
	driver = db.getDriverName()
	if driver != "" {
		t.Errorf("Expected empty string for unknown type, got: %s", driver)
	}
}

func TestDB_GetConnectionString(t *testing.T) {
	config := &parser.DatabaseConfig{
		Type: "sqlite",
		Path: "./test.db",
	}

	db := &DB{
		config: config,
		dbType: parser.DatabaseSQLite,
	}
	
	connStr := db.getConnectionString()
	if connStr != "./test.db" {
		t.Errorf("Expected './test.db', got: %s", connStr)
	}

	db.dbType = parser.DatabaseType("unknown")
	connStr = db.getConnectionString()
	if connStr != "" {
		t.Errorf("Expected empty string for unknown type, got: %s", connStr)
	}
}

func TestDB_BuildSelectQuery(t *testing.T) {
	schema := &parser.Schema{
		Models: map[string]*parser.Model{
			"User": {
				Name: "User",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID},
					{Name: "name", Type: parser.FieldTypeText},
				},
				UI: parser.UIModel{
					List: parser.UIList{
						Searchable: []string{"name"},
					},
				},
			},
		},
	}

	db := &DB{
		schema: schema,
		dbType: parser.DatabaseSQLite,
	}

	params := parser.QueryParams{
		Page:     1,
		PageSize: 10,
	}

	query, args := db.buildSelectQuery("User", params)
	expected := "SELECT * FROM \"User\" LIMIT 10 OFFSET 0"
	if query != expected {
		t.Errorf("Expected '%s', got '%s'", expected, query)
	}
	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(args))
	}

	params.Filters = []parser.Filter{
		{Field: "name", Operator: "=", Value: "John"},
	}

	query, args = db.buildSelectQuery("User", params)
	if !containsString(query, "WHERE") {
		t.Error("Expected query to contain WHERE clause")
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}

	params.Search = "test"
	query, args = db.buildSelectQuery("User", params)
	if !containsString(query, "LIKE") {
		t.Error("Expected query to contain LIKE clause for search")
	}

	params.Sort = []parser.SortField{
		{Field: "name", Desc: true},
	}
	query, args = db.buildSelectQuery("User", params)
	if !containsString(query, "ORDER BY") {
		t.Error("Expected query to contain ORDER BY clause")
	}
	if !containsString(query, "DESC") {
		t.Error("Expected query to contain DESC")
	}
}

func TestDB_BuildWhereClause(t *testing.T) {
	db := &DB{dbType: parser.DatabaseSQLite}

	filter := parser.Filter{Field: "name", Value: "John"}
	clause, arg := db.buildWhereClause(filter)
	if clause != "\"name\" = ?" {
		t.Errorf("Expected '\"name\" = ?', got '%s'", clause)
	}
	if arg != "John" {
		t.Errorf("Expected 'John', got %v", arg)
	}

	filter.Operator = "like"
	clause, arg = db.buildWhereClause(filter)
	if clause != "\"name\" LIKE ?" {
		t.Errorf("Expected '\"name\" LIKE ?', got '%s'", clause)
	}
	if arg != "%John%" {
		t.Errorf("Expected '%%John%%', got %v", arg)
	}

	filter.Operator = "in"
	filter.Value = []interface{}{"John", "Jane"}
	clause, arg = db.buildWhereClause(filter)
	if clause != "\"name\" IN (?,?)" {
		t.Errorf("Expected '\"name\" IN (?,?)', got '%s'", clause)
	}
}

func TestDB_Quote(t *testing.T) {
	db := &DB{dbType: parser.DatabaseSQLite}
	
	quoted := db.quote("table_name")
	if quoted != "\"table_name\"" {
		t.Errorf("Expected '\"table_name\"', got '%s'", quoted)
	}

	db.dbType = parser.DatabaseType("unknown")
	quoted = db.quote("table_name")
	if quoted != "table_name" {
		t.Errorf("Expected 'table_name', got '%s'", quoted)
	}
}

func TestDB_BuildInsertQuery(t *testing.T) {
	db := &DB{dbType: parser.DatabaseSQLite}

	data := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	query, args := db.buildInsertQuery("User", data)
	
	if !containsString(query, "INSERT INTO \"User\"") {
		t.Error("Expected query to contain INSERT INTO \"User\"")
	}
	if !containsString(query, "VALUES") {
		t.Error("Expected query to contain VALUES")
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestDB_BuildUpdateQuery(t *testing.T) {
	db := &DB{dbType: parser.DatabaseSQLite}

	data := map[string]interface{}{
		"name":  "John Updated",
		"email": "john.updated@example.com",
	}

	query, args := db.buildUpdateQuery("User", 123, data)
	
	if !containsString(query, "UPDATE \"User\" SET") {
		t.Error("Expected query to contain UPDATE \"User\" SET")
	}
	if !containsString(query, "WHERE id = ?") {
		t.Error("Expected query to contain WHERE id = ?")
	}
	if len(args) != 3 { // 2 fields + 1 ID
		t.Errorf("Expected 3 args, got %d", len(args))
	}
	if args[len(args)-1] != 123 {
		t.Error("Expected last arg to be the ID")
	}
}

func TestDB_BuildDeleteQuery(t *testing.T) {
	db := &DB{dbType: parser.DatabaseSQLite}

	query, args := db.buildDeleteQuery("User", 123)
	
	expected := "DELETE FROM \"User\" WHERE id = ?"
	if query != expected {
		t.Errorf("Expected '%s', got '%s'", expected, query)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
	if args[0] != 123 {
		t.Error("Expected arg to be the ID")
	}
}

func TestDB_Close(t *testing.T) {
	db := &DB{}

	err := db.Close()
	if err != nil {
		t.Errorf("Expected no error for nil connection, got: %v", err)
	}
}

func TestDB_DefaultMethods(t *testing.T) {
	db := &DB{}

	err := db.Connect()
	if err == nil {
		t.Error("Expected error for base Connect method")
	}

	err = db.CreateSchema(nil)
	if err == nil {
		t.Error("Expected error for base CreateSchema method")
	}

	_, err = db.Query("User", parser.QueryParams{})
	if err == nil {
		t.Error("Expected error for base Query method")
	}

	_, err = db.Get("User", 1)
	if err == nil {
		t.Error("Expected error for base Get method")
	}

	_, err = db.Create("User", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for base Create method")
	}

	err = db.Update("User", 1, map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for base Update method")
	}

	err = db.Delete("User", 1)
	if err == nil {
		t.Error("Expected error for base Delete method")
	}

	_, err = db.Count("User", []parser.Filter{})
	if err == nil {
		t.Error("Expected error for base Count method")
	}
}

func TestEscapeSQL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal string", "normal string"},
		{"string with 'quote", "string with ''quote"},
		{"'start quote", "''start quote"},
		{"end quote'", "end quote''"},
		{"multiple 'quotes' here", "multiple ''quotes'' here"},
	}

	for _, test := range tests {
		result := escapeSQL(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected '%s', got '%s'", test.input, test.expected, result)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 len(s) > len(substr) && 
		 (s[:len(substr)] == substr || 
		  s[len(s)-len(substr):] == substr ||
		  containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}