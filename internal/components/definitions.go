package components

import (
	"fmt"
	"os"
	"strings"

	"github.com/garterscopic/garterscopic/internal/config"
	"github.com/garterscopic/garterscopic/internal/yaml"
)

type Definition struct {
	Name     string                      `yaml:"name"`
	File     string                      `yaml:"file"`
	Style    string                      `yaml:"style"`
	Position string                      `yaml:"position"`
	Options  map[string]OptionDefinition `yaml:"options"`
}

type OptionDefinition struct {
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
	Default  any    `yaml:"default"`
}

type OptionBinding struct {
	From string `yaml:"from"`
}

type Definitions struct {
	Components map[string]ComponentDef `yaml:"components"`
}

type ComponentDef struct {
	File     string                      `yaml:"file"`
	Style    string                      `yaml:"style"`
	Position string                      `yaml:"position"`
	Options  map[string]OptionDefinition `yaml:"options"`
}

type Instance struct {
	Name     string                   `yaml:"name"`
	Options  map[string]any           `yaml:"options"`
	Position string                   `yaml:"position"`
	Bindings map[string]OptionBinding `yaml:"bindings"`
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
			Name:     name,
			File:     comp.File,
			Style:    comp.Style,
			Position: comp.Position,
			Options:  comp.Options,
		}
	}

	return result, nil
}

func (d *Definition) ValidateOptions(instance Instance) []error {
	var errs []error

	// Validate position if provided
	if instance.Position != "" {
		if !IsValidPosition(instance.Position) {
			errs = append(errs, fmt.Errorf("component '%s': invalid position '%s' (must be: top, bottom, left, right, center)", d.Name, instance.Position))
		}
	}

	// Check for unknown options
	for name := range instance.Options {
		if _, exists := d.Options[name]; !exists {
			errs = append(errs, fmt.Errorf("component '%s': unknown option '%s'", d.Name, name))
		}
	}

	// Validate declared options
	for name, optDef := range d.Options {
		val, exists := instance.Options[name]

		if optDef.Required && !exists {
			errs = append(errs, fmt.Errorf("component '%s': required option '%s' is missing (expected type: %s)", d.Name, name, optDef.Type))
			continue
		}

		if !exists {
			continue
		}

		if err := validateType(val, optDef.Type, d.Name, name); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (d *Definition) ResolvePosition(instance Instance) string {
	// Resolution order: instance position -> component definition default -> center
	if instance.Position != "" && IsValidPosition(instance.Position) {
		return instance.Position
	}
	if d.Position != "" && IsValidPosition(d.Position) {
		return d.Position
	}
	return "center"
}

func IsValidPosition(pos string) bool {
	switch strings.ToLower(pos) {
	case "top", "bottom", "left", "right", "center":
		return true
	}
	return false
}

func validateType(val any, expectedType, componentName, fieldName string) error {
	switch expectedType {
	case "string":
		if _, ok := val.(string); !ok {
			if _, ok := val.(int); ok {
				return nil // Allow coercion from int to string
			}
			return fmt.Errorf("component '%s', option '%s': expected string, got %T", componentName, fieldName, val)
		}
	case "number":
		switch val.(type) {
		case int, int32, int64, float32, float64:
		default:
			return fmt.Errorf("component '%s', option '%s': expected number, got %T", componentName, fieldName, val)
		}
	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("component '%s', option '%s': expected boolean, got %T", componentName, fieldName, val)
		}
	case "list":
		if _, ok := val.([]any); !ok {
			return fmt.Errorf("component '%s', option '%s': expected list, got %T", componentName, fieldName, val)
		}
	case "object":
		if _, ok := val.(map[string]any); !ok {
			return fmt.Errorf("component '%s', option '%s': expected object, got %T", componentName, fieldName, val)
		}
	case "date":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("component '%s', option '%s': expected date string, got %T", componentName, fieldName, val)
		}
	case "html":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("component '%s', option '%s': expected string for html type, got %T", componentName, fieldName, val)
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
