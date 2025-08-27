package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/yamlforge/yamlforge/internal/api"
	"github.com/yamlforge/yamlforge/internal/auth"
	"github.com/yamlforge/yamlforge/internal/database"
	"github.com/yamlforge/yamlforge/internal/parser"
	"github.com/yamlforge/yamlforge/internal/ui"
	"github.com/yamlforge/yamlforge/internal/validation"
)

type Server struct {
	config      *parser.Config
	db          database.Database
	schema      *parser.Schema
	router      *mux.Router
	authManager *auth.AuthManager
	validator   *validation.Validator
}

func New(config *parser.Config) *Server {
	return &Server{
		config: config,
		router: mux.NewRouter(),
	}
}

func (s *Server) Start(host string, port int) error {
	log.Println("Starting server initialization...")
	if err := s.initialize(); err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	log.Printf("Server starting on http://%s", addr)
	log.Printf("Routes registered, starting HTTP server...")

	return http.ListenAndServe(addr, s.router)
}

func (s *Server) initialize() error {
	log.Println("Creating database...")
	db, err := database.NewDatabase(&s.config.Database)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Println("Connecting to database...")
	if err := db.Connect(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.db = db

	log.Println("Loading schema...")
	schema, err := parser.LoadConfig(s.config)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	s.schema = schema
	s.validator = validation.New(schema)

	log.Println("Creating database schema...")
	if err := db.CreateSchema(schema); err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}

	if s.config.Server.Auth.Type != "none" {
		log.Println("Initializing authentication...")
		sqlDB, ok := db.(*database.SQLiteDB)
		if !ok {
			return fmt.Errorf("authentication requires SQLite database")
		}
		
		conn := sqlDB.GetConnection()
		if conn == nil {
			return fmt.Errorf("failed to get database connection for auth")
		}
		
		authManager, err := auth.New(&s.config.Server.Auth, conn)
		if err != nil {
			return fmt.Errorf("failed to initialize auth: %w", err)
		}
		s.authManager = authManager
	}

	log.Println("Templates loaded (hard-coded)")

	log.Println("Setting up routes...")
	s.setupRoutes()

	log.Println("Server initialization complete")
	return nil
}

func (s *Server) setupRoutes() {
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.debugMiddleware)

	if s.authManager != nil {
		s.router.Use(s.globalAuthMiddleware())
	}

	if s.authManager != nil {
		s.router.HandleFunc("/login", s.handleLogin).Methods("GET")
		s.router.HandleFunc("/api/auth/login", s.handleAuthLogin).Methods("POST")
	}

	if s.authManager != nil {
		s.router.HandleFunc("/logout", s.handleLogout).Methods("GET", "POST")
		s.router.HandleFunc("/api/auth/logout", s.handleAuthLogout).Methods("POST")
	}

	s.router.HandleFunc("/api/openapi", s.handleOpenAPI).Methods("GET")
	s.router.HandleFunc("/api/openapi.json", s.handleOpenAPI).Methods("GET")
	s.router.HandleFunc("/api/docs", s.handleSwaggerUI).Methods("GET")

	for modelName := range s.schema.Models {
		s.setupAPIRoutes(modelName)
	}

	s.router.HandleFunc("/", s.handleHome).Methods("GET")
	for modelName := range s.schema.Models {
		s.setupModelRoutes(modelName)
	}
}

func (s *Server) setupModelRoutes(modelName string) {
	basePath := "/" + strings.ToLower(modelName)

	s.router.HandleFunc(basePath, s.handleModelList(modelName)).Methods("GET")
	s.router.HandleFunc(basePath+"/new", s.handleModelNew(modelName)).Methods("GET")
	s.router.HandleFunc(basePath+"/{id}", s.handleModelView(modelName)).Methods("GET")
	s.router.HandleFunc(basePath+"/{id}/edit", s.handleModelEdit(modelName)).Methods("GET")
}

func (s *Server) setupAPIRoutes(modelName string) {
	basePath := "/api/" + strings.ToLower(modelName)

	s.router.HandleFunc(basePath, s.handleAPIList(modelName)).Methods("GET")
	s.router.HandleFunc(basePath, s.handleAPICreate(modelName)).Methods("POST")
	s.router.HandleFunc(basePath+"/{id}", s.handleAPIGet(modelName)).Methods("GET")
	s.router.HandleFunc(basePath+"/{id}", s.handleAPIUpdate(modelName)).Methods("PUT")
	s.router.HandleFunc(basePath+"/{id}", s.handleAPIDelete(modelName)).Methods("DELETE")
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	modelPermissions := make(map[string]bool)
	if s.authManager != nil && s.authManager.IsEnabled() {
		if user, ok := r.Context().Value("user").(*auth.User); ok {
			for modelName := range s.schema.Models {
				modelPermissions[modelName] = s.authManager.CheckPermission(user.Username, modelName, true)
			}
		}
	} else {
		for modelName := range s.schema.Models {
			modelPermissions[modelName] = true
		}
	}
	
	data := struct {
		Title            string
		Config           *parser.Config
		Models           map[string]*parser.Model
		ModelPermissions map[string]bool
	}{
		Title:            s.config.UI.Title,
		Config:           s.config,
		Models:           s.schema.Models,
		ModelPermissions: modelPermissions,
	}

	s.render(w, "home", data)
}

func (s *Server) handleModelList(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		model, ok := s.schema.GetModel(modelName)
		if !ok {
			http.NotFound(w, r)
			return
		}

		var canWrite bool
		if s.authManager != nil && s.authManager.IsEnabled() {
			if user, ok := r.Context().Value("user").(*auth.User); ok {
				canWrite = s.authManager.CheckPermission(user.Username, modelName, true)
			}
		} else {
			canWrite = true
		}

		data := struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			CanWrite  bool
		}{
			Title:     fmt.Sprintf("%s - %s", modelName, s.config.UI.Title),
			Config:    s.config,
			ModelName: modelName,
			Model:     model,
			CanWrite:  canWrite,
		}

		s.render(w, "list", data)
	}
}

func (s *Server) handleModelNew(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		model, ok := s.schema.GetModel(modelName)
		if !ok {
			http.NotFound(w, r)
			return
		}

		data := struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Action    string
		}{
			Title:     fmt.Sprintf("New %s - %s", modelName, s.config.UI.Title),
			Config:    s.config,
			ModelName: modelName,
			Model:     model,
			Action:    "create",
		}

		s.render(w, "form", data)
	}
}

func (s *Server) handleModelView(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		model, ok := s.schema.GetModel(modelName)
		if !ok {
			http.NotFound(w, r)
			return
		}

		record, err := s.db.Get(modelName, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		data := struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Record    map[string]any
		}{
			Title:     fmt.Sprintf("%s #%s - %s", modelName, id, s.config.UI.Title),
			Config:    s.config,
			ModelName: modelName,
			Model:     model,
			Record:    record,
		}

		s.render(w, "view", data)
	}
}

func (s *Server) handleModelEdit(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		model, ok := s.schema.GetModel(modelName)
		if !ok {
			http.NotFound(w, r)
			return
		}

		record, err := s.db.Get(modelName, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		data := struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Record    map[string]any
			Action    string
		}{
			Title:     fmt.Sprintf("Edit %s #%s - %s", modelName, id, s.config.UI.Title),
			Config:    s.config,
			ModelName: modelName,
			Model:     model,
			Record:    record,
			Action:    "update",
		}

		s.render(w, "form", data)
	}
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	var html string

	switch name {
	case "home":
		modelPermissions := make(map[string]bool)
		switch d := data.(type) {
		case struct {
			Title            string
			Config           *parser.Config
			Models           map[string]*parser.Model
			ModelPermissions map[string]bool
		}:
			modelPermissions = d.ModelPermissions
		}
		html = ui.GetHomeHTML(s.config, s.schema, modelPermissions)
	case "list":
		canWrite := false
		modelName := ""
		var model *parser.Model
		
		switch d := data.(type) {
		case struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			CanWrite  bool
		}:
			modelName = d.ModelName
			model = d.Model
			canWrite = d.CanWrite
		case struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
		}:
			modelName = d.ModelName
			model = d.Model
			canWrite = true
		}
		
		html = ui.GetListHTML(s.config, s.schema, modelName, model, canWrite)
	case "form":
		action := "create"
		recordID := "null"

		switch d := data.(type) {
		case struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Action    string
		}:
			action = d.Action
		case struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Record    map[string]any
			Action    string
		}:
			action = d.Action
			if d.Record != nil {
				if id, ok := d.Record["id"]; ok {
					recordID = fmt.Sprintf("%v", id)
				}
			}
		}

		if action == "update" {
			action = "edit"
		}
		var modelName string
		var model *parser.Model
		switch d := data.(type) {
		case struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Action    string
		}:
			modelName = d.ModelName
			model = d.Model
		case struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Record    map[string]any
			Action    string
		}:
			modelName = d.ModelName
			model = d.Model
		}
		html = ui.GetFormHTML(s.config, s.schema, modelName, model, action, recordID, s.extractRecordAsJSON(data))
	case "view":
		recordID := ""
		switch d := data.(type) {
		case struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Record    map[string]any
		}:
			if d.Record != nil {
				if id, ok := d.Record["id"].(string); ok {
					recordID = id
				}
			}
		}
		viewData := data.(struct {
			Title     string
			Config    *parser.Config
			ModelName string
			Model     *parser.Model
			Record    map[string]any
		})
		html = ui.GetViewHTML(s.config, s.schema, viewData.ModelName, viewData.Model, recordID, s.extractRecordAsJSON(data))
	default:
		log.Printf("Unknown template: %s", name)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) extractRecordAsJSON(data any) string {
	switch d := data.(type) {
	case struct {
		Title     string
		Config    *parser.Config
		ModelName string
		Model     *parser.Model
		Record    map[string]any
	}:
		if d.Record != nil {
			if jsonBytes, err := json.Marshal(d.Record); err == nil {
				return string(jsonBytes)
			}
		}
	case struct {
		Title     string
		Config    *parser.Config
		ModelName string
		Model     *parser.Model
		Record    map[string]any
		Action    string
	}:
		if d.Record != nil {
			if jsonBytes, err := json.Marshal(d.Record); err == nil {
				return string(jsonBytes)
			}
		}
	}
	return "null"
}


func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) debugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Debug-Mode", "true")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	html := ui.GetLoginHTML(s.config)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if s.authManager != nil {
		s.authManager.ClearAuthCookie(w)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		s.sendJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	user, err := s.authManager.Authenticate(loginRequest.Username, loginRequest.Password)
	if err != nil {
		s.sendJSON(w, http.StatusUnauthorized, map[string]any{
			"success": false,
			"error":   "Invalid credentials",
		})
		return
	}

	token, err := s.authManager.GenerateToken(user)
	if err != nil {
		s.sendJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "Failed to generate token",
		})
		return
	}

	s.authManager.SetAuthCookie(w, token)

	s.sendJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"token":   token,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if s.authManager != nil {
		s.authManager.ClearAuthCookie(w)
	}
	
	s.sendJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

func (s *Server) globalAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			publicPaths := []string{
				"/login",
				"/api/auth/login",
				"/api/docs",
				"/api/openapi.json",
			}

			isPublic := false
			for _, publicPath := range publicPaths {
				if r.URL.Path == publicPath {
					isPublic = true
					break
				}
			}
			
			if isPublic {
				next.ServeHTTP(w, r)
				return
			}

			token, err := s.authManager.GetTokenFromRequest(r)
			if err != nil {
				s.handleAuthError(w, r, "Authentication required")
				return
			}

			claims, err := s.authManager.ValidateToken(token)
			if err != nil {
				s.handleAuthError(w, r, "Invalid token")
				return
			}

			user, err := s.authManager.GetUserByID(claims.UserID)
			if err != nil {
				s.handleAuthError(w, r, "User not found")
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *Server) handleAuthError(w http.ResponseWriter, r *http.Request, message string) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		returnURL := r.URL.Path
		if r.URL.RawQuery != "" {
			returnURL += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, "/login?return="+returnURL, http.StatusSeeOther)
		return
	}
	s.sendJSON(w, http.StatusUnauthorized, map[string]any{
		"success": false,
		"error":   message,
	})
}

func (s *Server) authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return handler
}

func (s *Server) setupProtectedModelRoutes(modelName string) {
	basePath := "/" + strings.ToLower(modelName)
	
	s.router.HandleFunc(basePath, s.authMiddleware(s.handleModelList(modelName))).Methods("GET")
	s.router.HandleFunc(basePath+"/new", s.authMiddleware(s.handleModelNew(modelName))).Methods("GET")
	s.router.HandleFunc(basePath+"/{id}/edit", s.authMiddleware(s.handleModelEdit(modelName))).Methods("GET")
	s.router.HandleFunc(basePath+"/{id}", s.authMiddleware(s.handleModelView(modelName))).Methods("GET")
}

func (s *Server) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleAPIList(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authManager != nil && s.authManager.IsEnabled() {
			user, ok := r.Context().Value("user").(*auth.User)
			if !ok || !s.authManager.CheckPermission(user.Username, modelName, false) {
				s.sendJSON(w, http.StatusForbidden, map[string]any{
					"success": false,
					"error":   "You don't have permission to read this resource",
				})
				return
			}
		}

		params := s.parseQueryParams(r)
		results, err := s.db.Query(modelName, params)
		if err != nil {
			s.sendJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		total, err := s.db.Count(modelName, params.Filters)
		if err != nil {
			s.sendJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		totalPages := int(total) / params.PageSize
		if int(total)%params.PageSize > 0 {
			totalPages++
		}

		s.sendJSON(w, http.StatusOK, parser.APIResponse{
			Success: true,
			Data:    results,
			Meta: &parser.Meta{
				Page:       params.Page,
				PageSize:   params.PageSize,
				TotalCount: total,
				TotalPages: totalPages,
			},
		})
	}
}

func (s *Server) handleAPIGet(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authManager != nil && s.authManager.IsEnabled() {
			user, ok := r.Context().Value("user").(*auth.User)
			if !ok || !s.authManager.CheckPermission(user.Username, modelName, false) {
				s.sendJSON(w, http.StatusForbidden, map[string]any{
					"success": false,
					"error":   "You don't have permission to read this resource",
				})
				return
			}
		}

		vars := mux.Vars(r)
		id := vars["id"]

		result, err := s.db.Get(modelName, id)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				s.sendJSON(w, http.StatusNotFound, map[string]any{
					"success": false,
					"error":   "Record not found",
				})
			} else {
				s.sendJSON(w, http.StatusInternalServerError, map[string]any{
					"success": false,
					"error":   err.Error(),
				})
			}
			return
		}

		s.sendJSON(w, http.StatusOK, parser.APIResponse{
			Success: true,
			Data:    result,
		})
	}
}

func (s *Server) handleAPICreate(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authManager != nil && s.authManager.IsEnabled() {
			user, ok := r.Context().Value("user").(*auth.User)
			if !ok || !s.authManager.CheckPermission(user.Username, modelName, true) {
				s.sendJSON(w, http.StatusForbidden, map[string]any{
					"success": false,
					"error":   "You don't have permission to write to this resource",
				})
				return
			}
		}

		var data map[string]any
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			s.sendJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "Invalid JSON",
			})
			return
		}

		if err := s.validator.ValidateCreate(modelName, data); err != nil {
			s.sendJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		id, err := s.db.Create(modelName, data)
		if err != nil {
			s.sendJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		result, err := s.db.Get(modelName, id)
		if err != nil {
			s.sendJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		s.sendJSON(w, http.StatusCreated, parser.APIResponse{
			Success: true,
			Data:    result,
		})
	}
}

func (s *Server) handleAPIUpdate(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authManager != nil && s.authManager.IsEnabled() {
			user, ok := r.Context().Value("user").(*auth.User)
			if !ok || !s.authManager.CheckPermission(user.Username, modelName, true) {
				s.sendJSON(w, http.StatusForbidden, map[string]any{
					"success": false,
					"error":   "You don't have permission to write to this resource",
				})
				return
			}
		}

		vars := mux.Vars(r)
		id := vars["id"]

		var data map[string]any
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			s.sendJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "Invalid JSON",
			})
			return
		}

		data = s.filterEmptyPasswordFields(modelName, data)

		if err := s.validator.ValidateUpdate(modelName, data); err != nil {
			s.sendJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		if err := s.db.Update(modelName, id, data); err != nil {
			s.sendJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		result, err := s.db.Get(modelName, id)
		if err != nil {
			s.sendJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		s.sendJSON(w, http.StatusOK, parser.APIResponse{
			Success: true,
			Data:    result,
		})
	}
}

func (s *Server) handleAPIDelete(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authManager != nil && s.authManager.IsEnabled() {
			user, ok := r.Context().Value("user").(*auth.User)
			if !ok || !s.authManager.CheckPermission(user.Username, modelName, true) {
				s.sendJSON(w, http.StatusForbidden, map[string]any{
					"success": false,
					"error":   "You don't have permission to delete this resource",
				})
				return
			}
		}

		vars := mux.Vars(r)
		id := vars["id"]

		if err := s.db.Delete(modelName, id); err != nil {
			s.sendJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		s.sendJSON(w, http.StatusOK, parser.APIResponse{
			Success: true,
		})
	}
}

func (s *Server) parseQueryParams(r *http.Request) parser.QueryParams {
	params := parser.QueryParams{
		Page:     1,
		PageSize: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			params.PageSize = ps
		}
	}

	if sort := r.URL.Query().Get("sort"); sort != "" {
		for _, field := range strings.Split(sort, ",") {
			desc := false
			if strings.HasPrefix(field, "-") {
				desc = true
				field = field[1:]
			}
			params.Sort = append(params.Sort, parser.SortField{
				Field: field,
				Desc:  desc,
			})
		}
	}

	params.Search = r.URL.Query().Get("search")

	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "filter.") && len(values) > 0 {
			field := strings.TrimPrefix(key, "filter.")
			params.Filters = append(params.Filters, parser.Filter{
				Field:    field,
				Operator: "=",
				Value:    values[0],
			})
		}
	}

	return params
}

func (s *Server) filterEmptyPasswordFields(modelName string, data map[string]any) map[string]any {
	model, exists := s.schema.GetModel(modelName)
	if !exists {
		return data
	}

	for _, field := range model.Fields {
		if field.Type == parser.FieldTypePassword {
			if value, exists := data[field.Name]; exists {
				if value == nil || value == "" {
					delete(data, field.Name)
				}
			}
		}
	}

	return data
}

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	apiHandler := api.New(s.db, s.config, s.schema, s.authManager)
	spec := apiHandler.GenerateOpenAPI(r)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}

func (s *Server) handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>` + s.config.App.Name + ` - API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
    window.onload = function() {
        window.ui = SwaggerUIBundle({
            url: "/api/openapi.json",
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout"
        });
    };
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

