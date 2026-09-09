// Package yaml wraps gopkg.in/yaml.v3 with concise helpers.
package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Parse unmarshals YAML data into target, wrapping any error with context.
func Parse(data []byte, target any) error {
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("YAML parse error: %w", err)
	}
	return nil
}

func Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}
