package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/yamlforge/yamlforge/internal/parser"
	"github.com/yamlforge/yamlforge/internal/validation"
)

func TestServer_HandleAPIList(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	server.validator = validation.New(server.schema)
	
	req := httptest.NewRequest("GET", "/api/user", nil)
	w := httptest.NewRecorder()
	
	handler := server.handleAPIList("User")
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestServer_HandleAPIGet(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	
	mockDB := server.db.(*MockDatabase)
	id, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	
	req := httptest.NewRequest("GET", "/api/user/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": toString(id)})
	w := httptest.NewRecorder()
	
	handler := server.handleAPIGet("User")
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestServer_HandleAPIGet_NotFound(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	
	req := httptest.NewRequest("GET", "/api/user/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()
	
	handler := server.handleAPIGet("User")
	handler(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestServer_HandleAPICreate(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	server.validator = validation.New(server.schema)
	
	userData := map[string]interface{}{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}
	
	jsonData, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/api/user", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler := server.handleAPICreate("User")
	handler(w, req)
	
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
	
	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestServer_HandleAPICreate_InvalidJSON(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	server.validator = validation.New(server.schema)
	
	req := httptest.NewRequest("POST", "/api/user", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler := server.handleAPICreate("User")
	handler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestServer_HandleAPICreate_ValidationError(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	server.validator = validation.New(server.schema)
	
	userData := map[string]interface{}{
		"email": "jane@example.com",
	}
	
	jsonData, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/api/user", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler := server.handleAPICreate("User")
	handler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestServer_HandleAPIUpdate(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	server.validator = validation.New(server.schema)
	
	mockDB := server.db.(*MockDatabase)
	id, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	
	updateData := map[string]interface{}{
		"name": "John Smith",
	}
	
	jsonData, _ := json.Marshal(updateData)
	req := httptest.NewRequest("PUT", "/api/user/1", bytes.NewReader(jsonData))
	req = mux.SetURLVars(req, map[string]string{"id": toString(id)})
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler := server.handleAPIUpdate("User")
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestServer_HandleAPIUpdate_InvalidJSON(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	server.validator = validation.New(server.schema)
	
	req := httptest.NewRequest("PUT", "/api/user/1", strings.NewReader("invalid json"))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler := server.handleAPIUpdate("User")
	handler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestServer_HandleAPIDelete(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	
	mockDB := server.db.(*MockDatabase)
	id, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	
	req := httptest.NewRequest("DELETE", "/api/user/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": toString(id)})
	w := httptest.NewRecorder()
	
	handler := server.handleAPIDelete("User")
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response parser.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestServer_HandleOpenAPI(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	
	req := httptest.NewRequest("GET", "/api/openapi", nil)
	w := httptest.NewRecorder()
	
	server.handleOpenAPI(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "openapi") {
		t.Error("Expected response to contain OpenAPI specification")
	}
}

func TestServer_HandleModelEdit(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	
	mockDB := server.db.(*MockDatabase)
	id, _ := mockDB.Create("User", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	
	req := httptest.NewRequest("GET", "/user/1/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"id": toString(id)})
	w := httptest.NewRecorder()
	
	handler := server.handleModelEdit("User")
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "Edit User") {
		t.Error("Expected response to contain 'Edit User'")
	}
}

func TestServer_HandleModelEdit_NotFound(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	
	req := httptest.NewRequest("GET", "/user/999/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()
	
	handler := server.handleModelEdit("User")
	handler(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestServer_SetupRoutes(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	
	server.setupRoutes()
	
	if server.router == nil {
		t.Error("Expected router to be set up")
	}
}

func TestServer_HandleAPIList_DatabaseError(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	server.validator = validation.New(server.schema)
	
	mockDB := server.db.(*MockDatabase)
	mockDB.shouldError = true
	
	req := httptest.NewRequest("GET", "/api/user", nil)
	w := httptest.NewRecorder()
	
	handler := server.handleAPIList("User")
	handler(w, req)
	
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestServer_HandleAPICreate_DatabaseError(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	server.db = NewMockDatabase()
	server.validator = validation.New(server.schema)
	
	mockDB := server.db.(*MockDatabase)
	mockDB.shouldError = true
	
	userData := map[string]interface{}{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}
	
	jsonData, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/api/user", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler := server.handleAPICreate("User")
	handler(w, req)
	
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestServer_Render_ViewTemplate(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	
	data := struct {
		Title     string
		Config    *parser.Config
		ModelName string
		Model     *parser.Model
		Record    map[string]any
	}{
		Title:     "Test View",
		Config:    config,
		ModelName: "User",
		Model:     server.schema.Models["User"],
		Record:    map[string]any{"id": "1", "name": "Test User"},
	}
	
	w := httptest.NewRecorder()
	server.render(w, "view", data)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Error("Expected Content-Type to be text/html")
	}
}

func TestServer_Render_UnknownTemplate(t *testing.T) {
	config := createTestConfig()
	server := New(config)
	server.schema = createTestSchema()
	
	w := httptest.NewRecorder()
	server.render(w, "unknown", nil)
	
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}