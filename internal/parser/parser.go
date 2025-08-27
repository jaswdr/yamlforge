package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ParseConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	config = processConfig(config)

	return config, nil
}

func validateConfig(config *Config) error {
	if config.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if config.Database.Type != "sqlite" && config.Database.Type != "postgresql" && config.Database.Type != "mysql" {
		return fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}

	if config.Database.Type == "sqlite" && config.Database.Path == "" {
		return fmt.Errorf("database.path is required for SQLite")
	}

	if config.Database.Type != "sqlite" && config.Database.Connection == "" {
		return fmt.Errorf("database.connection is required for %s", config.Database.Type)
	}

	for modelName, model := range config.Models {
		if err := validateModel(modelName, model); err != nil {
			return err
		}
	}

	return nil
}

func validateModel(name string, model ModelConfig) error {
	if len(model.Fields) == 0 {
		return fmt.Errorf("model %s has no fields", name)
	}

	hasPrimary := false
	for fieldName, field := range model.Fields {
		if field.Primary {
			if hasPrimary {
				return fmt.Errorf("model %s has multiple primary keys", name)
			}
			hasPrimary = true
		}

		if err := validateField(name, fieldName, field); err != nil {
			return err
		}
	}

	if !hasPrimary {
		return fmt.Errorf("model %s has no primary key", name)
	}

	return nil
}

func validateField(modelName, fieldName string, field FieldConfig) error {
	fieldType := FieldType(field.Type)
	if !fieldType.IsValid() {
		return fmt.Errorf("invalid field type '%s' for %s.%s", field.Type, modelName, fieldName)
	}

	if fieldType == FieldTypeEnum && len(field.Options) == 0 {
		return fmt.Errorf("enum field %s.%s must have options", modelName, fieldName)
	}

	if fieldType == FieldTypeRelation && field.To == "" {
		return fmt.Errorf("relation field %s.%s must specify 'to' model", modelName, fieldName)
	}

	if fieldType == FieldTypeArray && field.Items == "" {
		return fmt.Errorf("array field %s.%s must specify 'items' type", modelName, fieldName)
	}

	if field.Min > field.Max && field.Max > 0 {
		return fmt.Errorf("field %s.%s has min > max", modelName, fieldName)
	}

	return nil
}

func processConfig(config *Config) *Config {
	for modelName, model := range config.Models {
		processedModel := model

		for fieldName, field := range model.Fields {
			processedField := field

			if field.Type == "id" && !field.Primary {
				processedField.Primary = true
			}

			if field.Type == "datetime" || field.Type == "date" || field.Type == "time" {
				if field.AutoNow || field.AutoNowAdd {
					processedField.Default = "CURRENT_TIMESTAMP"
				}
			}

			if field.Primary {
				processedField.Required = true
				processedField.Unique = true
			}

			processedModel.Fields[fieldName] = processedField
		}

		if processedModel.UI == nil {
			processedModel.UI = generateDefaultUI(processedModel)
		}

		if processedModel.Permissions == nil {
			processedModel.Permissions = &PermissionsConfig{
				Create: "authenticated",
				Read:   "all",
				Update: "authenticated",
				Delete: "authenticated",
			}
		}

		config.Models[modelName] = processedModel
	}

	return config
}

func generateDefaultUI(model ModelConfig) *UIModelConfig {
	var columns []string
	var sortable []string
	var searchable []string
	var formFields []string

	for fieldName, field := range model.Fields {
		if field.Type != "password" && !field.Primary {
			columns = append(columns, fieldName)

			if field.Type == "text" || field.Type == "email" || field.Type == "number" {
				searchable = append(searchable, fieldName)
			}

			if field.Type != "file" && field.Type != "image" {
				sortable = append(sortable, fieldName)
			}
		}

		if !field.Primary && !field.AutoNow && !field.AutoNowAdd {
			formFields = append(formFields, fieldName)
		}
	}

	if len(columns) > 5 {
		columns = columns[:5]
	}

	return &UIModelConfig{
		List: &UIListConfig{
			Columns:    columns,
			Sortable:   sortable,
			Searchable: searchable,
		},
		Form: &UIFormConfig{
			Fields: formFields,
		},
	}
}

func LoadConfig(config *Config) (*Schema, error) {
	schema := &Schema{
		Models: make(map[string]*Model),
	}

	for modelName, modelConfig := range config.Models {
		model := &Model{
			Name:   modelName,
			Fields: []Field{},
		}

		for fieldName, fieldConfig := range modelConfig.Fields {
			field := Field{
				Name:       fieldName,
				Type:       FieldType(fieldConfig.Type),
				Primary:    fieldConfig.Primary,
				Required:   fieldConfig.Required,
				Unique:     fieldConfig.Unique,
				Default:    fieldConfig.Default,
				AutoNow:    fieldConfig.AutoNow,
				AutoNowAdd: fieldConfig.AutoNowAdd,
				Nullable:   fieldConfig.Nullable,
				Index:      fieldConfig.Index,
				RelatedTo:  fieldConfig.To,
				OnDelete:   fieldConfig.OnDelete,
				ArrayType:  fieldConfig.Items,
			}

			if fieldConfig.Min > 0 {
				min := fieldConfig.Min
				field.Min = &min
			}
			if fieldConfig.Max > 0 {
				max := fieldConfig.Max
				field.Max = &max
			}

			field.Pattern = fieldConfig.Pattern
			field.Options = fieldConfig.Options

			model.Fields = append(model.Fields, field)
		}

		if modelConfig.Permissions != nil {
			model.Permissions = Permissions{
				Create: modelConfig.Permissions.Create,
				Read:   modelConfig.Permissions.Read,
				Update: modelConfig.Permissions.Update,
				Delete: modelConfig.Permissions.Delete,
			}
		}

		if modelConfig.UI != nil {
			model.UI = UIModel{
				List: UIList{
					Columns:    modelConfig.UI.List.Columns,
					Sortable:   modelConfig.UI.List.Sortable,
					Searchable: modelConfig.UI.List.Searchable,
				},
				Form: UIForm{
					Fields: modelConfig.UI.Form.Fields,
				},
			}
		}

		schema.Models[modelName] = model
	}

	return schema, nil
}

type Schema struct {
	Models map[string]*Model
}

func (s *Schema) GetModel(name string) (*Model, bool) {
	model, ok := s.Models[name]
	return model, ok
}

func (s *Schema) GetField(modelName, fieldName string) (*Field, bool) {
	model, ok := s.GetModel(modelName)
	if !ok {
		return nil, false
	}

	for _, field := range model.Fields {
		if field.Name == fieldName {
			return &field, true
		}
	}

	return nil, false
}

func SaveConfig(config *Config, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
