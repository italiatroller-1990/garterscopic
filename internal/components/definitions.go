package components

import (
	"fmt"
	"os"
	"strings"

	"github.com/garterscopic/garterscopic/internal/config"
	"github.com/garterscopic/garterscopic/internal/yaml"
)

type Definition struct {
	Name    string                     `yaml:"name"`
	File    string                     `yaml:"file"`
	Style   string                     `yaml:"style"`
	Options map[string]OptionDefinition `yaml:"options"`
}

type OptionDefinition struct {
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
	Default  any    `yaml:"default"`
}

type Definitions struct {
	Components map[string]ComponentDef `yaml:"components"`
}

type ComponentDef struct {
	File    string                     `yaml:"file"`
	Style   string                     `yaml:"style"`
	Options map[string]OptionDefinition `yaml:"options"`
}

type Instance struct {
	Name    string         `yaml:"name"`
	Options map[string]any `yaml:"options"`
}

func LoadDefinitions(cfg *config.SiteConfig) (map[string]Definition, error) {
	path := cfg.SourcePath("components", "definitions.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read component definitions: %w", err)
	}

	var defs Definitions
	if err := yaml.Parse(data, &defs); err != nil {
		return nil, fmt.Errorf("failed to parse component definitions: %w", err)
	}

	result := make(map[string]Definition)
	for name, comp := range defs.Components {
		result[name] = Definition{
			Name:    name,
			File:    comp.File,
			Style:   comp.Style,
			Options: comp.Options,
		}
	}

	return result, nil
}

func (d *Definition) ValidateOptions(instance Instance) []error {
	var errs []error

	for name, optDef := range d.Options {
		val, exists := instance.Options[name]

		if optDef.Required && !exists {
			errs = append(errs, fmt.Errorf("required option '%s' is missing", name))
			continue
		}

		if !exists {
			continue
		}

		if err := validateType(val, optDef.Type, name); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func validateType(val any, expectedType, fieldName string) error {
	switch expectedType {
	case "string":
		if _, ok := val.(string); !ok {
			if _, ok := val.(int); ok {
				return nil
			}
			return fmt.Errorf("option '%s' must be a string, got %T", fieldName, val)
		}
	case "number":
		switch val.(type) {
		case int, int32, int64, float32, float64:
		default:
			return fmt.Errorf("option '%s' must be a number, got %T", fieldName, val)
		}
	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("option '%s' must be a boolean, got %T", fieldName, val)
		}
	case "list":
		if _, ok := val.([]any); !ok {
			return fmt.Errorf("option '%s' must be a list, got %T", fieldName, val)
		}
	case "object":
		if _, ok := val.(map[string]any); !ok {
			return fmt.Errorf("option '%s' must be an object, got %T", fieldName, val)
		}
	case "date":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("option '%s' must be a date string, got %T", fieldName, val)
		}
	case "html":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("option '%s' must be a string for html type, got %T", fieldName, val)
		}
	}
	return nil
}

func (d *Definition) GetOptionType(name string) string {
	if opt, ok := d.Options[name]; ok {
		return opt.Type
	}
	return "string"
}

func LoadComponentHTML(cfg *config.SiteConfig, file string) (string, error) {
	path := cfg.SourcePath("components", file)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read component file %s: %w", file, err)
	}
	return string(data), nil
}

func IsComponentUsed(name string, instances []Instance) bool {
	for _, inst := range instances {
		if inst.Name == name {
			return true
		}
	}
	return false
}

func GetUsedStyles(defs map[string]Definition, instances []Instance) []string {
	var styles []string
	seen := make(map[string]bool)

	for _, inst := range instances {
		if def, ok := defs[inst.Name]; ok && def.Style != "" && !seen[def.Style] {
			styles = append(styles, def.Style)
			seen[def.Style] = true
		}
	}

	sortStrings(styles)
	return styles
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func NormalizeComponentName(name string) string {
	return strings.TrimSuffix(name, ".html")
}
