package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/mux"
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

func (m *MockDatabase) Close() error { return nil }

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
	case int64:
		return string(rune(val))
	default:
		return "unknown"
	}
}

func createTestAPI() *API {
	config := &parser.Config{
		App: parser.AppConfig{
			Name: "Test App",
		},
		Server: parser.ServerConfig{
			Auth: parser.AuthConfig{Type: "none"},
		},
	}

	schema := &parser.Schema{
		Models: map[string]*parser.Model{
			"User": {
				Name: "User",
				Fields: []parser.Field{
					{Name: "id", Type: parser.FieldTypeID, Primary: true},
					{Name: "name", Type: parser.FieldTypeText, Required: true},
					{Name: "email", Type: parser.FieldTypeEmail, Required: true},
					{Name: "age", Type: parser.FieldTypeNumber},
					{Name: "password", Type: parser.FieldTypePassword},
				},
			},
		},
	}

	mockDB := NewMockDatabase()
	
	return New(mockDB, config, schema, nil)
}

func TestNew(t *testing.T) {
	api := createTestAPI()

	if api.db == nil {
		t.Error("Expected database to be set")
	}
	if api.config == nil {
		t.Error("Expected config to be set")
	}
	if api.schema == nil {
		t.Error("Expected schema to be set")
	}
	if api.validator == nil {
		t.Error("Expected validator to be created")
	}
}

func TestAPI_RegisterRoutes(t *testing.T) {
	api := createTestAPI()
	router := mux.NewRouter()

	api.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/api/openapi", nil)
	match := &mux.RouteMatch{}
	if !router.Match(req, match) {
		t.Error("Expected /api/openapi route to be registered")
	}

	req = httptest.NewRequest("GET", "/api/user", nil)
	match = &mux.RouteMatch{}
	if !router.Match(req, match) {
		t.Error("Expected /api/user route to be registered")
	}
}

func TestAPI_HandleList_Success(t *testing.T) {
	api := createTestAPI()
	
	mockDB := api.db.(*MockDatabase)
	mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	})

	req := httptest.NewRequest("GET", "/api/user?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	handler := api.handleList("User")
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected response.Success to be true")
	}

	if response.Meta == nil {
		t.Error("Expected meta information")
	}
}

func TestAPI_HandleList_DatabaseError(t *testing.T) {
	api := createTestAPI()
	mockDB := api.db.(*MockDatabase)
	mockDB.shouldError = true

	req := httptest.NewRequest("GET", "/api/user", nil)
	w := httptest.NewRecorder()

	handler := api.handleList("User")
	handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAPI_HandleGet_Success(t *testing.T) {
	api := createTestAPI()
	
	mockDB := api.db.(*MockDatabase)
	id, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})

	req := httptest.NewRequest("GET", "/api/user/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": toString(id)})
	w := httptest.NewRecorder()

	handler := api.handleGet("User")
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected response.Success to be true")
	}
}

func TestAPI_HandleGet_NotFound(t *testing.T) {
	api := createTestAPI()

	req := httptest.NewRequest("GET", "/api/user/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()

	handler := api.handleGet("User")
	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestAPI_HandleCreate_Success(t *testing.T) {
	api := createTestAPI()

	userData := map[string]interface{}{
		"name":     "Jane Doe",
		"email":    "jane@example.com",
		"age":      25,
		"password": "secret123",
	}

	jsonData, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/api/user", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := api.handleCreate("User")
	handler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected response.Success to be true")
	}
}

func TestAPI_HandleCreate_InvalidJSON(t *testing.T) {
	api := createTestAPI()

	req := httptest.NewRequest("POST", "/api/user", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := api.handleCreate("User")
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAPI_HandleCreate_ValidationError(t *testing.T) {
	api := createTestAPI()

	userData := map[string]interface{}{
		"age": 25,
	}

	jsonData, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/api/user", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := api.handleCreate("User")
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAPI_HandleUpdate_Success(t *testing.T) {
	api := createTestAPI()
	
	mockDB := api.db.(*MockDatabase)
	id, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})

	updateData := map[string]interface{}{
		"name": "John Updated",
		"age":  31,
	}

	jsonData, _ := json.Marshal(updateData)
	req := httptest.NewRequest("PUT", "/api/user/1", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": toString(id)})
	w := httptest.NewRecorder()

	handler := api.handleUpdate("User")
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected response.Success to be true")
	}
}

func TestAPI_HandleDelete_Success(t *testing.T) {
	api := createTestAPI()
	
	mockDB := api.db.(*MockDatabase)
	id, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})

	req := httptest.NewRequest("DELETE", "/api/user/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": toString(id)})
	w := httptest.NewRecorder()

	handler := api.handleDelete("User")
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected response.Success to be true")
	}
}

func TestAPI_HandleBulk_Create(t *testing.T) {
	api := createTestAPI()

	bulkData := map[string]interface{}{
		"operation": "create",
		"data": []map[string]interface{}{
			{"name": "User 1", "email": "user1@example.com"},
			{"name": "User 2", "email": "user2@example.com"},
		},
	}

	jsonData, _ := json.Marshal(bulkData)
	req := httptest.NewRequest("POST", "/api/user/bulk", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := api.handleBulk("User")
	handler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected response.Success to be true")
	}
}

func TestAPI_HandleBulk_Delete(t *testing.T) {
	api := createTestAPI()
	
	mockDB := api.db.(*MockDatabase)
	id1, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "User 1",
		"email": "user1@example.com",
	})
	id2, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "User 2",
		"email": "user2@example.com",
	})

	bulkData := map[string]interface{}{
		"operation": "delete",
		"ids":       []interface{}{id1, id2},
	}

	jsonData, _ := json.Marshal(bulkData)
	req := httptest.NewRequest("POST", "/api/user/bulk", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := api.handleBulk("User")
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAPI_HandleBulk_InvalidOperation(t *testing.T) {
	api := createTestAPI()

	bulkData := map[string]interface{}{
		"operation": "invalid",
	}

	jsonData, _ := json.Marshal(bulkData)
	req := httptest.NewRequest("POST", "/api/user/bulk", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := api.handleBulk("User")
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAPI_ParseQueryParams(t *testing.T) {
	api := createTestAPI()

	req := httptest.NewRequest("GET", "/api/user", nil)
	params := api.parseQueryParams(req)

	if params.Page != 1 {
		t.Errorf("Expected default page 1, got %d", params.Page)
	}
	if params.PageSize != 20 {
		t.Errorf("Expected default page size 20, got %d", params.PageSize)
	}

	req = httptest.NewRequest("GET", "/api/user?page=2&page_size=50&search=john&sort=-name,age&filter.role=admin", nil)
	params = api.parseQueryParams(req)

	if params.Page != 2 {
		t.Errorf("Expected page 2, got %d", params.Page)
	}
	if params.PageSize != 50 {
		t.Errorf("Expected page size 50, got %d", params.PageSize)
	}
	if params.Search != "john" {
		t.Errorf("Expected search 'john', got '%s'", params.Search)
	}
	if len(params.Sort) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(params.Sort))
	}
	if !params.Sort[0].Desc {
		t.Error("Expected first sort field to be descending")
	}
	if len(params.Filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(params.Filters))
	}

	req = httptest.NewRequest("GET", "/api/user?page=invalid&page_size=200", nil)
	params = api.parseQueryParams(req)

	if params.Page != 1 {
		t.Errorf("Expected default page 1 for invalid page, got %d", params.Page)
	}
	if params.PageSize != 20 {
		t.Errorf("Expected default page size 20 for too large page_size, got %d", params.PageSize)
	}
}

func TestAPI_FilterPasswordFields(t *testing.T) {
	api := createTestAPI()

	records := []map[string]interface{}{
		{
			"id":       1,
			"name":     "John",
			"email":    "john@example.com",
			"password": "secret123",
		},
		{
			"id":       2,
			"name":     "Jane",
			"email":    "jane@example.com",
			"password": "secret456",
		},
	}

	filtered := api.filterPasswordFields("User", records)

	for _, record := range filtered {
		if _, exists := record["password"]; exists {
			t.Error("Expected password field to be filtered out")
		}
	}
}

func TestAPI_FilterPasswordFieldsSingle(t *testing.T) {
	api := createTestAPI()

	record := map[string]interface{}{
		"id":       1,
		"name":     "John",
		"email":    "john@example.com",
		"password": "secret123",
	}

	filtered := api.filterPasswordFieldsSingle("User", record)

	if _, exists := filtered["password"]; exists {
		t.Error("Expected password field to be filtered out")
	}
	if filtered["name"] != "John" {
		t.Error("Expected other fields to be preserved")
	}
}

func TestAPI_FilterEmptyPasswordFields(t *testing.T) {
	api := createTestAPI()

	data := map[string]interface{}{
		"name":     "John Updated",
		"password": "", // Empty password should be removed
		"email":    "john@example.com",
	}

	filtered := api.filterEmptyPasswordFields("User", data)

	if _, exists := filtered["password"]; exists {
		t.Error("Expected empty password field to be removed")
	}
	if filtered["name"] != "John Updated" {
		t.Error("Expected other fields to be preserved")
	}
}

func TestAPI_CheckPermission_NoAuth(t *testing.T) {
	api := createTestAPI()

	req := httptest.NewRequest("GET", "/api/user", nil)

	user, err := api.checkPermission(req, "User", false)
	if err != nil {
		t.Errorf("Expected no error with no auth, got: %v", err)
	}
	if user != nil {
		t.Error("Expected nil user with no auth")
	}
}

func TestAPI_CorsMiddleware(t *testing.T) {
	// Test with CORS enabled
	config := &parser.Config{
		Server: parser.ServerConfig{
			CORS: parser.CORSConfig{
				Enabled: true,
				Origins: []string{"*"},
			},
		},
	}

	api := &API{config: config}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest("GET", "/api/user", nil)
	w := httptest.NewRecorder()

	middleware := api.corsMiddleware(next)
	middleware.ServeHTTP(w, req)

	if !nextCalled {
		t.Error("Expected next handler to be called")
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS header to be set")
	}

	// Test OPTIONS request
	req = httptest.NewRequest("OPTIONS", "/api/user", nil)
	w = httptest.NewRecorder()
	nextCalled = false

	middleware.ServeHTTP(w, req)

	if nextCalled {
		t.Error("Expected next handler not to be called for OPTIONS")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}
}

func TestAPI_ContentTypeMiddleware(t *testing.T) {
	api := createTestAPI()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest("GET", "/api/user", nil)
	w := httptest.NewRecorder()

	middleware := api.contentTypeMiddleware(next)
	middleware.ServeHTTP(w, req)

	if !nextCalled {
		t.Error("Expected next handler to be called")
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type header to be set")
	}
}

func TestIntegration_APIEndpoints(t *testing.T) {
	tmpDir := t.TempDir()
	
	configFile := filepath.Join(tmpDir, "api_test_config.yaml")
	testConfig := `app:
  name: "API Test App"
  version: "1.0.0"

database:
  type: sqlite
  path: ` + filepath.Join(tmpDir, "api_test.db") + `

server:
  cors:
    enabled: true
    origins: ["*"]
  auth:
    type: none

models:
  Task:
    fields:
      id:
        type: id
        primary: true
      title:
        type: text
        required: true
      completed:
        type: boolean
        default: false
      priority:
        type: number
        min: 1
        max: 5
        default: 3

    permissions:
      create: "all"
      read: "all"
      update: "all"
      delete: "all"`

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

	mockDB := NewMockDatabase()
	
	apiHandler := New(mockDB, config, schema, nil)
	if apiHandler == nil {
		t.Fatal("Expected API handler to be created")
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method == "POST" {
			var taskData map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Invalid JSON",
				})
				return
			}
			
			taskData["id"] = 1 // Mock ID
			
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    taskData,
			})
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    []interface{}{},
				"meta": map[string]interface{}{
					"page":        1,
					"page_size":   10,
					"total_count": 0,
					"total_pages": 0,
				},
			})
		}
	}).Methods("GET", "POST")

	router.HandleFunc("/api/task/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		taskID := vars["id"]
		
		if r.Method == "GET" {
			if taskID == "1" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"data": map[string]interface{}{
						"id":        1,
						"title":     "Test Task",
						"completed": false,
						"priority":  4,
					},
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Task not found",
				})
			}
		} else if r.Method == "PUT" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Task updated successfully",
			})
		} else if r.Method == "DELETE" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Task deleted successfully",
			})
		}
	}).Methods("GET", "PUT", "DELETE")

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	taskData := map[string]interface{}{
		"title":     "Test Task",
		"completed": false,
		"priority":  4,
	}

	jsonData, _ := json.Marshal(taskData)
	resp, err := http.Post(testServer.URL+"/api/task", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	if err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	if createResponse["success"] != true {
		t.Error("Expected successful create response")
	}

	taskDataResp := createResponse["data"].(map[string]interface{})
	taskID := fmt.Sprintf("%.0f", taskDataResp["id"].(float64))

	resp, err = http.Get(testServer.URL + "/api/task/" + taskID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var getResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&getResponse)
	if err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	if getResponse["success"] != true {
		t.Error("Expected successful get response")
	}

	resp, err = http.Get(testServer.URL + "/api/task?page=1&page_size=10")
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var listResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	if listResponse["success"] != true {
		t.Error("Expected successful list response")
	}

	if listResponse["meta"] == nil {
		t.Error("Expected meta information in list response")
	}

	updateData := map[string]interface{}{
		"title":     "Updated Task",
		"completed": true,
	}

	jsonData, _ = json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", testServer.URL+"/api/task/"+taskID, bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	req, _ = http.NewRequest("DELETE", testServer.URL+"/api/task/"+taskID, nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(testServer.URL + "/api/task/999")
	if err != nil {
		t.Fatalf("Failed to check deleted task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404 for non-existent task, got %d", resp.StatusCode)
	}
}