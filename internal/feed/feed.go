package feed

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"
)

// Post is the flattened post metadata used to build the feed.
type Post struct {
	Title       string
	URL         string // absolute URL
	Route       string // site route, used if URL is empty
	Date        string
	Description string
	Author      string
}

// Generate produces an RSS 2.0 document for the given posts.
// Draft posts are excluded. Posts without a date sort last.
func Generate(posts []Post, siteTitle, baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")

	sorted := make([]Post, len(posts))
	copy(sorted, posts)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Date > sorted[j].Date
	})

	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<rss version="2.0">`)
	sb.WriteString("\n")
	sb.WriteString("<channel>\n")
	sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", html.EscapeString(siteTitle)))
	if baseURL != "" {
		sb.WriteString(fmt.Sprintf("  <link>%s</link>\n", html.EscapeString(baseURL+"/")))
	}
	sb.WriteString(fmt.Sprintf("  <description>%s</description>\n", html.EscapeString(siteTitle)))
	sb.WriteString(fmt.Sprintf("  <lastBuildDate>%s</lastBuildDate>\n", time.Now().UTC().Format(time.RFC1123Z)))

	for _, post := range sorted {
		link := post.URL
		if link == "" && baseURL != "" && post.Route != "" {
			link = baseURL + "/" + strings.TrimLeft(post.Route, "/")
		}
		if link == "" {
			continue
		}

		sb.WriteString("  <item>\n")
		sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", html.EscapeString(post.Title)))
		sb.WriteString(fmt.Sprintf("    <link>%s</link>\n", html.EscapeString(link)))
		if post.Date != "" {
			sb.WriteString(fmt.Sprintf("    <pubDate>%s</pubDate>\n", formatPubDate(post.Date)))
		}
		if post.Description != "" {
			sb.WriteString(fmt.Sprintf("    <description>%s</description>\n", html.EscapeString(post.Description)))
		}
		if post.Author != "" {
			sb.WriteString(fmt.Sprintf("    <author>%s</author>\n", html.EscapeString(post.Author)))
		}
		sb.WriteString("  </item>\n")
	}

	sb.WriteString("</channel>\n")
	sb.WriteString("</rss>\n")

	return sb.String()
}

// formatPubDate converts YYYY-MM-DD to RFC 1123Z (RSS 2.0 requires it).
func formatPubDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.UTC().Format(time.RFC1123Z)
}
