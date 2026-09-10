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

// --- Pagelist integration tests ---

func createPagelistSite(t *testing.T, tmpdir string, extraConfig string) {
	t.Helper()

	os.MkdirAll(filepath.Join(tmpdir, "pages", "blog"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "components"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "layouts"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "styles"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "page-types"), 0755)

	// Minimal site config.
	siteYaml := `name: Test
base_url: http://localhost:8080
build:
  source: .
  output: dist
default_layout: default
default_page_type: page
` + extraConfig
	os.WriteFile(filepath.Join(tmpdir, "site.yaml"), []byte(siteYaml), 0644)

	// Page type.
	os.WriteFile(filepath.Join(tmpdir, "page-types", "page.yaml"), []byte(`name: page
layout: default
fields:
  title:
    type: string
    required: true
  description:
    type: string
`), 0644)

	// Layout with post-list so pagelist results render.
	os.WriteFile(filepath.Join(tmpdir, "layouts", "default.yaml"), []byte(`name: default
components:
  - name: post-list
  - name: content
`), 0644)

	// Minimal components.
	os.WriteFile(filepath.Join(tmpdir, "components", "definitions.yaml"), []byte(`components:
  navbar:
    file: navbar.html
`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "components", "navbar.html"), []byte(`<nav>Nav</nav>`), 0644)

	// Global style.
	os.WriteFile(filepath.Join(tmpdir, "styles", "global.css"), []byte("body{}"), 0644)

	// Blog posts.
	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "post1.md"), []byte(`---
type: page
title: First Post
date: 2026-01-15
---

First post content.
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "post2.md"), []byte(`---
type: page
title: Second Post
date: 2026-02-20
---

Second post content.
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "post3.md"), []byte(`---
type: page
title: Third Post
date: 2026-03-10
---

Third post content.
`), 0644)

	// Draft post (should be excluded).
	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "draft.md"), []byte(`---
type: page
title: Draft Post
draft: true
---

Draft content.
`), 0644)

	// A page outside blog (should not appear in blog listing).
	os.WriteFile(filepath.Join(tmpdir, "pages", "about.md"), []byte(`---
type: page
title: About
---

About content.
`), 0644)
}

func TestPagelistBasic(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	// The listing page uses pagelist layout.
	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "index.md"), []byte(`---
type: page
layout: pagelist
title: Blog Index
pagelist-config:
  dir: blog
---

# Blog
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	result, _ := b.Build()

	if result.PagesGenerated == 0 {
		t.Fatal("expected pages to be generated")
	}

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "blog", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Should contain links to the 3 published blog posts.
	for _, want := range []string{"First Post", "Second Post", "Third Post"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected pagelist to contain %q", want)
		}
	}

	// Should NOT contain the draft post.
	if strings.Contains(content, "Draft Post") {
		t.Error("pagelist should not contain draft posts")
	}

	// Should NOT contain the about page.
	if strings.Contains(content, "About") {
		t.Error("pagelist should not contain pages from other directories")
	}
}

func TestPagelistExcludesCurrentPage(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	// The listing page itself.
	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "index.md"), []byte(`---
type: page
layout: pagelist
title: Blog Index
pagelist-config:
  dir: blog
---

# Blog
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	b.Build()

	data, _ := os.ReadFile(filepath.Join(tmpdir, "dist", "blog", "index.html"))
	content := string(data)

	// The listing page should not list itself.
	if strings.Contains(content, `<a href="/blog/">Blog Index</a>`) {
		t.Error("pagelist should not list itself")
	}
}

func TestPagelistSortByTitleAsc(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "index.md"), []byte(`---
type: page
layout: pagelist
title: Blog Index
pagelist-config:
  dir: blog
  sort: title
  order: asc
---

# Blog
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	b.Build()

	data, _ := os.ReadFile(filepath.Join(tmpdir, "dist", "blog", "index.html"))
	content := string(data)

	// In ascending title order: First, Second, Third.
	idxFirst := strings.Index(content, "First Post")
	idxSecond := strings.Index(content, "Second Post")
	idxThird := strings.Index(content, "Third Post")

	if idxFirst < 0 || idxSecond < 0 || idxThird < 0 {
		t.Fatal("expected all three posts in output")
	}
	if idxFirst > idxSecond || idxSecond > idxThird {
		t.Error("expected ascending title order: First < Second < Third")
	}
}

func TestPagelistSortByDateDesc(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "index.md"), []byte(`---
type: page
layout: pagelist
title: Blog Index
pagelist-config:
  dir: blog
  sort: date
  order: desc
---

# Blog
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	b.Build()

	data, _ := os.ReadFile(filepath.Join(tmpdir, "dist", "blog", "index.html"))
	content := string(data)

	// In descending date order: Third (Mar), Second (Feb), First (Jan).
	idxFirst := strings.Index(content, "First Post")
	idxSecond := strings.Index(content, "Second Post")
	idxThird := strings.Index(content, "Third Post")

	if idxFirst < 0 || idxSecond < 0 || idxThird < 0 {
		t.Fatal("expected all three posts in output")
	}
	if idxThird > idxSecond || idxSecond > idxFirst {
		t.Error("expected descending date order: Third > Second > First")
	}
}

func TestPagelistLimit(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "index.md"), []byte(`---
type: page
layout: pagelist
title: Blog Index
pagelist-config:
  dir: blog
  sort: title
  order: asc
  limit: 2
---

# Blog
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	b.Build()

	data, _ := os.ReadFile(filepath.Join(tmpdir, "dist", "blog", "index.html"))
	content := string(data)

	// Only 2 of 3 posts should appear.
	if strings.Contains(content, "First Post") && strings.Contains(content, "Second Post") && strings.Contains(content, "Third Post") {
		t.Error("expected limit to restrict to 2 posts")
	}
}

func TestPagelistNestedDir(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	// Add a nested post.
	os.MkdirAll(filepath.Join(tmpdir, "pages", "blog", "linux"), 0755)
	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "linux", "intro.md"), []byte(`---
type: page
title: Linux Intro
---

Linux content.
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "index.md"), []byte(`---
type: page
layout: pagelist
title: Blog Index
pagelist-config:
  dir: blog
---

# Blog
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	b.Build()

	data, _ := os.ReadFile(filepath.Join(tmpdir, "dist", "blog", "index.html"))
	content := string(data)

	// Should include the nested page.
	if !strings.Contains(content, "Linux Intro") {
		t.Error("expected nested page 'Linux Intro' in pagelist output")
	}
}

func TestPagelistDefaultDir(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	// No explicit dir: should default to the listing page's own directory.
	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "index.md"), []byte(`---
type: page
layout: pagelist
title: Blog Index
---

# Blog
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	b.Build()

	data, _ := os.ReadFile(filepath.Join(tmpdir, "dist", "blog", "index.html"))
	content := string(data)

	// Should list other pages in blog/ but not itself.
	for _, want := range []string{"First Post", "Second Post", "Third Post"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in default-dir pagelist", want)
		}
	}
}

func TestPagelistPathTraversalRejected(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	os.WriteFile(filepath.Join(tmpdir, "pages", "blog", "index.md"), []byte(`---
type: page
layout: pagelist
title: Blog Index
pagelist-config:
  dir: ../../etc
---

# Blog
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	result, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}

	// The page should still be generated (with empty list), not crash.
	if result.PagesGenerated == 0 {
		t.Fatal("expected page to be generated even with invalid dir")
	}

	data, _ := os.ReadFile(filepath.Join(tmpdir, "dist", "blog", "index.html"))
	content := string(data)

	// Should not contain any file paths from outside pages/.
	if strings.Contains(content, "/etc") {
		t.Error("pagelist should not expose files outside pages directory")
	}
}

func TestPagelistEmptyDir(t *testing.T) {
	tmpdir := t.TempDir()
	createPagelistSite(t, tmpdir, "")

	os.MkdirAll(filepath.Join(tmpdir, "pages", "empty"), 0755)
	os.WriteFile(filepath.Join(tmpdir, "pages", "empty", "index.md"), []byte(`---
type: page
layout: pagelist
title: Empty List
pagelist-config:
  dir: empty
---

# Empty
`), 0644)

	workingDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, _ := config.Load("site.yaml")
	b := New(cfg)
	b.Load()
	result, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}

	if result.PagesGenerated == 0 {
		t.Fatal("expected page to be generated for empty dir")
	}
}

// --- Responsive feature tests ---

func TestViewportMetaTagGenerated(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	_, _ = buildFeatureSite(t, tmpdir)

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, `<meta name="viewport" content="width=device-width, initial-scale=1">`) {
		t.Error("expected viewport meta tag in generated HTML")
	}
}

func TestViewportMetaTagNotDuplicated(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	// Create a component that includes a viewport meta tag in its body.
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

	// Component with a viewport tag in its HTML body.
	os.WriteFile(filepath.Join(tmpdir, "components", "hero.html"), []byte(`<meta name="viewport" content="width=device-width, initial-scale=1"><div>Hero</div>`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "layouts", "default.yaml"), []byte(`name: default
components:
  - name: hero
  - name: content
  - name: footer
`), 0644)

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
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Count occurrences of the viewport tag.
	count := strings.Count(content, `name="viewport"`)
	if count != 1 {
		t.Errorf("expected exactly 1 viewport meta tag, got %d", count)
	}
}

func TestResponsiveCSSGenerated(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
responsive:
  mobile: "690px"
  tablet: "820px"
  desktop: "1000px"
`)

	_, _ = buildFeatureSite(t, tmpdir)

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "styles.css"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, want := range []string{
		"--gs-mobile: 690px",
		"--gs-tablet: 820px",
		"--gs-desktop: 1000px",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("expected styles.css to contain %q", want)
		}
	}
}

func TestResponsiveCSSDefaultValues(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	_, _ = buildFeatureSite(t, tmpdir)

	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "styles.css"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, want := range []string{
		"--gs-mobile: 768px",
		"--gs-tablet: 1024px",
		"--gs-desktop: 1200px",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("expected styles.css to contain default %q", want)
		}
	}
}

func TestResponsiveInvalidOrderingFailsBuild(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
responsive:
  mobile: "900px"
  tablet: "500px"
  desktop: "1000px"
`)

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

	buildErr := b.Validate()
	if buildErr == nil {
		t.Error("expected validation error for invalid breakpoint ordering")
	}
}

func TestResponsiveInvalidCSSLengthFailsBuild(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
responsive:
  mobile: "mobile"
  tablet: "820px"
  desktop: "1000px"
`)

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

	buildErr := b.Validate()
	if buildErr == nil {
		t.Error("expected validation error for invalid CSS length")
	}
}

func TestSiteWithoutResponsiveStillBuilds(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, "")

	_, result := buildFeatureSite(t, tmpdir)

	if result.PagesGenerated == 0 {
		t.Fatal("expected pages to be generated without responsive config")
	}

	// Verify viewport tag is still generated.
	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `name="viewport"`) {
		t.Error("expected viewport meta tag even without responsive config")
	}
}

func TestResponsiveConfigDoesNotAffectUnrelatedBuilds(t *testing.T) {
	tmpdir := t.TempDir()
	createFeatureSite(t, tmpdir, `
responsive:
  mobile: "690px"
  tablet: "820px"
  desktop: "1000px"
`)

	_, result := buildFeatureSite(t, tmpdir)

	if result.PagesGenerated == 0 {
		t.Fatal("expected pages to be generated")
	}

	// Verify the site still works correctly with responsive config.
	data, err := os.ReadFile(filepath.Join(tmpdir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Core functionality should still work.
	if !strings.Contains(content, `<link rel="stylesheet" href="/styles.css">`) {
		t.Error("expected styles.css link in output")
	}
	if !strings.Contains(content, `<meta charset="utf-8">`) {
		t.Error("expected charset meta tag in output")
	}
	if !strings.Contains(content, `<meta name="viewport"`) {
		t.Error("expected viewport meta tag in output")
	}
}
