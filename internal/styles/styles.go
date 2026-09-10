// Package styles collects, concatenates, and copies CSS stylesheets.
package styles

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/italiatroller-1990/garterscopic/internal/config"
)

// GenerateResponsiveCSS returns a CSS snippet defining custom properties
// for the configured responsive breakpoints. The variables use the --gs-
// prefix to avoid collisions with user-defined properties.
func GenerateResponsiveCSS(cfg *config.SiteConfig) string {
	return fmt.Sprintf(`:root {
    --gs-mobile: %s;
    --gs-tablet: %s;
    --gs-desktop: %s;
}
`, cfg.Responsive.Mobile, cfg.Responsive.Tablet, cfg.Responsive.Desktop)
}

// CollectStyles deduplicates and validates the list of stylesheet paths,
// returning only files that exist on disk.
func CollectStyles(cfg *config.SiteConfig, componentStyles []string, globalStyles []string) ([]string, error) {
	allStyles := make([]string, 0)
	seen := make(map[string]bool)

	for _, style := range globalStyles {
		if !seen[style] {
			allStyles = append(allStyles, style)
			seen[style] = true
		}
	}

	for _, style := range componentStyles {
		if !seen[style] {
			allStyles = append(allStyles, style)
			seen[style] = true
		}
	}

	var validated []string
	for _, style := range allStyles {
		fullPath := cfg.SourcePath(style)
		if _, err := os.Stat(fullPath); err == nil {
			validated = append(validated, style)
		}
	}

	return validated, nil
}

// LoadStyles reads and concatenates all listed stylesheets into a single byte slice.
func LoadStyles(styles []string, sourceDir string) ([]byte, error) {
	var result []byte

	for _, style := range styles {
		path := filepath.Join(sourceDir, style)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read stylesheet %s: %w", style, err)
		}
		result = append(result, data...)
		result = append(result, '\n')
	}

	return result, nil
}

// CopyStyles copies each listed stylesheet from sourceDir to destDir,
// preserving relative directory structure.
func CopyStyles(styles []string, sourceDir, destDir string) error {
	for _, style := range styles {
		srcPath := filepath.Join(sourceDir, style)
		destPath := filepath.Join(destDir, style)

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		src, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("failed to open stylesheet %s: %w", style, err)
		}

		dest, err := os.Create(destPath)
		if err != nil {
			src.Close()
			return fmt.Errorf("failed to create stylesheet %s: %w", style, err)
		}

		if _, err := io.Copy(dest, src); err != nil {
			src.Close()
			dest.Close()
			return fmt.Errorf("failed to copy stylesheet %s: %w", style, err)
		}

		src.Close()
		dest.Close()
	}

	return nil
}
