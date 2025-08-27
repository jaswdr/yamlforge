package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yamlforge/yamlforge/internal/parser"
)

func createTestAPIForOpenAPI() *API {
	config := &parser.Config{
		App: parser.AppConfig{
			Name:        "Test API",
			Version:     "1.0.0",
			Description: "Test API Description",
		},
		Server: parser.ServerConfig{
			Port: 8080,
			Host: "localhost",
			Auth: parser.AuthConfig{Type: "jwt"},
		},
	}

	schema := &parser.Schema{
		Models: map[string]*parser.Model{
			"User": {
				Name: "User",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID, Primary: true},
					{Name: "name", Type: parser.FieldTypeText, Required: true, Min: &[]int{2}[0], Max: &[]int{100}[0]},
					{Name: "email", Type: parser.FieldTypeEmail, Required: true},
					{Name: "age", Type: parser.FieldTypeNumber, Min: &[]int{0}[0], Max: &[]int{120}[0]},
					{Name: "active", Type: parser.FieldTypeBoolean, Default: true},
					{Name: "role", Type: parser.FieldTypeEnum, Options: []string{"user", "admin"}, Default: "user"},
					{Name: "website", Type: parser.FieldTypeURL},
					{Name: "password", Type: parser.FieldTypePassword, Required: true},
					{Name: "avatar", Type: parser.FieldTypeFile},
					{Name: "created_at", Type: parser.FieldTypeDatetime, AutoNowAdd: true},
					{Name: "tags", Type: parser.FieldTypeArray, ArrayType: "text"},
					{Name: "location", Type: parser.FieldTypeLocation},
				},
			},
			"Post": {
				Name: "Post",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID, Primary: true},
					{Name: "title", Type: parser.FieldTypeText, Required: true},
					{Name: "user_id", Type: parser.FieldTypeRelation, RelatedTo: "User"},
				},
			},
		},
	}

	return New(NewMockDatabase(), config, schema, nil)
}

func TestGenerateOpenAPI_BasicStructure(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	req.Host = "test.example.com"
	
	spec := api.GenerateOpenAPI(req)

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("Expected OpenAPI version '3.0.3', got: %s", spec.OpenAPI)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got: %s", spec.Info.Title)
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got: %s", spec.Info.Version)
	}

	if spec.Info.Description != "Test API Description" {
		t.Errorf("Expected description 'Test API Description', got: %s", spec.Info.Description)
	}

	if len(spec.Servers) != 1 {
		t.Fatalf("Expected 1 server, got: %d", len(spec.Servers))
	}

	if spec.Servers[0].URL != "http://test.example.com/api" {
		t.Errorf("Expected server URL 'http://test.example.com/api', got: %s", spec.Servers[0].URL)
	}
}

func TestGenerateOpenAPI_HTTPSDetection(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	req.Host = "secure.example.com"
	req.TLS = nil
	
	spec := api.GenerateOpenAPI(req)

	if spec.Servers[0].URL != "http://secure.example.com/api" {
		t.Errorf("Expected HTTP server URL, got: %s", spec.Servers[0].URL)
	}
}

func TestGenerateOpenAPI_AuthenticationSchemes(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	spec := api.GenerateOpenAPI(req)

	if len(spec.Components.SecuritySchemes) != 2 {
		t.Fatalf("Expected 2 security schemes, got: %d", len(spec.Components.SecuritySchemes))
	}

	// Check bearer auth
	bearerAuth, exists := spec.Components.SecuritySchemes["bearerAuth"]
	if !exists {
		t.Fatal("Expected bearerAuth security scheme")
	}
	if bearerAuth.Type != "http" {
		t.Errorf("Expected bearer auth type 'http', got: %s", bearerAuth.Type)
	}
	if bearerAuth.Scheme != "bearer" {
		t.Errorf("Expected bearer auth scheme 'bearer', got: %s", bearerAuth.Scheme)
	}
	if bearerAuth.BearerFormat != "JWT" {
		t.Errorf("Expected bearer format 'JWT', got: %s", bearerAuth.BearerFormat)
	}

	// Check cookie auth
	cookieAuth, exists := spec.Components.SecuritySchemes["cookieAuth"]
	if !exists {
		t.Fatal("Expected cookieAuth security scheme")
	}
	if cookieAuth.Type != "apiKey" {
		t.Errorf("Expected cookie auth type 'apiKey', got: %s", cookieAuth.Type)
	}
	if cookieAuth.In != "cookie" {
		t.Errorf("Expected cookie auth in 'cookie', got: %s", cookieAuth.In)
	}
	if cookieAuth.Name != "auth_token" {
		t.Errorf("Expected cookie name 'auth_token', got: %s", cookieAuth.Name)
	}
}

func TestGenerateOpenAPI_NoAuth(t *testing.T) {
	api := createTestAPIForOpenAPI()
	api.config.Server.Auth.Type = "none"
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	spec := api.GenerateOpenAPI(req)

	if len(spec.Components.SecuritySchemes) != 0 {
		t.Errorf("Expected no security schemes for no auth, got: %d", len(spec.Components.SecuritySchemes))
	}

	// Check that auth endpoints are not included
	if _, exists := spec.Paths["/auth/login"]; exists {
		t.Error("Expected no login endpoint without auth")
	}
}

func TestGenerateOpenAPI_ModelPaths(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	spec := api.GenerateOpenAPI(req)

	// Check User model paths
	userPath := "/user"
	if _, exists := spec.Paths[userPath]; !exists {
		t.Errorf("Expected path %s to exist", userPath)
	}

	userItemPath := "/user/{id}"
	if _, exists := spec.Paths[userItemPath]; !exists {
		t.Errorf("Expected path %s to exist", userItemPath)
	}

	// Check operations
	userPathItem := spec.Paths[userPath]
	
	// Check GET (list)
	if _, exists := userPathItem["get"]; !exists {
		t.Error("Expected GET operation for user list")
	}

	// Check POST (create)
	if _, exists := userPathItem["post"]; !exists {
		t.Error("Expected POST operation for user create")
	}

	userItemPathItem := spec.Paths[userItemPath]
	
	// Check GET (single item)
	if _, exists := userItemPathItem["get"]; !exists {
		t.Error("Expected GET operation for single user")
	}

	// Check PUT (update)
	if _, exists := userItemPathItem["put"]; !exists {
		t.Error("Expected PUT operation for user update")
	}

	// Check DELETE
	if _, exists := userItemPathItem["delete"]; !exists {
		t.Error("Expected DELETE operation for user")
	}
}

func TestGenerateOpenAPI_AuthEndpoints(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	spec := api.GenerateOpenAPI(req)

	// Check login endpoint
	loginPath := "/auth/login"
	if _, exists := spec.Paths[loginPath]; !exists {
		t.Errorf("Expected path %s to exist", loginPath)
	}

	loginPathItem := spec.Paths[loginPath]
	loginOp := loginPathItem["post"]
	
	if loginOp.Summary != "Login" {
		t.Errorf("Expected login summary 'Login', got: %s", loginOp.Summary)
	}
	if loginOp.OperationID != "login" {
		t.Errorf("Expected login operation ID 'login', got: %s", loginOp.OperationID)
	}

	// Check logout endpoint
	logoutPath := "/auth/logout"
	if _, exists := spec.Paths[logoutPath]; !exists {
		t.Errorf("Expected path %s to exist", logoutPath)
	}

	// Check current user endpoint
	mePath := "/auth/me"
	if _, exists := spec.Paths[mePath]; !exists {
		t.Errorf("Expected path %s to exist", mePath)
	}
}

func TestGenerateModelSchema(t *testing.T) {
	api := createTestAPIForOpenAPI()
	model := api.schema.Models["User"]
	
	schema := api.generateModelSchema(model)

	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got: %s", schema.Type)
	}

	if len(schema.Properties) == 0 {
		t.Error("Expected schema to have properties")
	}

	// Check specific field properties
	nameProperty, exists := schema.Properties["name"]
	if !exists {
		t.Fatal("Expected 'name' property to exist")
	}
	if nameProperty.Type != "string" {
		t.Errorf("Expected name type 'string', got: %s", nameProperty.Type)
	}
	if nameProperty.MinLength == nil || *nameProperty.MinLength != 2 {
		t.Error("Expected name to have minLength 2")
	}
	if nameProperty.MaxLength == nil || *nameProperty.MaxLength != 100 {
		t.Error("Expected name to have maxLength 100")
	}

	// Check required fields
	requiredCount := 0
	for _, required := range schema.Required {
		if required == "name" || required == "email" || required == "password" {
			requiredCount++
		}
	}
	if requiredCount != 3 {
		t.Errorf("Expected 3 required fields, found %d", requiredCount)
	}
}

func TestGenerateInputSchema(t *testing.T) {
	api := createTestAPIForOpenAPI()
	model := api.schema.Models["User"]
	
	schema := api.generateInputSchema(model)

	// Input schema should exclude primary key and auto fields
	if _, exists := schema.Properties["id"]; exists {
		t.Error("Expected input schema to exclude primary key")
	}
	if _, exists := schema.Properties["created_at"]; exists {
		t.Error("Expected input schema to exclude auto fields")
	}

	// Should include regular fields
	if _, exists := schema.Properties["name"]; !exists {
		t.Error("Expected input schema to include name field")
	}
}

func TestFieldToSchema_AllTypes(t *testing.T) {
	api := createTestAPIForOpenAPI()

	tests := []struct {
		field    parser.Field
		expected map[string]string
	}{
		{
			field:    parser.Field{Type: parser.FieldTypeID},
			expected: map[string]string{"type": "integer", "format": "int64"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeText},
			expected: map[string]string{"type": "string"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeEmail},
			expected: map[string]string{"type": "string", "format": "email"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeURL},
			expected: map[string]string{"type": "string", "format": "uri"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypePassword},
			expected: map[string]string{"type": "string", "format": "password"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeUUID},
			expected: map[string]string{"type": "string", "format": "uuid"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeNumber},
			expected: map[string]string{"type": "number"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeBoolean},
			expected: map[string]string{"type": "boolean"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeDatetime},
			expected: map[string]string{"type": "string", "format": "date-time"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeDate},
			expected: map[string]string{"type": "string", "format": "date"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeTime},
			expected: map[string]string{"type": "string", "format": "time"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeArray},
			expected: map[string]string{"type": "array"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeFile},
			expected: map[string]string{"type": "string", "format": "binary"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeRelation},
			expected: map[string]string{"type": "integer", "format": "int64"},
		},
		{
			field:    parser.Field{Type: parser.FieldTypeLocation},
			expected: map[string]string{"type": "object"},
		},
	}

	for _, test := range tests {
		schema := api.fieldToSchema(test.field)
		
		if schema.Type != test.expected["type"] {
			t.Errorf("For field type %s, expected type %s, got %s", 
				test.field.Type, test.expected["type"], schema.Type)
		}
		
		if expectedFormat, hasFormat := test.expected["format"]; hasFormat {
			if schema.Format != expectedFormat {
				t.Errorf("For field type %s, expected format %s, got %s", 
					test.field.Type, expectedFormat, schema.Format)
			}
		}
	}
}

func TestFieldToSchema_EnumField(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	field := parser.Field{
		Type:    parser.FieldTypeEnum,
		Options: []string{"option1", "option2", "option3"},
	}
	
	schema := api.fieldToSchema(field)
	
	if schema.Type != "string" {
		t.Errorf("Expected enum type 'string', got: %s", schema.Type)
	}
	
	if len(schema.Enum) != 3 {
		t.Errorf("Expected 3 enum options, got: %d", len(schema.Enum))
	}
	
	expectedOptions := []string{"option1", "option2", "option3"}
	for i, option := range expectedOptions {
		if i >= len(schema.Enum) || schema.Enum[i] != option {
			t.Errorf("Expected enum option %s at index %d", option, i)
		}
	}
}

func TestFieldToSchema_Constraints(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	// Test min/max constraints
	field := parser.Field{
		Type: parser.FieldTypeText,
		Min:  &[]int{5}[0],
		Max:  &[]int{100}[0],
	}
	
	schema := api.fieldToSchema(field)
	
	if schema.MinLength == nil || *schema.MinLength != 5 {
		t.Error("Expected minLength constraint to be set")
	}
	if schema.MaxLength == nil || *schema.MaxLength != 100 {
		t.Error("Expected maxLength constraint to be set")
	}
	
	// Test number constraints
	numberField := parser.Field{
		Type: parser.FieldTypeNumber,
		Min:  &[]int{10}[0],
		Max:  &[]int{200}[0],
	}
	
	numberSchema := api.fieldToSchema(numberField)
	
	if numberSchema.Minimum == nil || *numberSchema.Minimum != 10 {
		t.Error("Expected minimum constraint to be set")
	}
	if numberSchema.Maximum == nil || *numberSchema.Maximum != 200 {
		t.Error("Expected maximum constraint to be set")
	}
}

func TestFieldToSchema_DefaultValue(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	field := parser.Field{
		Type:    parser.FieldTypeText,
		Default: "default value",
	}
	
	schema := api.fieldToSchema(field)
	
	if schema.Default != "default value" {
		t.Errorf("Expected default value 'default value', got: %v", schema.Default)
	}
}

func TestFieldToSchema_LocationField(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	field := parser.Field{Type: parser.FieldTypeLocation}
	schema := api.fieldToSchema(field)
	
	if schema.Type != "object" {
		t.Errorf("Expected location type 'object', got: %s", schema.Type)
	}
	
	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 location properties, got: %d", len(schema.Properties))
	}
	
	latProp, exists := schema.Properties["lat"]
	if !exists || latProp.Type != "number" {
		t.Error("Expected lat property to be number")
	}
	
	lngProp, exists := schema.Properties["lng"]
	if !exists || lngProp.Type != "number" {
		t.Error("Expected lng property to be number")
	}
}

func TestHandleOpenAPI(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	w := httptest.NewRecorder()
	
	handler := api.HandleOpenAPI()
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got: %s", contentType)
	}
	
	var spec OpenAPISpec
	err := json.Unmarshal(w.Body.Bytes(), &spec)
	if err != nil {
		t.Fatalf("Failed to parse OpenAPI spec: %v", err)
	}
	
	if spec.OpenAPI != "3.0.3" {
		t.Errorf("Expected OpenAPI version '3.0.3', got: %s", spec.OpenAPI)
	}
}

func TestHandleSwaggerUI(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	req := httptest.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()
	
	handler := api.HandleSwaggerUI()
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected content type 'text/html; charset=utf-8', got: %s", contentType)
	}
	
	body := w.Body.String()
	
	// Check for essential Swagger UI elements
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Expected HTML document")
	}
	if !strings.Contains(body, "swagger-ui") {
		t.Error("Expected Swagger UI elements")
	}
	if !strings.Contains(body, api.config.App.Name) {
		t.Error("Expected app name in title")
	}
	if !strings.Contains(body, "/api/openapi.json") {
		t.Error("Expected OpenAPI spec URL")
	}
}

func TestOpenAPISpec_CompleteStructure(t *testing.T) {
	api := createTestAPIForOpenAPI()
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	req.Host = "api.example.com"
	
	spec := api.GenerateOpenAPI(req)

	// Test all major sections exist
	if spec.Info.Title == "" {
		t.Error("Expected info section to have title")
	}
	
	if len(spec.Servers) == 0 {
		t.Error("Expected at least one server")
	}
	
	if len(spec.Paths) == 0 {
		t.Error("Expected paths to be defined")
	}
	
	if len(spec.Components.Schemas) == 0 {
		t.Error("Expected component schemas to be defined")
	}

	// Check that both User and Post models have schemas
	if _, exists := spec.Components.Schemas["User"]; !exists {
		t.Error("Expected User schema to exist")
	}
	if _, exists := spec.Components.Schemas["UserInput"]; !exists {
		t.Error("Expected UserInput schema to exist")
	}
	if _, exists := spec.Components.Schemas["Post"]; !exists {
		t.Error("Expected Post schema to exist")
	}
	if _, exists := spec.Components.Schemas["PostInput"]; !exists {
		t.Error("Expected PostInput schema to exist")
	}

	// Test specific operation details
	userPath := spec.Paths["/user"]["get"]
	if len(userPath.Parameters) == 0 {
		t.Error("Expected list operation to have parameters")
	}
	
	// Check for pagination parameters
	hasPageParam := false
	hasPageSizeParam := false
	for _, param := range userPath.Parameters {
		if param.Name == "page" {
			hasPageParam = true
		}
		if param.Name == "page_size" {
			hasPageSizeParam = true
		}
	}
	if !hasPageParam {
		t.Error("Expected page parameter")
	}
	if !hasPageSizeParam {
		t.Error("Expected page_size parameter")
	}
}