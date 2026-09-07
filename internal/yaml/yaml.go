package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func Parse(data []byte, target any) error {
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("YAML parse error: %w", err)
	}
	return nil
}

func Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}
