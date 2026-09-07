package config

import (
	"os"
	"testing"
)

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

	tmpfile, err := os.CreateTemp("", "site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
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
	content := `
name: Minimal Site
`

	tmpfile, err := os.CreateTemp("", "site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
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
