package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/config"
)

func TestBuildSite(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "garterscopic_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	createTestSite(t, tmpdir)

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(workingDir)

	cfg, err := config.Load("site.yaml")
	if err != nil {
		t.Fatal(err)
	}

	builder := New(cfg)
	if err := builder.Load(); err != nil {
		t.Fatal(err)
	}
	if buildErr := builder.Validate(); buildErr != nil {
		t.Fatal(buildErr)
	}
	if _, err := builder.Build(); err != nil {
		t.Fatal(err)
	}

	index, err := os.ReadFile(filepath.Join(tmpdir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), `<link rel="stylesheet" href="/styles.css">`) {
		t.Error("expected generated HTML to link styles.css")
	}
	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "styles.css")); err != nil {
		t.Fatal(err)
	}
}

func TestAtomicBuild(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "garterscopic_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	createTestSite(t, tmpdir)
}

func createTestSite(t *testing.T, tmpdir string) {
	os.MkdirAll(filepath.Join(tmpdir, "components"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "layouts"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "page-types"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "pages"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "styles"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "assets", "js"), 0755)

	os.WriteFile(filepath.Join(tmpdir, "site.yaml"), []byte(`name: Test Site
base_url: http://localhost:8080
build:
  source: .
  output: dist
default_layout: default
default_page_type: page
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "components", "definitions.yaml"), []byte(`components:
  navbar:
    file: navbar.html
    style: styles/navbar.css
    options:
      logo:
        type: string
  footer:
    file: footer.html
    style: styles/footer.css
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "components", "navbar.html"), []byte(`<nav class="navbar">{{ logo }}</nav>`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "components", "footer.html"), []byte(`<footer>Footer</footer>`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "layouts", "default.yaml"), []byte(`name: default
components:
  - name: navbar
    options:
      logo: Test Site
  - name: content
  - name: footer
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "page-types", "page.yaml"), []byte(`name: page
layout: default
fields:
  title:
    type: string
    required: true
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "pages", "index.md"), []byte(`---
type: page
title: Home
---

# Welcome

This is the home page.
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "styles", "global.css"), []byte(`body { margin: 0; }`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "styles", "navbar.css"), []byte(`.navbar { background: #333; }`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "styles", "footer.css"), []byte(`.footer { background: #666; }`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "assets", "js", "main.js"), []byte(`console.log("test");`), 0644)
}
