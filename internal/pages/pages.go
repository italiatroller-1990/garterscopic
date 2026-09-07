package pages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/italiatroller-1990/garterscopic/internal/config"
	"github.com/italiatroller-1990/garterscopic/internal/routing"
	"github.com/italiatroller-1990/garterscopic/internal/yaml"
)

type Page struct {
	SourcePath string
	Route      string
	Type       string
	Metadata   map[string]any
	Body       string
	RawContent string
}

const frontmatterDelim = "---"

func ParsePages(cfg *config.SiteConfig) ([]Page, error) {
	pagesDir := cfg.SourcePath("pages")

	var pages []Page

	if err := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		relPath, err := filepath.Rel(pagesDir, path)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read page %s: %w", path, err)
		}

		page, err := parsePage(string(data), path, relPath)
		if err != nil {
			return fmt.Errorf("failed to parse page %s: %w", path, err)
		}

		if page.Type == "" {
			page.Type = cfg.DefaultPageType
		}

		pages = append(pages, page)
		return nil
	}); err != nil {
		return nil, err
	}

	for i := range pages {
		pages[i].Route = routing.GenerateRoute(pages[i].SourcePath, cfg.SourcePath("pages"))
	}

	return pages, nil
}

func parsePage(content string, sourcePath, relPath string) (Page, error) {
	page := Page{
		SourcePath: sourcePath,
	}

	if !strings.HasPrefix(content, frontmatterDelim) {
		page.Body = content
		page.Metadata = make(map[string]any)
		return page, nil
	}

	lines := strings.Split(content, "\n")
	endIdx := -1

	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontmatterDelim {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		page.Body = content
		page.Metadata = make(map[string]any)
		return page, nil
	}

	frontmatter := strings.Join(lines[1:endIdx], "\n")
	if err := yaml.Parse([]byte(frontmatter), &page.Metadata); err != nil {
		return Page{}, fmt.Errorf("invalid frontmatter: %w", err)
	}

	bodyLines := lines[endIdx+1:]
	if len(bodyLines) > 0 && bodyLines[0] == "" {
		bodyLines = bodyLines[1:]
	}
	page.Body = strings.Join(bodyLines, "\n")
	page.RawContent = page.Body

	if pageType, ok := page.Metadata["type"].(string); ok {
		page.Type = pageType
	}

	return page, nil
}

func (p *Page) RelativePath(baseDir string) string {
	rel, _ := filepath.Rel(baseDir, p.SourcePath)
	return rel
}
