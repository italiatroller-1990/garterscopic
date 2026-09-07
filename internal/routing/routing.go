package routing

import (
	"path"
	"path/filepath"
	"strings"
)

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

func JoinRoute(base, part string) string {
	base = strings.TrimSuffix(base, "/")
	part = strings.TrimPrefix(part, "/")
	return base + "/" + part
}

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
