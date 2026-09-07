package renderer

import (
	"testing"

	"github.com/garterscopic/garterscopic/internal/components"
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
