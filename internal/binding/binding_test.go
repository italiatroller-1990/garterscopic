package binding

import "testing"

func TestResolveFrontmatter(t *testing.T) {
	resolver := NewResolver(
		map[string]any{
			"title": "Hello",
			"tags":  []any{"a", "b"},
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

func TestResolvePosts(t *testing.T) {
	content := ContentInfo{
		Posts: []map[string]any{
			{"title": "Post 1", "route": "/blog/post1/"},
			{"title": "Post 2", "route": "/blog/post2/"},
		},
	}

	resolver := NewResolverWithContent(nil, PageInfo{}, SiteInfo{}, content)

	val, err := resolver.Resolve("posts.all")
	if err != nil {
		t.Fatal(err)
	}

	posts, ok := val.([]map[string]any)
	if !ok {
		t.Fatalf("expected []map[string]any, got %T", val)
	}
	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}
}

func TestResolveSections(t *testing.T) {
	content := ContentInfo{
		Sections: []string{"blog", "guides"},
		SectionPosts: map[string][]map[string]any{
			"blog": {
				{"title": "Blog Post", "route": "/blog/post/"},
			},
			"guides": {
				{"title": "Guide", "route": "/guides/guide/"},
			},
		},
	}

	resolver := NewResolverWithContent(nil, PageInfo{}, SiteInfo{}, content)

	val, err := resolver.Resolve("sections.all")
	if err != nil {
		t.Fatal(err)
	}
	sections, ok := val.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", val)
	}
	if len(sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(sections))
	}

	val, err = resolver.Resolve("sections.blog")
	if err != nil {
		t.Fatal(err)
	}
	posts, ok := val.([]map[string]any)
	if !ok {
		t.Fatalf("expected []map[string]any, got %T", val)
	}
	if len(posts) != 1 {
		t.Errorf("expected 1 post in blog section, got %d", len(posts))
	}
}

func TestResolveTags(t *testing.T) {
	content := ContentInfo{
		Tags: []string{"linux", "tutorial"},
		TagIndex: map[string][]map[string]any{
			"linux": {
				{"title": "Linux Post", "route": "/blog/linux/"},
			},
			"tutorial": {
				{"title": "Tutorial 1", "route": "/blog/tutorial1/"},
				{"title": "Tutorial 2", "route": "/guides/tutorial2/"},
			},
		},
	}

	resolver := NewResolverWithContent(nil, PageInfo{}, SiteInfo{}, content)

	val, err := resolver.Resolve("tags.all")
	if err != nil {
		t.Fatal(err)
	}
	tags, ok := val.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", val)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}

	val, err = resolver.Resolve("tags.linux")
	if err != nil {
		t.Fatal(err)
	}
	posts, ok := val.([]map[string]any)
	if !ok {
		t.Fatalf("expected []map[string]any, got %T", val)
	}
	if len(posts) != 1 {
		t.Errorf("expected 1 linux post, got %d", len(posts))
	}

	val, err = resolver.Resolve("tags.tutorial")
	if err != nil {
		t.Fatal(err)
	}
	posts, ok = val.([]map[string]any)
	if !ok {
		t.Fatalf("expected []map[string]any, got %T", val)
	}
	if len(posts) != 2 {
		t.Errorf("expected 2 tutorial posts, got %d", len(posts))
	}
}

func TestResolveUnknownTag(t *testing.T) {
	content := ContentInfo{
		Tags:     []string{"linux"},
		TagIndex: map[string][]map[string]any{},
	}

	resolver := NewResolverWithContent(nil, PageInfo{}, SiteInfo{}, content)

	_, err := resolver.Resolve("tags.nonexistent")
	if err == nil {
		t.Error("expected error for unknown tag")
	}
}

func TestResolveUnknownSection(t *testing.T) {
	content := ContentInfo{
		Sections:     []string{"blog"},
		SectionPosts: map[string][]map[string]any{},
	}

	resolver := NewResolverWithContent(nil, PageInfo{}, SiteInfo{}, content)

	_, err := resolver.Resolve("sections.nonexistent")
	if err == nil {
		t.Error("expected error for unknown section")
	}
}

func TestResolvePostsUnknownField(t *testing.T) {
	content := ContentInfo{}

	resolver := NewResolverWithContent(nil, PageInfo{}, SiteInfo{}, content)

	_, err := resolver.Resolve("posts.nonexistent")
	if err == nil {
		t.Error("expected error for unknown posts field")
	}
}
