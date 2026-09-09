package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	content := `
name: Test Site
base_url: http://localhost:8080
build:
  source: .
  output: dist
default_layout: default
default_page_type: page
`

	cfg, err := Load(writeTemp(t, "site.yaml", content))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Name != "Test Site" {
		t.Errorf("expected name 'Test Site', got '%s'", cfg.Name)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("expected base_url 'http://localhost:8080', got '%s'", cfg.BaseURL)
	}
	if cfg.Build.Output != "dist" {
		t.Errorf("expected output 'dist', got '%s'", cfg.Build.Output)
	}
	if cfg.DefaultLayout != "default" {
		t.Errorf("expected default_layout 'default', got '%s'", cfg.DefaultLayout)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg, err := Load(writeTemp(t, "site.yaml", "name: Minimal Site\n"))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Build.Source != "." {
		t.Errorf("expected default source '.', got '%s'", cfg.Build.Source)
	}
	if cfg.Build.Output != "dist" {
		t.Errorf("expected default output 'dist', got '%s'", cfg.Build.Output)
	}
	if cfg.DefaultLayout != "default" {
		t.Errorf("expected default layout 'default', got '%s'", cfg.DefaultLayout)
	}
	if cfg.DefaultPageType != "page" {
		t.Errorf("expected default page_type 'page', got '%s'", cfg.DefaultPageType)
	}
	if cfg.Language != "en" {
		t.Errorf("expected default language 'en', got '%s'", cfg.Language)
	}
	if cfg.Pagination.PerPage != 10 {
		t.Errorf("expected default per_page 10, got %d", cfg.Pagination.PerPage)
	}
	if cfg.Feed.Path != "/feed.xml" {
		t.Errorf("expected default feed path '/feed.xml', got '%s'", cfg.Feed.Path)
	}
}

func TestSourcePath(t *testing.T) {
	cfg := &SiteConfig{
		Build: BuildConfig{Source: "/project"},
	}

	path := cfg.SourcePath("components", "test.html")
	expected := "/project/components/test.html"
	if path != expected {
		t.Errorf("expected '%s', got '%s'", expected, path)
	}
}

func TestOutputPath(t *testing.T) {
	cfg := &SiteConfig{
		Build: BuildConfig{Output: "public"},
	}

	path := cfg.OutputPath("index.html")
	expected := "public/index.html"
	if path != expected {
		t.Errorf("expected '%s', got '%s'", expected, path)
	}
}

// The legacy nav-links list form must keep working.
func TestLinksLegacyList(t *testing.T) {
	content := `
name: Test
links:
  - name: Home
    url: /
  - name: About
    url: /about/
`
	cfg, err := Load(writeTemp(t, "site.yaml", content))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(cfg.Links.Entries) != 2 {
		t.Fatalf("expected 2 link entries, got %d", len(cfg.Links.Entries))
	}
	if cfg.Links.Entries[0].Name != "Home" || cfg.Links.Entries[0].URL != "/" {
		t.Errorf("unexpected first link: %+v", cfg.Links.Entries[0])
	}
	if cfg.Links.FailOnBroken {
		t.Error("expected fail_on_broken to be false by default")
	}
}

func TestLinksWithFailOnBroken(t *testing.T) {
	content := `
name: Test
links:
  fail_on_broken: true
  entries:
    - name: Home
      url: /
`
	cfg, err := Load(writeTemp(t, "site.yaml", content))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if !cfg.Links.FailOnBroken {
		t.Error("expected fail_on_broken true")
	}
	if len(cfg.Links.Entries) != 1 || cfg.Links.Entries[0].URL != "/" {
		t.Errorf("unexpected entries: %+v", cfg.Links.Entries)
	}
}

func TestFeatureSections(t *testing.T) {
	content := `
name: Test
sitemap:
  enabled: true
feed:
  enabled: true
  path: /rss.xml
  title: My Feed
taxonomy:
  tags: true
  categories: true
pagination:
  enabled: true
  per_page: 5
gallery:
  enabled: true
`
	cfg, err := Load(writeTemp(t, "site.yaml", content))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if !cfg.Sitemap.Enabled {
		t.Error("expected sitemap.enabled true")
	}
	if !cfg.Feed.Enabled {
		t.Error("expected feed.enabled true")
	}
	if cfg.Feed.Path != "/rss.xml" {
		t.Errorf("expected feed path '/rss.xml', got '%s'", cfg.Feed.Path)
	}
	if cfg.Feed.Title != "My Feed" {
		t.Errorf("expected feed title 'My Feed', got '%s'", cfg.Feed.Title)
	}
	if !cfg.Taxonomy.Tags || !cfg.Taxonomy.Categories {
		t.Error("expected taxonomy tags/categories enabled")
	}
	if !cfg.Pagination.Enabled || cfg.Pagination.PerPage != 5 {
		t.Errorf("unexpected pagination: %+v", cfg.Pagination)
	}
	if !cfg.Gallery.Enabled {
		t.Error("expected gallery.enabled true")
	}
}
