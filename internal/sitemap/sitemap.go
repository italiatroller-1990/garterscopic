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

// Generate creates a sitemap.xml from pages
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

		fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(page.Route, "/")

		entry := Entry{
			Location:     fullURL,
			LastModified: time.Now().Format("2006-01-02"),
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
