package pagetypes

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/garterscopic/garterscopic/internal/config"
	"github.com/garterscopic/garterscopic/internal/yaml"
)

type PageType struct {
	Name   string              `yaml:"name"`
	Layout string              `yaml:"layout"`
	Fields map[string]FieldDef `yaml:"fields"`
}

type FieldDef struct {
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
	Default  any    `yaml:"default"`
}

type rawPageType struct {
	Name   string              `yaml:"name"`
	Layout string              `yaml:"layout"`
	Fields map[string]FieldDef `yaml:"fields"`
}

func LoadPageTypes(cfg *config.SiteConfig) (map[string]PageType, error) {
	dir := cfg.SourcePath("page-types")

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]PageType), nil
		}
		return nil, fmt.Errorf("failed to read page-types directory: %w", err)
	}

	types := make(map[string]PageType)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read page type %s: %w", path, err)
		}

		var raw rawPageType
		if err := yaml.Parse(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse page type %s: %w", path, err)
		}

		if raw.Fields == nil {
			raw.Fields = make(map[string]FieldDef)
		}

		types[raw.Name] = PageType{
			Name:   raw.Name,
			Layout: raw.Layout,
			Fields: raw.Fields,
		}
	}

	return types, nil
}

func (pt *PageType) ApplyDefaults(metadata map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range metadata {
		result[k] = v
	}

	for name, field := range pt.Fields {
		if _, exists := result[name]; !exists && field.Default != nil {
			result[name] = field.Default
		}
	}

	return result
}

func (pt *PageType) Validate(metadata map[string]any) []error {
	var errs []error

	for name, field := range pt.Fields {
		value, exists := metadata[name]

		if field.Required && !exists {
			errs = append(errs, fmt.Errorf("required field '%s' is missing", name))
			continue
		}

		if !exists {
			continue
		}

		if err := validateFieldType(value, field.Type, name); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func validateFieldType(value any, expectedType, fieldName string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			if _, ok := value.(int); ok {
				return nil
			}
			return fmt.Errorf("field '%s' must be a string, got %T", fieldName, value)
		}
	case "number":
		switch value.(type) {
		case int, int32, int64, float32, float64:
		default:
			return fmt.Errorf("field '%s' must be a number, got %T", fieldName, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' must be a boolean, got %T", fieldName, value)
		}
	case "list":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("field '%s' must be a list, got %T", fieldName, value)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("field '%s' must be an object, got %T", fieldName, value)
		}
	case "date":
		switch value.(type) {
		case string, time.Time:
		default:
			return fmt.Errorf("field '%s' must be a date string (YYYY-MM-DD), got %T", fieldName, value)
		}
	case "html":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' must be a string for html type, got %T", fieldName, value)
		}
	}
	return nil
}
