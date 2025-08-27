package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yamlforge/yamlforge/internal/parser"
)

type Validator struct {
	schema *parser.Schema
}

func New(schema *parser.Schema) *Validator {
	return &Validator{
		schema: schema,
	}
}

func (v *Validator) ValidateCreate(modelName string, data map[string]any) error {
	model, ok := v.schema.GetModel(modelName)
	if !ok {
		return fmt.Errorf("model %s not found", modelName)
	}

	for _, field := range model.Fields {
		if field.Primary && field.Type == parser.FieldTypeID {
			continue
		}

		value, exists := data[field.Name]

		if field.Required && !exists {
			return parser.ValidationError{
				Field:   field.Name,
				Message: "field is required",
			}
		}

		if exists {
			if err := v.validateField(field, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) ValidateUpdate(modelName string, data map[string]any) error {
	model, ok := v.schema.GetModel(modelName)
	if !ok {
		return fmt.Errorf("model %s not found", modelName)
	}

	for fieldName, value := range data {
		field, found := v.getField(model, fieldName)
		if !found {
			return parser.ValidationError{
				Field:   fieldName,
				Message: "field does not exist",
			}
		}

		if field.Primary {
			return parser.ValidationError{
				Field:   fieldName,
				Message: "cannot update primary key",
			}
		}

		if err := v.validateField(*field, value); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateField(field parser.Field, value any) error {
	if value == nil {
		if field.Nullable {
			return nil
		}
		if field.Required {
			return parser.ValidationError{
				Field:   field.Name,
				Message: "field is required",
			}
		}
		return nil
	}

	switch field.Type {
	case parser.FieldTypeText, parser.FieldTypePassword:
		return v.validateText(field, value)
	case parser.FieldTypeNumber:
		return v.validateNumber(field, value)
	case parser.FieldTypeBoolean:
		return v.validateBoolean(field, value)
	case parser.FieldTypeEmail:
		return v.validateEmail(field, value)
	case parser.FieldTypeURL:
		return v.validateURL(field, value)
	case parser.FieldTypeEnum:
		return v.validateEnum(field, value)
	case parser.FieldTypeDatetime, parser.FieldTypeDate, parser.FieldTypeTime:
		return v.validateDatetime(field, value)
	}

	return nil
}

func (v *Validator) validateText(field parser.Field, value any) error {
	str, ok := value.(string)
	if !ok {
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a string",
		}
	}

	if field.Min != nil && len(str) < *field.Min {
		return parser.ValidationError{
			Field:   field.Name,
			Message: fmt.Sprintf("must be at least %d characters", *field.Min),
		}
	}

	if field.Max != nil && len(str) > *field.Max {
		return parser.ValidationError{
			Field:   field.Name,
			Message: fmt.Sprintf("must be at most %d characters", *field.Max),
		}
	}

	if field.Pattern != "" {
		matched, err := regexp.MatchString(field.Pattern, str)
		if err != nil {
			return parser.ValidationError{
				Field:   field.Name,
				Message: "invalid pattern",
			}
		}
		if !matched {
			return parser.ValidationError{
				Field:   field.Name,
				Message: "does not match required pattern",
			}
		}
	}

	return nil
}

func (v *Validator) validateNumber(field parser.Field, value any) error {
	var num float64

	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	default:
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a number",
		}
	}

	if field.Min != nil && num < float64(*field.Min) {
		return parser.ValidationError{
			Field:   field.Name,
			Message: fmt.Sprintf("must be at least %d", *field.Min),
		}
	}

	if field.Max != nil && num > float64(*field.Max) {
		return parser.ValidationError{
			Field:   field.Name,
			Message: fmt.Sprintf("must be at most %d", *field.Max),
		}
	}

	return nil
}

func (v *Validator) validateBoolean(field parser.Field, value any) error {
	_, ok := value.(bool)
	if !ok {
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a boolean",
		}
	}

	return nil
}

func (v *Validator) validateEmail(field parser.Field, value any) error {
	str, ok := value.(string)
	if !ok {
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a string",
		}
	}

	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, str)
	if !matched {
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a valid email address",
		}
	}

	return nil
}

func (v *Validator) validateURL(field parser.Field, value any) error {
	str, ok := value.(string)
	if !ok {
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a string",
		}
	}

	urlRegex := `^https?://[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(urlRegex, str)
	if !matched {
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a valid URL",
		}
	}

	return nil
}

func (v *Validator) validateEnum(field parser.Field, value any) error {
	str, ok := value.(string)
	if !ok {
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a string",
		}
	}

	for _, option := range field.Options {
		if str == option {
			return nil
		}
	}

	return parser.ValidationError{
		Field:   field.Name,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(field.Options, ", ")),
	}
}

func (v *Validator) validateDatetime(field parser.Field, value any) error {
	_, ok := value.(string)
	if !ok {
		return parser.ValidationError{
			Field:   field.Name,
			Message: "must be a string",
		}
	}

	return nil
}

func (v *Validator) getField(model *parser.Model, fieldName string) (*parser.Field, bool) {
	for _, field := range model.Fields {
		if field.Name == fieldName {
			return &field, true
		}
	}
	return nil, false
}

