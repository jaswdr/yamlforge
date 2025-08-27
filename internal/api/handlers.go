package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/yamlforge/yamlforge/internal/auth"
	"github.com/yamlforge/yamlforge/internal/database"
	"github.com/yamlforge/yamlforge/internal/parser"
	"github.com/yamlforge/yamlforge/internal/validation"
)

type API struct {
	db          database.Database
	config      *parser.Config
	schema      *parser.Schema
	validator   *validation.Validator
	authManager *auth.AuthManager
}

func New(db database.Database, config *parser.Config, schema *parser.Schema, authManager *auth.AuthManager) *API {
	return &API{
		db:          db,
		config:      config,
		schema:      schema,
		validator:   validation.New(schema),
		authManager: authManager,
	}
}

func (api *API) RegisterRoutes(router *mux.Router) {
	apiRouter := router.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("/openapi", api.HandleOpenAPI()).Methods("GET")
	apiRouter.HandleFunc("/openapi.json", api.HandleOpenAPI()).Methods("GET")
	apiRouter.HandleFunc("/docs", api.HandleSwaggerUI()).Methods("GET")

	for modelName := range api.schema.Models {
		api.registerModelRoutes(apiRouter, modelName)
	}

	apiRouter.Use(api.corsMiddleware)
	apiRouter.Use(api.contentTypeMiddleware)
}

func (api *API) registerModelRoutes(router *mux.Router, modelName string) {
	basePath := "/" + strings.ToLower(modelName)

	router.HandleFunc(basePath, api.handleList(modelName)).Methods("GET")
	router.HandleFunc(basePath, api.handleCreate(modelName)).Methods("POST")
	router.HandleFunc(basePath+"/{id}", api.handleGet(modelName)).Methods("GET")
	router.HandleFunc(basePath+"/{id}", api.handleUpdate(modelName)).Methods("PUT")
	router.HandleFunc(basePath+"/{id}", api.handleDelete(modelName)).Methods("DELETE")
	router.HandleFunc(basePath+"/bulk", api.handleBulk(modelName)).Methods("POST")
}

func (api *API) handleList(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := api.checkPermission(r, modelName, false)
		if err != nil {
			api.sendError(w, http.StatusForbidden, err.Error())
			return
		}

		params := api.parseQueryParams(r)

		results, err := api.db.Query(modelName, params)
		if err != nil {
			api.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		total, err := api.db.Count(modelName, params.Filters)
		if err != nil {
			api.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		totalPages := int(total) / params.PageSize
		if int(total)%params.PageSize > 0 {
			totalPages++
		}

		api.sendResponse(w, http.StatusOK, parser.APIResponse{
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

func (api *API) handleGet(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := api.checkPermission(r, modelName, false)
		if err != nil {
			api.sendError(w, http.StatusForbidden, err.Error())
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		result, err := api.db.Get(modelName, id)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				api.sendError(w, http.StatusNotFound, "Record not found")
			} else {
				api.sendError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		api.sendResponse(w, http.StatusOK, parser.APIResponse{
			Success: true,
			Data:    result,
		})
	}
}

func (api *API) handleCreate(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := api.checkPermission(r, modelName, true)
		if err != nil {
			api.sendError(w, http.StatusForbidden, err.Error())
			return
		}

		var data map[string]any
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			api.sendError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		if err := api.validator.ValidateCreate(modelName, data); err != nil {
			api.sendError(w, http.StatusBadRequest, err.Error())
			return
		}

		id, err := api.db.Create(modelName, data)
		if err != nil {
			api.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		result, err := api.db.Get(modelName, id)
		if err != nil {
			api.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		api.sendResponse(w, http.StatusCreated, parser.APIResponse{
			Success: true,
			Data:    result,
		})
	}
}

func (api *API) handleUpdate(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := api.checkPermission(r, modelName, true)
		if err != nil {
			api.sendError(w, http.StatusForbidden, err.Error())
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		var data map[string]any
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			api.sendError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		data = api.filterEmptyPasswordFields(modelName, data)

		if err := api.validator.ValidateUpdate(modelName, data); err != nil {
			api.sendError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := api.db.Update(modelName, id, data); err != nil {
			api.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		result, err := api.db.Get(modelName, id)
		if err != nil {
			api.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		api.sendResponse(w, http.StatusOK, parser.APIResponse{
			Success: true,
			Data:    result,
		})
	}
}

func (api *API) handleDelete(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := api.checkPermission(r, modelName, true)
		if err != nil {
			api.sendError(w, http.StatusForbidden, err.Error())
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		if err := api.db.Delete(modelName, id); err != nil {
			api.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		api.sendResponse(w, http.StatusOK, parser.APIResponse{
			Success: true,
		})
	}
}

func (api *API) handleBulk(modelName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Operation string           `json:"operation"`
			Data      []map[string]any `json:"data"`
			IDs       []any            `json:"ids"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			api.sendError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		switch request.Operation {
		case "create":
			results := []any{}
			for _, item := range request.Data {
				if err := api.validator.ValidateCreate(modelName, item); err != nil {
					api.sendError(w, http.StatusBadRequest, err.Error())
					return
				}

				id, err := api.db.Create(modelName, item)
				if err != nil {
					api.sendError(w, http.StatusInternalServerError, err.Error())
					return
				}

				result, err := api.db.Get(modelName, id)
				if err != nil {
					api.sendError(w, http.StatusInternalServerError, err.Error())
					return
				}

				results = append(results, result)
			}

			api.sendResponse(w, http.StatusCreated, parser.APIResponse{
				Success: true,
				Data:    results,
			})

		case "delete":
			for _, id := range request.IDs {
				if err := api.db.Delete(modelName, id); err != nil {
					api.sendError(w, http.StatusInternalServerError, err.Error())
					return
				}
			}

			api.sendResponse(w, http.StatusOK, parser.APIResponse{
				Success: true,
			})

		default:
			api.sendError(w, http.StatusBadRequest, "Invalid operation")
		}
	}
}

func (api *API) parseQueryParams(r *http.Request) parser.QueryParams {
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

func (api *API) sendResponse(w http.ResponseWriter, status int, response parser.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (api *API) sendError(w http.ResponseWriter, status int, message string) {
	api.sendResponse(w, status, parser.APIResponse{
		Success: false,
		Error:   message,
	})
}

func (api *API) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if api.config.Server.CORS.Enabled {
			origins := api.config.Server.CORS.Origins
			if len(origins) > 0 && origins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin := r.Header.Get("Origin"); origin != "" {
				for _, allowed := range origins {
					if allowed == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (api *API) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (api *API) filterPasswordFields(modelName string, records []map[string]any) []map[string]any {
	model, exists := api.schema.GetModel(modelName)
	if !exists {
		return records
	}

	var passwordFields []string
	for _, field := range model.Fields {
		if field.Type == parser.FieldTypePassword {
			passwordFields = append(passwordFields, field.Name)
		}
	}

	if len(passwordFields) == 0 {
		return records
	}
	filteredRecords := make([]map[string]any, len(records))
	for i, record := range records {
		filteredRecord := make(map[string]any)
		for key, value := range record {
			isPassword := false
			for _, pwField := range passwordFields {
				if key == pwField {
					isPassword = true
					break
				}
			}
			if !isPassword {
				filteredRecord[key] = value
			}
		}
		filteredRecords[i] = filteredRecord
	}

	return filteredRecords
}

func (api *API) filterPasswordFieldsSingle(modelName string, record map[string]any) map[string]any {
	model, exists := api.schema.GetModel(modelName)
	if !exists {
		return record
	}

	var passwordFields []string
	for _, field := range model.Fields {
		if field.Type == parser.FieldTypePassword {
			passwordFields = append(passwordFields, field.Name)
		}
	}

	if len(passwordFields) == 0 {
		return record
	}
	filteredRecord := make(map[string]any)
	for key, value := range record {
		isPassword := false
		for _, pwField := range passwordFields {
			if key == pwField {
				isPassword = true
				break
			}
		}
		if !isPassword {
			filteredRecord[key] = value
		}
	}

	return filteredRecord
}

func (api *API) filterEmptyPasswordFields(modelName string, data map[string]any) map[string]any {
	model, exists := api.schema.GetModel(modelName)
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

func (api *API) checkPermission(r *http.Request, modelName string, write bool) (*auth.User, error) {
	if api.authManager == nil || !api.authManager.IsEnabled() {
		return nil, nil
	}

	user, err := api.authManager.GetUserFromToken(r)
	if err != nil {
		return nil, err
	}

	if !api.authManager.CheckPermission(user.Username, modelName, write) {
		action := "read"
		if write {
			action = "write"
		}
		return nil, fmt.Errorf("you don't have permission to %s this resource", action)
	}

	return user, nil
}
