package robots

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	content := Generate("https://example.com", "sitemap.xml")

	for _, want := range []string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /dist/",
		"Sitemap: https://example.com/sitemap.xml",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("expected robots.txt to contain %q", want)
		}
	}
}

func TestGenerateNoSitemap(t *testing.T) {
	content := Generate("https://example.com", "")

	if strings.Contains(content, "Sitemap:") {
		t.Error("expected no Sitemap line when sitemapFile is empty")
	}
}

func TestGenerateNoBaseURL(t *testing.T) {
	content := Generate("", "sitemap.xml")

	if strings.Contains(content, "Sitemap:") {
		t.Error("expected no Sitemap line when baseURL is empty")
	}
}

func TestGenerateMinimal(t *testing.T) {
	content := Generate("", "")

	if !strings.Contains(content, "User-agent: *") {
		t.Error("expected User-agent directive")
	}
}
