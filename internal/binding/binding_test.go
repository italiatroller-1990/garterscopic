package binding

import "testing"

func TestResolveFrontmatter(t *testing.T) {
	resolver := NewResolver(
		map[string]any{
			"title": "Hello",
			"tags": []any{"a", "b"},
		},
		PageInfo{},
		SiteInfo{},
	)

	val, err := resolver.Resolve("frontmatter.title")
	if err != nil {
		t.Fatal(err)
	}
	if val != "Hello" {
		t.Errorf("expected 'Hello', got %v", val)
	}

	val, err = resolver.Resolve("frontmatter.tags")
	if err != nil {
		t.Fatal(err)
	}
	tags, ok := val.([]any)
	if !ok || len(tags) != 2 {
		t.Errorf("expected 2 tags, got %v", val)
	}
}

func TestResolvePage(t *testing.T) {
	resolver := NewResolver(
		nil,
		PageInfo{Route: "/about/", URL: "/about/", Type: "page"},
		SiteInfo{},
	)

	tests := []struct {
		path     string
		expected any
	}{
		{"page.route", "/about/"},
		{"page.url", "/about/"},
		{"page.type", "page"},
	}

	for _, tt := range tests {
		val, err := resolver.Resolve(tt.path)
		if err != nil {
			t.Errorf("Resolve(%s) error: %v", tt.path, err)
			continue
		}
		if val != tt.expected {
			t.Errorf("Resolve(%s) = %v, want %v", tt.path, val, tt.expected)
		}
	}
}

func TestResolveSite(t *testing.T) {
	resolver := NewResolver(
		nil,
		PageInfo{},
		SiteInfo{Name: "My Site", BaseURL: "https://example.com"},
	)

	val, err := resolver.Resolve("site.name")
	if err != nil {
		t.Fatal(err)
	}
	if val != "My Site" {
		t.Errorf("expected 'My Site', got %v", val)
	}

	val, err = resolver.Resolve("site.base_url")
	if err != nil {
		t.Fatal(err)
	}
	if val != "https://example.com" {
		t.Errorf("expected 'https://example.com', got %v", val)
	}
}

func TestResolveUnknownNamespace(t *testing.T) {
	resolver := NewResolver(nil, PageInfo{}, SiteInfo{})
	_, err := resolver.Resolve("unknown.field")
	if err == nil {
		t.Error("expected error for unknown namespace")
	}
}

func TestApplyBindings(t *testing.T) {
	resolver := NewResolver(
		map[string]any{
			"title": "My Title",
		},
		PageInfo{Route: "/"},
		SiteInfo{Name: "Test"},
	)

	options := map[string]any{
		"values": []any{
			"frontmatter.title",
			"site.name",
		},
	}

	result, err := ApplyBindings(options, resolver)
	if err != nil {
		t.Fatal(err)
	}

	if result["title"] != "My Title" {
		t.Errorf("expected title to be 'My Title', got %v", result["title"])
	}
	if result["name"] != "Test" {
		t.Errorf("expected name to be 'Test', got %v", result["name"])
	}
}

func TestApplyBindingsWithFrom(t *testing.T) {
	resolver := NewResolver(
		map[string]any{
			"date": "2026-01-01",
		},
		PageInfo{},
		SiteInfo{},
	)

	options := map[string]any{
		"from": map[string]any{
			"postDate": "frontmatter.date",
		},
	}

	result, err := ApplyBindings(options, resolver)
	if err != nil {
		t.Fatal(err)
	}

	if result["postDate"] != "2026-01-01" {
		t.Errorf("expected postDate to be '2026-01-01', got %v", result["postDate"])
	}
}
