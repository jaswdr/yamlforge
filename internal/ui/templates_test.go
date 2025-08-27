package ui

import (
	"strings"
	"testing"

	"github.com/yamlforge/yamlforge/internal/parser"
)

func createTestConfig() *parser.Config {
	return &parser.Config{
		App: parser.AppConfig{
			Name: "Test App",
		},
		Server: parser.ServerConfig{
			Auth: parser.AuthConfig{
				Type: "jwt",
			},
		},
	}
}

func createTestSchema() *parser.Schema {
	return &parser.Schema{
		Models: map[string]*parser.Model{
			"User": {
				Name: "User",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID, Primary: true},
					{Name: "name", Type: parser.FieldTypeText, Required: true},
					{Name: "email", Type: parser.FieldTypeEmail, Required: true},
					{Name: "password", Type: parser.FieldTypePassword},
					{Name: "age", Type: parser.FieldTypeNumber, Min: &[]int{0}[0], Max: &[]int{120}[0]},
					{Name: "active", Type: parser.FieldTypeBoolean, Default: true},
					{Name: "role", Type: parser.FieldTypeEnum, Options: []string{"user", "admin"}, Default: "user"},
					{Name: "bio", Type: parser.FieldTypeMarkdown},
					{Name: "created_at", Type: parser.FieldTypeDatetime, AutoNowAdd: true},
				},
				UI: parser.UIModel{
					List: parser.UIList{
						Columns:    []string{"name", "email"},
						Sortable:   []string{"name", "created_at"},
						Searchable: []string{"name", "email"},
					},
					Form: parser.UIForm{
						Fields: []string{"name", "email", "age", "active", "role", "bio"},
					},
				},
			},
			"Post": {
				Name: "Post",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID, Primary: true},
					{Name: "title", Type: parser.FieldTypeText, Required: true},
				},
			},
		},
	}
}

func TestFormatFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name", "Name"},
		{"first_name", "First Name"},
		{"email_address", "Email Address"},
		{"created_at", "Created At"},
		{"UPPERCASE", "Uppercase"},
		{"camelCase", "CamelCase"},
	}

	for _, test := range tests {
		result := formatFieldName(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected '%s', got '%s'", test.input, test.expected, result)
		}
	}
}

func TestGetHomeHTML(t *testing.T) {
	config := createTestConfig()
	schema := createTestSchema()
	modelPermissions := map[string]bool{
		"User": true,
		"Post": false,
	}

	html := GetHomeHTML(config, schema, modelPermissions)

	// Check basic structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Expected HTML to contain DOCTYPE")
	}
	if !strings.Contains(html, config.App.Name) {
		t.Error("Expected HTML to contain app name")
	}
	if !strings.Contains(html, "Dashboard") {
		t.Error("Expected HTML to contain Dashboard title")
	}

	// Check that models are included
	if !strings.Contains(html, "User") {
		t.Error("Expected HTML to contain User model")
	}
	if !strings.Contains(html, "Post") {
		t.Error("Expected HTML to contain Post model")
	}

	// Check auth-specific content
	if !strings.Contains(html, "Logout") {
		t.Error("Expected HTML to contain Logout link for JWT auth")
	}

	// Check permissions (User should have Add New button, Post should not)
	if !strings.Contains(html, `href="/user/new"`) {
		t.Error("Expected HTML to contain add new link for User")
	}
}

func TestGetHomeHTML_NoAuth(t *testing.T) {
	config := createTestConfig()
	config.Server.Auth.Type = "none"
	schema := createTestSchema()

	html := GetHomeHTML(config, schema, nil)

	// Should not contain logout link without auth
	if strings.Contains(html, "Logout") {
		t.Error("Expected HTML to not contain Logout link without auth")
	}
}

func TestGetListHTML(t *testing.T) {
	config := createTestConfig()
	schema := createTestSchema()
	model := schema.Models["User"]

	html := GetListHTML(config, schema, "User", model, true)

	// Check basic structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Expected HTML to contain DOCTYPE")
	}
	if !strings.Contains(html, "User - " + config.App.Name) {
		t.Error("Expected HTML to contain page title")
	}

	// Check that model columns are included
	if !strings.Contains(html, "Name") {
		t.Error("Expected HTML to contain Name column")
	}
	if !strings.Contains(html, "Email") {
		t.Error("Expected HTML to contain Email column")
	}

	// Check Add New button (canWrite = true)
	if !strings.Contains(html, `href="/user/new"`) {
		t.Error("Expected HTML to contain Add New button")
	}

	// Check JavaScript data
	if !strings.Contains(html, `const modelName = 'user'`) {
		t.Error("Expected HTML to contain model name in JavaScript")
	}
	if !strings.Contains(html, `const canWrite = true`) {
		t.Error("Expected HTML to contain canWrite flag")
	}
}

func TestGetListHTML_NoWritePermission(t *testing.T) {
	config := createTestConfig()
	schema := createTestSchema()
	model := schema.Models["User"]

	html := GetListHTML(config, schema, "User", model, false)

	// Should not contain Add New button
	if strings.Contains(html, `href="/user/new"`) {
		t.Error("Expected HTML to not contain Add New button without write permission")
	}
	if !strings.Contains(html, `const canWrite = false`) {
		t.Error("Expected HTML to contain canWrite = false")
	}
}

func TestGetFormHTML_Create(t *testing.T) {
	config := createTestConfig()
	schema := createTestSchema()
	model := schema.Models["User"]

	html := GetFormHTML(config, schema, "User", model, "create", "", "null")

	// Check basic structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Expected HTML to contain DOCTYPE")
	}
	if !strings.Contains(html, "New User") {
		t.Error("Expected HTML to contain New User title")
	}

	// Check form fields
	if !strings.Contains(html, `id="name"`) {
		t.Error("Expected HTML to contain name field")
	}
	if !strings.Contains(html, `id="email"`) {
		t.Error("Expected HTML to contain email field")
	}
	if !strings.Contains(html, `id="age"`) {
		t.Error("Expected HTML to contain age field")
	}
	if !strings.Contains(html, `id="role"`) {
		t.Error("Expected HTML to contain role field")
	}

	// Check submit button
	if !strings.Contains(html, "Create") {
		t.Error("Expected HTML to contain Create button")
	}

	// Check JavaScript
	if !strings.Contains(html, `const action = 'create'`) {
		t.Error("Expected HTML to contain action = 'create'")
	}
}

func TestGetFormHTML_Edit(t *testing.T) {
	config := createTestConfig()
	schema := createTestSchema()
	model := schema.Models["User"]
	recordJSON := `{"id": 1, "name": "John Doe", "email": "john@example.com"}`

	html := GetFormHTML(config, schema, "User", model, "edit", "1", recordJSON)

	// Check title
	if !strings.Contains(html, "Edit User 1") {
		t.Error("Expected HTML to contain Edit User title")
	}

	// Check submit button
	if !strings.Contains(html, "Update") {
		t.Error("Expected HTML to contain Update button")
	}

	// Check JavaScript
	if !strings.Contains(html, `const action = 'edit'`) {
		t.Error("Expected HTML to contain action = 'edit'")
	}
	if !strings.Contains(html, recordJSON) {
		t.Error("Expected HTML to contain record data")
	}
}

func TestGetViewHTML(t *testing.T) {
	config := createTestConfig()
	schema := createTestSchema()
	model := schema.Models["User"]
	recordJSON := `{"id": 1, "name": "John Doe", "email": "john@example.com"}`

	html := GetViewHTML(config, schema, "User", model, "1", recordJSON)

	// Check basic structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Expected HTML to contain DOCTYPE")
	}
	if !strings.Contains(html, "User Details") {
		t.Error("Expected HTML to contain User Details title")
	}

	// Check action buttons
	if !strings.Contains(html, `href="/user/1/edit"`) {
		t.Error("Expected HTML to contain Edit link")
	}
	if !strings.Contains(html, `onclick="deleteRecord('user', '1')"`) {
		t.Error("Expected HTML to contain Delete button")
	}
	if !strings.Contains(html, `href="/user"`) {
		t.Error("Expected HTML to contain Back to List link")
	}

	// Check record data
	if !strings.Contains(html, recordJSON) {
		t.Error("Expected HTML to contain record data")
	}
}

func TestGenerateFormField_Text(t *testing.T) {
	field := &parser.Field{
		Name:     "name",
		Type:     parser.FieldTypeText,
		Required: true,
		Min:      &[]int{3}[0],
		Max:      &[]int{50}[0],
		Default:  "Default Name",
	}

	html := generateFormField(field)

	if !strings.Contains(html, `id="name"`) {
		t.Error("Expected field to have correct ID")
	}
	if !strings.Contains(html, `type="text"`) {
		t.Error("Expected field to be text type")
	}
	if !strings.Contains(html, `required`) {
		t.Error("Expected field to be required")
	}
	if !strings.Contains(html, `minlength="3"`) {
		t.Error("Expected field to have min length")
	}
	if !strings.Contains(html, `maxlength="50"`) {
		t.Error("Expected field to have max length")
	}
	if !strings.Contains(html, `data-default="Default Name"`) {
		t.Error("Expected field to have default value")
	}
	if !strings.Contains(html, "Name*") {
		t.Error("Expected field label to have required asterisk")
	}
}

func TestGenerateFormField_Email(t *testing.T) {
	field := &parser.Field{
		Name:     "email",
		Type:     parser.FieldTypeEmail,
		Required: true,
	}

	html := generateFormField(field)

	if !strings.Contains(html, `type="email"`) {
		t.Error("Expected field to be email type")
	}
}

func TestGenerateFormField_Password(t *testing.T) {
	field := &parser.Field{
		Name:     "password",
		Type:     parser.FieldTypePassword,
		Required: true,
		Min:      &[]int{8}[0],
	}

	html := generateFormField(field)

	if !strings.Contains(html, `type="password"`) {
		t.Error("Expected field to be password type")
	}
	// Password fields should not have default value
	if strings.Contains(html, `data-default`) {
		t.Error("Expected password field to not have default value")
	}
}

func TestGenerateFormField_Number(t *testing.T) {
	field := &parser.Field{
		Name:    "age",
		Type:    parser.FieldTypeNumber,
		Min:     &[]int{0}[0],
		Max:     &[]int{120}[0],
		Default: 25,
	}

	html := generateFormField(field)

	if !strings.Contains(html, `type="number"`) {
		t.Error("Expected field to be number type")
	}
	if !strings.Contains(html, `min="0"`) {
		t.Error("Expected field to have min value")
	}
	if !strings.Contains(html, `max="120"`) {
		t.Error("Expected field to have max value")
	}
	if !strings.Contains(html, `data-default="25"`) {
		t.Error("Expected field to have default value")
	}
}

func TestGenerateFormField_Boolean(t *testing.T) {
	field := &parser.Field{
		Name:    "active",
		Type:    parser.FieldTypeBoolean,
		Default: true,
	}

	html := generateFormField(field)

	if !strings.Contains(html, `type="checkbox"`) {
		t.Error("Expected field to be checkbox type")
	}
	if !strings.Contains(html, `data-default="true"`) {
		t.Error("Expected field to have default value")
	}
}

func TestGenerateFormField_Enum(t *testing.T) {
	field := &parser.Field{
		Name:    "role",
		Type:    parser.FieldTypeEnum,
		Options: []string{"user", "admin", "moderator"},
		Default: "user",
	}

	html := generateFormField(field)

	if !strings.Contains(html, `<select`) {
		t.Error("Expected field to be select element")
	}
	if !strings.Contains(html, `<option value="user">User</option>`) {
		t.Error("Expected field to have user option")
	}
	if !strings.Contains(html, `<option value="admin">Admin</option>`) {
		t.Error("Expected field to have admin option")
	}
	if !strings.Contains(html, `data-default="user"`) {
		t.Error("Expected field to have default value")
	}
}

func TestGenerateFormField_Textarea(t *testing.T) {
	field := &parser.Field{
		Name:    "bio",
		Type:    parser.FieldTypeMarkdown,
		Default: "Default bio",
	}

	html := generateFormField(field)

	if !strings.Contains(html, `<textarea`) {
		t.Error("Expected field to be textarea element")
	}
	if !strings.Contains(html, `data-default="Default bio"`) {
		t.Error("Expected field to have default value")
	}
}

func TestGenerateFormField_Date(t *testing.T) {
	field := &parser.Field{
		Name: "birth_date",
		Type: parser.FieldTypeDate,
	}

	html := generateFormField(field)

	if !strings.Contains(html, `type="date"`) {
		t.Error("Expected field to be date type")
	}
}

func TestGenerateFormField_Datetime(t *testing.T) {
	field := &parser.Field{
		Name: "created_at",
		Type: parser.FieldTypeDatetime,
	}

	html := generateFormField(field)

	if !strings.Contains(html, `type="datetime-local"`) {
		t.Error("Expected field to be datetime-local type")
	}
}

func TestGenerateFormField_Time(t *testing.T) {
	field := &parser.Field{
		Name: "start_time",
		Type: parser.FieldTypeTime,
	}

	html := generateFormField(field)

	if !strings.Contains(html, `type="time"`) {
		t.Error("Expected field to be time type")
	}
}

func TestRequiredStar(t *testing.T) {
	if requiredStar(true) != "*" {
		t.Error("Expected asterisk for required field")
	}
	if requiredStar(false) != "" {
		t.Error("Expected no asterisk for non-required field")
	}
}

func TestBuildModelInfoJSON(t *testing.T) {
	model := &parser.Model{
		Fields: []parser.Field{
			{Name: "id", Type: parser.FieldTypeID},
			{Name: "role", Type: parser.FieldTypeEnum, Options: []string{"user", "admin"}},
			{Name: "password", Type: parser.FieldTypePassword},
		},
	}

	jsonStr := buildModelInfoJSON(model)

	if !strings.Contains(jsonStr, `"fields"`) {
		t.Error("Expected JSON to contain fields")
	}
	if !strings.Contains(jsonStr, `"type":"enum"`) {
		t.Error("Expected JSON to contain enum type")
	}
	if !strings.Contains(jsonStr, `"user":"User"`) {
		t.Error("Expected JSON to contain enum options")
	}
	if !strings.Contains(jsonStr, `"type":"password"`) {
		t.Error("Expected JSON to contain password type")
	}
}

func TestGetLoginHTML(t *testing.T) {
	config := createTestConfig()

	html := GetLoginHTML(config)

	// Check basic structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Expected HTML to contain DOCTYPE")
	}
	if !strings.Contains(html, config.App.Name + " - Login") {
		t.Error("Expected HTML to contain app name in title")
	}
	if !strings.Contains(html, config.App.Name) {
		t.Error("Expected HTML to contain app name in header")
	}

	// Check form elements
	if !strings.Contains(html, `id="username"`) {
		t.Error("Expected HTML to contain username field")
	}
	if !strings.Contains(html, `id="password"`) {
		t.Error("Expected HTML to contain password field")
	}
	if !strings.Contains(html, `id="loginForm"`) {
		t.Error("Expected HTML to contain login form")
	}

	// Check JavaScript
	if !strings.Contains(html, "/api/auth/login") {
		t.Error("Expected HTML to contain login API endpoint")
	}
}

func TestGetListHTML_DefaultColumns(t *testing.T) {
	config := createTestConfig()
	schema := createTestSchema()
	
	model := &parser.Model{
		Name: "SimpleModel",
		Fields: []parser.Field{
			{Name: "id", Type: parser.FieldTypeID, Primary: true},
			{Name: "title", Type: parser.FieldTypeText},
			{Name: "description", Type: parser.FieldTypeText},
			{Name: "created_at", Type: parser.FieldTypeDatetime, AutoNowAdd: true},
			{Name: "updated_at", Type: parser.FieldTypeDatetime, AutoNow: true},
		},
		UI: parser.UIModel{
			List: parser.UIList{}, // Empty list config
		},
	}

	html := GetListHTML(config, schema, "SimpleModel", model, true)

	// Should contain generated columns (excluding id, auto fields)
	if !strings.Contains(html, "Title") {
		t.Error("Expected HTML to contain Title column from default generation")
	}
	if !strings.Contains(html, "Description") {
		t.Error("Expected HTML to contain Description column from default generation")
	}
}

func TestGetFormHTML_DefaultFields(t *testing.T) {
	config := createTestConfig()
	schema := createTestSchema()
	
	model := &parser.Model{
		Name: "SimpleModel",
		Fields: []parser.Field{
			{Name: "id", Type: parser.FieldTypeID, Primary: true},
			{Name: "title", Type: parser.FieldTypeText, Required: true},
			{Name: "created_at", Type: parser.FieldTypeDatetime, AutoNowAdd: true},
		},
		UI: parser.UIModel{
			Form: parser.UIForm{Fields: []string{}}, // Empty form fields
		},
	}

	html := GetFormHTML(config, schema, "SimpleModel", model, "create", "", "null")

	// Should contain generated form fields (excluding id, auto fields)
	if !strings.Contains(html, `id="title"`) {
		t.Error("Expected HTML to contain title field from default generation")
	}
	// Should not contain auto fields
	if strings.Contains(html, `id="created_at"`) {
		t.Error("Expected HTML to not contain auto field")
	}
}