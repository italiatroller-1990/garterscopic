package pagetypes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/config"
)

func TestLoadPageTypes(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "page-types"), 0755)
	os.WriteFile(filepath.Join(dir, "page-types", "page.yaml"), []byte(`name: page
layout: default
fields:
  title:
    type: string
    required: true
  draft:
    type: boolean
    default: false
`), 0644)

	cfg := &config.SiteConfig{Build: config.BuildConfig{Source: dir}}
	types, err := LoadPageTypes(cfg)
	if err != nil {
		t.Fatal(err)
	}

	pt, ok := types["page"]
	if !ok {
		t.Fatal("expected 'page' type to exist")
	}
	if pt.Layout != "default" {
		t.Errorf("expected layout 'default', got %q", pt.Layout)
	}
	if len(pt.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(pt.Fields))
	}
	if !pt.Fields["title"].Required {
		t.Error("expected 'title' to be required")
	}
	if pt.Fields["draft"].Default != false {
		t.Error("expected 'draft' default to be false")
	}
}

func TestLoadPageTypesMissing(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.SiteConfig{Build: config.BuildConfig{Source: dir}}
	types, err := LoadPageTypes(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 0 {
		t.Errorf("expected empty map, got %d types", len(types))
	}
}

func TestApplyDefaults(t *testing.T) {
	pt := &PageType{
		Fields: map[string]FieldDef{
			"title": {Type: "string", Required: true},
			"color": {Type: "string", Default: "blue"},
		},
	}

	metadata := map[string]any{"title": "Hello"}
	result := pt.ApplyDefaults(metadata)

	if result["title"] != "Hello" {
		t.Error("expected existing value preserved")
	}
	if result["color"] != "blue" {
		t.Error("expected default applied")
	}
}

func TestApplyDefaultsDoesNotOverride(t *testing.T) {
	pt := &PageType{
		Fields: map[string]FieldDef{
			"color": {Type: "string", Default: "blue"},
		},
	}

	metadata := map[string]any{"color": "red"}
	result := pt.ApplyDefaults(metadata)

	if result["color"] != "red" {
		t.Error("expected existing value not overridden by default")
	}
}

func TestValidate(t *testing.T) {
	pt := &PageType{
		Fields: map[string]FieldDef{
			"title": {Type: "string", Required: true},
			"count": {Type: "number"},
			"tags":  {Type: "list"},
		},
	}

	errs := pt.Validate(map[string]any{
		"title": "Hello",
		"count": 42,
		"tags":  []any{"a", "b"},
	})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateMissingRequired(t *testing.T) {
	pt := &PageType{
		Fields: map[string]FieldDef{
			"title": {Type: "string", Required: true},
		},
	}

	errs := pt.Validate(map[string]any{})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Error() != "required field 'title' is missing" {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestValidateWrongType(t *testing.T) {
	pt := &PageType{
		Fields: map[string]FieldDef{
			"count": {Type: "number"},
		},
	}

	errs := pt.Validate(map[string]any{"count": "not a number"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestValidateFieldType(t *testing.T) {
	pt := &PageType{
		Fields: map[string]FieldDef{
			"name": {Type: "string"},
			"age":  {Type: "number"},
			"ok":   {Type: "boolean"},
			"list": {Type: "list"},
			"obj":  {Type: "object"},
			"date": {Type: "date"},
			"html": {Type: "html"},
		},
	}

	// All correct types should pass.
	errs := pt.Validate(map[string]any{
		"name": "test",
		"age":  25,
		"ok":   true,
		"list": []any{"a"},
		"obj":  map[string]any{"k": "v"},
		"date": "2024-01-15",
		"html": "<p>hi</p>",
	})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateIntCoercionToString(t *testing.T) {
	pt := &PageType{
		Fields: map[string]FieldDef{
			"name": {Type: "string"},
		},
	}

	errs := pt.Validate(map[string]any{"name": 42})
	if len(errs) != 0 {
		t.Errorf("expected int-to-string coercion, got %d errors: %v", len(errs), errs)
	}
}
