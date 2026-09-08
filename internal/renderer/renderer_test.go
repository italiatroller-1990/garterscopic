package renderer

import (
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/components"
)

func TestRenderComponentHTML(t *testing.T) {
	r := New(map[string]components.Definition{})

	html := `<h1>{{ title }}</h1><p>{{ content }}</p>`
	options := map[string]any{
		"title":   "Hello World",
		"content": "This is a test",
	}

	result, err := r.RenderComponentHTML("test", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<h1>Hello World</h1><p>This is a test</p>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderWithMissingPlaceholder(t *testing.T) {
	r := New(map[string]components.Definition{})

	html := `<h1>{{ title }}</h1><p>{{ missing }}</p>`
	options := map[string]any{
		"title": "Hello",
	}

	result, err := r.RenderComponentHTML("test", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<h1>Hello</h1><p>{{ missing }}</p>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderWithEscaping(t *testing.T) {
	r := New(map[string]components.Definition{})

	html := `<h1>{{ title }}</h1>`
	options := map[string]any{
		"title": "<script>alert('xss')</script>",
	}

	result, err := r.RenderComponentHTML("test", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<h1>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</h1>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderListType(t *testing.T) {
	r := New(map[string]components.Definition{})

	html := `<ul>{{ items }}</ul>`
	itemsList := []any{"apple", "banana", "cherry"}
	options := map[string]any{
		"items": itemsList,
	}

	result, err := r.RenderComponentHTML("test", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestRenderHTMLType(t *testing.T) {
	defs := map[string]components.Definition{
		"hero": {
			Name: "hero",
			Options: map[string]components.OptionDefinition{
				"body": {Type: "html"},
			},
		},
	}
	r := New(defs)

	html := `<div>{{ body }}</div>`
	options := map[string]any{
		"body": "<strong>Bold</strong>",
	}

	result, err := r.RenderComponentHTML("hero", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<div><strong>Bold</strong></div>"
	if result != expected {
		t.Errorf("html type: expected %q, got %q", expected, result)
	}
}

func TestRenderStringTypeEscapesHTML(t *testing.T) {
	defs := map[string]components.Definition{
		"card": {
			Name: "card",
			Options: map[string]components.OptionDefinition{
				"title": {Type: "string"},
			},
		},
	}
	r := New(defs)

	html := `<h1>{{ title }}</h1>`
	options := map[string]any{
		"title": "<script>alert('xss')</script>",
	}

	result, err := r.RenderComponentHTML("card", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if contains(result, "<script>") {
		t.Errorf("string type should escape HTML, got: %q", result)
	}

	if !contains(result, "&lt;script&gt;") {
		t.Errorf("expected escaped HTML, got: %q", result)
	}
}

func TestRenderBooleanType(t *testing.T) {
	r := New(map[string]components.Definition{})

	tests := []struct {
		value    bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}

	for _, tt := range tests {
		html := `<div>{{ flag }}</div>`
		options := map[string]any{
			"flag": tt.value,
		}

		result, err := r.RenderComponentHTML("test", html, options)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !contains(result, tt.expected) {
			t.Errorf("boolean render: expected %q in %q", tt.expected, result)
		}
	}
}

func TestRenderNumberType(t *testing.T) {
	r := New(map[string]components.Definition{})

	tests := []struct {
		value    any
		expected string
	}{
		{42, "42"},
		{3.14, "3.14"},
		{int64(999), "999"},
	}

	for _, tt := range tests {
		html := `<span>{{ count }}</span>`
		options := map[string]any{
			"count": tt.value,
		}

		result, err := r.RenderComponentHTML("test", html, options)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !contains(result, tt.expected) {
			t.Errorf("number render: expected %q in %q", tt.expected, result)
		}
	}
}

func TestRenderPlaceholderWithWhitespace(t *testing.T) {
	r := New(map[string]components.Definition{})

	tests := []struct {
		html     string
		expected string
	}{
		{`<div>{{ title }}</div>`, "Hello"},
		{`<div>{{title}}</div>`, "Hello"},
		{`<div>{{  title  }}</div>`, "Hello"},
		{`<div>{{ title   }}</div>`, "Hello"},
	}

	for _, tt := range tests {
		options := map[string]any{
			"title": "Hello",
		}

		result, err := r.RenderComponentHTML("test", tt.html, options)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !contains(result, tt.expected) {
			t.Errorf("whitespace handling: expected %q in %q", tt.expected, result)
		}
	}
}

func TestRenderMultiplePlaceholders(t *testing.T) {
	r := New(map[string]components.Definition{})

	html := `<h1>{{ title }}</h1><p>{{ subtitle }}</p><span>{{ author }}</span>`
	options := map[string]any{
		"title":    "My Article",
		"subtitle": "A great read",
		"author":   "Alice",
	}

	result, err := r.RenderComponentHTML("test", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(result, "My Article") || !contains(result, "A great read") || !contains(result, "Alice") {
		t.Errorf("multiple placeholders not all replaced: %q", result)
	}
}

func TestRenderDateType(t *testing.T) {
	r := New(map[string]components.Definition{})

	html := `<time>{{ published }}</time>`
	options := map[string]any{
		"published": "2024-01-15",
	}

	result, err := r.RenderComponentHTML("test", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(result, "2024-01-15") {
		t.Errorf("date type: expected date string in %q", result)
	}
}

func TestRenderObjectType(t *testing.T) {
	r := New(map[string]components.Definition{})

	html := `<div>{{ metadata }}</div>`
	options := map[string]any{
		"metadata": map[string]any{
			"text": "Important",
			"key":  "value",
		},
	}

	result, err := r.RenderComponentHTML("test", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("object type: expected non-empty result")
	}
}

func TestRenderLinksType(t *testing.T) {
	defs := map[string]components.Definition{
		"navbar": {
			Name: "navbar",
			Options: map[string]components.OptionDefinition{
				"links": {Type: "links"},
			},
		},
	}
	r := New(defs)

	html := `<ul>{{ links }}</ul>`
	linksList := []any{
		map[string]any{"name": "Home", "url": "/"},
		map[string]any{"name": "About", "url": "/about"},
		map[string]any{"name": "Contact", "url": "/contact"},
	}
	options := map[string]any{
		"links": linksList,
	}

	result, err := r.RenderComponentHTML("navbar", html, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(result, "<li><a href=\"/\">Home</a></li>") {
		t.Errorf("links type: expected Home link, got %q", result)
	}
	if !contains(result, "<li><a href=\"/about\">About</a></li>") {
		t.Errorf("links type: expected About link, got %q", result)
	}
	if !contains(result, "<li><a href=\"/contact\">Contact</a></li>") {
		t.Errorf("links type: expected Contact link, got %q", result)
	}
}

func contains(s, substr string) bool {
	// Simple substring check
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
