// Package routing converts page source paths to URL routes and output file paths.
package routing

import (
	"path"
	"path/filepath"
	"strings"
)

// GenerateRoute converts a page source path to a URL route. Pages in the
// pages directory become URL paths with trailing slashes. The root page
// (pages/index.md) maps to "/".
func GenerateRoute(sourcePath, pagesDir string) string {
	relPath, err := filepath.Rel(pagesDir, sourcePath)
	if err != nil {
		return "/"
	}

	relPath = strings.TrimSuffix(relPath, ".md")

	if relPath == "" || relPath == "index.md" {
		return "/"
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	var routeParts []string

	for _, part := range parts {
		if part == "" || part == "index" {
			continue
		}
		routeParts = append(routeParts, part)
	}

	if len(routeParts) == 0 {
		return "/"
	}

	return "/" + strings.Join(routeParts, "/") + "/"
}

// OutputPath converts a URL route to the filesystem output path.
// "/" becomes "index.html"; "/about/" becomes "about/index.html".
func OutputPath(route string) string {
	if route == "/" {
		return "index.html"
	}

	route = strings.TrimPrefix(route, "/")
	route = strings.TrimSuffix(route, "/")

	parts := strings.Split(route, "/")
	var outputParts []string

	for i, part := range parts {
		if i == len(parts)-1 && part != "" {
			outputParts = append(outputParts, part, "index.html")
		} else if part != "" {
			outputParts = append(outputParts, part)
		}
	}

	if len(outputParts) == 0 {
		return "index.html"
	}

	return filepath.Join(outputParts...)
}

// EnsureUniqueRoutes scans pages for duplicate routes and returns the file
// paths of any pages that share a route with an earlier page.
func EnsureUniqueRoutes(pages []struct {
	Route string
	Path  string
}) (duplicates []string) {
	seen := make(map[string]bool)

	for _, p := range pages {
		if seen[p.Route] {
			duplicates = append(duplicates, p.Path)
		}
		seen[p.Route] = true
	}

	return duplicates
}

// JoinRoute concatenates a base route and a part with a single slash,
// trimming redundant slashes.
func JoinRoute(base, part string) string {
	base = strings.TrimSuffix(base, "/")
	part = strings.TrimPrefix(part, "/")
	return base + "/" + part
}

// CleanRoute normalizes a route path: cleans dot segments, ensures a leading
// slash, and removes trailing slashes (except for the root "/").
func CleanRoute(route string) string {
	route = path.Clean(route)
	if route == "." {
		return "/"
	}
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	if route != "/" && strings.HasSuffix(route, "/") {
		route = strings.TrimSuffix(route, "/")
	}
	return route
}
