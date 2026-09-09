package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/config"
)

// createFeatureSite builds a site with posts exercising tags/categories.
func createFeatureSite(t *testing.T, tmpdir string, extraConfig string) {
	t.Helper()

	createTestSite(t, tmpdir)

	os.MkdirAll(filepath.Join(tmpdir, "posts", "blog"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "page-types"), 0755)

	os.WriteFile(filepath.Join(tmpdir, "page-types", "post.yaml"), []byte(`name: post
layout: default
fields:
  title:
    type: string
    required: true
  date:
    type: date
    required: true
  tags:
    type: list
  categories:
    type: list
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "posts", "blog", "hello.md"), []byte(`---
type: post
title: Hello
description: First post
date: 2026-01-01
author: Jane
tags: [linux, tutorial]
categories: [development]
---

Hello world.
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "posts", "blog", "second.md"), []byte(`---
type: post
title: Second
description: Second post
date: 2026-02-01
tags: [linux]
categories: [development]
---

Second post.
`), 0644)

	// Listing layout so auto-generated pages show their posts.
	os.WriteFile(filepath.Join(tmpdir, "layouts", "default.yaml"), []byte(`name: default
components:
  - name: navbar
    options:
      logo: Test Site
  - name: post-list
  - name: content
  - name: footer
`), 0644)

	siteYaml := `name: Test Site
base_url: http://localhost:8080
build:
  source: .
  output: dist
` + extraConfig

	os.WriteFile(filepath.Join(tmpdir, "site.yaml"), []byte(siteYaml), 0644)
}

func buildFeatureSite(t *testing.T, tmpdir string) (*Builder, *BuildResult) {
	t.Helper()

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, err := config.Load("site.yaml")
	if err != nil {
		t.Fatal(err)
	}

	b := New(cfg)
	if err := b.Load(); err != nil {
		t.Fatal(err)
	}
	if verr := b.Validate(); verr != nil {
		t.Fatalf("validation failed: %v", verr)
	}
	result, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	return b, result
}

func TestFeedGeneration(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
feed:
  enabled: true
  path: /feed.xml
`)

	_, _ = buildFeatureSite(t, tmpdir)

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "feed.xml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, want := range []string{"<rss", "<title>Hello</title>", "<title>Second</title>", "http://localhost:8080/blog/hello/", "<author>Jane</author>"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected feed to contain %s", want)
		}
	}
}

func TestFeedDisabledByDefault(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	_, _ = buildFeatureSite(t, tmpdir)

	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "feed.xml")); !os.IsNotExist(err) {
		t.Error("expected no feed.xml when feed is disabled")
	}
}

func TestSitemapOptIn(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
sitemap:
  enabled: true
`)

	_, _ = buildFeatureSite(t, tmpdir)

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "sitemap.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "http://localhost:8080/blog/hello/") {
		t.Error("expected sitemap to contain post URL")
	}

	// Robots references sitemap only when enabled.
	robots, err := os.ReadFile(filepath.Join(tmpdir, "dist", "robots.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(robots), "Sitemap:") {
		t.Error("expected robots.txt to reference sitemap when enabled")
	}
}

func TestSitemapDisabledByDefault(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	_, _ = buildFeatureSite(t, tmpdir)

	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "sitemap.xml")); !os.IsNotExist(err) {
		t.Error("expected no sitemap.xml by default")
	}
}

func TestTaxonomyTagsOptIn(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
taxonomy:
  tags: true
`)

	_, _ = buildFeatureSite(t, tmpdir)

	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "tags", "index.html")); err != nil {
		t.Error("expected /tags/ index when taxonomy.tags enabled")
	}
	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "tags", "linux", "index.html")); err != nil {
		t.Error("expected /tags/linux/ when taxonomy.tags enabled")
	}
}

func TestTaxonomyOffByDefault(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	_, _ = buildFeatureSite(t, tmpdir)

	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "tags")); !os.IsNotExist(err) {
		t.Error("expected no /tags/ by default")
	}
}

func TestTaxonomyCategories(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
taxonomy:
  categories: true
`)

	_, _ = buildFeatureSite(t, tmpdir)

	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "categories", "development", "index.html")); err != nil {
		t.Error("expected /categories/development/ when taxonomy.categories enabled")
	}
	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "categories", "index.html")); err != nil {
		t.Error("expected /categories/ index")
	}
}

func TestPagination(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
pagination:
  enabled: true
  per_page: 1
`)

	_, _ = buildFeatureSite(t, tmpdir)

	// 2 posts, per_page=1 -> /posts/ (page 1) and /posts/page/2/
	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "posts", "index.html")); err != nil {
		t.Error("expected /posts/ first page")
	}
	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "posts", "page", "2", "index.html")); err != nil {
		t.Error("expected /posts/page/2/")
	}

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "posts", "page", "2", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Hello") {
		t.Error("expected page 2 to contain the oldest post")
	}
}

func TestGalleryOptIn(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
gallery:
  enabled: true
`)

	_, _ = buildFeatureSite(t, tmpdir)

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "components", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	// Metadata is available to templates; page renders via default layout.
	if len(data) == 0 {
		t.Error("expected gallery page content")
	}
}

func TestGalleryOffByDefault(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	_, _ = buildFeatureSite(t, tmpdir)

	if _, err := os.Stat(filepath.Join(tmpdir, "dist", "components")); !os.IsNotExist(err) {
		t.Error("expected no /components/ by default")
	}
}

func TestComponentSlots(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	// Card component with a default slot.
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
  card:
    file: card.html
    options:
      title:
        type: string
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "components", "card.html"), []byte(`<div class="card"><h2>{{ title }}</h2><slot /></div>`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "layouts", "default.yaml"), []byte("name: default\ncomponents:\n  - name: card\n    options:\n      title: Featured\n      slots:\n        default: <p>Slotted body</p>\n  - name: content\n  - name: footer\n"), 0644)

	os.WriteFile(filepath.Join(tmpdir, "pages", "about.md"), []byte("---\ntype: page\ntitle: About\n---\n\nAbout body.\n"), 0644)

	_, _ = buildFeatureSite(t, tmpdir)

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "<h2>Featured</h2>") {
		t.Error("expected card title rendered")
	}
	if !strings.Contains(content, "<p>Slotted body</p>") {
		t.Error("expected slot content injected")
	}
	if strings.Contains(content, "<slot") {
		t.Error("expected slot tag removed from output")
	}
}

func TestRebuildRoutesPartial(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	b, _ := buildFeatureSite(t, tmpdir)

	// Modify the index page and rebuild just that route.
	os.WriteFile(filepath.Join(tmpdir, "pages", "index.md"), []byte("---\ntype: page\ntitle: Home v2\n---\n\nUpdated body.\n"), 0644)

	if err := b.Load(); err != nil {
		t.Fatal(err)
	}

	result, err := b.RebuildRoutes([]string{"/"})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.RebuiltRoutes) != 1 || result.RebuiltRoutes[0] != "/" {
		t.Fatalf("expected only '/' rebuilt, got %v", result.RebuiltRoutes)
	}

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Home v2") {
		t.Error("expected updated page content after partial rebuild")
	}
}

func TestRebuildRoutesFallsBackToFull(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	b, _ := buildFeatureSite(t, tmpdir)

	// A brand-new route must trigger a full rebuild.
	result, err := b.RebuildRoutes([]string{"/new-page/"})
	if err != nil {
		t.Fatal(err)
	}
	if result.RebuiltRoutes != nil {
		t.Errorf("expected full rebuild (no partial routes), got %v", result.RebuiltRoutes)
	}
}

func TestPagesAffectedBy(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	b, _ := buildFeatureSite(t, tmpdir)

	affected := b.PagesAffectedBy(filepath.Join("components", "navbar.html"))
	found := false
	for _, route := range affected {
		if route == "/" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected '/' among pages affected by navbar.html, got %v", affected)
	}
}
