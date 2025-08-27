package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yamlforge/yamlforge/internal/parser"
)

func createTestSchema() *parser.Schema {
	return &parser.Schema{
		Models: map[string]*parser.Model{
			"User": {
				Name: "User",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID, Primary: true},
					{Name: "name", Type: parser.FieldTypeText, Required: true, Min: &[]int{3}[0], Max: &[]int{50}[0]},
					{Name: "email", Type: parser.FieldTypeEmail, Required: true, Unique: true},
					{Name: "age", Type: parser.FieldTypeNumber, Min: &[]int{0}[0], Max: &[]int{120}[0]},
					{Name: "active", Type: parser.FieldTypeBoolean, Default: true},
					{Name: "role", Type: parser.FieldTypeEnum, Options: []string{"user", "admin", "moderator"}},
					{Name: "website", Type: parser.FieldTypeURL},
					{Name: "password", Type: parser.FieldTypePassword, Required: true, Min: &[]int{8}[0]},
					{Name: "created_at", Type: parser.FieldTypeDatetime, AutoNowAdd: true},
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

func TestNew(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	if validator.schema != schema {
		t.Error("Expected validator to store schema reference")
	}
}

func TestValidateCreate_ValidData(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	data := map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"age":      30,
		"active":   true,
		"role":     "user",
		"password": "password123",
		"website":  "https://johndoe.com",
	}

	err := validator.ValidateCreate("User", data)
	if err != nil {
		t.Errorf("Expected no error for valid data, got: %v", err)
	}
}

func TestValidateCreate_ModelNotFound(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	data := map[string]interface{}{"name": "test"}

	err := validator.ValidateCreate("NonExistent", data)
	if err == nil {
		t.Fatal("Expected error for non-existent model")
	}
	if err.Error() != "model NonExistent not found" {
		t.Errorf("Expected 'model NonExistent not found', got: %s", err.Error())
	}
}

func TestValidateCreate_SkipPrimaryKeyID(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	data := map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}

	err := validator.ValidateCreate("User", data)
	if err != nil {
		t.Errorf("Expected no error when ID is missing (auto-generated), got: %v", err)
	}
}

func TestValidateCreate_RequiredFieldMissing(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	data := map[string]interface{}{
		"email": "john@example.com",
	}

	err := validator.ValidateCreate("User", data)
	if err == nil {
		t.Fatal("Expected error for missing required field")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Field != "name" {
		t.Errorf("Expected error for 'name' field, got: %s", validationErr.Field)
	}
}

func TestValidateUpdate_ValidData(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	data := map[string]interface{}{
		"name":   "Jane Doe",
		"active": false,
	}

	err := validator.ValidateUpdate("User", data)
	if err != nil {
		t.Errorf("Expected no error for valid update data, got: %v", err)
	}
}

func TestValidateUpdate_ModelNotFound(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	data := map[string]interface{}{"name": "test"}

	err := validator.ValidateUpdate("NonExistent", data)
	if err == nil {
		t.Fatal("Expected error for non-existent model")
	}
	if err.Error() != "model NonExistent not found" {
		t.Errorf("Expected 'model NonExistent not found', got: %s", err.Error())
	}
}

func TestValidateUpdate_NonExistentField(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	data := map[string]interface{}{
		"nonexistent_field": "value",
	}

	err := validator.ValidateUpdate("User", data)
	if err == nil {
		t.Fatal("Expected error for non-existent field")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Field != "nonexistent_field" {
		t.Errorf("Expected error for 'nonexistent_field', got: %s", validationErr.Field)
	}
}

func TestValidateUpdate_PrimaryKeyUpdate(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	data := map[string]interface{}{
		"id": 123, // trying to update primary key
	}

	err := validator.ValidateUpdate("User", data)
	if err == nil {
		t.Fatal("Expected error for primary key update")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "cannot update primary key" {
		t.Errorf("Expected 'cannot update primary key', got: %s", validationErr.Message)
	}
}

func TestValidateField_Text_ValidString(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "name",
		Type: parser.FieldTypeText,
		Min:  &[]int{3}[0],
		Max:  &[]int{50}[0],
	}

	err := validator.validateField(field, "Valid Name")
	if err != nil {
		t.Errorf("Expected no error for valid text, got: %v", err)
	}
}

func TestValidateField_Text_TooShort(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "name",
		Type: parser.FieldTypeText,
		Min:  &[]int{5}[0],
	}

	err := validator.validateField(field, "Hi")
	if err == nil {
		t.Fatal("Expected error for text too short")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be at least 5 characters" {
		t.Errorf("Expected 'must be at least 5 characters', got: %s", validationErr.Message)
	}
}

func TestValidateField_Text_TooLong(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "name",
		Type: parser.FieldTypeText,
		Max:  &[]int{5}[0],
	}

	err := validator.validateField(field, "This is too long")
	if err == nil {
		t.Fatal("Expected error for text too long")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be at most 5 characters" {
		t.Errorf("Expected 'must be at most 5 characters', got: %s", validationErr.Message)
	}
}

func TestValidateField_Text_NotString(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "name",
		Type: parser.FieldTypeText,
	}

	err := validator.validateField(field, 123)
	if err == nil {
		t.Fatal("Expected error for non-string value")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be a string" {
		t.Errorf("Expected 'must be a string', got: %s", validationErr.Message)
	}
}

func TestValidateField_Text_Pattern(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name:    "code",
		Type:    parser.FieldTypeText,
		Pattern: "^[A-Z]{3}$", // Must be 3 uppercase letters
	}

	// Valid pattern
	err := validator.validateField(field, "ABC")
	if err != nil {
		t.Errorf("Expected no error for valid pattern, got: %v", err)
	}

	// Invalid pattern
	err = validator.validateField(field, "abc")
	if err == nil {
		t.Fatal("Expected error for invalid pattern")
	}
}

func TestValidateField_Number_ValidNumber(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "age",
		Type: parser.FieldTypeNumber,
		Min:  &[]int{0}[0],
		Max:  &[]int{120}[0],
	}

	tests := []interface{}{
		25,      // int
		int64(30), // int64
		25.5,    // float64
	}

	for _, value := range tests {
		err := validator.validateField(field, value)
		if err != nil {
			t.Errorf("Expected no error for valid number %v, got: %v", value, err)
		}
	}
}

func TestValidateField_Number_TooSmall(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "age",
		Type: parser.FieldTypeNumber,
		Min:  &[]int{18}[0],
	}

	err := validator.validateField(field, 10)
	if err == nil {
		t.Fatal("Expected error for number too small")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be at least 18" {
		t.Errorf("Expected 'must be at least 18', got: %s", validationErr.Message)
	}
}

func TestValidateField_Number_TooBig(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "age",
		Type: parser.FieldTypeNumber,
		Max:  &[]int{100}[0],
	}

	err := validator.validateField(field, 150)
	if err == nil {
		t.Fatal("Expected error for number too big")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be at most 100" {
		t.Errorf("Expected 'must be at most 100', got: %s", validationErr.Message)
	}
}

func TestValidateField_Number_NotNumber(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "age",
		Type: parser.FieldTypeNumber,
	}

	err := validator.validateField(field, "not a number")
	if err == nil {
		t.Fatal("Expected error for non-number value")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be a number" {
		t.Errorf("Expected 'must be a number', got: %s", validationErr.Message)
	}
}

func TestValidateField_Boolean_Valid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "active",
		Type: parser.FieldTypeBoolean,
	}

	tests := []bool{true, false}

	for _, value := range tests {
		err := validator.validateField(field, value)
		if err != nil {
			t.Errorf("Expected no error for boolean %v, got: %v", value, err)
		}
	}
}

func TestValidateField_Boolean_Invalid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "active",
		Type: parser.FieldTypeBoolean,
	}

	err := validator.validateField(field, "not a boolean")
	if err == nil {
		t.Fatal("Expected error for non-boolean value")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be a boolean" {
		t.Errorf("Expected 'must be a boolean', got: %s", validationErr.Message)
	}
}

func TestValidateField_Email_Valid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "email",
		Type: parser.FieldTypeEmail,
	}

	validEmails := []string{
		"test@example.com",
		"user.name+tag@domain.com",
		"firstname.lastname@company.co.uk",
	}

	for _, email := range validEmails {
		err := validator.validateField(field, email)
		if err != nil {
			t.Errorf("Expected no error for valid email %s, got: %v", email, err)
		}
	}
}

func TestValidateField_Email_Invalid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "email",
		Type: parser.FieldTypeEmail,
	}

	invalidEmails := []string{
		"invalid-email",
		"@domain.com",
		"user@",
		"user@domain",
		"user name@domain.com",
	}

	for _, email := range invalidEmails {
		err := validator.validateField(field, email)
		if err == nil {
			t.Errorf("Expected error for invalid email %s", email)
		}
	}
}

func TestValidateField_Email_NotString(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "email",
		Type: parser.FieldTypeEmail,
	}

	err := validator.validateField(field, 123)
	if err == nil {
		t.Fatal("Expected error for non-string email")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be a string" {
		t.Errorf("Expected 'must be a string', got: %s", validationErr.Message)
	}
}

func TestValidateField_URL_Valid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "website",
		Type: parser.FieldTypeURL,
	}

	validUrls := []string{
		"https://example.com",
		"http://subdomain.domain.com/path",
		"https://domain.com:8080/path?query=value",
	}

	for _, url := range validUrls {
		err := validator.validateField(field, url)
		if err != nil {
			t.Errorf("Expected no error for valid URL %s, got: %v", url, err)
		}
	}
}

func TestValidateField_URL_Invalid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "website",
		Type: parser.FieldTypeURL,
	}

	invalidUrls := []string{
		"not-a-url",
		"ftp://example.com", // Only http/https allowed by the regex
		"example.com",       // Missing protocol
	}

	for _, url := range invalidUrls {
		err := validator.validateField(field, url)
		if err == nil {
			t.Errorf("Expected error for invalid URL %s", url)
		}
	}
}

func TestValidateField_Enum_Valid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name:    "role",
		Type:    parser.FieldTypeEnum,
		Options: []string{"user", "admin", "moderator"},
	}

	for _, option := range field.Options {
		err := validator.validateField(field, option)
		if err != nil {
			t.Errorf("Expected no error for valid enum option %s, got: %v", option, err)
		}
	}
}

func TestValidateField_Enum_Invalid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name:    "role",
		Type:    parser.FieldTypeEnum,
		Options: []string{"user", "admin", "moderator"},
	}

	err := validator.validateField(field, "invalid_role")
	if err == nil {
		t.Fatal("Expected error for invalid enum option")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	expectedMessage := "must be one of: user, admin, moderator"
	if validationErr.Message != expectedMessage {
		t.Errorf("Expected '%s', got: %s", expectedMessage, validationErr.Message)
	}
}

func TestValidateField_Enum_NotString(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name:    "role",
		Type:    parser.FieldTypeEnum,
		Options: []string{"user", "admin"},
	}

	err := validator.validateField(field, 123)
	if err == nil {
		t.Fatal("Expected error for non-string enum value")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be a string" {
		t.Errorf("Expected 'must be a string', got: %s", validationErr.Message)
	}
}

func TestValidateField_Datetime_Valid(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "created_at",
		Type: parser.FieldTypeDatetime,
	}

	validDatetimes := []string{
		"2023-01-01T12:00:00Z",
		"2023-12-31 23:59:59",
		"2023-06-15T14:30:00.000Z",
	}

	for _, datetime := range validDatetimes {
		err := validator.validateField(field, datetime)
		if err != nil {
			t.Errorf("Expected no error for datetime %s, got: %v", datetime, err)
		}
	}
}

func TestValidateField_Datetime_NotString(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "created_at",
		Type: parser.FieldTypeDatetime,
	}

	err := validator.validateField(field, 123)
	if err == nil {
		t.Fatal("Expected error for non-string datetime")
	}

	validationErr, ok := err.(parser.ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got: %T", err)
	}
	if validationErr.Message != "must be a string" {
		t.Errorf("Expected 'must be a string', got: %s", validationErr.Message)
	}
}

func TestValidateField_NullableField(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name:     "description",
		Type:     parser.FieldTypeText,
		Nullable: true,
		Required: false,
	}

	err := validator.validateField(field, nil)
	if err != nil {
		t.Errorf("Expected no error for nil value on nullable field, got: %v", err)
	}
}

func TestValidateField_UnsupportedType(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "unknown",
		Type: parser.FieldType("unknown_type"),
	}

	err := validator.validateField(field, "some value")
	if err != nil {
		t.Errorf("Expected no error for unsupported field type, got: %v", err)
	}
}

func TestGetField(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field, found := validator.getField(schema.Models["User"], "name")
	if !found {
		t.Fatal("Expected to find 'name' field")
	}
	if field.Name != "name" {
		t.Errorf("Expected field name 'name', got: %s", field.Name)
	}

	_, found = validator.getField(schema.Models["User"], "nonexistent")
	if found {
		t.Error("Expected not to find non-existent field")
	}
}

func TestValidateField_Password(t *testing.T) {
	schema := createTestSchema()
	validator := New(schema)

	field := parser.Field{
		Name: "password",
		Type: parser.FieldTypePassword,
		Min:  &[]int{8}[0],
	}

	// Valid password
	err := validator.validateField(field, "password123")
	if err != nil {
		t.Errorf("Expected no error for valid password, got: %v", err)
	}

	err = validator.validateField(field, "pass")
	if err == nil {
		t.Fatal("Expected error for too short password")
	}
}

func TestIntegration_ValidationWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	
	configFile := filepath.Join(tmpDir, "validation_test.yaml")
	testConfig := `app:
  name: "Validation Test App"

database:
  type: sqlite
  path: ` + filepath.Join(tmpDir, "validation_test.db") + `

server:
  auth:
    type: none

models:
  Product:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text
        required: true
        min: 3
        max: 50
      price:
        type: number
        required: true
        min: 0
      category:
        type: enum
        options: ["electronics", "clothing", "books"]
        required: true
      email:
        type: email
        required: true
      website:
        type: url
      active:
        type: boolean
        default: true`

	err := os.WriteFile(configFile, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	config, err := parser.ParseConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	schema, err := parser.LoadConfig(config)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	validator := New(schema)

	validData := map[string]interface{}{
		"name":     "Test Product",
		"price":    29.99,
		"category": "electronics",
		"email":    "test@example.com",
		"website":  "https://example.com",
		"active":   true,
	}

	err = validator.ValidateCreate("Product", validData)
	if err != nil {
		t.Errorf("Expected no error for valid data, got: %v", err)
	}

	invalidData := map[string]interface{}{
		"price": 29.99,
	}

	err = validator.ValidateCreate("Product", invalidData)
	if err == nil {
		t.Error("Expected error for missing required fields")
	}

	invalidLengthData := map[string]interface{}{
		"name":     "AB",
		"price":    29.99,
		"category": "electronics",
		"email":    "test@example.com",
	}

	err = validator.ValidateCreate("Product", invalidLengthData)
	if err == nil {
		t.Error("Expected error for invalid length")
	}

	invalidEnumData := map[string]interface{}{
		"name":     "Test Product",
		"price":    29.99,
		"category": "invalid_category",
		"email":    "test@example.com",
	}

	err = validator.ValidateCreate("Product", invalidEnumData)
	if err == nil {
		t.Error("Expected error for invalid enum")
	}

	invalidEmailData := map[string]interface{}{
		"name":     "Test Product",
		"price":    29.99,
		"category": "electronics",
		"email":    "not-an-email",
	}

	err = validator.ValidateCreate("Product", invalidEmailData)
	if err == nil {
		t.Error("Expected error for invalid email")
	}

	invalidURLData := map[string]interface{}{
		"name":     "Test Product",
		"price":    29.99,
		"category": "electronics",
		"email":    "test@example.com",
		"website":  "not-a-url",
	}

	err = validator.ValidateCreate("Product", invalidURLData)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}