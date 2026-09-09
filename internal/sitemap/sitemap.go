package sitemap

import (
	"fmt"
	"strings"
	"time"

	"github.com/italiatroller-1990/garterscopic/internal/pages"
)

// Entry represents a single sitemap entry
type Entry struct {
	Location     string
	LastModified string
	ChangeFreq   string
	Priority     string
}

// Generate creates a sitemap.xml from pages.
// Draft pages are excluded. Pages with a canonical URL use it instead of
// the derived route URL.
func Generate(pages []pages.Page, baseURL string) string {
	var entries []Entry

	for _, page := range pages {
		// Skip draft pages
		if page.Draft {
			continue
		}

		if baseURL == "" || page.Route == "" {
			continue
		}

		location := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(page.Route, "/")
		if page.Canonical != "" {
			location = page.Canonical
		}

		entry := Entry{
			Location:     location,
			LastModified: lastModified(page),
			ChangeFreq:   "weekly",
			Priority:     "0.8",
		}

		// Blog posts have higher priority than regular pages
		if page.Type == "post" || page.Section == "blog" {
			entry.Priority = "0.9"
		}

		// Homepage has highest priority
		if page.Route == "/" || page.Route == "/index.html" {
			entry.Priority = "1.0"
			entry.ChangeFreq = "daily"
		}

		entries = append(entries, entry)
	}

	return renderXML(entries)
}

// lastModified prefers the page's own date over the build time so output
// stays stable for unchanged content.
func lastModified(page pages.Page) string {
	if page.Date != "" {
		return page.Date
	}
	if updated, ok := page.Metadata["updated"].(string); ok && updated != "" {
		return updated
	}
	return time.Now().Format("2006-01-02")
}

func renderXML(entries []Entry) string {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	sb.WriteString("\n")

	for _, entry := range entries {
		sb.WriteString("  <url>\n")
		sb.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", entry.Location))
		if entry.LastModified != "" {
			sb.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", entry.LastModified))
		}
		if entry.ChangeFreq != "" {
			sb.WriteString(fmt.Sprintf("    <changefreq>%s</changefreq>\n", entry.ChangeFreq))
		}
		if entry.Priority != "" {
			sb.WriteString(fmt.Sprintf("    <priority>%s</priority>\n", entry.Priority))
		}
		sb.WriteString("  </url>\n")
	}

	sb.WriteString("</urlset>\n")

	return sb.String()
}
