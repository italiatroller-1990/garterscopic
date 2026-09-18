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
// for the configured responsive breakpoints, plus a small responsive
// foundation (fluid media, overflow guards, and stacking media queries).
// The variables use the --gs- prefix to avoid collisions with
// user-defined properties.
//
// Note: CSS custom properties cannot be used in @media conditions, so the
// media queries below interpolate the configured breakpoint values directly.
func GenerateResponsiveCSS(cfg *config.SiteConfig) string {
	mobile := cfg.Responsive.Mobile
	if mobile == "" {
		mobile = "768px"
	}
	tablet := cfg.Responsive.Tablet
	if tablet == "" {
		tablet = "1024px"
	}
	desktop := cfg.Responsive.Desktop
	if desktop == "" {
		desktop = "1200px"
	}
	return fmt.Sprintf(`:root {
    --gs-mobile: %s;
    --gs-tablet: %s;
    --gs-desktop: %s;
}

/* Garterscopic responsive foundation: look good on all devices. */

/* Responsive media: scale down if needed, never scale up beyond natural size. */
img, video, svg, canvas {
    max-width: 100%%;
    height: auto;
}
picture, figure {
    max-width: 100%%;
}
iframe {
    max-width: 100%%;
}

/* Guard against horizontal overflow from long words, tables, and code. */
body {
    overflow-x: hidden;
}
h1, h2, h3, h4, h5, h6, p {
    overflow-wrap: break-word;
}
table {
    display: block;
    max-width: 100%%;
    overflow-x: auto;
}
pre, code {
    white-space: pre-wrap;
    word-wrap: break-word;
}
pre {
    max-width: 100%%;
    overflow-x: auto;
}

/* Fluid type: follows the viewport instead of staying fixed. */
h1 {
    font-size: clamp(1.75rem, 1.25rem + 2.5vw, 2.5rem);
}

/* Tablet and below: tighten page gutters and hero rhythm. */
@media screen and (max-width: %s) {
    .container {
        padding-left: 16px;
        padding-right: 16px;
    }
    .hero {
        padding: 3rem 1.5rem;
    }
    .hero-title {
        font-size: clamp(2rem, 1.25rem + 4vw, 2.5rem);
    }
}

/* Mobile: stack navigation and content instead of squeezing them. */
@media screen and (max-width: %s) {
    .container {
        padding-left: 12px;
        padding-right: 12px;
    }
    .navbar {
        flex-direction: column;
        align-items: stretch;
        gap: 0.75rem;
        padding: 1rem;
    }
    .navbar-links {
        flex-direction: column;
        align-items: stretch;
        flex-wrap: wrap;
        gap: 0.5rem;
    }
    .hero {
        padding: 2rem 1rem;
    }
    .hero-title {
        font-size: clamp(1.75rem, 1rem + 8vw, 2.25rem);
    }
    .hero-subtitle {
        font-size: 1rem;
    }
    .card {
        margin: 0.75rem 0;
    }
    .footer {
        padding: 1.5rem 1rem;
    }
}
`, mobile, tablet, desktop, tablet, mobile)
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
