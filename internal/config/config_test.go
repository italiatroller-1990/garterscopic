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

func TestLinksHighlightActive(t *testing.T) {
	content := `
name: Test
links:
  highlight_active: true
  entries:
    - name: Home
      url: /
    - name: About
      url: /about/
`
	cfg, err := Load(writeTemp(t, "site.yaml", content))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if !cfg.Links.HighlightActive {
		t.Error("expected highlight_active true")
	}
	if len(cfg.Links.Entries) != 2 {
		t.Errorf("unexpected entries: %+v", cfg.Links.Entries)
	}
}

func TestLinksHighlightActiveDefaultsFalse(t *testing.T) {
	content := `
name: Test
links:
  - name: Home
    url: /
`
	cfg, err := Load(writeTemp(t, "site.yaml", content))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Links.HighlightActive {
		t.Error("expected highlight_active to be false by default")
	}
}

func TestLinksHighlightColor(t *testing.T) {
	content := `
name: Test
links:
  highlight_active: true
  highlight_color: "#ffffff"
  entries:
    - name: Home
      url: /
`
	cfg, err := Load(writeTemp(t, "site.yaml", content))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Links.HighlightColor != "#ffffff" {
		t.Errorf("expected highlight_color %q, got %q", "#ffffff", cfg.Links.HighlightColor)
	}
}

func TestNormalizeHighlightColor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty", input: "", want: ""},
		{name: "hex six", input: "#ffffff", want: "#ffffff"},
		{name: "hex three", input: "#fff", want: "#fff"},
		{name: "hex uppercase", input: "#FFAA00", want: "#FFAA00"},
		{name: "rgb triplet", input: "12, 69, 11", want: "rgb(12, 69, 11)"},
		{name: "rgb no spaces", input: "12,69,11", want: "rgb(12, 69, 11)"},
		{name: "rgb boundary", input: "0, 0, 0", want: "rgb(0, 0, 0)"},
		{name: "rgb max", input: "255,255,255", want: "rgb(255, 255, 255)"},
		{name: "named color rejected", input: "red", wantErr: true},
		{name: "short hex rejected", input: "#12", wantErr: true},
		{name: "long hex rejected", input: "#12345", wantErr: true},
		{name: "non-hex rejected", input: "#gggggg", wantErr: true},
		{name: "missing hash rejected", input: "ffffff", wantErr: true},
		{name: "rgb out of range", input: "300, 0, 0", wantErr: true},
		{name: "rgb too few parts", input: "12, 69", wantErr: true},
		{name: "rgb non-numeric", input: "a, b, c", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeHighlightColor(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %q, got %q", tt.input, got)
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("unexpected error for %q: %v", tt.input, err)
				} else if got != tt.want {
					t.Errorf("expected %q, got %q", tt.want, got)
				}
			}
		})
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

func TestResponsiveConfigParsed(t *testing.T) {
	content := `
name: Test
responsive:
  mobile: "690px"
  tablet: "820px"
  desktop: "1000px"
  viewport: "1200"
`
	cfg, err := Load(writeTemp(t, "site.yaml", content))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Responsive.Mobile != "690px" {
		t.Errorf("expected mobile '690px', got '%s'", cfg.Responsive.Mobile)
	}
	if cfg.Responsive.Tablet != "820px" {
		t.Errorf("expected tablet '820px', got '%s'", cfg.Responsive.Tablet)
	}
	if cfg.Responsive.Desktop != "1000px" {
		t.Errorf("expected desktop '1000px', got '%s'", cfg.Responsive.Desktop)
	}
	if cfg.Responsive.Viewport != "1200" {
		t.Errorf("expected viewport '1200', got '%s'", cfg.Responsive.Viewport)
	}
}

func TestResponsiveConfigDefaults(t *testing.T) {
	cfg, err := Load(writeTemp(t, "site.yaml", "name: Test\n"))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Responsive.Mobile != "768px" {
		t.Errorf("expected default mobile '768px', got '%s'", cfg.Responsive.Mobile)
	}
	if cfg.Responsive.Tablet != "1024px" {
		t.Errorf("expected default tablet '1024px', got '%s'", cfg.Responsive.Tablet)
	}
	if cfg.Responsive.Desktop != "1200px" {
		t.Errorf("expected default desktop '1200px', got '%s'", cfg.Responsive.Desktop)
	}
	if cfg.Responsive.Viewport != "device-width" {
		t.Errorf("expected default viewport 'device-width', got '%s'", cfg.Responsive.Viewport)
	}
}

func TestResponsiveValidationOrdering(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SiteConfig
		wantErr bool
	}{
		{
			name: "valid ordering",
			cfg: SiteConfig{
				Responsive: ResponsiveConfig{Mobile: "690px", Tablet: "820px", Desktop: "1000px"},
			},
			wantErr: false,
		},
		{
			name: "default ordering",
			cfg: SiteConfig{
				Responsive: ResponsiveConfig{Mobile: "768px", Tablet: "1024px", Desktop: "1200px"},
			},
			wantErr: false,
		},
		{
			name: "invalid - mobile >= tablet",
			cfg: SiteConfig{
				Responsive: ResponsiveConfig{Mobile: "900px", Tablet: "500px", Desktop: "1000px"},
			},
			wantErr: true,
		},
		{
			name: "invalid - tablet >= desktop",
			cfg: SiteConfig{
				Responsive: ResponsiveConfig{Mobile: "690px", Tablet: "1200px", Desktop: "1000px"},
			},
			wantErr: true,
		},
		{
			name: "invalid - mobile == tablet",
			cfg: SiteConfig{
				Responsive: ResponsiveConfig{Mobile: "800px", Tablet: "800px", Desktop: "1000px"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateResponsive()
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestResponsiveValidationInvalidCSSLength(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"no unit", "768"},
		{"letters", "mobile"},
		{"empty", ""},
		{"negative", "-100px"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := SiteConfig{
				Responsive: ResponsiveConfig{
					Mobile:  tt.value,
					Tablet:  "820px",
					Desktop: "1000px",
				},
			}
			err := cfg.ValidateResponsive()
			if err == nil {
				t.Errorf("expected validation error for %q, got none", tt.value)
			}
		})
	}
}

func TestResponsiveValidationRemUnits(t *testing.T) {
	cfg := SiteConfig{
		Responsive: ResponsiveConfig{Mobile: "30rem", Tablet: "50rem", Desktop: "80rem"},
	}
	if err := cfg.ValidateResponsive(); err != nil {
		t.Errorf("expected rem units to be valid, got: %v", err)
	}
}

func TestIsCSSLength(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{"768px", true},
		{"10.5rem", true},
		{"50em", true},
		{"100vw", true},
		{"80%", true},
		{"  768px  ", true},
		{"768", false},
		{"px", false},
		{"", false},
		{"abc", false},
	}

	for _, tt := range tests {
		if got := isValidCSSLength(tt.value); got != tt.valid {
			t.Errorf("isValidCSSLength(%q) = %v, want %v", tt.value, got, tt.valid)
		}
	}
}
