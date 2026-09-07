package pages

import (
	"os"
	"path/filepath"
	"testing"
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

	cfg := &testSiteConfig{source: tmpdir}

	pages, err := parsePagesForTest(cfg)
	if err != nil {
		t.Fatalf("failed to parse pages: %v", err)
	}

	if len(pages) != 3 {
		t.Errorf("expected 3 pages, got %d", len(pages))
	}
}

type testSiteConfig struct {
	source string
}

func (c *testSiteConfig) SourcePath(parts ...string) string {
	return filepath.Join(append([]string{c.source}, parts...)...)
}

func parsePagesForTest(cfg *testSiteConfig) ([]Page, error) {
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
