package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yamlforge/yamlforge/internal/parser"
)

type AuthManager struct {
	config      *parser.AuthConfig
	db          *sql.DB
	jwtKey      []byte
	expires     time.Duration
	permissions map[string]map[string]parser.EntityPermission // username -> model -> permissions
}

type User struct {
	ID          int64                                        `json:"id"`
	Username    string                                       `json:"username"`
	Email       string                                       `json:"email"`
	Password    string                                       `json:"-"`
	Role        string                                       `json:"role"`
	Active      bool                                         `json:"active"`
	CreatedAt   time.Time                                    `json:"created_at"`
	Permissions map[string]parser.EntityPermission          `json:"permissions,omitempty"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func New(config *parser.AuthConfig, db *sql.DB) (*AuthManager, error) {
	if config.Type != "jwt" {
		return nil, fmt.Errorf("unsupported auth type: %s", config.Type)
	}

	if config.Secret == "" {
		config.Secret = generateRandomSecret()
	}

	expires := 24 * time.Hour
	if config.Expires != "" {
		d, err := time.ParseDuration(config.Expires)
		if err == nil {
			expires = d
		}
	}

	am := &AuthManager{
		config:      config,
		db:          db,
		jwtKey:      []byte(config.Secret),
		expires:     expires,
		permissions: make(map[string]map[string]parser.EntityPermission),
	}

	for _, user := range config.Users {
		if user.Permissions != nil {
			for entity, perm := range user.Permissions {
				if !perm.Read && !perm.Write {
					return nil, fmt.Errorf("user %s has no permissions (neither read nor write) for entity %s", user.Username, entity)
				}
			}
			am.permissions[user.Username] = user.Permissions
		}
	}

	if err := am.initAuthTables(); err != nil {
		return nil, err
	}

	return am, nil
}

func (am *AuthManager) initAuthTables() error {
	createUserTable := `
	CREATE TABLE IF NOT EXISTS auth_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'user',
		active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := am.db.Exec(createUserTable); err != nil {
		return err
	}

	var count int
	err := am.db.QueryRow("SELECT COUNT(*) FROM auth_users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		if len(am.config.Users) > 0 {
			for _, user := range am.config.Users {
				hashedPassword := hashPassword(user.Password)
				email := user.Email
				if email == "" {
					email = user.Username + "@example.com"
				}
				role := user.Role
				if role == "" {
					role = "user"
				}
				active := user.Active
				if !user.Active && user.Username != "" {
					active = true
				}
				
				_, err = am.db.Exec(
					"INSERT INTO auth_users (username, email, password, role, active) VALUES (?, ?, ?, ?, ?)",
					user.Username, email, hashedPassword, role, active,
				)
				if err != nil {
					return fmt.Errorf("failed to create user %s: %w", user.Username, err)
				}
			}
		} else {
			hashedPassword := hashPassword("admin123")
			_, err = am.db.Exec(
				"INSERT INTO auth_users (username, email, password, role) VALUES (?, ?, ?, ?)",
				"admin", "admin@example.com", hashedPassword, "admin",
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (am *AuthManager) Authenticate(username, password string) (*User, error) {
	var user User
	var hashedPassword string

	query := `
		SELECT id, username, email, password, role, active, created_at 
		FROM auth_users 
		WHERE (username = ? OR email = ?) AND active = 1
	`

	err := am.db.QueryRow(query, username, username).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword,
		&user.Role, &user.Active, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if !verifyPassword(password, hashedPassword) {
		return nil, errors.New("invalid credentials")
	}

	if perms, exists := am.permissions[user.Username]; exists {
		user.Permissions = perms
	}

	return &user, nil
}

func (am *AuthManager) GenerateToken(user *User) (string, error) {
	expirationTime := time.Now().Add(am.expires)
	
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (am *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (am *AuthManager) GetTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	return "", errors.New("no token found")
}

func (am *AuthManager) SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(am.expires.Seconds()),
	})
}

func (am *AuthManager) ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func (am *AuthManager) CreateUser(username, email, password, role string) error {
	hashedPassword := hashPassword(password)
	
	_, err := am.db.Exec(
		"INSERT INTO auth_users (username, email, password, role) VALUES (?, ?, ?, ?)",
		username, email, hashedPassword, role,
	)
	
	return err
}

func (am *AuthManager) GetUserByID(id int64) (*User, error) {
	var user User
	
	query := `
		SELECT id, username, email, role, active, created_at 
		FROM auth_users 
		WHERE id = ?
	`
	
	err := am.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.Role, &user.Active, &user.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if perms, exists := am.permissions[user.Username]; exists {
		user.Permissions = perms
	}
	
	return &user, nil
}

func (am *AuthManager) CheckPermission(username, modelName string, write bool) bool {
	var role string
	err := am.db.QueryRow("SELECT role FROM auth_users WHERE username = ?", username).Scan(&role)
	if err == nil && role == "admin" {
		return true
	}
	
	if userPerms, exists := am.permissions[username]; exists {
		if modelPerm, exists := userPerms[modelName]; exists {
			if write {
				return modelPerm.Write
			}
			return modelPerm.Read
		}
	}
	
	return false
}

func (am *AuthManager) GetUserFromToken(r *http.Request) (*User, error) {
	token, err := am.GetTokenFromRequest(r)
	if err != nil {
		return nil, err
	}
	
	claims, err := am.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	
	return am.GetUserByID(claims.UserID)
}

func (am *AuthManager) IsEnabled() bool {
	return am.config.Type != "none"
}

func hashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func verifyPassword(password, hashedPassword string) bool {
	return hashPassword(password) == hashedPassword
}

func generateRandomSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}