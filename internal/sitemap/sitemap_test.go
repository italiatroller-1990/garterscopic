package sitemap

import (
	"strings"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/pages"
)

func TestGenerateBasic(t *testing.T) {
	pgs := []pages.Page{
		{Route: "/", Type: "page", Date: "2026-01-01"},
		{Route: "/about/", Type: "page", Date: "2026-01-02"},
	}

	xml := Generate(pgs, "http://example.com")

	for _, want := range []string{
		"<loc>http://example.com/</loc>",
		"<loc>http://example.com/about/</loc>",
		"<priority>1.0</priority>",
		"<lastmod>2026-01-02</lastmod>",
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected sitemap to contain %s", want)
		}
	}
}

func TestGenerateSkipsDrafts(t *testing.T) {
	pgs := []pages.Page{
		{Route: "/", Type: "page"},
		{Route: "/draft/", Type: "page", Draft: true},
	}

	xml := Generate(pgs, "http://example.com")

	if strings.Contains(xml, "/draft/") {
		t.Error("expected draft pages to be excluded from sitemap")
	}
}

func TestGenerateSkipsEmptyBaseURL(t *testing.T) {
	pgs := []pages.Page{{Route: "/", Type: "page"}}

	xml := Generate(pgs, "")

	if strings.Contains(xml, "<loc>") {
		t.Error("expected no entries without a base_url")
	}
}

func TestGenerateCanonicalURL(t *testing.T) {
	pgs := []pages.Page{
		{
			Route:     "/original/",
			Type:      "page",
			Canonical: "https://other.example.com/canonical/",
		},
	}

	xml := Generate(pgs, "http://example.com")

	if strings.Contains(xml, "http://example.com/original/") {
		t.Error("expected canonical URL to replace the route URL")
	}
	if !strings.Contains(xml, "<loc>https://other.example.com/canonical/</loc>") {
		t.Error("expected canonical URL in sitemap")
	}
}

func TestGeneratePostPriority(t *testing.T) {
	pgs := []pages.Page{
		{Route: "/blog/hello/", Type: "post", Date: "2026-02-03"},
	}

	xml := Generate(pgs, "http://example.com")

	if !strings.Contains(xml, "<priority>0.9</priority>") {
		t.Error("expected posts to have priority 0.9")
	}
}
