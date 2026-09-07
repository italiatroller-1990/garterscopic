package layouts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/garterscopic/garterscopic/internal/config"
	"github.com/garterscopic/garterscopic/internal/yaml"
)

type Position string

const (
	PositionTop    Position = "top"
	PositionBottom Position = "bottom"
	PositionLeft   Position = "left"
	PositionRight  Position = "right"
	PositionCenter Position = "center"
)

func ValidPosition(p Position) bool {
	switch p {
	case PositionTop, PositionBottom, PositionLeft, PositionRight, PositionCenter, "":
		return true
	}
	return false
}

type ComponentInstance struct {
	Name     string                 `yaml:"name"`
	Options  map[string]any         `yaml:"options"`
	Position Position               `yaml:"position"`
}

func (c *ComponentInstance) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var single string
	if err := unmarshal(&single); err == nil {
		c.Name = single
		c.Position = PositionCenter
		return nil
	}

	type rawComponent struct {
		Name     string         `yaml:"name"`
		Options  map[string]any `yaml:"options"`
		Position string         `yaml:"position"`
	}
	var raw rawComponent
	if err := unmarshal(&raw); err != nil {
		return err
	}
	c.Name = raw.Name
	c.Options = raw.Options
	if raw.Position == "" {
		c.Position = PositionCenter
	} else {
		c.Position = Position(raw.Position)
	}
	return nil
}

type Layout struct {
	Name       string                 `yaml:"name"`
	Components []ComponentInstance    `yaml:"components"`
	Inherit    string                 `yaml:"inherit"`
}

type rawLayout struct {
	Name       string              `yaml:"name"`
	Components []ComponentInstance `yaml:"components"`
	Inherit    string              `yaml:"inherit"`
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

	rawLayouts := make(map[string]rawLayout)
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
			raw.Components = []ComponentInstance{}
		}

		rawLayouts[raw.Name] = raw
	}

	resolved := make(map[string]Layout)
	visiting := make(map[string]bool)

	for name := range rawLayouts {
		if _, ok := resolved[name]; !ok {
			if err := resolveLayout(name, rawLayouts, resolved, visiting); err != nil {
				return nil, err
			}
		}
	}

	return resolved, nil
}

func resolveLayout(name string, raw map[string]rawLayout, resolved map[string]Layout, visiting map[string]bool) error {
	if visiting[name] {
		chain := []string{}
		for k := range visiting {
			if visiting[k] {
				chain = append(chain, k)
			}
		}
		chain = append(chain, name)
		return fmt.Errorf("circular layout inheritance detected:\n  %s", filepath.Join(chain...))
	}

	if layout, ok := resolved[name]; ok {
		_ = layout
		return nil
	}

	r, ok := raw[name]
	if !ok {
		return fmt.Errorf("layout '%s' not found", name)
	}

	if r.Inherit != "" {
		visiting[name] = true
		if err := resolveLayout(r.Inherit, raw, resolved, visiting); err != nil {
			return err
		}
		delete(visiting, name)

		parent, ok := resolved[r.Inherit]
		if !ok {
			return fmt.Errorf("parent layout '%s' not found for '%s'", r.Inherit, name)
		}

		merged := mergeLayouts(parent, r)
		resolved[name] = merged
	} else {
		resolved[name] = Layout{
			Name:       r.Name,
			Components: r.Components,
		}
	}

	return nil
}

func mergeLayouts(parent Layout, child rawLayout) Layout {
	return Layout{
		Name:       child.Name,
		Components: child.Components,
		Inherit:    child.Inherit,
	}
}

func (l *Layout) HasComponent(name string) bool {
	for _, c := range l.Components {
		if c.Name == name {
			return true
		}
	}
	return false
}

func (l *Layout) GetComponentByPosition(position Position) []ComponentInstance {
	var result []ComponentInstance
	for _, c := range l.Components {
		if c.Position == position {
			result = append(result, c)
		}
	}
	return result
}

func (l *Layout) GroupByPosition() map[Position][]ComponentInstance {
	groups := make(map[Position][]ComponentInstance)
	for _, c := range l.Components {
		pos := c.Position
		if pos == "" {
			pos = PositionCenter
		}
		groups[pos] = append(groups[pos], c)
	}
	return groups
}

func NormalizeInstance(inst ComponentInstance) ComponentInstance {
	if inst.Position == "" {
		inst.Position = PositionCenter
	}
	return inst
}
