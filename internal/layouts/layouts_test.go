package layouts

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidPosition(t *testing.T) {
	tests := []struct {
		position Position
		expected bool
	}{
		{PositionTop, true},
		{PositionBottom, true},
		{PositionLeft, true},
		{PositionRight, true},
		{PositionCenter, true},
		{"", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		if got := ValidPosition(tt.position); got != tt.expected {
			t.Errorf("ValidPosition(%q) = %v, want %v", tt.position, got, tt.expected)
		}
	}
}

func TestNormalizeInstance(t *testing.T) {
	inst := ComponentInstance{Name: "test"}
	inst = NormalizeInstance(inst)
	if inst.Position != PositionCenter {
		t.Errorf("expected default position to be center, got %q", inst.Position)
	}

	inst2 := ComponentInstance{Name: "test", Position: PositionTop}
	inst2 = NormalizeInstance(inst2)
	if inst2.Position != PositionTop {
		t.Errorf("expected position to be top, got %q", inst2.Position)
	}
}

func TestGroupByPosition(t *testing.T) {
	layout := Layout{
		Components: []ComponentInstance{
			{Name: "a", Position: PositionTop},
			{Name: "b", Position: PositionCenter},
			{Name: "c", Position: PositionCenter},
			{Name: "d", Position: PositionBottom},
		},
	}

	groups := layout.GroupByPosition()
	if len(groups[PositionTop]) != 1 {
		t.Errorf("expected 1 top component, got %d", len(groups[PositionTop]))
	}
	if len(groups[PositionCenter]) != 2 {
		t.Errorf("expected 2 center components, got %d", len(groups[PositionCenter]))
	}
	if len(groups[PositionBottom]) != 1 {
		t.Errorf("expected 1 bottom component, got %d", len(groups[PositionBottom]))
	}
}

func TestLoadLayoutsWithInheritance(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "layouts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	layoutsDir := filepath.Join(tmpdir, "layouts")
	os.MkdirAll(layoutsDir, 0755)

	os.WriteFile(filepath.Join(layoutsDir, "default.yaml"), []byte(`name: default
components:
  - name: navbar
    position: top
  - name: content
    position: center
  - name: footer
    position: bottom
`), 0644)

	os.WriteFile(filepath.Join(layoutsDir, "blog.yaml"), []byte(`name: blog
inherit: default.yaml
components:
  - name: post-header
    position: top
  - name: content
    position: center
  - name: footer
    position: bottom
`), 0644)

	cfg := &testSiteConfig{source: tmpdir}

	loaded, err := loadLayoutsForTest(cfg)
	if err != nil {
		t.Fatalf("failed to load layouts: %v", err)
	}

	if _, ok := loaded["default"]; !ok {
		t.Error("default layout not loaded")
	}
	if _, ok := loaded["blog"]; !ok {
		t.Error("blog layout not loaded")
	}
}

func TestLoadLayoutsCircularInheritance(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "layouts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	layoutsDir := filepath.Join(tmpdir, "layouts")
	os.MkdirAll(layoutsDir, 0755)

	os.WriteFile(filepath.Join(layoutsDir, "a.yaml"), []byte(`name: a
inherit: b.yaml
components: []
`), 0644)

	os.WriteFile(filepath.Join(layoutsDir, "b.yaml"), []byte(`name: b
inherit: a.yaml
components: []
`), 0644)

	cfg := &testSiteConfig{source: tmpdir}

	_, err = loadLayoutsForTest(cfg)
	if err == nil {
		t.Fatal("expected error for circular inheritance")
	}
}

type testSiteConfig struct {
	source string
}

func (c *testSiteConfig) SourcePath(parts ...string) string {
	return filepath.Join(append([]string{c.source}, parts...)...)
}

func loadLayoutsForTest(cfg *testSiteConfig) (map[string]Layout, error) {
	dir := cfg.SourcePath("layouts")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	rawLayouts := make(map[string]rawLayout)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var raw rawLayout
		if err := yamlParseForTest(data, &raw); err != nil {
			return nil, err
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

func yamlParseForTest(data []byte, target any) error {
	return yaml.Unmarshal(data, target)
}
