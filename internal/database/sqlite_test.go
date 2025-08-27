package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yamlforge/yamlforge/internal/parser"
)

func createTestSQLiteDB(t *testing.T) (*SQLiteDB, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	config := &parser.DatabaseConfig{
		Type: "sqlite",
		Path: dbPath,
	}

	db, err := NewSQLite(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite database: %v", err)
	}

	sqliteDB, ok := db.(*SQLiteDB)
	if !ok {
		t.Fatalf("Expected SQLiteDB, got: %T", db)
	}

	return sqliteDB, dbPath
}

func createTestSchema() *parser.Schema {
	return &parser.Schema{
		Models: map[string]*parser.Model{
			"User": {
				Name: "User",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID, Primary: true},
					{Name: "name", Type: parser.FieldTypeText, Required: true, Max: &[]int{100}[0]},
					{Name: "email", Type: parser.FieldTypeEmail, Required: true, Unique: true},
					{Name: "age", Type: parser.FieldTypeNumber, Min: &[]int{0}[0], Max: &[]int{120}[0]},
					{Name: "active", Type: parser.FieldTypeBoolean, Default: true},
					{Name: "role", Type: parser.FieldTypeEnum, Options: []string{"user", "admin"}, Default: "user"},
					{Name: "created_at", Type: parser.FieldTypeDatetime, AutoNowAdd: true, Default: "CURRENT_TIMESTAMP"},
					{Name: "updated_at", Type: parser.FieldTypeDatetime, AutoNow: true, Default: "CURRENT_TIMESTAMP"},
					{Name: "description", Type: parser.FieldTypeText, Nullable: true},
					{Name: "avatar", Type: parser.FieldTypeFile},
					{Name: "tags", Type: parser.FieldTypeArray, ArrayType: "text"},
				},
				UI: parser.UIModel{
					List: parser.UIList{
						Searchable: []string{"name", "email"},
					},
				},
			},
			"Post": {
				Name: "Post",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID, Primary: true},
					{Name: "title", Type: parser.FieldTypeText, Required: true},
					{Name: "user_id", Type: parser.FieldTypeRelation, RelatedTo: "User", OnDelete: "cascade", Index: true},
				},
			},
		},
	}
}

func TestNewSQLite(t *testing.T) {
	config := &parser.DatabaseConfig{
		Type: "sqlite",
		Path: "./test.db",
	}

	db, err := NewSQLite(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	sqliteDB, ok := db.(*SQLiteDB)
	if !ok {
		t.Errorf("Expected SQLiteDB, got: %T", db)
	}

	if sqliteDB.config != config {
		t.Error("Expected config to be stored")
	}

	if sqliteDB.dbType != parser.DatabaseSQLite {
		t.Error("Expected dbType to be SQLite")
	}
}

func TestSQLiteDB_Connect(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	err := db.Connect()
	if err != nil {
		t.Fatalf("Expected no error connecting, got: %v", err)
	}

	if db.conn == nil {
		t.Fatal("Expected connection to be established")
	}

	var foreignKeys int
	err = db.conn.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("Failed to check foreign keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Error("Expected foreign keys to be enabled")
	}

	db.Close()
}

func TestSQLiteDB_Close(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	err := db.Close()
	if err != nil {
		t.Errorf("Expected no error closing nil connection, got: %v", err)
	}

	err = db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Expected no error closing connection, got: %v", err)
	}
}

func TestSQLiteDB_CreateSchema(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	schema := createTestSchema()
	err = db.CreateSchema(schema)
	if err != nil {
		t.Fatalf("Expected no error creating schema, got: %v", err)
	}

	tables := []string{"User", "Post"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err = db.conn.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table existence: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected table %s to exist", table)
		}
	}

	var indexCount int
	query := "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'"
	err = db.conn.QueryRow(query).Scan(&indexCount)
	if err != nil {
		t.Fatalf("Failed to check indexes: %v", err)
	}
	if indexCount == 0 {
		t.Error("Expected at least one index to be created")
	}
}

func TestSQLiteDB_BuildColumnDefinition(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	tests := []struct {
		field    parser.Field
		expected string
	}{
		{
			field:    parser.Field{Name: "id", Type: parser.FieldTypeID, Primary: true},
			expected: "\"id\" INTEGER PRIMARY KEY AUTOINCREMENT",
		},
		{
			field:    parser.Field{Name: "name", Type: parser.FieldTypeText, Required: true, Unique: true},
			expected: "\"name\" TEXT NOT NULL UNIQUE",
		},
		{
			field:    parser.Field{Name: "count", Type: parser.FieldTypeNumber, Default: 0},
			expected: "\"count\" INTEGER DEFAULT 0",
		},
		{
			field:    parser.Field{Name: "active", Type: parser.FieldTypeBoolean, Default: true},
			expected: "\"active\" BOOLEAN DEFAULT 1",
		},
		{
			field:    parser.Field{Name: "created", Type: parser.FieldTypeDatetime, Default: "CURRENT_TIMESTAMP"},
			expected: "\"created\" DATETIME DEFAULT CURRENT_TIMESTAMP",
		},
	}

	for _, test := range tests {
		result := db.buildColumnDefinition(test.field)
		if result != test.expected {
			t.Errorf("For field %+v, expected '%s', got '%s'", test.field, test.expected, result)
		}
	}
}

func TestSQLiteDB_GetSQLiteType(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	tests := []struct {
		field    parser.Field
		expected string
	}{
		{parser.Field{Type: parser.FieldTypeID}, "INTEGER"},
		{parser.Field{Type: parser.FieldTypeNumber}, "INTEGER"},
		{parser.Field{Type: parser.FieldTypeText}, "TEXT"},
		{parser.Field{Type: parser.FieldTypeText, Max: &[]int{50}[0]}, "VARCHAR(50)"},
		{parser.Field{Type: parser.FieldTypeBoolean}, "BOOLEAN"},
		{parser.Field{Type: parser.FieldTypeDatetime}, "DATETIME"},
		{parser.Field{Type: parser.FieldTypeFile}, "TEXT"},
		{parser.Field{Type: parser.FieldTypeArray}, "TEXT"},
		{parser.Field{Type: parser.FieldTypeRelation}, "INTEGER"},
	}

	for _, test := range tests {
		result := db.getSQLiteType(test.field)
		if result != test.expected {
			t.Errorf("For field type %s, expected '%s', got '%s'", test.field.Type, test.expected, result)
		}
	}
}

func TestSQLiteDB_FormatDefaultValue(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	tests := []struct {
		value     interface{}
		fieldType parser.FieldType
		expected  string
	}{
		{"CURRENT_TIMESTAMP", parser.FieldTypeDatetime, "CURRENT_TIMESTAMP"},
		{"hello", parser.FieldTypeText, "'hello'"},
		{42, parser.FieldTypeNumber, "42"},
		{int64(42), parser.FieldTypeNumber, "42"},
		{3.14, parser.FieldTypeNumber, "3.14"},
		{true, parser.FieldTypeBoolean, "1"},
		{false, parser.FieldTypeBoolean, "0"},
		{nil, parser.FieldTypeText, "NULL"},
		{[]string{"invalid"}, parser.FieldTypeText, "NULL"},
	}

	for _, test := range tests {
		result := db.formatDefaultValue(test.value, test.fieldType)
		if result != test.expected {
			t.Errorf("For value %v, expected '%s', got '%s'", test.value, test.expected, result)
		}
	}
}

func TestSQLiteDB_CRUD_Operations(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	schema := createTestSchema()
	err = db.CreateSchema(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	userData := map[string]interface{}{
		"name":   "John Doe",
		"email":  "john@example.com",
		"age":    30,
		"active": true,
		"role":   "user",
	}

	userID, err := db.Create("User", userData)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if userID == nil {
		t.Fatal("Expected user ID to be returned")
	}

	user, err := db.Get("User", userID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user["name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got: %v", user["name"])
	}

	updateData := map[string]interface{}{
		"name": "John Updated",
		"age":  31,
	}

	err = db.Update("User", userID, updateData)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	updatedUser, err := db.Get("User", userID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser["name"] != "John Updated" {
		t.Errorf("Expected updated name 'John Updated', got: %v", updatedUser["name"])
	}

	params := parser.QueryParams{
		Page:     1,
		PageSize: 10,
	}

	users, err := db.Query("User", params)
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got: %d", len(users))
	}

	count, err := db.Count("User", []parser.Filter{})
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got: %d", count)
	}

	err = db.Delete("User", userID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err = db.Get("User", userID)
	if err != sql.ErrNoRows {
		t.Error("Expected user to be deleted")
	}
}

func TestSQLiteDB_Query_WithFilters(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	schema := createTestSchema()
	err = db.CreateSchema(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	users := []map[string]interface{}{
		{"name": "Alice", "email": "alice@example.com", "age": 25, "role": "user"},
		{"name": "Bob", "email": "bob@example.com", "age": 30, "role": "admin"},
		{"name": "Charlie", "email": "charlie@example.com", "age": 35, "role": "user"},
	}

	for _, user := range users {
		_, err = db.Create("User", user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	params := parser.QueryParams{
		Filters: []parser.Filter{
			{Field: "role", Operator: "=", Value: "user"},
		},
	}

	results, err := db.Query("User", params)
	if err != nil {
		t.Fatalf("Failed to query with filter: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 users with role 'user', got: %d", len(results))
	}

	params = parser.QueryParams{
		Search: "Alice",
	}

	results, err = db.Query("User", params)
	if err != nil {
		t.Fatalf("Failed to query with search: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 user matching 'Alice', got: %d", len(results))
	}

	params = parser.QueryParams{
		Sort: []parser.SortField{
			{Field: "age", Desc: true},
		},
	}

	results, err = db.Query("User", params)
	if err != nil {
		t.Fatalf("Failed to query with sort: %v", err)
	}

	if len(results) > 0 {
		firstUser := results[0]
		if firstUser["name"] != "Charlie" {
			t.Errorf("Expected Charlie to be first (oldest), got: %v", firstUser["name"])
		}
	}
}

func TestSQLiteDB_Count_WithFilters(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	schema := createTestSchema()
	err = db.CreateSchema(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	users := []map[string]interface{}{
		{"name": "Alice", "email": "alice@example.com", "role": "user"},
		{"name": "Bob", "email": "bob@example.com", "role": "admin"},
		{"name": "Charlie", "email": "charlie@example.com", "role": "user"},
	}

	for _, user := range users {
		_, err = db.Create("User", user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	filters := []parser.Filter{
		{Field: "role", Operator: "=", Value: "user"},
	}

	count, err := db.Count("User", filters)
	if err != nil {
		t.Fatalf("Failed to count with filter: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2 for role 'user', got: %d", count)
	}
}

func TestSQLiteDB_GetConnection(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	conn := db.GetConnection()
	if conn == nil {
		t.Error("Expected connection to be returned")
	}

	var version string
	err = conn.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		t.Errorf("Failed to query SQLite version: %v", err)
	}
}

func TestSQLiteDB_BuildWhereClause(t *testing.T) {
	db, _ := createTestSQLiteDB(t)

	tests := []struct {
		filter   parser.Filter
		expected string
	}{
		{
			filter:   parser.Filter{Field: "name", Value: "John"},
			expected: "\"name\" = ?",
		},
		{
			filter:   parser.Filter{Field: "name", Operator: "like", Value: "John"},
			expected: "\"name\" LIKE ?",
		},
		{
			filter:   parser.Filter{Field: "id", Operator: "in", Value: []interface{}{1, 2, 3}},
			expected: "\"id\" IN (?,?,?)",
		},
		{
			filter:   parser.Filter{Field: "age", Operator: ">", Value: 18},
			expected: "\"age\" > ?",
		},
	}

	for _, test := range tests {
		clause, _ := db.buildWhereClause(test.filter)
		if clause != test.expected {
			t.Errorf("For filter %+v, expected '%s', got '%s'", test.filter, test.expected, clause)
		}
	}
}

func TestSQLiteDB_ScanRow(t *testing.T) {
	db, dbPath := createTestSQLiteDB(t)
	defer os.Remove(dbPath)

	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	_, err = db.conn.Exec("CREATE TABLE test (id INTEGER, name TEXT, data BLOB)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.conn.Exec("INSERT INTO test (id, name, data) VALUES (?, ?, ?)", 
		1, "test", []byte("binary data"))
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	rows, err := db.conn.Query("SELECT * FROM test")
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Expected at least one row")
	}

	result, err := db.scanRow(rows)
	if err != nil {
		t.Fatalf("Failed to scan row: %v", err)
	}

	if result["id"] != int64(1) {
		t.Errorf("Expected id 1, got: %v", result["id"])
	}
	if result["name"] != "test" {
		t.Errorf("Expected name 'test', got: %v", result["name"])
	}
	if result["data"] != "binary data" {
		t.Errorf("Expected data 'binary data', got: %v", result["data"])
	}
}

func TestIntegration_CompleteWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	
	configFile := filepath.Join(tmpDir, "integration_test_config.yaml")
	testConfig := `app:
  name: "Integration Test App"
  version: "1.0.0"
  description: "Test application for integration testing"

database:
  type: sqlite
  path: ` + filepath.Join(tmpDir, "test.db") + `

server:
  port: 0
  host: "localhost"
  cors:
    enabled: true
    origins: ["*"]
  auth:
    type: none

ui:
  theme: "light"
  title: "Test App"
  layout: "sidebar"

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true
        min: 2
        max: 100
      email:
        type: email
        required: true
        unique: true
      age:
        type: number
        min: 0
        max: 150
      active:
        type: boolean
        default: true
      role:
        type: enum
        options: ["user", "admin", "moderator"]
        default: "user"
      created_at:
        type: datetime
        auto_now_add: true
      updated_at:
        type: datetime
        auto_now: true

    ui:
      list:
        columns: ["name", "email", "role", "active"]
        sortable: ["name", "email", "created_at"]
        searchable: ["name", "email"]
      form:
        fields: ["name", "email", "age", "active", "role"]

    permissions:
      create: "all"
      read: "all"
      update: "all"
      delete: "all"

  Post:
    fields:
      id:
        type: id
        primary: true
      title:
        type: text
        required: true
        min: 5
        max: 200
      content:
        type: markdown
      published:
        type: boolean
        default: false
      author_id:
        type: relation
        to: "User"
        required: true
      created_at:
        type: datetime
        auto_now_add: true

    ui:
      list:
        columns: ["title", "published", "created_at"]
        sortable: ["title", "created_at"]
        searchable: ["title", "content"]
      form:
        fields: ["title", "content", "published", "author_id"]`

	err := os.WriteFile(configFile, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := parser.ParseConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if config.App.Name != "Integration Test App" {
		t.Errorf("Expected app name 'Integration Test App', got: %s", config.App.Name)
	}

	if len(config.Models) != 2 {
		t.Errorf("Expected 2 models, got: %d", len(config.Models))
	}

	schema, err := parser.LoadConfig(config)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	userModel, exists := schema.GetModel("User")
	if !exists {
		t.Fatal("Expected User model to exist")
	}

	if len(userModel.Fields) != 8 {
		t.Errorf("Expected 8 fields in User model, got: %d", len(userModel.Fields))
	}

	db, err := NewDatabase(&config.Database)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	err = db.CreateSchema(schema)
	if err != nil {
		t.Fatalf("Failed to create database schema: %v", err)
	}

	userData := map[string]interface{}{
		"name":   "John Doe",
		"email":  "john@example.com",
		"age":    30,
		"active": true,
		"role":   "admin",
	}

	userID, err := db.Create("User", userData)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user, err := db.Get("User", userID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user["name"] != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got: %v", user["name"])
	}
	if user["email"] != "john@example.com" {
		t.Errorf("Expected user email 'john@example.com', got: %v", user["email"])
	}

	postData := map[string]interface{}{
		"title":     "Test Post",
		"content":   "This is a test post content",
		"published": true,
		"author_id": userID,
	}

	postID, err := db.Create("Post", postData)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	updateData := map[string]interface{}{
		"name": "John Updated",
		"age":  31,
	}

	err = db.Update("User", userID, updateData)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	updatedUser, err := db.Get("User", userID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser["name"] != "John Updated" {
		t.Errorf("Expected updated name 'John Updated', got: %v", updatedUser["name"])
	}

	params := parser.QueryParams{
		Page:     1,
		PageSize: 10,
	}

	users, err := db.Query("User", params)
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got: %d", len(users))
	}

	count, err := db.Count("User", []parser.Filter{})
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got: %d", count)
	}

	err = db.Delete("Post", postID)
	if err != nil {
		t.Fatalf("Failed to delete post: %v", err)
	}

	err = db.Delete("User", userID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err = db.Get("User", userID)
	if err == nil {
		t.Error("Expected user to be deleted")
	}
}

func TestIntegration_PerformanceBasics(t *testing.T) {
	
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "perf_test.yaml")
	
	testConfig := `app:
  name: "Performance Test"
database:
  type: sqlite
  path: ` + filepath.Join(tmpDir, "perf.db") + `
server:
  auth:
    type: none
models:
  Item:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true`

	err := os.WriteFile(configFile, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	config, err := parser.ParseConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	schema, err := parser.LoadConfig(config)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	db, err := NewDatabase(&config.Database)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	err = db.CreateSchema(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	start := time.Now()
	itemCount := 100
	
	for i := 0; i < itemCount; i++ {
		itemData := map[string]interface{}{
			"name": fmt.Sprintf("Item %d", i),
		}
		
		_, err = db.Create("Item", itemData)
		if err != nil {
			t.Fatalf("Failed to create item %d: %v", i, err)
		}
	}
	
	createDuration := time.Since(start)
	
	if createDuration > 5*time.Second {
		t.Errorf("Creating %d records took too long: %v", itemCount, createDuration)
	}

	start = time.Now()
	
	params := parser.QueryParams{
		Page:     1,
		PageSize: 50,
	}
	
	items, err := db.Query("Item", params)
	if err != nil {
		t.Fatalf("Failed to query items: %v", err)
	}
	
	queryDuration := time.Since(start)
	
	if len(items) != 50 {
		t.Errorf("Expected 50 items, got %d", len(items))
	}
	
	if queryDuration > 1*time.Second {
		t.Errorf("Querying records took too long: %v", queryDuration)
	}

	count, err := db.Count("Item", []parser.Filter{})
	if err != nil {
		t.Fatalf("Failed to count items: %v", err)
	}
	
	if count != int64(itemCount) {
		t.Errorf("Expected count %d, got %d", itemCount, count)
	}
}