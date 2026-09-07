package layouts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/garterscopic/garterscopic/internal/components"
	"github.com/garterscopic/garterscopic/internal/config"
	"github.com/garterscopic/garterscopic/internal/yaml"
)

type Layout struct {
	Name       string                `yaml:"name"`
	Components []components.Instance `yaml:"components"`
}

type rawLayout struct {
	Name       string          `yaml:"name"`
	Components []components.Instance `yaml:"components"`
}

func LoadLayouts(cfg *config.SiteConfig) (map[string]Layout, error) {
	dir := cfg.SourcePath("layouts")

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]Layout), nil
		}
		return nil, fmt.Errorf("failed to read layouts directory: %w", err)
	}

	layouts := make(map[string]Layout)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read layout %s: %w", path, err)
		}

		var raw rawLayout
		if err := yaml.Parse(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse layout %s: %w", path, err)
		}

		if raw.Components == nil {
			raw.Components = []components.Instance{}
		}

		layouts[raw.Name] = Layout{
			Name:       raw.Name,
			Components: raw.Components,
		}
	}

	return layouts, nil
}

func (l *Layout) HasComponent(name string) bool {
	for _, c := range l.Components {
		if c.Name == name {
			return true
		}
	}
	return false
}
