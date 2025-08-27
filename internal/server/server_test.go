package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yamlforge/yamlforge/internal/parser"
)

type MockDatabase struct {
	data        map[string]map[string]interface{}
	nextID      int64
	shouldError bool
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		data:   make(map[string]map[string]interface{}),
		nextID: 1,
	}
}

func (m *MockDatabase) Connect() error {
	if m.shouldError {
		return parser.ValidationError{Message: "connection failed"}
	}
	return nil
}

func (m *MockDatabase) Close() error {
	return nil
}

func (m *MockDatabase) CreateSchema(schema *parser.Schema) error {
	if m.shouldError {
		return parser.ValidationError{Message: "schema creation failed"}
	}
	return nil
}

func (m *MockDatabase) Query(model string, params parser.QueryParams) ([]map[string]interface{}, error) {
	if m.shouldError {
		return nil, parser.ValidationError{Message: "query failed"}
	}
	
	var results []map[string]interface{}
	for _, record := range m.data {
		if record["_model"] == model {
			filtered := make(map[string]interface{})
			for k, v := range record {
				if !strings.HasPrefix(k, "_") {
					filtered[k] = v
				}
			}
			results = append(results, filtered)
		}
	}
	
	return results, nil
}

func (m *MockDatabase) Get(model string, id interface{}) (map[string]interface{}, error) {
	if m.shouldError {
		return nil, parser.ValidationError{Message: "get failed"}
	}
	
	key := model + "_" + toString(id)
	if record, exists := m.data[key]; exists {
		filtered := make(map[string]interface{})
		for k, v := range record {
			if !strings.HasPrefix(k, "_") {
				filtered[k] = v
			}
		}
		return filtered, nil
	}
	
	return nil, sql.ErrNoRows
}

func (m *MockDatabase) Create(model string, data map[string]interface{}) (interface{}, error) {
	if m.shouldError {
		return nil, parser.ValidationError{Message: "create failed"}
	}
	
	id := m.nextID
	m.nextID++
	
	record := make(map[string]interface{})
	for k, v := range data {
		record[k] = v
	}
	record["id"] = id
	record["_model"] = model
	
	key := model + "_" + toString(id)
	m.data[key] = record
	
	return id, nil
}

func (m *MockDatabase) Update(model string, id interface{}, data map[string]interface{}) error {
	if m.shouldError {
		return parser.ValidationError{Message: "update failed"}
	}
	
	key := model + "_" + toString(id)
	if record, exists := m.data[key]; exists {
		for k, v := range data {
			record[k] = v
		}
		return nil
	}
	
	return parser.ValidationError{Message: "record not found"}
}

func (m *MockDatabase) Delete(model string, id interface{}) error {
	if m.shouldError {
		return parser.ValidationError{Message: "delete failed"}
	}
	
	key := model + "_" + toString(id)
	if _, exists := m.data[key]; exists {
		delete(m.data, key)
	}
	
	return nil
}

func (m *MockDatabase) Count(model string, filters []parser.Filter) (int64, error) {
	if m.shouldError {
		return 0, parser.ValidationError{Message: "count failed"}
	}
	
	count := int64(0)
	for _, record := range m.data {
		if record["_model"] == model {
			count++
		}
	}
	
	return count, nil
}

func (m *MockDatabase) BeginTx() (*sql.Tx, error) {
	return nil, nil
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val))
	case int64:
		return string(rune(val))
	case float64:
		return string(rune(int(val)))
	default:
		return "unknown"
	}
}

func createTestConfig() *parser.Config {
	return &parser.Config{
		App: parser.AppConfig{
			Name:    "Test App",
			Version: "1.0.0",
		},
		Database: parser.DatabaseConfig{
			Type: "sqlite",
			Path: ":memory:",
		},
		Server: parser.ServerConfig{
			Port: 8080,
			Host: "localhost",
			CORS: parser.CORSConfig{
				Enabled: true,
				Origins: []string{"*"},
			},
			Auth: parser.AuthConfig{
				Type: "none",
			},
		},
		UI: parser.UIConfig{
			Theme:  "light",
			Title:  "Test App",
			Layout: "sidebar",
		},
		Models: map[string]parser.ModelConfig{
			"User": {
				Fields: map[string]parser.FieldConfig{
					"id": {
						Type:    "id",
						Primary: true,
					},
					"name": {
						Type:     "text",
						Required: true,
						Min:      3,
						Max:      50,
					},
					"email": {
						Type:     "email",
						Required: true,
						Unique:   true,
					},
				},
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
					{Name: "name", Type: parser.FieldTypeText, Required: true, Min: &[]int{3}[0], Max: &[]int{50}[0]},
					{Name: "email", Type: parser.FieldTypeEmail, Required: true, Unique: true},
				},
				UI: parser.UIModel{
					List: parser.UIList{
						Columns:    []string{"name", "email"},
						Sortable:   []string{"name", "email"},
						Searchable: []string{"name", "email"},
					},
					Form: parser.UIForm{
						Fields: []string{"name", "email"},
					},
				},
			},
		},
	}
}

func TestNew(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	if server == nil {
		t.Fatal("Expected server to be created")
	}
	
	if server.config != config {
		t.Error("Expected config to be set")
	}
	
	if server.router == nil {
		t.Error("Expected router to be initialized")
	}
}

func TestServer_Initialize_Success(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	
	if server.config == nil {
		t.Error("Expected config to be set")
	}
	
	if server.router == nil {
		t.Error("Expected router to be initialized")
	}
}

func TestServer_HandleHome(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.setupRoutes()
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	server.handleHome(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "Test App") {
		t.Error("Expected response to contain app name")
	}
}

func TestServer_HandleModelList(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.setupRoutes()
	
	req := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()
	
	handler := server.handleModelList("User")
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "User") {
		t.Error("Expected response to contain model name")
	}
}

func TestServer_HandleModelList_NotFound(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.setupRoutes()
	
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	
	handler := server.handleModelList("NonExistent")
	handler(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestServer_HandleModelNew(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.setupRoutes()
	
	req := httptest.NewRequest("GET", "/user/new", nil)
	w := httptest.NewRecorder()
	
	handler := server.handleModelNew("User")
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "New User") {
		t.Error("Expected response to contain 'New User'")
	}
}

func TestServer_HandleModelView(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	
	mockDB := server.db.(*MockDatabase)
	_, _ = mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	
	req := httptest.NewRequest("GET", "/user/1", nil)
	req = req.WithContext(req.Context())
	w := httptest.NewRecorder()
	
	handler := server.handleModelView("User")
	
	mockDB.shouldError = true
	handler(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestServer_LoggingMiddleware(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})
	
	wrappedHandler := server.loggingMiddleware(testHandler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got %s", w.Body.String())
	}
}

func TestServer_DebugMiddleware(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	wrappedHandler := server.debugMiddleware(testHandler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("X-Debug-Mode") != "true" {
		t.Error("Expected X-Debug-Mode header to be set")
	}
}

func TestServer_ParseQueryParams(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	req := httptest.NewRequest("GET", "/test?page=2&page_size=50&sort=name,-age&search=test&filter.status=active", nil)
	
	params := server.parseQueryParams(req)
	
	if params.Page != 2 {
		t.Errorf("Expected page 2, got %d", params.Page)
	}
	
	if params.PageSize != 50 {
		t.Errorf("Expected page size 50, got %d", params.PageSize)
	}
	
	if params.Search != "test" {
		t.Errorf("Expected search 'test', got %s", params.Search)
	}
	
	if len(params.Sort) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(params.Sort))
	}
	
	if len(params.Filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(params.Filters))
	}
	
	if params.Filters[0].Field != "status" {
		t.Errorf("Expected filter field 'status', got %s", params.Filters[0].Field)
	}
}

func TestServer_ParseQueryParams_Defaults(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	req := httptest.NewRequest("GET", "/test", nil)
	
	params := server.parseQueryParams(req)
	
	if params.Page != 1 {
		t.Errorf("Expected default page 1, got %d", params.Page)
	}
	
	if params.PageSize != 20 {
		t.Errorf("Expected default page size 20, got %d", params.PageSize)
	}
}

func TestServer_FilterEmptyPasswordFields(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = &parser.Schema{
		Models: map[string]*parser.Model{
			"User": {
				Fields: []parser.Field{
					{Name: "password", Type: parser.FieldTypePassword},
					{Name: "name", Type: parser.FieldTypeText},
				},
			},
		},
	}
	
	data := map[string]interface{}{
		"name":     "John",
		"password": "",
		"email":    "john@example.com",
	}
	
	filtered := server.filterEmptyPasswordFields("User", data)
	
	if _, exists := filtered["password"]; exists {
		t.Error("Expected empty password field to be removed")
	}
	
	if filtered["name"] != "John" {
		t.Error("Expected non-password fields to be preserved")
	}
}

func TestServer_SendJSON(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}
	
	server.sendJSON(w, http.StatusOK, data)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "test") {
		t.Error("Expected response body to contain the data")
	}
}

func TestServer_ExtractRecordAsJSON(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	data := struct {
		Title     string
		Config    *parser.Config
		ModelName string
		Model     *parser.Model
		Record    map[string]any
	}{
		Record: map[string]any{
			"id":   1,
			"name": "Test",
		},
	}
	
	result := server.extractRecordAsJSON(data)
	
	if result == "null" {
		t.Error("Expected JSON string, got null")
	}
	
	if !strings.Contains(result, "Test") {
		t.Error("Expected JSON to contain record data")
	}
}

func TestServer_ExtractRecordAsJSON_Nil(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	data := struct {
		Title     string
		Config    *parser.Config
		ModelName string
		Model     *parser.Model
		Record    map[string]any
	}{
		Record: nil,
	}
	
	result := server.extractRecordAsJSON(data)
	
	if result != "null" {
		t.Errorf("Expected 'null', got %s", result)
	}
}

func TestServer_HandleSwaggerUI(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	
	req := httptest.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()
	
	server.handleSwaggerUI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Error("Expected Content-Type to be text/html")
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "swagger") {
		t.Error("Expected response to contain Swagger UI")
	}
	
	if !strings.Contains(body, config.App.Name) {
		t.Error("Expected response to contain app name in title")
	}
}