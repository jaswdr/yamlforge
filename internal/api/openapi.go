package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/yamlforge/yamlforge/internal/parser"
)

type OpenAPISpec struct {
	OpenAPI    string              `json:"openapi"`
	Info       OpenAPIInfo         `json:"info"`
	Servers    []OpenAPIServer     `json:"servers"`
	Paths      map[string]PathItem `json:"paths"`
	Components OpenAPIComponents   `json:"components"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type PathItem map[string]Operation

type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema *Schema `json:"schema,omitempty"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
	Enum        []string           `json:"enum,omitempty"`
	Default     any                `json:"default,omitempty"`
	Minimum     *int               `json:"minimum,omitempty"`
	Maximum     *int               `json:"maximum,omitempty"`
	MinLength   *int               `json:"minLength,omitempty"`
	MaxLength   *int               `json:"maxLength,omitempty"`
}

type OpenAPIComponents struct {
	Schemas         map[string]*Schema        `json:"schemas"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty"`
	In           string `json:"in,omitempty"`
	Name         string `json:"name,omitempty"`
}

func (api *API) GenerateOpenAPI(r *http.Request) *OpenAPISpec {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = fmt.Sprintf("%s:%d", api.config.Server.Host, api.config.Server.Port)
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       api.config.App.Name,
			Description: api.config.App.Description,
			Version:     api.config.App.Version,
		},
		Servers: []OpenAPIServer{
			{
				URL:         fmt.Sprintf("%s://%s/api", scheme, host),
				Description: "API Server",
			},
		},
		Paths: make(map[string]PathItem),
		Components: OpenAPIComponents{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	if api.config.Server.Auth.Type == "jwt" {
		spec.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT authentication. Use the /api/auth/login endpoint to obtain a token.",
		}
		spec.Components.SecuritySchemes["cookieAuth"] = SecurityScheme{
			Type:        "apiKey",
			In:          "cookie",
			Name:        "auth_token",
			Description: "Cookie-based JWT authentication",
		}
	}

	var securityReq []map[string][]string
	if api.config.Server.Auth.Type == "jwt" {
		securityReq = []map[string][]string{
			{"bearerAuth": {}},
			{"cookieAuth": {}},
		}
	}

	for modelName, model := range api.schema.Models {
		spec.Components.Schemas[modelName] = api.generateModelSchema(model)
		spec.Components.Schemas[modelName+"Input"] = api.generateInputSchema(model)

		basePath := "/" + strings.ToLower(modelName)

		spec.Paths[basePath] = PathItem{
			"get": Operation{
				Tags:        []string{modelName},
				Summary:     fmt.Sprintf("List %s", modelName),
				Description: fmt.Sprintf("Get a paginated list of %s", modelName),
				OperationID: fmt.Sprintf("list%s", modelName),
				Security:    securityReq,
				Parameters: []Parameter{
					{
						Name:        "page",
						In:          "query",
						Description: "Page number",
						Schema:      &Schema{Type: "integer", Default: 1},
					},
					{
						Name:        "page_size",
						In:          "query",
						Description: "Items per page",
						Schema:      &Schema{Type: "integer", Default: 20},
					},
					{
						Name:        "search",
						In:          "query",
						Description: "Search query",
						Schema:      &Schema{Type: "string"},
					},
					{
						Name:        "sort",
						In:          "query",
						Description: "Sort fields (prefix with - for descending)",
						Schema:      &Schema{Type: "string"},
					},
				},
				Responses: map[string]Response{
					"200": {
						Description: "Successful response",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"data": {
											Type:  "array",
											Items: &Schema{Ref: fmt.Sprintf("#/components/schemas/%s", modelName)},
										},
										"meta": {
											Type: "object",
											Properties: map[string]*Schema{
												"page":        {Type: "integer"},
												"page_size":   {Type: "integer"},
												"total_count": {Type: "integer"},
												"total_pages": {Type: "integer"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"post": Operation{
				Tags:        []string{modelName},
				Summary:     fmt.Sprintf("Create %s", modelName),
				Description: fmt.Sprintf("Create a new %s", modelName),
				OperationID: fmt.Sprintf("create%s", modelName),
				Security:    securityReq,
				RequestBody: &RequestBody{
					Required: true,
					Content: map[string]MediaType{
						"application/json": {
							Schema: &Schema{Ref: fmt.Sprintf("#/components/schemas/%sInput", modelName)},
						},
					},
				},
				Responses: map[string]Response{
					"201": {
						Description: "Created successfully",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"data":    {Ref: fmt.Sprintf("#/components/schemas/%s", modelName)},
									},
								},
							},
						},
					},
					"400": {
						Description: "Bad request",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"error":   {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
		}

		spec.Paths[basePath+"/{id}"] = PathItem{
			"get": Operation{
				Tags:        []string{modelName},
				Summary:     fmt.Sprintf("Get %s", modelName),
				Description: fmt.Sprintf("Get a single %s by ID", modelName),
				OperationID: fmt.Sprintf("get%s", modelName),
				Security:    securityReq,
				Parameters: []Parameter{
					{
						Name:        "id",
						In:          "path",
						Description: fmt.Sprintf("%s ID", modelName),
						Required:    true,
						Schema:      &Schema{Type: "string"},
					},
				},
				Responses: map[string]Response{
					"200": {
						Description: "Successful response",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"data":    {Ref: fmt.Sprintf("#/components/schemas/%s", modelName)},
									},
								},
							},
						},
					},
					"404": {
						Description: "Not found",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"error":   {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			"put": Operation{
				Tags:        []string{modelName},
				Summary:     fmt.Sprintf("Update %s", modelName),
				Description: fmt.Sprintf("Update an existing %s", modelName),
				OperationID: fmt.Sprintf("update%s", modelName),
				Security:    securityReq,
				Parameters: []Parameter{
					{
						Name:        "id",
						In:          "path",
						Description: fmt.Sprintf("%s ID", modelName),
						Required:    true,
						Schema:      &Schema{Type: "string"},
					},
				},
				RequestBody: &RequestBody{
					Required: true,
					Content: map[string]MediaType{
						"application/json": {
							Schema: &Schema{Ref: fmt.Sprintf("#/components/schemas/%sInput", modelName)},
						},
					},
				},
				Responses: map[string]Response{
					"200": {
						Description: "Updated successfully",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"data":    {Ref: fmt.Sprintf("#/components/schemas/%s", modelName)},
									},
								},
							},
						},
					},
					"404": {
						Description: "Not found",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"error":   {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			"delete": Operation{
				Tags:        []string{modelName},
				Summary:     fmt.Sprintf("Delete %s", modelName),
				Description: fmt.Sprintf("Delete a %s by ID", modelName),
				OperationID: fmt.Sprintf("delete%s", modelName),
				Security:    securityReq,
				Parameters: []Parameter{
					{
						Name:        "id",
						In:          "path",
						Description: fmt.Sprintf("%s ID", modelName),
						Required:    true,
						Schema:      &Schema{Type: "string"},
					},
				},
				Responses: map[string]Response{
					"200": {
						Description: "Deleted successfully",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
									},
								},
							},
						},
					},
					"404": {
						Description: "Not found",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"error":   {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	if api.config.Server.Auth.Type == "jwt" {
		spec.Paths["/auth/login"] = PathItem{
			"post": Operation{
				Tags:        []string{"Authentication"},
				Summary:     "Login",
				Description: "Authenticate and receive a JWT token",
				OperationID: "login",
				RequestBody: &RequestBody{
					Required: true,
					Content: map[string]MediaType{
						"application/json": {
							Schema: &Schema{
								Type:     "object",
								Required: []string{"username", "password"},
								Properties: map[string]*Schema{
									"username": {Type: "string", Description: "Username or email"},
									"password": {Type: "string", Format: "password", Description: "User password"},
								},
							},
						},
					},
				},
				Responses: map[string]Response{
					"200": {
						Description: "Successful authentication",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"token":   {Type: "string", Description: "JWT token"},
										"user": {
											Type: "object",
											Properties: map[string]*Schema{
												"id":       {Type: "integer"},
												"username": {Type: "string"},
												"email":    {Type: "string"},
												"role":     {Type: "string"},
											},
										},
									},
								},
							},
						},
					},
					"401": {
						Description: "Invalid credentials",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"error":   {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
		}

		spec.Paths["/auth/logout"] = PathItem{
			"post": Operation{
				Tags:        []string{"Authentication"},
				Summary:     "Logout",
				Description: "Logout and invalidate the session",
				OperationID: "logout",
				Security:    securityReq,
				Responses: map[string]Response{
					"200": {
						Description: "Successfully logged out",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"message": {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
		}

		spec.Paths["/auth/me"] = PathItem{
			"get": Operation{
				Tags:        []string{"Authentication"},
				Summary:     "Get current user",
				Description: "Get the currently authenticated user's information",
				OperationID: "getCurrentUser",
				Security:    securityReq,
				Responses: map[string]Response{
					"200": {
						Description: "Current user information",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"user": {
											Type: "object",
											Properties: map[string]*Schema{
												"id":       {Type: "integer"},
												"username": {Type: "string"},
												"email":    {Type: "string"},
												"role":     {Type: "string"},
											},
										},
									},
								},
							},
						},
					},
					"401": {
						Description: "Not authenticated",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"error":   {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	return spec
}

func (api *API) generateModelSchema(model *parser.Model) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   []string{},
	}

	for _, field := range model.Fields {
		fieldSchema := api.fieldToSchema(field)
		schema.Properties[field.Name] = fieldSchema

		if field.Required && !field.AutoNow && !field.AutoNowAdd {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	return schema
}

func (api *API) generateInputSchema(model *parser.Model) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   []string{},
	}

	for _, field := range model.Fields {
		if field.Primary || field.AutoNow || field.AutoNowAdd {
			continue
		}

		fieldSchema := api.fieldToSchema(field)
		schema.Properties[field.Name] = fieldSchema

		if field.Required {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	return schema
}

func (api *API) fieldToSchema(field parser.Field) *Schema {
	schema := &Schema{}

	switch field.Type {
	case parser.FieldTypeID:
		schema.Type = "integer"
		schema.Format = "int64"
	case parser.FieldTypeText, parser.FieldTypeEmail, parser.FieldTypePhone,
		parser.FieldTypeURL, parser.FieldTypeSlug, parser.FieldTypePassword,
		parser.FieldTypeColor, parser.FieldTypeMarkdown, parser.FieldTypeJSON,
		parser.FieldTypeCurrency, parser.FieldTypeIP, parser.FieldTypeUUID,
		parser.FieldTypeDuration:
		schema.Type = "string"
		if field.Type == parser.FieldTypeEmail {
			schema.Format = "email"
		} else if field.Type == parser.FieldTypeURL {
			schema.Format = "uri"
		} else if field.Type == parser.FieldTypeUUID {
			schema.Format = "uuid"
		} else if field.Type == parser.FieldTypePassword {
			schema.Format = "password"
		}
		schema.MinLength = field.Min
		schema.MaxLength = field.Max
	case parser.FieldTypeNumber:
		schema.Type = "number"
		schema.Minimum = field.Min
		schema.Maximum = field.Max
	case parser.FieldTypeBoolean:
		schema.Type = "boolean"
	case parser.FieldTypeDatetime, parser.FieldTypeDate, parser.FieldTypeTime:
		schema.Type = "string"
		if field.Type == parser.FieldTypeDatetime {
			schema.Format = "date-time"
		} else if field.Type == parser.FieldTypeDate {
			schema.Format = "date"
		} else if field.Type == parser.FieldTypeTime {
			schema.Format = "time"
		}
	case parser.FieldTypeEnum:
		schema.Type = "string"
		schema.Enum = field.Options
	case parser.FieldTypeArray:
		schema.Type = "array"
		schema.Items = &Schema{Type: "string"}
	case parser.FieldTypeFile, parser.FieldTypeImage:
		schema.Type = "string"
		schema.Format = "binary"
	case parser.FieldTypeRelation:
		schema.Type = "integer"
		schema.Format = "int64"
	case parser.FieldTypeLocation:
		schema.Type = "object"
		schema.Properties = map[string]*Schema{
			"lat": {Type: "number"},
			"lng": {Type: "number"},
		}
	}

	if field.Default != nil {
		schema.Default = field.Default
	}

	return schema
}

func (api *API) HandleOpenAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spec := api.GenerateOpenAPI(r)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(spec)
	}
}

func (api *API) HandleSwaggerUI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>` + api.config.App.Name + ` - API Documentation</title>
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
}

