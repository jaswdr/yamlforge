package parser

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFieldType_String(t *testing.T) {
	fieldType := FieldTypeText
	if fieldType.String() != "text" {
		t.Errorf("Expected 'text', got: %s", fieldType.String())
	}
}

func TestFieldType_IsValid(t *testing.T) {
	tests := []struct {
		fieldType FieldType
		expected  bool
	}{
		{FieldTypeText, true},
		{FieldTypeNumber, true},
		{FieldTypeBoolean, true},
		{FieldTypeDatetime, true},
		{FieldTypeDate, true},
		{FieldTypeTime, true},
		{FieldTypeID, true},
		{FieldTypeEmail, true},
		{FieldTypePassword, true},
		{FieldTypePhone, true},
		{FieldTypeURL, true},
		{FieldTypeSlug, true},
		{FieldTypeEnum, true},
		{FieldTypeColor, true},
		{FieldTypeFile, true},
		{FieldTypeImage, true},
		{FieldTypeMarkdown, true},
		{FieldTypeJSON, true},
		{FieldTypeArray, true},
		{FieldTypeRelation, true},
		{FieldTypeCurrency, true},
		{FieldTypeLocation, true},
		{FieldTypeIP, true},
		{FieldTypeUUID, true},
		{FieldTypeDuration, true},
		{FieldType("invalid"), false},
	}

	for _, test := range tests {
		result := test.fieldType.IsValid()
		if result != test.expected {
			t.Errorf("For field type %s, expected %v, got %v", test.fieldType, test.expected, result)
		}
	}
}

func TestFieldType_SQLType(t *testing.T) {
	tests := []struct {
		fieldType FieldType
		expected  string
	}{
		{FieldTypeID, "INTEGER"},
		{FieldTypeNumber, "INTEGER"},
		{FieldTypeText, "TEXT"},
		{FieldTypeEmail, "TEXT"},
		{FieldTypePassword, "TEXT"},
		{FieldTypePhone, "TEXT"},
		{FieldTypeURL, "TEXT"},
		{FieldTypeSlug, "TEXT"},
		{FieldTypeEnum, "TEXT"},
		{FieldTypeColor, "TEXT"},
		{FieldTypeMarkdown, "TEXT"},
		{FieldTypeJSON, "TEXT"},
		{FieldTypeCurrency, "TEXT"},
		{FieldTypeIP, "TEXT"},
		{FieldTypeUUID, "TEXT"},
		{FieldTypeDuration, "TEXT"},
		{FieldTypeBoolean, "BOOLEAN"},
		{FieldTypeDatetime, "DATETIME"},
		{FieldTypeDate, "DATETIME"},
		{FieldTypeTime, "DATETIME"},
		{FieldTypeFile, "TEXT"},
		{FieldTypeImage, "TEXT"},
		{FieldTypeArray, "TEXT"},
		{FieldTypeRelation, "INTEGER"},
		{FieldTypeLocation, "TEXT"},
		{FieldType("unknown"), "TEXT"}, // Default case
	}

	for _, test := range tests {
		result := test.fieldType.SQLType()
		if result != test.expected {
			t.Errorf("For field type %s, expected SQL type %s, got %s", test.fieldType, test.expected, result)
		}
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "username",
		Message: "field is required",
	}

	expected := "username: field is required"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestRequiredRule_Validate(t *testing.T) {
	rule := RequiredRule{}

	// Test with nil value
	err := rule.Validate(nil)
	if err == nil {
		t.Error("Expected error for nil value")
	}

	// Test with empty string
	err = rule.Validate("")
	if err == nil {
		t.Error("Expected error for empty string")
	}

	// Test with valid value
	err = rule.Validate("valid value")
	if err != nil {
		t.Errorf("Expected no error for valid value, got: %v", err)
	}
}

func TestMinLengthRule_Validate(t *testing.T) {
	rule := MinLengthRule{Min: 5}

	// Test with non-string value (should not validate)
	err := rule.Validate(123)
	if err != nil {
		t.Errorf("Expected no error for non-string value, got: %v", err)
	}

	// Test with string too short
	err = rule.Validate("abc")
	if err == nil {
		t.Error("Expected error for string too short")
	}

	// Test with string long enough
	err = rule.Validate("abcdef")
	if err != nil {
		t.Errorf("Expected no error for valid string, got: %v", err)
	}
}

func TestMaxLengthRule_Validate(t *testing.T) {
	rule := MaxLengthRule{Max: 5}

	// Test with non-string value (should not validate)
	err := rule.Validate(123)
	if err != nil {
		t.Errorf("Expected no error for non-string value, got: %v", err)
	}

	// Test with string too long
	err = rule.Validate("abcdefgh")
	if err == nil {
		t.Error("Expected error for string too long")
	}

	// Test with string short enough
	err = rule.Validate("abc")
	if err != nil {
		t.Errorf("Expected no error for valid string, got: %v", err)
	}
}

func TestTimestamp(t *testing.T) {
	now := time.Now()
	ts := Timestamp{
		CreatedAt: now,
		UpdatedAt: now,
	}

	if ts.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if ts.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

// Test all enum types and constants
func TestDatabaseType(t *testing.T) {
	if DatabaseSQLite != "sqlite" {
		t.Errorf("Expected DatabaseSQLite to be 'sqlite', got: %s", DatabaseSQLite)
	}
}

func TestAuthType(t *testing.T) {
	tests := []struct {
		authType AuthType
		expected string
	}{
		{AuthNone, "none"},
		{AuthBasic, "basic"},
		{AuthJWT, "jwt"},
	}

	for _, test := range tests {
		if string(test.authType) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.authType))
		}
	}
}

func TestUITheme(t *testing.T) {
	tests := []struct {
		theme    UITheme
		expected string
	}{
		{UIThemeLight, "light"},
		{UIThemeDark, "dark"},
		{UIThemeAuto, "auto"},
	}

	for _, test := range tests {
		if string(test.theme) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.theme))
		}
	}
}

func TestUILayout(t *testing.T) {
	tests := []struct {
		layout   UILayout
		expected string
	}{
		{UILayoutSidebar, "sidebar"},
		{UILayoutTopbar, "topbar"},
		{UILayoutMinimal, "minimal"},
	}

	for _, test := range tests {
		if string(test.layout) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.layout))
		}
	}
}

func TestRelationType(t *testing.T) {
	tests := []struct {
		relationType RelationType
		expected     string
	}{
		{RelationCascade, "cascade"},
		{RelationRestrict, "restrict"},
		{RelationSetNull, "set_null"},
	}

	for _, test := range tests {
		if string(test.relationType) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.relationType))
		}
	}
}

func TestAPIResponse(t *testing.T) {
	response := APIResponse{
		Success: true,
		Data:    map[string]interface{}{"id": 1, "name": "test"},
		Error:   "",
		Meta: &Meta{
			Page:       1,
			PageSize:   20,
			TotalCount: 100,
			TotalPages: 5,
		},
	}

	if !response.Success {
		t.Error("Expected Success to be true")
	}

	if response.Meta.TotalPages != 5 {
		t.Errorf("Expected TotalPages to be 5, got %d", response.Meta.TotalPages)
	}
}

func TestQueryParams(t *testing.T) {
	params := QueryParams{
		Page:     2,
		PageSize: 50,
		Sort: []SortField{
			{Field: "name", Desc: false},
			{Field: "created_at", Desc: true},
		},
		Filters: []Filter{
			{Field: "status", Operator: "=", Value: "active"},
			{Field: "name", Operator: "like", Value: "test"},
		},
		Search: "search term",
	}

	if params.Page != 2 {
		t.Errorf("Expected Page to be 2, got %d", params.Page)
	}

	if len(params.Sort) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(params.Sort))
	}

	if params.Sort[1].Desc != true {
		t.Error("Expected second sort field to be descending")
	}

	if len(params.Filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(params.Filters))
	}
}

func TestRequestContext(t *testing.T) {
	ctx := RequestContext{
		UserID:   "user123",
		Role:     "admin",
		TenantID: "tenant456",
	}

	if ctx.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got %s", ctx.UserID)
	}

	if ctx.Role != "admin" {
		t.Errorf("Expected Role 'admin', got %s", ctx.Role)
	}
}

// Test struct initialization and field assignments
func TestStructInitialization(t *testing.T) {
	// Test Config struct
	config := Config{
		App: AppConfig{
			Name:        "Test App",
			Version:     "1.0.0",
			Description: "Test Description",
		},
		Database: DatabaseConfig{
			Type:       "sqlite",
			Path:       "./test.db",
			Connection: "",
		},
		Server: ServerConfig{
			Port: 3000,
			Host: "localhost",
			CORS: CORSConfig{
				Enabled: true,
				Origins: []string{"http://localhost:3000"},
			},
			Auth: AuthConfig{
				Type:    "jwt",
				Secret:  "secret",
				Expires: "24h",
				Users: []UserConfig{
					{
						Username: "admin",
						Password: "password",
						Email:    "admin@test.com",
						Role:     "admin",
						Active:   true,
						Permissions: map[string]EntityPermission{
							"users": {Read: true, Write: true},
						},
					},
				},
			},
		},
		UI: UIConfig{
			Theme:  "light",
			Title:  "Test App",
			Logo:   "logo.png",
			Layout: "sidebar",
		},
		Models: map[string]ModelConfig{
			"User": {
				Fields: map[string]FieldConfig{
					"id": {
						Type:       "id",
						Primary:    true,
						Required:   true,
						Unique:     true,
						Min:        0,
						Max:        0,
						Pattern:    "",
						Options:    []string{},
						Default:    nil,
						AutoNow:    false,
						AutoNowAdd: false,
						Nullable:   false,
						Index:      false,
						To:         "",
						OnDelete:   "",
						Items:      "",
					},
				},
				UI: &UIModelConfig{
					List: &UIListConfig{
						Columns:    []string{"id", "name"},
						Sortable:   []string{"name"},
						Searchable: []string{"name"},
					},
					Form: &UIFormConfig{
						Fields: []string{"name"},
					},
				},
				Permissions: &PermissionsConfig{
					Create: "authenticated",
					Read:   "all",
					Update: "owner",
					Delete: "admin",
				},
			},
		},
	}

	// Basic verification that all fields can be set
	if config.App.Name != "Test App" {
		t.Error("Failed to set App.Name")
	}

	if len(config.Server.Auth.Users) != 1 {
		t.Error("Failed to set Auth.Users")
	}

	user := config.Server.Auth.Users[0]
	if !user.Permissions["users"].Read {
		t.Error("Failed to set nested permissions")
	}

	if len(config.Models) != 1 {
		t.Error("Failed to set Models map")
	}
}

func TestFieldTypeValidationComprehensive(t *testing.T) {
	testCases := []struct {
		fieldType FieldType
		sqlType   string
		valid     bool
	}{
		{FieldTypeID, "INTEGER", true},
		{FieldTypeText, "TEXT", true},
		{FieldTypeNumber, "INTEGER", true},
		{FieldTypeBoolean, "BOOLEAN", true},
		{FieldTypeEmail, "TEXT", true},
		{FieldTypeDatetime, "DATETIME", true},
		{FieldTypeEnum, "TEXT", true},
		{FieldTypeRelation, "INTEGER", true},
		{FieldTypeArray, "TEXT", true},
		{FieldTypeMarkdown, "TEXT", true},
		{"invalid_type", "", false},
	}

	for _, tc := range testCases {
		if tc.fieldType.IsValid() != tc.valid {
			t.Errorf("Field type %s validity check failed", tc.fieldType)
		}

		if tc.valid && tc.fieldType.SQLType() != tc.sqlType {
			t.Errorf("Field type %s expected SQL type %s, got %s", tc.fieldType, tc.sqlType, tc.fieldType.SQLType())
		}
	}
}

func TestValidationRulesComprehensive(t *testing.T) {
	// Test RequiredRule
	requiredRule := RequiredRule{}

	err := requiredRule.Validate("")
	if err == nil {
		t.Error("Expected required rule to fail on empty string")
	}

	err = requiredRule.Validate(nil)
	if err == nil {
		t.Error("Expected required rule to fail on nil")
	}

	err = requiredRule.Validate("valid")
	if err != nil {
		t.Errorf("Expected required rule to pass on valid string, got: %v", err)
	}

	// Test MinLengthRule
	minRule := MinLengthRule{Min: 5}

	err = minRule.Validate("hi")
	if err == nil {
		t.Error("Expected min length rule to fail on short string")
	}

	err = minRule.Validate("long enough")
	if err != nil {
		t.Errorf("Expected min length rule to pass on long string, got: %v", err)
	}

	// Test MaxLengthRule
	maxRule := MaxLengthRule{Max: 5}

	err = maxRule.Validate("too long")
	if err == nil {
		t.Error("Expected max length rule to fail on long string")
	}

	err = maxRule.Validate("ok")
	if err != nil {
		t.Errorf("Expected max length rule to pass on short string, got: %v", err)
	}
}

func TestAPIResponseStructure(t *testing.T) {
	// Test successful response
	successResponse := APIResponse{
		Success: true,
		Data:    map[string]interface{}{"id": 1, "name": "Test"},
	}

	jsonData, err := json.Marshal(successResponse)
	if err != nil {
		t.Fatalf("Failed to marshal API response: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal API response: %v", err)
	}

	if parsed["success"] != true {
		t.Error("Expected success field to be true")
	}

	// Test error response
	errorResponse := APIResponse{
		Success: false,
		Error:   "Test error",
	}

	jsonData, err = json.Marshal(errorResponse)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if parsed["success"] != false {
		t.Error("Expected success field to be false")
	}

	if parsed["error"] != "Test error" {
		t.Error("Expected error field to contain error message")
	}
}

func TestQueryParametersHandling(t *testing.T) {
	queryParams := QueryParams{
		Page:     2,
		PageSize: 50,
		Search:   "test search",
		Sort: []SortField{
			{Field: "name", Desc: false},
			{Field: "created_at", Desc: true},
		},
		Filters: []Filter{
			{Field: "active", Operator: "=", Value: true},
			{Field: "role", Operator: "in", Value: []string{"admin", "user"}},
		},
	}

	// Verify query parameters structure
	if queryParams.Page != 2 {
		t.Errorf("Expected page 2, got %d", queryParams.Page)
	}

	if len(queryParams.Sort) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(queryParams.Sort))
	}

	if queryParams.Sort[0].Field != "name" || queryParams.Sort[0].Desc != false {
		t.Error("First sort field incorrect")
	}

	if queryParams.Sort[1].Field != "created_at" || queryParams.Sort[1].Desc != true {
		t.Error("Second sort field incorrect")
	}

	if len(queryParams.Filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(queryParams.Filters))
	}
}

func TestValidationErrorHandling(t *testing.T) {
	validationError := ValidationError{
		Field:   "name",
		Message: "field is required",
	}

	errorString := validationError.Error()
	expectedError := "name: field is required"

	if errorString != expectedError {
		t.Errorf("Expected error string '%s', got '%s'", expectedError, errorString)
	}
}

func TestTimestampStructure(t *testing.T) {
	now := time.Now()
	timestamp := Timestamp{
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Verify timestamp can be marshaled to JSON
	jsonData, err := json.Marshal(timestamp)
	if err != nil {
		t.Fatalf("Failed to marshal timestamp: %v", err)
	}

	var parsed Timestamp
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal timestamp: %v", err)
	}

	if parsed.CreatedAt.IsZero() {
		t.Error("Expected created_at to be set")
	}

	if parsed.UpdatedAt.IsZero() {
		t.Error("Expected updated_at to be set")
	}
}