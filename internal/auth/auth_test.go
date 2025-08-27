package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yamlforge/yamlforge/internal/parser"
)

func createTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return db
}

func TestNew_JWT(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "admin",
				Password: "admin123",
				Email:    "admin@test.com",
				Role:     "admin",
				Active:   true,
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if authManager == nil {
		t.Fatal("Expected AuthManager to be created")
	}
	if authManager.config != config {
		t.Error("Expected config to be stored")
	}
	if string(authManager.jwtKey) != "test-secret" {
		t.Error("Expected JWT key to be set")
	}
}

func TestNew_UnsupportedType(t *testing.T) {
	config := &parser.AuthConfig{
		Type: "oauth",
	}

	db := createTestDB(t)
	defer db.Close()

	_, err := New(config, db)
	if err == nil {
		t.Fatal("Expected error for unsupported auth type")
	}

	expected := "unsupported auth type: oauth"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got: %s", expected, err.Error())
	}
}

func TestNew_AutoGenerateSecret(t *testing.T) {
	config := &parser.AuthConfig{
		Type: "jwt",
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Secret == "" {
		t.Error("Expected secret to be auto-generated")
	}
	if len(authManager.jwtKey) == 0 {
		t.Error("Expected JWT key to be set")
	}
}

func TestNew_ParseExpiry(t *testing.T) {
	config := &parser.AuthConfig{
		Type:    "jwt",
		Secret:  "test-secret",
		Expires: "2h",
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if authManager.expires != 2*time.Hour {
		t.Errorf("Expected 2 hour expiry, got: %v", authManager.expires)
	}
}

func TestNew_InvalidExpiry(t *testing.T) {
	config := &parser.AuthConfig{
		Type:    "jwt",
		Secret:  "test-secret",
		Expires: "invalid",
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if authManager.expires != 24*time.Hour {
		t.Errorf("Expected 24 hour default expiry, got: %v", authManager.expires)
	}
}

func TestNew_UserPermissions(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "user1",
				Password: "pass123",
				Email:    "user1@test.com",
				Permissions: map[string]parser.EntityPermission{
					"posts": {Read: true, Write: false},
					"users": {Read: true, Write: true},
				},
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	userPerms, exists := authManager.permissions["user1"]
	if !exists {
		t.Fatal("Expected user permissions to be stored")
	}

	if !userPerms["posts"].Read {
		t.Error("Expected read permission for posts")
	}
	if userPerms["posts"].Write {
		t.Error("Expected no write permission for posts")
	}
	if !userPerms["users"].Write {
		t.Error("Expected write permission for users")
	}
}

func TestNew_InvalidUserPermissions(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "user1",
				Password: "pass123",
				Email:    "user1@test.com",
				Permissions: map[string]parser.EntityPermission{
					"posts": {Read: false, Write: false},
				},
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	_, err := New(config, db)
	if err == nil {
		t.Fatal("Expected error for user with no permissions")
	}

	if !containsString(err.Error(), "has no permissions") {
		t.Errorf("Expected error about no permissions, got: %s", err.Error())
	}
}

func TestInitAuthTables(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "admin",
				Password: "admin123",
				Email:    "admin@test.com",
				Role:     "admin",
				Active:   true,
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	authManager := &AuthManager{
		config: config,
		db:     db,
	}

	err := authManager.initAuthTables()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='auth_users'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	if count != 1 {
		t.Error("Expected auth_users table to be created")
	}

	// Check that user was inserted
	err = db.QueryRow("SELECT COUNT(*) FROM auth_users").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}
	if count != 1 {
		t.Error("Expected 1 user to be inserted")
	}
}

func TestInitAuthTables_DefaultAdmin(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
	}

	db := createTestDB(t)
	defer db.Close()

	authManager := &AuthManager{
		config: config,
		db:     db,
	}

	err := authManager.initAuthTables()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var username string
	err = db.QueryRow("SELECT username FROM auth_users WHERE role='admin'").Scan(&username)
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}
	if username != "admin" {
		t.Errorf("Expected default admin username, got: %s", username)
	}
}

func TestAuthenticate_Success(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "testuser",
				Password: "testpass",
				Email:    "test@example.com",
				Role:     "user",
				Active:   true,
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	user, err := authManager.Authenticate("testuser", "testpass")
	if err != nil {
		t.Fatalf("Expected successful authentication, got: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got: %s", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got: %s", user.Email)
	}
	if user.Role != "user" {
		t.Errorf("Expected role 'user', got: %s", user.Role)
	}
}

func TestAuthenticate_ByEmail(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "testuser",
				Password: "testpass",
				Email:    "test@example.com",
				Role:     "user",
				Active:   true,
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	user, err := authManager.Authenticate("test@example.com", "testpass")
	if err != nil {
		t.Fatalf("Expected successful authentication by email, got: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got: %s", user.Username)
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "testuser",
				Password: "testpass",
				Email:    "test@example.com",
				Active:   true,
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	_, err = authManager.Authenticate("testuser", "wrongpass")
	if err == nil {
		t.Fatal("Expected error for wrong password")
	}
	if err.Error() != "invalid credentials" {
		t.Errorf("Expected 'invalid credentials', got: %s", err.Error())
	}

	_, err = authManager.Authenticate("wronguser", "testpass")
	if err == nil {
		t.Fatal("Expected error for wrong username")
	}
}

func TestGenerateToken(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	user := &User{
		ID:       123,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	token, err := authManager.GenerateToken(user)
	if err != nil {
		t.Fatalf("Expected no error generating token, got: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestValidateToken(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	user := &User{
		ID:       123,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	token, err := authManager.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := authManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected valid token, got: %v", err)
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got: %s", claims.Username)
	}
	if claims.UserID != 123 {
		t.Errorf("Expected user ID 123, got: %d", claims.UserID)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	_, err = authManager.ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}

	otherAuth, _ := New(&parser.AuthConfig{Type: "jwt", Secret: "other-secret"}, createTestDB(t))
	user := &User{ID: 123, Username: "testuser"}
	tokenWithWrongSecret, _ := otherAuth.GenerateToken(user)

	_, err = authManager.ValidateToken(tokenWithWrongSecret)
	if err == nil {
		t.Fatal("Expected error for token with wrong secret")
	}
}

func TestGetTokenFromRequest_Cookie(t *testing.T) {
	config := &parser.AuthConfig{Type: "jwt", Secret: "test-secret"}
	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test-token"})

	token, err := authManager.GetTokenFromRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if token != "test-token" {
		t.Errorf("Expected 'test-token', got: %s", token)
	}
}

func TestGetTokenFromRequest_Authorization(t *testing.T) {
	config := &parser.AuthConfig{Type: "jwt", Secret: "test-secret"}
	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	token, err := authManager.GetTokenFromRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if token != "test-token" {
		t.Errorf("Expected 'test-token', got: %s", token)
	}
}

func TestGetTokenFromRequest_NoToken(t *testing.T) {
	config := &parser.AuthConfig{Type: "jwt", Secret: "test-secret"}
	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)

	_, err = authManager.GetTokenFromRequest(req)
	if err == nil {
		t.Fatal("Expected error for request with no token")
	}
	if err.Error() != "no token found" {
		t.Errorf("Expected 'no token found', got: %s", err.Error())
	}
}

func TestSetAuthCookie(t *testing.T) {
	config := &parser.AuthConfig{Type: "jwt", Secret: "test-secret"}
	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	w := httptest.NewRecorder()
	authManager.SetAuthCookie(w, "test-token")

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != "auth_token" {
		t.Errorf("Expected cookie name 'auth_token', got: %s", cookie.Name)
	}
	if cookie.Value != "test-token" {
		t.Errorf("Expected cookie value 'test-token', got: %s", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Error("Expected cookie to be HttpOnly")
	}
}

func TestClearAuthCookie(t *testing.T) {
	config := &parser.AuthConfig{Type: "jwt", Secret: "test-secret"}
	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	w := httptest.NewRecorder()
	authManager.ClearAuthCookie(w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != "auth_token" {
		t.Errorf("Expected cookie name 'auth_token', got: %s", cookie.Name)
	}
	if cookie.Value != "" {
		t.Errorf("Expected empty cookie value, got: %s", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("Expected MaxAge -1, got: %d", cookie.MaxAge)
	}
}

func TestCreateUser(t *testing.T) {
	config := &parser.AuthConfig{Type: "jwt", Secret: "test-secret"}
	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	err = authManager.CreateUser("newuser", "new@example.com", "password123", "user")
	if err != nil {
		t.Fatalf("Expected no error creating user, got: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM auth_users WHERE username=?", "newuser").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check user creation: %v", err)
	}
	if count != 1 {
		t.Error("Expected user to be created in database")
	}
}

func TestGetUserByID(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "testuser",
				Password: "testpass",
				Email:    "test@example.com",
				Role:     "user",
				Active:   true,
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	var userID int64
	err = db.QueryRow("SELECT id FROM auth_users WHERE username=?", "testuser").Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	user, err := authManager.GetUserByID(userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got: %s", user.Username)
	}
	if user.ID != userID {
		t.Errorf("Expected user ID %d, got: %d", userID, user.ID)
	}
}

func TestCheckPermission(t *testing.T) {
	config := &parser.AuthConfig{
		Type:   "jwt",
		Secret: "test-secret",
		Users: []parser.UserConfig{
			{
				Username: "admin",
				Password: "admin123",
				Email:    "admin@example.com",
				Role:     "admin",
				Active:   true,
			},
			{
				Username: "user",
				Password: "user123",
				Email:    "user@example.com",
				Role:     "user",
				Active:   true,
				Permissions: map[string]parser.EntityPermission{
					"posts": {Read: true, Write: false},
					"users": {Read: true, Write: true},
				},
			},
		},
	}

	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	hasPermission := authManager.CheckPermission("admin", "posts", true)
	if !hasPermission {
		t.Error("Expected admin to have write permission for posts")
	}

	hasPermission = authManager.CheckPermission("user", "posts", false)
	if !hasPermission {
		t.Error("Expected user to have read permission for posts")
	}

	hasPermission = authManager.CheckPermission("user", "posts", true)
	if hasPermission {
		t.Error("Expected user to not have write permission for posts")
	}

	hasPermission = authManager.CheckPermission("user", "users", true)
	if !hasPermission {
		t.Error("Expected user to have write permission for users")
	}

	hasPermission = authManager.CheckPermission("nonexistent", "posts", false)
	if hasPermission {
		t.Error("Expected non-existent user to not have permissions")
	}
}

func TestIsEnabled(t *testing.T) {
	config := &parser.AuthConfig{Type: "jwt"}
	db := createTestDB(t)
	defer db.Close()

	authManager, err := New(config, db)
	if err != nil {
		t.Fatalf("Failed to create auth manager: %v", err)
	}

	if !authManager.IsEnabled() {
		t.Error("Expected auth to be enabled for JWT type")
	}

	authManager.config.Type = "none"
	if authManager.IsEnabled() {
		t.Error("Expected auth to be disabled for none type")
	}
}

func TestHashPassword(t *testing.T) {
	password := "test123"
	hash := hashPassword(password)

	if hash == "" {
		t.Error("Expected non-empty hash")
	}
	if hash == password {
		t.Error("Expected hash to be different from password")
	}

	hash2 := hashPassword(password)
	if hash != hash2 {
		t.Error("Expected same password to generate same hash")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "test123"
	hash := hashPassword(password)

	if !verifyPassword(password, hash) {
		t.Error("Expected password verification to succeed")
	}

	if verifyPassword("wrong", hash) {
		t.Error("Expected password verification to fail for wrong password")
	}
}

func TestGenerateRandomSecret(t *testing.T) {
	secret1 := generateRandomSecret()
	secret2 := generateRandomSecret()

	if secret1 == "" {
		t.Error("Expected non-empty secret")
	}
	if secret1 == secret2 {
		t.Error("Expected different secrets to be generated")
	}
	if len(secret1) != 64 { // 32 bytes * 2 for hex encoding
		t.Errorf("Expected 64 character secret, got %d", len(secret1))
	}
}

func TestIntegration_AuthenticationFlow(t *testing.T) {
	
	tmpDir := t.TempDir()
	
	configFile := filepath.Join(tmpDir, "auth_test.yaml")
	testConfig := `app:
  name: "Auth Test App"

database:
  type: sqlite
  path: ` + filepath.Join(tmpDir, "auth_test.db") + `

server:
  auth:
    type: jwt
    secret: "test-secret"
    expires: "1h"
    users:
      - username: testuser
        password: testpass
        email: test@example.com
        role: admin
        active: true

models:
  User:
    fields:
      id:
        type: id
        primary: true
      name:
        type: text`

	err := os.WriteFile(configFile, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	config, err := parser.ParseConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if config.Server.Auth.Type != "jwt" {
		t.Errorf("Expected auth type 'jwt', got: %s", config.Server.Auth.Type)
	}

	if len(config.Server.Auth.Users) != 1 {
		t.Errorf("Expected 1 user, got: %d", len(config.Server.Auth.Users))
	}

	user := config.Server.Auth.Users[0]
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got: %s", user.Username)
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0
}