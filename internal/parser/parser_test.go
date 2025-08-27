package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "valid_config.yaml")
	
	validConfig := `app:
  name: "Test App"
  version: "1.0.0"
  description: "A test application"

database:
  type: sqlite
  path: "./test.db"

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true
      email:
        type: email
        required: true
        unique: true`

	err := os.WriteFile(configFile, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := ParseConfig(configFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.App.Name != "Test App" {
		t.Errorf("Expected app name 'Test App', got: %s", config.App.Name)
	}

	if config.Database.Type != "sqlite" {
		t.Errorf("Expected database type 'sqlite', got: %s", config.Database.Type)
	}

	if len(config.Models) != 1 {
		t.Errorf("Expected 1 model, got: %d", len(config.Models))
	}

	user, exists := config.Models["User"]
	if !exists {
		t.Fatal("Expected User model to exist")
	}

	if len(user.Fields) != 3 {
		t.Errorf("Expected 3 fields, got: %d", len(user.Fields))
	}
}

func TestParseConfig_InvalidFile(t *testing.T) {
	_, err := ParseConfig("nonexistent_file.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestParseConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")
	
	invalidYAML := `invalid: yaml: content:
  - item1
    - nested_incorrectly`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err = ParseConfig(configFile)
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}
}

func TestValidateConfig_MissingAppName(t *testing.T) {
	config := &Config{
		App:      AppConfig{},
		Database: DatabaseConfig{Type: "sqlite", Path: "./test.db"},
	}

	err := validateConfig(config)
	if err == nil {
		t.Fatal("Expected error for missing app name")
	}
	if err.Error() != "app.name is required" {
		t.Errorf("Expected 'app.name is required', got: %s", err.Error())
	}
}

func TestValidateConfig_UnsupportedDatabase(t *testing.T) {
	config := &Config{
		App:      AppConfig{Name: "Test App"},
		Database: DatabaseConfig{Type: "redis"},
	}

	err := validateConfig(config)
	if err == nil {
		t.Fatal("Expected error for unsupported database")
	}
	if err.Error() != "unsupported database type: redis" {
		t.Errorf("Expected 'unsupported database type: redis', got: %s", err.Error())
	}
}

func TestValidateConfig_SQLiteMissingPath(t *testing.T) {
	config := &Config{
		App:      AppConfig{Name: "Test App"},
		Database: DatabaseConfig{Type: "sqlite"},
	}

	err := validateConfig(config)
	if err == nil {
		t.Fatal("Expected error for missing SQLite path")
	}
	if err.Error() != "database.path is required for SQLite" {
		t.Errorf("Expected 'database.path is required for SQLite', got: %s", err.Error())
	}
}

func TestValidateConfig_PostgreSQLMissingConnection(t *testing.T) {
	config := &Config{
		App:      AppConfig{Name: "Test App"},
		Database: DatabaseConfig{Type: "postgresql"},
	}

	err := validateConfig(config)
	if err == nil {
		t.Fatal("Expected error for missing PostgreSQL connection")
	}
	if err.Error() != "database.connection is required for postgresql" {
		t.Errorf("Expected 'database.connection is required for postgresql', got: %s", err.Error())
	}
}

func TestValidateModel_NoFields(t *testing.T) {
	model := ModelConfig{
		Fields: map[string]FieldConfig{},
	}

	err := validateModel("TestModel", model)
	if err == nil {
		t.Fatal("Expected error for model with no fields")
	}
	if err.Error() != "model TestModel has no fields" {
		t.Errorf("Expected 'model TestModel has no fields', got: %s", err.Error())
	}
}

func TestValidateModel_NoPrimaryKey(t *testing.T) {
	model := ModelConfig{
		Fields: map[string]FieldConfig{
			"name": {Type: "text"},
		},
	}

	err := validateModel("TestModel", model)
	if err == nil {
		t.Fatal("Expected error for model with no primary key")
	}
	if err.Error() != "model TestModel has no primary key" {
		t.Errorf("Expected 'model TestModel has no primary key', got: %s", err.Error())
	}
}

func TestValidateModel_MultiplePrimaryKeys(t *testing.T) {
	model := ModelConfig{
		Fields: map[string]FieldConfig{
			"id":   {Type: "id", Primary: true},
			"uuid": {Type: "text", Primary: true},
		},
	}

	err := validateModel("TestModel", model)
	if err == nil {
		t.Fatal("Expected error for model with multiple primary keys")
	}
	if err.Error() != "model TestModel has multiple primary keys" {
		t.Errorf("Expected 'model TestModel has multiple primary keys', got: %s", err.Error())
	}
}

func TestValidateField_InvalidType(t *testing.T) {
	field := FieldConfig{Type: "invalid_type"}

	err := validateField("TestModel", "testField", field)
	if err == nil {
		t.Fatal("Expected error for invalid field type")
	}
}

func TestValidateField_EnumMissingOptions(t *testing.T) {
	field := FieldConfig{Type: "enum"}

	err := validateField("TestModel", "testField", field)
	if err == nil {
		t.Fatal("Expected error for enum field missing options")
	}
	if err.Error() != "enum field TestModel.testField must have options" {
		t.Errorf("Expected 'enum field TestModel.testField must have options', got: %s", err.Error())
	}
}

func TestValidateField_RelationMissingTo(t *testing.T) {
	field := FieldConfig{Type: "relation"}

	err := validateField("TestModel", "testField", field)
	if err == nil {
		t.Fatal("Expected error for relation field missing 'to'")
	}
	if err.Error() != "relation field TestModel.testField must specify 'to' model" {
		t.Errorf("Expected 'relation field TestModel.testField must specify 'to' model', got: %s", err.Error())
	}
}

func TestValidateField_ArrayMissingItems(t *testing.T) {
	field := FieldConfig{Type: "array"}

	err := validateField("TestModel", "testField", field)
	if err == nil {
		t.Fatal("Expected error for array field missing 'items'")
	}
	if err.Error() != "array field TestModel.testField must specify 'items' type" {
		t.Errorf("Expected 'array field TestModel.testField must specify 'items' type', got: %s", err.Error())
	}
}

func TestValidateField_MinMaxInvalid(t *testing.T) {
	field := FieldConfig{Type: "text", Min: 10, Max: 5}

	err := validateField("TestModel", "testField", field)
	if err == nil {
		t.Fatal("Expected error for min > max")
	}
	if err.Error() != "field TestModel.testField has min > max" {
		t.Errorf("Expected 'field TestModel.testField has min > max', got: %s", err.Error())
	}
}

func TestProcessConfig_IDFieldPrimaryKey(t *testing.T) {
	config := &Config{
		Models: map[string]ModelConfig{
			"User": {
				Fields: map[string]FieldConfig{
					"id": {Type: "id", Primary: false},
				},
			},
		},
	}

	processed := processConfig(config)
	
	idField := processed.Models["User"].Fields["id"]
	if !idField.Primary {
		t.Error("Expected ID field to be set as primary key")
	}
}

func TestProcessConfig_DatetimeDefaults(t *testing.T) {
	config := &Config{
		Models: map[string]ModelConfig{
			"User": {
				Fields: map[string]FieldConfig{
					"id":         {Type: "id", Primary: true},
					"created_at": {Type: "datetime", AutoNowAdd: true},
					"updated_at": {Type: "datetime", AutoNow: true},
				},
			},
		},
	}

	processed := processConfig(config)
	
	createdField := processed.Models["User"].Fields["created_at"]
	if createdField.Default != "CURRENT_TIMESTAMP" {
		t.Error("Expected auto_now_add field to have CURRENT_TIMESTAMP default")
	}

	updatedField := processed.Models["User"].Fields["updated_at"]
	if updatedField.Default != "CURRENT_TIMESTAMP" {
		t.Error("Expected auto_now field to have CURRENT_TIMESTAMP default")
	}
}

func TestProcessConfig_PrimaryKeyRequiredUnique(t *testing.T) {
	config := &Config{
		Models: map[string]ModelConfig{
			"User": {
				Fields: map[string]FieldConfig{
					"id": {Type: "id", Primary: true, Required: false, Unique: false},
				},
			},
		},
	}

	processed := processConfig(config)
	
	idField := processed.Models["User"].Fields["id"]
	if !idField.Required {
		t.Error("Expected primary key field to be required")
	}
	if !idField.Unique {
		t.Error("Expected primary key field to be unique")
	}
}

func TestProcessConfig_DefaultUI(t *testing.T) {
	config := &Config{
		Models: map[string]ModelConfig{
			"User": {
				Fields: map[string]FieldConfig{
					"id":   {Type: "id", Primary: true},
					"name": {Type: "text"},
				},
				UI: nil,
			},
		},
	}

	processed := processConfig(config)
	
	ui := processed.Models["User"].UI
	if ui == nil {
		t.Fatal("Expected UI to be generated")
	}
	if len(ui.List.Columns) == 0 {
		t.Error("Expected default list columns to be generated")
	}
}

func TestProcessConfig_DefaultPermissions(t *testing.T) {
	config := &Config{
		Models: map[string]ModelConfig{
			"User": {
				Fields: map[string]FieldConfig{
					"id": {Type: "id", Primary: true},
				},
				Permissions: nil,
			},
		},
	}

	processed := processConfig(config)
	
	permissions := processed.Models["User"].Permissions
	if permissions == nil {
		t.Fatal("Expected permissions to be generated")
	}
	if permissions.Create != "authenticated" {
		t.Error("Expected default create permission to be 'authenticated'")
	}
	if permissions.Read != "all" {
		t.Error("Expected default read permission to be 'all'")
	}
}

func TestGenerateDefaultUI(t *testing.T) {
	model := ModelConfig{
		Fields: map[string]FieldConfig{
			"id":       {Type: "id", Primary: true},
			"name":     {Type: "text"},
			"email":    {Type: "email"},
			"password": {Type: "password"},
			"created":  {Type: "datetime", AutoNowAdd: true},
		},
	}

	ui := generateDefaultUI(model)
	
	for _, col := range ui.List.Columns {
		if col == "password" {
			t.Error("Expected password field to be excluded from list columns")
		}
	}

	for _, col := range ui.List.Columns {
		if col == "id" {
			t.Error("Expected primary key field to be excluded from list columns")
		}
	}

	for _, field := range ui.Form.Fields {
		if field == "id" || field == "created" {
			t.Error("Expected auto fields to be excluded from form")
		}
	}
}

func TestLoadConfig(t *testing.T) {
	config := &Config{
		Models: map[string]ModelConfig{
			"User": {
				Fields: map[string]FieldConfig{
					"id":   {Type: "id", Primary: true},
					"name": {Type: "text", Required: true, Min: 3, Max: 50},
				},
				Permissions: &PermissionsConfig{
					Create: "admin",
					Read:   "all",
					Update: "owner",
					Delete: "admin",
				},
				UI: &UIModelConfig{
					List: &UIListConfig{
						Columns:    []string{"name"},
						Sortable:   []string{"name"},
						Searchable: []string{"name"},
					},
					Form: &UIFormConfig{
						Fields: []string{"name"},
					},
				},
			},
		},
	}

	schema, err := LoadConfig(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(schema.Models) != 1 {
		t.Errorf("Expected 1 model, got: %d", len(schema.Models))
	}

	user, exists := schema.Models["User"]
	if !exists {
		t.Fatal("Expected User model to exist")
	}

	if len(user.Fields) != 2 {
		t.Errorf("Expected 2 fields, got: %d", len(user.Fields))
	}

	var nameField *Field
	for _, field := range user.Fields {
		if field.Name == "name" {
			nameField = &field
			break
		}
	}
	if nameField == nil {
		t.Error("Expected to find 'name' field")
		return
	}
	if nameField.Min == nil || *nameField.Min != 3 {
		t.Error("Expected min constraint to be preserved")
	}
	if nameField.Max == nil || *nameField.Max != 50 {
		t.Error("Expected max constraint to be preserved")
	}
}

func TestSchema_GetModel(t *testing.T) {
	schema := &Schema{
		Models: map[string]*Model{
			"User": {Name: "User"},
		},
	}

	model, exists := schema.GetModel("User")
	if !exists {
		t.Fatal("Expected model to exist")
	}
	if model.Name != "User" {
		t.Errorf("Expected model name 'User', got: %s", model.Name)
	}

	_, exists = schema.GetModel("NonExistent")
	if exists {
		t.Error("Expected model to not exist")
	}
}

func TestSchema_GetField(t *testing.T) {
	schema := &Schema{
		Models: map[string]*Model{
			"User": {
				Name: "User",
				Fields: []Field{
					{Name: "id", Type: FieldTypeID},
					{Name: "name", Type: FieldTypeText},
				},
			},
		},
	}

	field, exists := schema.GetField("User", "name")
	if !exists {
		t.Fatal("Expected field to exist")
	}
	if field.Name != "name" {
		t.Errorf("Expected field name 'name', got: %s", field.Name)
	}

	_, exists = schema.GetField("User", "nonexistent")
	if exists {
		t.Error("Expected field to not exist")
	}

	_, exists = schema.GetField("NonExistentModel", "name")
	if exists {
		t.Error("Expected field to not exist for non-existent model")
	}
}

func TestSaveConfig(t *testing.T) {
	config := &Config{
		App: AppConfig{
			Name:    "Test App",
			Version: "1.0.0",
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "./test.db",
		},
	}

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "saved_config.yaml")

	err := SaveConfig(config, configFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Expected config file to be created")
	}

	parsedConfig, err := ParseConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to parse saved config: %v", err)
	}

	if parsedConfig.App.Name != config.App.Name {
		t.Errorf("Expected app name %s, got %s", config.App.Name, parsedConfig.App.Name)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.App.Name != "My Application" {
		t.Errorf("Expected default app name 'My Application', got: %s", config.App.Name)
	}

	if config.Database.Type != "sqlite" {
		t.Errorf("Expected default database type 'sqlite', got: %s", config.Database.Type)
	}

	if config.Server.Port != 8080 {
		t.Errorf("Expected default server port 8080, got: %d", config.Server.Port)
	}

	if !config.Server.CORS.Enabled {
		t.Error("Expected default CORS to be enabled")
	}

	if config.Models == nil {
		t.Error("Expected models map to be initialized")
	}
}

func TestComprehensiveConfigurationValidation(t *testing.T) {
	tmpDir := t.TempDir()
	
	configPath := filepath.Join(tmpDir, "comprehensive_config.yaml")
	testConfig := `app:
  name: "Comprehensive Test App"
  version: "1.0.0"
  description: "A comprehensive test application"

database:
  type: sqlite
  path: ` + filepath.Join(tmpDir, "test.db") + `

server:
  port: 0  # Let the system assign a port
  host: "127.0.0.1"
  cors:
    enabled: true
    origins: ["*"]
  auth:
    type: none

ui:
  theme: light
  title: "Test Application"
  layout: sidebar

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true
        min: 3
        max: 50
      email:
        type: email
        required: true
        unique: true
      age:
        type: number
        min: 0
        max: 120
      role:
        type: enum
        options: ["user", "admin"]
        default: "user"
      active:
        type: boolean
        default: true
      created_at:
        type: datetime
        auto_now_add: true
      updated_at:
        type: datetime
        auto_now: true
      bio:
        type: text
        nullable: true
        max: 500
    ui:
      list:
        columns: ["name", "email", "role", "active"]
        sortable: ["name", "email", "created_at"]
        searchable: ["name", "email"]
      form:
        fields: ["name", "email", "age", "role", "bio"]
    permissions:
      create: "all"
      read: "all"
      update: "owner"
      delete: "admin"

  Post:
    fields:
      id:
        type: id
        primary: true
      title:
        type: text
        required: true
        min: 10
        max: 100
      content:
        type: markdown
        required: true
      author_id:
        type: relation
        to: "User"
        required: true
        on_delete: cascade
      published:
        type: boolean
        default: false
      tags:
        type: array
        items: text
      views:
        type: number
        default: 0
      created_at:
        type: datetime
        auto_now_add: true
      updated_at:
        type: datetime
        auto_now: true
    ui:
      list:
        columns: ["title", "author_id", "published", "views"]
        sortable: ["title", "created_at", "views"]
        searchable: ["title", "content"]
      form:
        fields: ["title", "content", "author_id", "published", "tags"]`

	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := ParseConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}

	if config.App.Name != "Comprehensive Test App" {
		t.Errorf("Expected app name 'Comprehensive Test App', got %s", config.App.Name)
	}

	if len(config.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(config.Models))
	}

	userModel, exists := config.Models["User"]
	if !exists {
		t.Fatal("User model not found")
	}

	if len(userModel.Fields) != 9 {
		t.Errorf("Expected 9 fields in User model, got %d", len(userModel.Fields))
	}

	postModel, exists := config.Models["Post"]
	if !exists {
		t.Fatal("Post model not found")
	}

	authorField, exists := postModel.Fields["author_id"]
	if !exists {
		t.Fatal("author_id field not found in Post model")
	}

	if authorField.Type != "relation" {
		t.Errorf("Expected author_id to be relation type, got %s", authorField.Type)
	}

	if authorField.To != "User" {
		t.Errorf("Expected author_id to relate to User, got %s", authorField.To)
	}
}

func TestSchemaGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "schema_test.yaml")
	
	testConfig := `app:
  name: "Schema Test App"
  version: "1.0.0"

database:
  type: sqlite
  path: ` + filepath.Join(tmpDir, "test.db") + `

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true
        min: 3
        max: 50
      email:
        type: email
        required: true
        unique: true`

	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := ParseConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}

	schema, err := LoadConfig(config)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	if len(schema.Models) != 1 {
		t.Errorf("Expected 1 model in schema, got %d", len(schema.Models))
	}

	userModel, exists := schema.GetModel("User")
	if !exists {
		t.Fatal("User model not found in schema")
	}

	if len(userModel.Fields) != 3 {
		t.Errorf("Expected 3 fields in User schema, got %d", len(userModel.Fields))
	}

	var nameField *Field
	for _, field := range userModel.Fields {
		if field.Name == "name" {
			nameField = &field
			break
		}
	}

	if nameField == nil {
		t.Fatal("name field not found")
	}

	if nameField.Type != FieldTypeText {
		t.Errorf("Expected name field to be text type, got %v", nameField.Type)
	}

	if !nameField.Required {
		t.Error("Expected name field to be required")
	}

	if nameField.Min == nil || *nameField.Min != 3 {
		t.Error("Expected name field min length to be 3")
	}

	if nameField.Max == nil || *nameField.Max != 50 {
		t.Error("Expected name field max length to be 50")
	}
}

func TestCompleteModelWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "workflow_test.yaml")
	
	testConfig := `app:
  name: "Workflow Test App"
  version: "1.0.0"

database:
  type: sqlite
  path: ` + filepath.Join(tmpDir, "test.db") + `

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true
        min: 3
        max: 50
      email:
        type: email
        required: true
        unique: true
      age:
        type: number
        min: 0
        max: 120
      role:
        type: enum
        options: ["user", "admin"]
        default: "user"
      active:
        type: boolean
        default: true
      created_at:
        type: datetime
        auto_now_add: true
      updated_at:
        type: datetime
        auto_now: true
      bio:
        type: text
        nullable: true
        max: 500
    ui:
      list:
        columns: ["name", "email", "role", "active"]
        sortable: ["name", "email", "created_at"]
        searchable: ["name", "email"]
      form:
        fields: ["name", "email", "age", "role", "bio"]
    permissions:
      create: "all"
      read: "all"
      update: "owner"
      delete: "admin"`

	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := ParseConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}

	schema, err := LoadConfig(config)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	userModel, exists := schema.GetModel("User")
	if !exists {
		t.Fatal("User model not found")
	}

	expectedFields := []string{"id", "name", "email", "age", "role", "active", "created_at", "updated_at", "bio"}
	actualFields := make(map[string]bool)

	for _, field := range userModel.Fields {
		actualFields[field.Name] = true
	}

	for _, expected := range expectedFields {
		if !actualFields[expected] {
			t.Errorf("Expected field '%s' not found in User model", expected)
		}
	}

	if len(userModel.UI.List.Columns) == 0 {
		t.Error("Expected UI list columns to be configured")
	}

	if len(userModel.UI.List.Searchable) == 0 {
		t.Error("Expected searchable fields to be configured")
	}

	if len(userModel.UI.Form.Fields) == 0 {
		t.Error("Expected form fields to be configured")
	}

	if userModel.Permissions.Create == "" {
		t.Error("Expected create permission to be set")
	}

	if userModel.Permissions.Read == "" {
		t.Error("Expected read permission to be set")
	}
}

func TestParseConfig_EmptyConfiguration(t *testing.T) {
	tmpDir := t.TempDir()
	emptyConfigPath := filepath.Join(tmpDir, "empty.yaml")

	err := os.WriteFile(emptyConfigPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty config: %v", err)
	}

	config, err := ParseConfig(emptyConfigPath)
	if err != nil {
		t.Fatalf("Expected empty config to use defaults, got error: %v", err)
	}

	if config.App.Name != "My Application" {
		t.Errorf("Expected default app name 'My Application', got: %s", config.App.Name)
	}

	if config.Database.Type != "sqlite" {
		t.Errorf("Expected default database type 'sqlite', got: %s", config.Database.Type)
	}
}

func TestParseConfig_InvalidYAMLContent(t *testing.T) {
	tmpDir := t.TempDir()
	invalidConfigPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `app:
  name: "Test App"
  version: "1.0.0"
invalid_yaml: [unclosed bracket`

	err := os.WriteFile(invalidConfigPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}

	_, err = ParseConfig(invalidConfigPath)
	if err == nil {
		t.Error("Expected parsing invalid YAML to fail")
	}
}

func TestParseConfig_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	incompleteConfigPath := filepath.Join(tmpDir, "incomplete.yaml")

	incompleteYAML := `database:
  type: sqlite
  path: "./test.db"

models:
  User:
    fields:
      id:
        type: id
        primary: true`

	err := os.WriteFile(incompleteConfigPath, []byte(incompleteYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create incomplete config: %v", err)
	}

	config, err := ParseConfig(incompleteConfigPath)
	if err != nil {
		t.Fatalf("Expected incomplete config to use defaults, got error: %v", err)
	}

	if config.App.Name != "My Application" {
		t.Errorf("Expected default app name 'My Application', got: %s", config.App.Name)
	}
}

func TestParseConfig_InvalidFieldTypes(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFieldConfigPath := filepath.Join(tmpDir, "invalid_field.yaml")

	invalidFieldYAML := `app:
  name: "Test App"
  version: "1.0.0"

database:
  type: sqlite
  path: "./test.db"

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: invalid_type
        required: true`

	err := os.WriteFile(invalidFieldConfigPath, []byte(invalidFieldYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid field config: %v", err)
	}

	_, err = ParseConfig(invalidFieldConfigPath)
	if err == nil {
		t.Error("Expected parsing config with invalid field type to fail")
	}
}

func TestParseConfig_InvalidRelations(t *testing.T) {
	tmpDir := t.TempDir()
	relationConfigPath := filepath.Join(tmpDir, "relations.yaml")

	relationYAML := `app:
  name: "Test App"
  version: "1.0.0"

database:
  type: sqlite
  path: "./test.db"

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true
      post_id:
        type: relation
        to: "NonExistentModel"
        required: true`

	err := os.WriteFile(relationConfigPath, []byte(relationYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create relation config: %v", err)
	}

	config, err := ParseConfig(relationConfigPath)
	if err != nil {
		t.Fatalf("Failed to parse relation config: %v", err)
	}

	userModel := config.Models["User"]
	postIDField, exists := userModel.Fields["post_id"]
	if !exists {
		t.Error("Expected post_id relation field to exist")
	}

	if postIDField.To != "NonExistentModel" {
		t.Errorf("Expected relation to NonExistentModel, got %s", postIDField.To)
	}
}

func TestPerformance_LargeModel(t *testing.T) {
	tmpDir := t.TempDir()
	largeConfigPath := filepath.Join(tmpDir, "large.yaml")

	var fieldsYAML strings.Builder
	fieldsYAML.WriteString("      id:\n        type: id\n        primary: true\n")

	for i := 1; i < 50; i++ {
		fieldsYAML.WriteString(fmt.Sprintf(`      field_%d:
        type: text
        required: false
        max: 100
`, i))
	}

	largeYAML := fmt.Sprintf(`app:
  name: "Large Test App"
  version: "1.0.0"

database:
  type: sqlite
  path: "./test.db"

models:
  LargeModel:
    fields:
%s`, fieldsYAML.String())

	err := os.WriteFile(largeConfigPath, []byte(largeYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create large config: %v", err)
	}

	start := time.Now()
	config, err := ParseConfig(largeConfigPath)
	parseTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to parse large config: %v", err)
	}

	largeModel := config.Models["LargeModel"]
	if len(largeModel.Fields) != 50 {
		t.Errorf("Expected 50 fields, got %d", len(largeModel.Fields))
	}

	start = time.Now()
	schema, err := LoadConfig(config)
	loadTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to load large schema: %v", err)
	}

	model, exists := schema.GetModel("LargeModel")
	if !exists {
		t.Fatal("Large model not found in schema")
	}

	if len(model.Fields) != 50 {
		t.Errorf("Expected 50 fields in schema, got %d", len(model.Fields))
	}

	t.Logf("Parse time for 50-field model: %v", parseTime)
	t.Logf("Schema load time for 50-field model: %v", loadTime)

	if parseTime > time.Second {
		t.Errorf("Parsing took too long: %v", parseTime)
	}

	if loadTime > time.Second {
		t.Errorf("Schema loading took too long: %v", loadTime)
	}
}

func TestPerformance_MultipleModels(t *testing.T) {
	tmpDir := t.TempDir()
	multiConfigPath := filepath.Join(tmpDir, "multi.yaml")

	var modelsYAML strings.Builder
	
	for i := 1; i <= 20; i++ {
		modelsYAML.WriteString(fmt.Sprintf(`  Model_%d:
    fields:
      id:
        type: id
        primary: true
      name_%d:
        type: text
        required: true
      email_%d:
        type: email
        unique: true
      count_%d:
        type: number
        default: 0
      active_%d:
        type: boolean
        default: true
`, i, i, i, i, i))
	}

	multiYAML := fmt.Sprintf(`app:
  name: "Multi Model Test App"
  version: "1.0.0"

database:
  type: sqlite
  path: "./test.db"

models:
%s`, modelsYAML.String())

	err := os.WriteFile(multiConfigPath, []byte(multiYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create multi model config: %v", err)
	}

	start := time.Now()
	config, err := ParseConfig(multiConfigPath)
	parseTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to parse multi model config: %v", err)
	}

	if len(config.Models) != 20 {
		t.Errorf("Expected 20 models, got %d", len(config.Models))
	}

	start = time.Now()
	schema, err := LoadConfig(config)
	loadTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to load multi model schema: %v", err)
	}

	if len(schema.Models) != 20 {
		t.Errorf("Expected 20 models in schema, got %d", len(schema.Models))
	}

	for i := 1; i <= 20; i++ {
		modelName := fmt.Sprintf("Model_%d", i)
		model, exists := schema.GetModel(modelName)
		if !exists {
			t.Errorf("Model %s not found", modelName)
			continue
		}

		if len(model.Fields) != 5 {
			t.Errorf("Expected 5 fields in %s, got %d", modelName, len(model.Fields))
		}
	}

	t.Logf("Parse time for 20 models: %v", parseTime)
	t.Logf("Schema load time for 20 models: %v", loadTime)

	if parseTime > 2*time.Second {
		t.Errorf("Multi-model parsing took too long: %v", parseTime)
	}

	if loadTime > 2*time.Second {
		t.Errorf("Multi-model schema loading took too long: %v", loadTime)
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	
	tmpDir := t.TempDir()
	invalidConfigFile := filepath.Join(tmpDir, "invalid_config.yaml")
	
	invalidConfig := `app:
  name: "Test App"
  invalid_yaml: [unclosed_bracket`
	err := os.WriteFile(invalidConfigFile, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	_, err = ParseConfig(invalidConfigFile)
	if err == nil {
		t.Error("Expected error for invalid YAML configuration")
	}

	explicitEmptyConfigFile := filepath.Join(tmpDir, "explicit_empty.yaml")
	explicitEmptyConfig := `app:
  name: ""
database:
  type: sqlite
  path: "test.db"`

	err = os.WriteFile(explicitEmptyConfigFile, []byte(explicitEmptyConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create explicit empty config file: %v", err)
	}

	_, err = ParseConfig(explicitEmptyConfigFile)
	if err == nil {
		t.Error("Expected error for explicitly empty app name")
	}

	invalidModelConfig := filepath.Join(tmpDir, "invalid_model.yaml")
	badModelConfig := `app:
  name: "Test App"
database:
  type: sqlite
  path: "test.db"
models:
  User:
    fields:
      name:
        type: text
        # Missing primary key`

	err = os.WriteFile(invalidModelConfig, []byte(badModelConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid model config: %v", err)
	}

	_, err = ParseConfig(invalidModelConfig)
	if err == nil {
		t.Error("Expected error for model with no primary key")
	}
}