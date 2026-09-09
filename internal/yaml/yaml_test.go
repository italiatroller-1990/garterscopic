package yaml

import (
	"testing"
)

func TestParse(t *testing.T) {
	data := []byte(`name: test
count: 42
tags:
  - a
  - b
`)

	var result struct {
		Name  string   `yaml:"name"`
		Count int      `yaml:"count"`
		Tags  []string `yaml:"tags"`
	}

	if err := Parse(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Name != "test" {
		t.Errorf("expected name 'test', got %q", result.Name)
	}
	if result.Count != 42 {
		t.Errorf("expected count 42, got %d", result.Count)
	}
	if len(result.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(result.Tags))
	}
}

func TestParseInvalid(t *testing.T) {
	data := []byte(`: invalid yaml {{{`)

	var result map[string]any
	if err := Parse(data, &result); err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestMarshal(t *testing.T) {
	input := map[string]any{
		"name":  "test",
		"count": 42,
	}

	data, err := Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty output")
	}

	// Verify round-trip
	var output map[string]any
	if err := Parse(data, &output); err != nil {
		t.Fatal(err)
	}
	if output["name"] != "test" {
		t.Errorf("expected name 'test', got %v", output["name"])
	}
}

func TestParseNested(t *testing.T) {
	data := []byte(`
components:
  card:
    file: card.html
    options:
      title:
        type: string
        required: true
`)

	var result struct {
		Components map[string]struct {
			File    string `yaml:"file"`
			Options map[string]struct {
				Type     string `yaml:"type"`
				Required bool   `yaml:"required"`
			} `yaml:"options"`
		} `yaml:"components"`
	}

	if err := Parse(data, &result); err != nil {
		t.Fatal(err)
	}

	card, ok := result.Components["card"]
	if !ok {
		t.Fatal("expected 'card' component")
	}
	if card.File != "card.html" {
		t.Errorf("expected file 'card.html', got %q", card.File)
	}
	title, ok := card.Options["title"]
	if !ok {
		t.Fatal("expected 'title' option")
	}
	if !title.Required {
		t.Error("expected 'title' to be required")
	}
}
