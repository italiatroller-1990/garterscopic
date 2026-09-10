package styles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/config"
)

func TestCollectStyles(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "styles"), 0755)
	os.WriteFile(filepath.Join(dir, "styles", "global.css"), []byte("body{}"), 0644)
	os.WriteFile(filepath.Join(dir, "styles", "nav.css"), []byte(".nav{}"), 0644)

	cfg := &config.SiteConfig{Build: config.BuildConfig{Source: dir}}

	styles, err := CollectStyles(cfg, []string{"styles/nav.css"}, []string{"styles/global.css"})
	if err != nil {
		t.Fatal(err)
	}
	if len(styles) != 2 {
		t.Fatalf("expected 2 styles, got %d", len(styles))
	}
}

func TestCollectStylesDeduplicates(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "styles"), 0755)
	os.WriteFile(filepath.Join(dir, "styles", "a.css"), []byte("x"), 0644)

	cfg := &config.SiteConfig{Build: config.BuildConfig{Source: dir}}

	styles, err := CollectStyles(cfg, []string{"styles/a.css"}, []string{"styles/a.css"})
	if err != nil {
		t.Fatal(err)
	}
	if len(styles) != 1 {
		t.Errorf("expected 1 deduplicated style, got %d", len(styles))
	}
}

func TestCollectStylesSkipsMissing(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "styles"), 0755)
	os.WriteFile(filepath.Join(dir, "styles", "exists.css"), []byte("x"), 0644)

	cfg := &config.SiteConfig{Build: config.BuildConfig{Source: dir}}

	styles, err := CollectStyles(cfg, []string{"styles/exists.css", "styles/missing.css"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(styles) != 1 {
		t.Errorf("expected 1 style (missing skipped), got %d", len(styles))
	}
}

func TestLoadStyles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.css"), []byte("body{}"), 0644)
	os.WriteFile(filepath.Join(dir, "b.css"), []byte(".x{}"), 0644)

	data, err := LoadStyles([]string{"a.css", "b.css"}, dir)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !filepath.IsAbs(dir) {
		// Relative path test
		_ = content
	}
	if len(data) == 0 {
		t.Error("expected non-empty CSS output")
	}
}

func TestLoadStylesEmpty(t *testing.T) {
	data, err := LoadStyles(nil, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty output, got %d bytes", len(data))
	}
}

func TestCopyStyles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	os.WriteFile(filepath.Join(src, "main.css"), []byte("body{}"), 0644)

	err := CopyStyles([]string{"main.css"}, src, filepath.Join(dst, "out"))
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "out", "main.css"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "body{}" {
		t.Errorf("expected 'body{}', got %q", string(data))
	}
}

func TestGenerateResponsiveCSS(t *testing.T) {
	cfg := &config.SiteConfig{
		Responsive: config.ResponsiveConfig{
			Mobile:  "690px",
			Tablet:  "820px",
			Desktop: "1000px",
		},
	}

	css := GenerateResponsiveCSS(cfg)

	for _, want := range []string{
		"--gs-mobile: 690px",
		"--gs-tablet: 820px",
		"--gs-desktop: 1000px",
		":root {",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("expected CSS to contain %q, got:\n%s", want, css)
		}
	}
}

func TestGenerateResponsiveCSSDefaults(t *testing.T) {
	cfg := &config.SiteConfig{
		Responsive: config.ResponsiveConfig{
			Mobile:  "768px",
			Tablet:  "1024px",
			Desktop: "1200px",
		},
	}

	css := GenerateResponsiveCSS(cfg)

	for _, want := range []string{
		"--gs-mobile: 768px",
		"--gs-tablet: 1024px",
		"--gs-desktop: 1200px",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("expected CSS to contain %q, got:\n%s", want, css)
		}
	}
}

func TestGenerateResponsiveCSSRemUnits(t *testing.T) {
	cfg := &config.SiteConfig{
		Responsive: config.ResponsiveConfig{
			Mobile:  "30rem",
			Tablet:  "50rem",
			Desktop: "80rem",
		},
	}

	css := GenerateResponsiveCSS(cfg)

	for _, want := range []string{
		"--gs-mobile: 30rem",
		"--gs-tablet: 50rem",
		"--gs-desktop: 80rem",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("expected CSS to contain %q, got:\n%s", want, css)
		}
	}
}
