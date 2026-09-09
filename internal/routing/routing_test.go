package routing

import (
	"path/filepath"
	"testing"
)

func TestGenerateRoute(t *testing.T) {
	pagesDir := filepath.Join("site", "pages")

	tests := []struct {
		source string
		want   string
	}{
		{filepath.Join(pagesDir, "index.md"), "/"},
		{filepath.Join(pagesDir, "about.md"), "/about/"},
		{filepath.Join(pagesDir, "blog", "post.md"), "/blog/post/"},
		{filepath.Join(pagesDir, "a", "b", "c.md"), "/a/b/c/"},
		{filepath.Join(pagesDir, "index", "index.md"), "/"},
	}

	for _, tt := range tests {
		if got := GenerateRoute(tt.source, pagesDir); got != tt.want {
			t.Errorf("GenerateRoute(%q, %q) = %q, want %q", tt.source, pagesDir, got, tt.want)
		}
	}
}

func TestOutputPath(t *testing.T) {
	tests := []struct {
		route string
		want  string
	}{
		{"/", "index.html"},
		{"/about/", filepath.Join("about", "index.html")},
		{"/blog/post/", filepath.Join("blog", "post", "index.html")},
	}

	for _, tt := range tests {
		if got := OutputPath(tt.route); got != tt.want {
			t.Errorf("OutputPath(%q) = %q, want %q", tt.route, got, tt.want)
		}
	}
}

func TestEnsureUniqueRoutes(t *testing.T) {
	pages := []struct {
		Route string
		Path  string
	}{
		{Route: "/", Path: "index.md"},
		{Route: "/about/", Path: "about.md"},
		{Route: "/", Path: "other.md"},
	}

	dups := EnsureUniqueRoutes(pages)
	if len(dups) != 1 || dups[0] != "other.md" {
		t.Errorf("expected [other.md], got %v", dups)
	}
}

func TestEnsureUniqueRoutesNoDuplicates(t *testing.T) {
	pages := []struct {
		Route string
		Path  string
	}{
		{Route: "/", Path: "index.md"},
		{Route: "/about/", Path: "about.md"},
	}

	dups := EnsureUniqueRoutes(pages)
	if len(dups) != 0 {
		t.Errorf("expected no duplicates, got %v", dups)
	}
}

func TestJoinRoute(t *testing.T) {
	tests := []struct {
		base, part, want string
	}{
		{"/blog", "post", "/blog/post"},
		{"/blog/", "post", "/blog/post"},
		{"/", "about", "/about"},
		{"/blog", "post/extra", "/blog/post/extra"},
	}

	for _, tt := range tests {
		if got := JoinRoute(tt.base, tt.part); got != tt.want {
			t.Errorf("JoinRoute(%q, %q) = %q, want %q", tt.base, tt.part, got, tt.want)
		}
	}
}

func TestCleanRoute(t *testing.T) {
	tests := []struct {
		route, want string
	}{
		{"/", "/"},
		{"/about/", "/about"},
		{"about", "/about"},
		{".", "/"},
		{"//foo//bar//", "/foo/bar"},
	}

	for _, tt := range tests {
		if got := CleanRoute(tt.route); got != tt.want {
			t.Errorf("CleanRoute(%q) = %q, want %q", tt.route, got, tt.want)
		}
	}
}
