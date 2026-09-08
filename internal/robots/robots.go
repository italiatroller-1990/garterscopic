package robots

import (
	"fmt"
	"strings"
)

// Generate creates a robots.txt file
func Generate(baseURL string, sitemapFile string) string {
	var sb strings.Builder

	sb.WriteString("# Garterscopic robots.txt\n")
	sb.WriteString("# Generated automatically\n\n")

	sb.WriteString("User-agent: *\n")
	sb.WriteString("Allow: /\n")
	sb.WriteString("Disallow: /dist/\n")
	sb.WriteString("Disallow: /*.tmp\n\n")

	// Add sitemap URL if baseURL is provided
	if baseURL != "" && sitemapFile != "" {
		sitemapURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(sitemapFile, "/")
		sb.WriteString(fmt.Sprintf("Sitemap: %s\n", sitemapURL))
	}

	return sb.String()
}
