package pages

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/config"
)

func TestParseFrontmatter(t *testing.T) {
	content := `---
type: page
title: Test Page
description: A test page
---

# Hello

This is content.
`

	page, err := parsePage(content, "test.md", "test.md")
	if err != nil {
		t.Fatalf("failed to parse page: %v", err)
	}

	if page.Type != "page" {
		t.Errorf("expected type 'page', got '%s'", page.Type)
	}
	if page.Metadata["title"] != "Test Page" {
		t.Errorf("expected title 'Test Page', got '%v'", page.Metadata["title"])
	}
	if page.Metadata["description"] != "A test page" {
		t.Errorf("expected description 'A test page', got '%v'", page.Metadata["description"])
	}
	if page.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestParsePageWithoutFrontmatter(t *testing.T) {
	content := `# Just Content

Some text.
`

	page, err := parsePage(content, "test.md", "test.md")
	if err != nil {
		t.Fatalf("failed to parse page: %v", err)
	}

	if page.Type != "" {
		t.Errorf("expected empty type, got '%s'", page.Type)
	}
	if page.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestParsePages(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "pages_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	pagesDir := filepath.Join(tmpdir, "pages")
	os.MkdirAll(pagesDir, 0755)

	os.WriteFile(filepath.Join(pagesDir, "index.md"), []byte(`---
type: page
title: Home
---

# Home`), 0644)

	os.WriteFile(filepath.Join(pagesDir, "about.md"), []byte(`---
type: page
title: About
---

# About`), 0644)

	os.MkdirAll(filepath.Join(pagesDir, "posts"), 0755)
	os.WriteFile(filepath.Join(pagesDir, "posts", "hello.md"), []byte(`---
type: post
title: Hello
date: 2024-01-01
---

# Hello`), 0644)

	cfg := &config.SiteConfig{
		Build: config.BuildConfig{Source: tmpdir},
	}

	pages, err := parsePagesForTest(cfg)
	if err != nil {
		t.Fatalf("failed to parse pages: %v", err)
	}

	if len(pages) != 3 {
		t.Errorf("expected 3 pages, got %d", len(pages))
	}
}

func TestParsePosts(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "posts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	postsDir := filepath.Join(tmpdir, "posts")
	os.MkdirAll(filepath.Join(postsDir, "blog"), 0755)
	os.MkdirAll(filepath.Join(postsDir, "guides"), 0755)

	os.WriteFile(filepath.Join(postsDir, "blog", "first.md"), []byte(`---
type: post
title: First Post
date: 2024-01-15
tags:
  - intro
  - welcome
---

# First Post

Content here.`), 0644)

	os.WriteFile(filepath.Join(postsDir, "blog", "second.md"), []byte(`---
type: post
title: Second Post
date: 2024-01-10
tags:
  - tutorial
---

# Second Post

Content here.`), 0644)

	os.WriteFile(filepath.Join(postsDir, "guides", "linux.md"), []byte(`---
type: post
title: Linux Guide
date: 2024-01-20
tags:
  - linux
  - tutorial
---

# Linux Guide

Content here.`), 0644)

	cfg := &config.SiteConfig{
		Build: config.BuildConfig{Source: tmpdir},
	}

	posts, err := ParsePosts(cfg)
	if err != nil {
		t.Fatalf("failed to parse posts: %v", err)
	}

	if len(posts) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(posts))
	}

	for _, p := range posts {
		if p.Section == "" {
			t.Errorf("post %s has empty section", p.Slug)
		}
		if p.Date == "" {
			t.Errorf("post %s has empty date", p.Slug)
		}
	}
}

func TestSectionDetection(t *testing.T) {
	tests := []struct {
		relPath  string
		expected string
	}{
		{"blog/post.md", "blog"},
		{"guides/linux.md", "guides"},
		{"post.md", ""},
		{"news/2024/update.md", "news"},
	}

	for _, tt := range tests {
		result := detectSection(tt.relPath)
		if result != tt.expected {
			t.Errorf("detectSection(%q) = %q, want %q", tt.relPath, result, tt.expected)
		}
	}
}

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]any
		expected []string
	}{
		{
			name:     "with tags",
			metadata: map[string]any{"tags": []any{"linux", "tutorial"}},
			expected: []string{"linux", "tutorial"},
		},
		{
			name:     "empty tags",
			metadata: map[string]any{"tags": []any{}},
			expected: nil,
		},
		{
			name:     "no tags field",
			metadata: map[string]any{"title": "Test"},
			expected: nil,
		},
		{
			name:     "tags not a list",
			metadata: map[string]any{"tags": "linux"},
			expected: nil,
		},
		{
			name:     "tags with mixed types",
			metadata: map[string]any{"tags": []any{"linux", 123, true}},
			expected: []string{"linux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTags(tt.metadata)
			if len(result) != len(tt.expected) {
				t.Errorf("extractTags() returned %v, want %v", result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("extractTags()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestExtractCategories(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]any
		expected []string
	}{
		{
			name:     "with categories",
			metadata: map[string]any{"categories": []any{"development", "gamedev"}},
			expected: []string{"development", "gamedev"},
		},
		{
			name:     "no categories field",
			metadata: map[string]any{"title": "Test"},
			expected: nil,
		},
		{
			name:     "categories not a list",
			metadata: map[string]any{"categories": "development"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractStringList(tt.metadata, "categories")
			if len(result) != len(tt.expected) {
				t.Errorf("extractStringList() returned %v, want %v", result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("extractStringList()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestAllCategoriesAndIndex(t *testing.T) {
	posts := []Page{
		{Categories: []string{"development", "tutorial"}},
		{Categories: []string{"development"}},
		{Categories: nil},
	}

	categories := AllCategories(posts)
	if len(categories) != 2 || categories[0] != "development" || categories[1] != "tutorial" {
		t.Errorf("AllCategories() = %v, want [development tutorial]", categories)
	}

	index := CategoryIndex(posts)
	if len(index["development"]) != 2 {
		t.Errorf("CategoryIndex()['development'] has %d posts, want 2", len(index["development"]))
	}
	if len(index["tutorial"]) != 1 {
		t.Errorf("CategoryIndex()['tutorial'] has %d posts, want 1", len(index["tutorial"]))
	}

	devPosts := PostsByCategory(posts, "development")
	if len(devPosts) != 2 {
		t.Errorf("PostsByCategory('development') returned %d posts, want 2", len(devPosts))
	}
}

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]any
		expected string
	}{
		{
			name:     "with date",
			metadata: map[string]any{"date": "2024-01-15"},
			expected: "2024-01-15",
		},
		{
			name:     "no date",
			metadata: map[string]any{"title": "Test"},
			expected: "",
		},
		{
			name:     "date not a string",
			metadata: map[string]any{"date": 123},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDate(tt.metadata)
			if result != tt.expected {
				t.Errorf("extractDate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractDraft(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]any
		expected bool
	}{
		{
			name:     "draft true",
			metadata: map[string]any{"draft": true},
			expected: true,
		},
		{
			name:     "draft false",
			metadata: map[string]any{"draft": false},
			expected: false,
		},
		{
			name:     "no draft field",
			metadata: map[string]any{"title": "Test"},
			expected: false,
		},
		{
			name:     "draft not a bool",
			metadata: map[string]any{"draft": "yes"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDraft(tt.metadata)
			if result != tt.expected {
				t.Errorf("extractDraft() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDeriveSlug(t *testing.T) {
	tests := []struct {
		relPath  string
		expected string
	}{
		{"blog/hello-world.md", "hello-world"},
		{"guides/linux-basics.md", "linux-basics"},
		{"post.md", "post"},
		{"nested/deep/post.md", "post"},
	}

	for _, tt := range tests {
		result := deriveSlug(tt.relPath)
		if result != tt.expected {
			t.Errorf("deriveSlug(%q) = %q, want %q", tt.relPath, result, tt.expected)
		}
	}
}

func TestGeneratePostRoute(t *testing.T) {
	tests := []struct {
		name     string
		section  string
		source   string
		postsDir string
		expected string
	}{
		{
			name:     "post in section",
			section:  "blog",
			source:   "/tmp/posts/blog/hello.md",
			postsDir: "/tmp/posts",
			expected: "/blog/hello/",
		},
		{
			name:     "post without section",
			section:  "",
			source:   "/tmp/posts/hello.md",
			postsDir: "/tmp/posts",
			expected: "/hello/",
		},
		{
			name:     "nested post",
			section:  "guides",
			source:   "/tmp/posts/guides/linux/basics.md",
			postsDir: "/tmp/posts",
			expected: "/guides/linux/basics/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePostRoute(tt.section, tt.source, tt.postsDir)
			if result != tt.expected {
				t.Errorf("generatePostRoute() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSections(t *testing.T) {
	posts := []Page{
		{Section: "blog", Metadata: map[string]any{"title": "Post 1"}},
		{Section: "guides", Metadata: map[string]any{"title": "Post 2"}},
		{Section: "blog", Metadata: map[string]any{"title": "Post 3"}},
		{Section: "news", Metadata: map[string]any{"title": "Post 4"}},
	}

	sections := Sections(posts)
	expected := []string{"blog", "guides", "news"}

	if len(sections) != len(expected) {
		t.Fatalf("Sections() returned %d items, want %d", len(sections), len(expected))
	}

	for i := range sections {
		if sections[i] != expected[i] {
			t.Errorf("Sections()[%d] = %q, want %q", i, sections[i], expected[i])
		}
	}
}

func TestPostsBySection(t *testing.T) {
	posts := []Page{
		{Section: "blog", Slug: "post1"},
		{Section: "guides", Slug: "post2"},
		{Section: "blog", Slug: "post3"},
	}

	blogPosts := PostsBySection(posts, "blog")
	if len(blogPosts) != 2 {
		t.Fatalf("PostsBySection('blog') returned %d posts, want 2", len(blogPosts))
	}

	guidePosts := PostsBySection(posts, "guides")
	if len(guidePosts) != 1 {
		t.Fatalf("PostsBySection('guides') returned %d posts, want 1", len(guidePosts))
	}

	emptyPosts := PostsBySection(posts, "nonexistent")
	if len(emptyPosts) != 0 {
		t.Fatalf("PostsBySection('nonexistent') returned %d posts, want 0", len(emptyPosts))
	}
}

func TestPostsByTag(t *testing.T) {
	posts := []Page{
		{Tags: []string{"linux", "tutorial"}, Slug: "post1"},
		{Tags: []string{"css", "tutorial"}, Slug: "post2"},
		{Tags: []string{"linux", "advanced"}, Slug: "post3"},
	}

	linuxPosts := PostsByTag(posts, "linux")
	if len(linuxPosts) != 2 {
		t.Fatalf("PostsByTag('linux') returned %d posts, want 2", len(linuxPosts))
	}

	tutorialPosts := PostsByTag(posts, "tutorial")
	if len(tutorialPosts) != 2 {
		t.Fatalf("PostsByTag('tutorial') returned %d posts, want 2", len(tutorialPosts))
	}

	emptyPosts := PostsByTag(posts, "nonexistent")
	if len(emptyPosts) != 0 {
		t.Fatalf("PostsByTag('nonexistent') returned %d posts, want 0", len(emptyPosts))
	}
}

func TestAllTags(t *testing.T) {
	posts := []Page{
		{Tags: []string{"linux", "tutorial"}},
		{Tags: []string{"css", "tutorial"}},
		{Tags: []string{"linux", "advanced"}},
	}

	tags := AllTags(posts)
	expected := []string{"advanced", "css", "linux", "tutorial"}

	if len(tags) != len(expected) {
		t.Fatalf("AllTags() returned %d tags, want %d", len(tags), len(expected))
	}

	for i := range tags {
		if tags[i] != expected[i] {
			t.Errorf("AllTags()[%d] = %q, want %q", i, tags[i], expected[i])
		}
	}
}

func TestTagIndex(t *testing.T) {
	posts := []Page{
		{Tags: []string{"linux", "tutorial"}, Slug: "post1"},
		{Tags: []string{"css", "tutorial"}, Slug: "post2"},
		{Tags: []string{"linux", "advanced"}, Slug: "post3"},
	}

	index := TagIndex(posts)

	if len(index["linux"]) != 2 {
		t.Errorf("TagIndex()['linux'] has %d posts, want 2", len(index["linux"]))
	}
	if len(index["tutorial"]) != 2 {
		t.Errorf("TagIndex()['tutorial'] has %d posts, want 2", len(index["tutorial"]))
	}
	if len(index["css"]) != 1 {
		t.Errorf("TagIndex()['css'] has %d posts, want 1", len(index["css"]))
	}
	if len(index["advanced"]) != 1 {
		t.Errorf("TagIndex()['advanced'] has %d posts, want 1", len(index["advanced"]))
	}
}

func TestParsePostsWithSections(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "posts_sections_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	postsDir := filepath.Join(tmpdir, "posts")
	os.MkdirAll(filepath.Join(postsDir, "blog"), 0755)
	os.MkdirAll(filepath.Join(postsDir, "guides"), 0755)
	os.MkdirAll(filepath.Join(postsDir, "news"), 0755)

	os.WriteFile(filepath.Join(postsDir, "blog", "post1.md"), []byte(`---
type: post
title: Blog Post 1
date: 2024-01-15
tags:
  - intro
---`), 0644)

	os.WriteFile(filepath.Join(postsDir, "blog", "post2.md"), []byte(`---
type: post
title: Blog Post 2
date: 2024-01-10
tags:
  - tutorial
---`), 0644)

	os.WriteFile(filepath.Join(postsDir, "guides", "linux.md"), []byte(`---
type: post
title: Linux Guide
date: 2024-01-20
tags:
  - linux
  - tutorial
---`), 0644)

	os.WriteFile(filepath.Join(postsDir, "guides", "css.md"), []byte(`---
type: post
title: CSS Guide
date: 2024-01-25
tags:
  - css
---`), 0644)

	os.WriteFile(filepath.Join(postsDir, "news", "update.md"), []byte(`---
type: post
title: Site Update
date: 2024-01-30
tags:
  - news
---`), 0644)

	cfg := &config.SiteConfig{
		Build: config.BuildConfig{Source: tmpdir},
	}

	posts, err := ParsePosts(cfg)
	if err != nil {
		t.Fatalf("failed to parse posts: %v", err)
	}

	if len(posts) != 5 {
		t.Fatalf("expected 5 posts, got %d", len(posts))
	}

	sections := Sections(posts)
	if len(sections) != 3 {
		t.Errorf("expected 3 sections, got %d", len(sections))
	}

	blogPosts := PostsBySection(posts, "blog")
	if len(blogPosts) != 2 {
		t.Errorf("expected 2 blog posts, got %d", len(blogPosts))
	}

	guidePosts := PostsBySection(posts, "guides")
	if len(guidePosts) != 2 {
		t.Errorf("expected 2 guide posts, got %d", len(guidePosts))
	}

	newsPosts := PostsBySection(posts, "news")
	if len(newsPosts) != 1 {
		t.Errorf("expected 1 news post, got %d", len(newsPosts))
	}

	tutorialPosts := PostsByTag(posts, "tutorial")
	if len(tutorialPosts) != 2 {
		t.Errorf("expected 2 tutorial posts, got %d", len(tutorialPosts))
	}

	for _, p := range posts {
		if p.Route == "" {
			t.Errorf("post %s has empty route", p.Slug)
		}
		if p.Section == "" {
			t.Errorf("post %s has empty section", p.Slug)
		}
	}
}

func TestPostRoutesAreSorted(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "posts_sort_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	postsDir := filepath.Join(tmpdir, "posts")
	os.MkdirAll(filepath.Join(postsDir, "blog"), 0755)

	os.WriteFile(filepath.Join(postsDir, "blog", "post-c.md"), []byte(`---
type: post
title: Post C
date: 2024-01-01
---`), 0644)

	os.WriteFile(filepath.Join(postsDir, "blog", "post-a.md"), []byte(`---
type: post
title: Post A
date: 2024-01-03
---`), 0644)

	os.WriteFile(filepath.Join(postsDir, "blog", "post-b.md"), []byte(`---
type: post
title: Post B
date: 2024-01-02
---`), 0644)

	cfg := &config.SiteConfig{
		Build: config.BuildConfig{Source: tmpdir},
	}

	posts, err := ParsePosts(cfg)
	if err != nil {
		t.Fatalf("failed to parse posts: %v", err)
	}

	if len(posts) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(posts))
	}

	for i := 1; i < len(posts); i++ {
		if posts[i].Date > posts[i-1].Date {
			t.Errorf("posts not sorted by date: %s (index %d) > %s (index %d)",
				posts[i].Date, i, posts[i-1].Date, i-1)
		}
	}
}

func TestDraftPosts(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "drafts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	postsDir := filepath.Join(tmpdir, "posts")
	os.MkdirAll(filepath.Join(postsDir, "blog"), 0755)

	os.WriteFile(filepath.Join(postsDir, "blog", "published.md"), []byte(`---
type: post
title: Published Post
date: 2024-01-15
draft: false
---`), 0644)

	os.WriteFile(filepath.Join(postsDir, "blog", "draft-post.md"), []byte(`---
type: post
title: Draft Post
date: 2024-01-10
draft: true
---`), 0644)

	cfg := &config.SiteConfig{
		Build: config.BuildConfig{Source: tmpdir},
	}

	posts, err := ParsePosts(cfg)
	if err != nil {
		t.Fatalf("failed to parse posts: %v", err)
	}

	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}

	for _, p := range posts {
		if p.Slug == "draft-post" && !p.Draft {
			t.Error("expected draft-post to be a draft")
		}
		if p.Slug == "published" && p.Draft {
			t.Error("expected published to not be a draft")
		}
	}
}

func TestMixedPagesAndPosts(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "mixed_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	pagesDir := filepath.Join(tmpdir, "pages")
	os.MkdirAll(pagesDir, 0755)
	os.WriteFile(filepath.Join(pagesDir, "about.md"), []byte(`---
type: page
title: About
---
# About`), 0644)

	postsDir := filepath.Join(tmpdir, "posts")
	os.MkdirAll(filepath.Join(postsDir, "blog"), 0755)
	os.WriteFile(filepath.Join(postsDir, "blog", "post1.md"), []byte(`---
type: post
title: Post 1
date: 2024-01-15
tags:
  - intro
---
# Post 1`), 0644)

	cfg := &config.SiteConfig{
		Build: config.BuildConfig{Source: tmpdir},
	}

	pages, err := ParsePages(cfg)
	if err != nil {
		t.Fatalf("failed to parse pages: %v", err)
	}

	posts, err := ParsePosts(cfg)
	if err != nil {
		t.Fatalf("failed to parse posts: %v", err)
	}

	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}
	if len(posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(posts))
	}

	if pages[0].Section != "" {
		t.Error("page should not have a section")
	}
	if posts[0].Section != "blog" {
		t.Errorf("expected post section 'blog', got '%s'", posts[0].Section)
	}
}

func TestBackwardsCompatibility(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "compat_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	pagesDir := filepath.Join(tmpdir, "pages")
	os.MkdirAll(pagesDir, 0755)

	os.WriteFile(filepath.Join(pagesDir, "index.md"), []byte(`---
type: page
title: Home
---
# Home`), 0644)

	os.WriteFile(filepath.Join(pagesDir, "about.md"), []byte(`---
type: page
title: About
---
# About`), 0644)

	cfg := &config.SiteConfig{
		Build: config.BuildConfig{Source: tmpdir},
	}

	pages, err := ParsePages(cfg)
	if err != nil {
		t.Fatalf("failed to parse pages: %v", err)
	}

	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}

	for _, p := range pages {
		if p.Type != "page" {
			t.Errorf("expected type 'page', got '%s'", p.Type)
		}
		if p.Route == "" {
			t.Error("page has empty route")
		}
	}
}

func TestParsePostsNoPostsDirectory(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "no_posts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	cfg := &config.SiteConfig{
		Build: config.BuildConfig{Source: tmpdir},
	}

	posts, err := ParsePosts(cfg)
	if err != nil {
		t.Fatalf("failed to parse posts: %v", err)
	}

	if len(posts) != 0 {
		t.Errorf("expected 0 posts, got %d", len(posts))
	}
}

type testSiteConfig struct {
	source string
}

func (c *testSiteConfig) SourcePath(parts ...string) string {
	return filepath.Join(append([]string{c.source}, parts...)...)
}

func parsePagesForTest(cfg *config.SiteConfig) ([]Page, error) {
	pagesDir := cfg.SourcePath("pages")

	var pages []Page

	err := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		relPath, _ := filepath.Rel(pagesDir, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		page, err := parsePage(string(data), path, relPath)
		if err != nil {
			return err
		}

		if page.Type == "" {
			page.Type = "page"
		}

		pages = append(pages, page)
		return nil
	})

	return pages, err
}

// --- Pagelist tests ---

func TestFilterPagesByDir(t *testing.T) {
	pagesDir := t.TempDir()

	pages := []Page{
		{SourcePath: filepath.Join(pagesDir, "blog", "index.md"), Route: "/blog/", Metadata: map[string]any{"title": "Blog Index"}},
		{SourcePath: filepath.Join(pagesDir, "blog", "post1.md"), Route: "/blog/post1/", Metadata: map[string]any{"title": "Post 1"}},
		{SourcePath: filepath.Join(pagesDir, "blog", "post2.md"), Route: "/blog/post2/", Metadata: map[string]any{"title": "Post 2"}},
		{SourcePath: filepath.Join(pagesDir, "about.md"), Route: "/about/", Metadata: map[string]any{"title": "About"}},
	}

	result, err := FilterPagesByDir(pages, pagesDir, "blog", filepath.Join(pagesDir, "blog", "index.md"))
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(result))
	}
	if result[0].Metadata["title"] != "Post 1" {
		t.Errorf("expected 'Post 1', got %v", result[0].Metadata["title"])
	}
	if result[1].Metadata["title"] != "Post 2" {
		t.Errorf("expected 'Post 2', got %v", result[1].Metadata["title"])
	}
}

func TestFilterPagesByDirExcludesCurrentPage(t *testing.T) {
	pagesDir := t.TempDir()

	pages := []Page{
		{SourcePath: filepath.Join(pagesDir, "blog", "index.md"), Route: "/blog/"},
		{SourcePath: filepath.Join(pagesDir, "blog", "post1.md"), Route: "/blog/post1/"},
	}

	result, err := FilterPagesByDir(pages, pagesDir, "blog", filepath.Join(pagesDir, "blog", "index.md"))
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 page (excluding index), got %d", len(result))
	}
	if result[0].SourcePath != filepath.Join(pagesDir, "blog", "post1.md") {
		t.Errorf("expected post1.md, got %s", result[0].SourcePath)
	}
}

func TestFilterPagesByDirEmpty(t *testing.T) {
	pagesDir := t.TempDir()

	pages := []Page{
		{SourcePath: filepath.Join(pagesDir, "blog", "index.md"), Route: "/blog/"},
	}

	result, err := FilterPagesByDir(pages, pagesDir, "blog", filepath.Join(pagesDir, "blog", "index.md"))
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 pages, got %d", len(result))
	}
}

func TestFilterPagesByDirMissing(t *testing.T) {
	pagesDir := t.TempDir()

	pages := []Page{
		{SourcePath: filepath.Join(pagesDir, "about.md"), Route: "/about/"},
	}

	result, err := FilterPagesByDir(pages, pagesDir, "nonexistent", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 pages for missing dir, got %d", len(result))
	}
}

func TestFilterPagesByDirPathTraversal(t *testing.T) {
	pagesDir := t.TempDir()

	pages := []Page{
		{SourcePath: filepath.Join(pagesDir, "secret.md"), Route: "/secret/"},
	}

	_, err := FilterPagesByDir(pages, pagesDir, "../../etc", "")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestFilterPagesByDirExcludesDrafts(t *testing.T) {
	pagesDir := t.TempDir()

	pages := []Page{
		{SourcePath: filepath.Join(pagesDir, "blog", "draft.md"), Route: "/blog/draft/", Draft: true},
		{SourcePath: filepath.Join(pagesDir, "blog", "published.md"), Route: "/blog/published/"},
	}

	result, err := FilterPagesByDir(pages, pagesDir, "blog", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 published page, got %d", len(result))
	}
}

func TestFilterPagesByDirExcludesAutoGenerated(t *testing.T) {
	pagesDir := t.TempDir()

	pages := []Page{
		{SourcePath: filepath.Join(pagesDir, "blog", "real.md"), Route: "/blog/real/"},
		{SourcePath: "", Route: "/blog/", AutoGenerated: true},
	}

	result, err := FilterPagesByDir(pages, pagesDir, "blog", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 non-auto page, got %d", len(result))
	}
}

func TestFilterPagesByDirDefaultDir(t *testing.T) {
	pagesDir := t.TempDir()

	// When dir is empty, pages in the same directory as the listing page are returned.
	pages := []Page{
		{SourcePath: filepath.Join(pagesDir, "blog", "index.md"), Route: "/blog/"},
		{SourcePath: filepath.Join(pagesDir, "blog", "post1.md"), Route: "/blog/post1/"},
	}

	result, err := FilterPagesByDir(pages, pagesDir, "", filepath.Join(pagesDir, "blog", "index.md"))
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 page, got %d", len(result))
	}
}

func TestSortPagesByTitle(t *testing.T) {
	pages := []Page{
		{Title: "Charlie", Metadata: map[string]any{"title": "Charlie"}},
		{Title: "Alpha", Metadata: map[string]any{"title": "Alpha"}},
		{Title: "Bravo", Metadata: map[string]any{"title": "Bravo"}},
	}

	SortPages(pages, "title", "asc")

	expected := []string{"Alpha", "Bravo", "Charlie"}
	for i, want := range expected {
		if pages[i].Title != want {
			t.Errorf("position %d: expected %q, got %q", i, want, pages[i].Title)
		}
	}
}

func TestSortPagesByTitleDesc(t *testing.T) {
	pages := []Page{
		{Title: "Alpha", Metadata: map[string]any{"title": "Alpha"}},
		{Title: "Charlie", Metadata: map[string]any{"title": "Charlie"}},
		{Title: "Bravo", Metadata: map[string]any{"title": "Bravo"}},
	}

	SortPages(pages, "title", "desc")

	expected := []string{"Charlie", "Bravo", "Alpha"}
	for i, want := range expected {
		if pages[i].Title != want {
			t.Errorf("position %d: expected %q, got %q", i, want, pages[i].Title)
		}
	}
}

func TestSortPagesByDate(t *testing.T) {
	pages := []Page{
		{Date: "2026-01-01", Metadata: map[string]any{"title": "A"}},
		{Date: "2026-03-01", Metadata: map[string]any{"title": "C"}},
		{Date: "2026-02-01", Metadata: map[string]any{"title": "B"}},
	}

	SortPages(pages, "date", "asc")

	expected := []string{"A", "B", "C"}
	for i, want := range expected {
		if pages[i].Metadata["title"] != want {
			t.Errorf("position %d: expected %q, got %v", i, want, pages[i].Metadata["title"])
		}
	}
}

func TestSortPagesByDateDesc(t *testing.T) {
	pages := []Page{
		{Date: "2026-01-01", Metadata: map[string]any{"title": "A"}},
		{Date: "2026-03-01", Metadata: map[string]any{"title": "C"}},
		{Date: "2026-02-01", Metadata: map[string]any{"title": "B"}},
	}

	SortPages(pages, "date", "desc")

	expected := []string{"C", "B", "A"}
	for i, want := range expected {
		if pages[i].Metadata["title"] != want {
			t.Errorf("position %d: expected %q, got %v", i, want, pages[i].Metadata["title"])
		}
	}
}

func TestSortPagesByPath(t *testing.T) {
	pages := []Page{
		{SourcePath: "/c.md"},
		{SourcePath: "/a.md"},
		{SourcePath: "/b.md"},
	}

	SortPages(pages, "", "asc")

	expected := []string{"/a.md", "/b.md", "/c.md"}
	for i, want := range expected {
		if pages[i].SourcePath != want {
			t.Errorf("position %d: expected %q, got %q", i, want, pages[i].SourcePath)
		}
	}
}

func TestSortPagesDeterministic(t *testing.T) {
	// Pages with equal sort values should preserve original order (stable).
	pages := []Page{
		{Date: "2026-01-01", Metadata: map[string]any{"title": "First"}},
		{Date: "2026-01-01", Metadata: map[string]any{"title": "Second"}},
		{Date: "2026-01-01", Metadata: map[string]any{"title": "Third"}},
	}

	SortPages(pages, "date", "asc")

	expected := []string{"First", "Second", "Third"}
	for i, want := range expected {
		if pages[i].Metadata["title"] != want {
			t.Errorf("position %d: expected %q, got %v (stable sort broken)", i, want, pages[i].Metadata["title"])
		}
	}
}

func TestPagelistPageData(t *testing.T) {
	p := Page{
		Metadata:   map[string]any{"title": "My Page", "description": "A page"},
		Date:       "2026-01-15",
		Tags:       []string{"linux", "tutorial"},
		Route:      "/blog/my-page/",
		SourcePath: "pages/blog/my-page.md",
	}

	data := PagelistPageData(p)

	if data["title"] != "My Page" {
		t.Errorf("expected title 'My Page', got %v", data["title"])
	}
	if data["route"] != "/blog/my-page/" {
		t.Errorf("expected route '/blog/my-page/', got %v", data["route"])
	}
	if data["date"] != "2026-01-15" {
		t.Errorf("expected date '2026-01-15', got %v", data["date"])
	}
	if data["path"] != "pages/blog/my-page.md" {
		t.Errorf("expected path 'pages/blog/my-page.md', got %v", data["path"])
	}
}

func TestPagelistConfigParsing(t *testing.T) {
	content := `---
type: page
layout: pagelist
pagelist-config:
  dir: blog
  sort: date
  order: desc
  limit: 5
  pagination: true
  per_page: 10
---

# Blog

List of posts.
`
	tmpdir := t.TempDir()
	pagesDir := filepath.Join(tmpdir, "pages")
	os.MkdirAll(pagesDir, 0755)
	path := filepath.Join(pagesDir, "blog", "index.md")
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(content), 0644)

	cfg := &config.SiteConfig{Build: config.BuildConfig{Source: tmpdir}}
	result, err := ParsePages(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 page, got %d", len(result))
	}

	page := result[0]
	if page.Layout != "pagelist" {
		t.Errorf("expected layout 'pagelist', got %q", page.Layout)
	}
	if page.PagelistConfig == nil {
		t.Fatal("expected pagelist-config to be parsed")
	}
	if page.PagelistConfig.Dir != "blog" {
		t.Errorf("expected dir 'blog', got %q", page.PagelistConfig.Dir)
	}
	if page.PagelistConfig.Sort != "date" {
		t.Errorf("expected sort 'date', got %q", page.PagelistConfig.Sort)
	}
	if page.PagelistConfig.Order != "desc" {
		t.Errorf("expected order 'desc', got %q", page.PagelistConfig.Order)
	}
	if page.PagelistConfig.Limit != 5 {
		t.Errorf("expected limit 5, got %d", page.PagelistConfig.Limit)
	}
	if !page.PagelistConfig.Pagination {
		t.Error("expected pagination true")
	}
	if page.PagelistConfig.PerPage != 10 {
		t.Errorf("expected per_page 10, got %d", page.PagelistConfig.PerPage)
	}
}

func TestPagelistConfigMissing(t *testing.T) {
	content := `---
type: page
title: Simple Page
---

# Hello
`
	tmpdir := t.TempDir()
	pagesDir := filepath.Join(tmpdir, "pages")
	os.MkdirAll(pagesDir, 0755)
	path := filepath.Join(pagesDir, "simple.md")
	os.WriteFile(path, []byte(content), 0644)

	cfg := &config.SiteConfig{Build: config.BuildConfig{Source: tmpdir}}
	result, err := ParsePages(cfg)
	if err != nil {
		t.Fatal(err)
	}

	page := result[0]
	if page.Layout != "" {
		t.Errorf("expected empty layout, got %q", page.Layout)
	}
	if page.PagelistConfig != nil {
		t.Error("expected nil pagelist-config")
	}
}
