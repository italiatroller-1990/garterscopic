package renderer

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/italiatroller-1990/garterscopic/internal/components"
	"github.com/italiatroller-1990/garterscopic/internal/pages"
)

var placeholderRegex = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)

type Renderer struct {
	Definitions map[string]components.Definition
}

func New(defs map[string]components.Definition) *Renderer {
	return &Renderer{Definitions: defs}
}

func (r *Renderer) RenderComponentHTML(compName string, htmlContent string, options map[string]any) (string, error) {
	result := htmlContent

	result = placeholderRegex.ReplaceAllStringFunc(result, func(match string) string {
		matches := placeholderRegex.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		key := matches[1]

		value, exists := options[key]
		if !exists {
			return match
		}

		optType := "string"
		if def, ok := r.Definitions[compName]; ok {
			optType = def.GetOptionType(key)
		}

		return r.formatValue(value, optType)
	})

	return result, nil
}

func (r *Renderer) formatValue(value any, optType string) string {
	switch optType {
	case "html":
		return r.formatHTMLValue(value)
	case "string", "number", "date":
		return r.formatStringValue(value)
	case "boolean":
		return r.formatBoolValue(value)
	case "list":
		return r.formatListValue(value)
	case "links":
		return r.formatLinksValue(value)
	case "object":
		return r.formatObjectValue(value)
	default:
		return r.formatStringValue(value)
	}
}

func (r *Renderer) formatHTMLValue(value any) string {
	return fmt.Sprintf("%v", value)
}

func (r *Renderer) formatStringValue(value any) string {
	return html.EscapeString(fmt.Sprintf("%v", value))
}

func (r *Renderer) formatBoolValue(value any) string {
	if b, ok := value.(bool); ok {
		if b {
			return "true"
		}
		return "false"
	}
	return "false"
}

func (r *Renderer) formatListValue(value any) string {
	list, ok := value.([]any)
	if !ok {
		return html.EscapeString(fmt.Sprintf("%v", value))
	}

	var items []string
	for _, item := range list {
		items = append(items, html.EscapeString(fmt.Sprintf("%v", item)))
	}
	return strings.Join(items, ", ")
}

func (r *Renderer) formatLinksValue(value any) string {
	list, ok := value.([]any)
	if !ok {
		return html.EscapeString(fmt.Sprintf("%v", value))
	}

	var items []string
	for _, item := range list {
		link, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name, _ := link["name"].(string)
		url, _ := link["url"].(string)

		name = html.EscapeString(name)
		url = html.EscapeString(url)

		items = append(items, fmt.Sprintf("<li><a href=\"%s\">%s</a></li>", url, name))
	}
	return strings.Join(items, "")
}

func (r *Renderer) formatObjectValue(value any) string {
	obj, ok := value.(map[string]any)
	if !ok {
		return html.EscapeString(fmt.Sprintf("%v", value))
	}

	if text, ok := obj["text"].(string); ok {
		return html.EscapeString(text)
	}

	var parts []string
	for k, v := range obj {
		parts = append(parts, html.EscapeString(k)+"="+html.EscapeString(fmt.Sprintf("%v", v)))
	}
	return strings.Join(parts, " ")
}

// RenderSEOMetaTags generates meta tags for a page
func (r *Renderer) RenderSEOMetaTags(page *pages.Page, baseURL string) string {
	var tags []string

	// Meta charset (always included)
	tags = append(tags, `<meta charset="utf-8">`)

	// Viewport (always included for mobile)
	tags = append(tags, `<meta name="viewport" content="width=device-width, initial-scale=1">`)

	// Language
	tags = append(tags, fmt.Sprintf(`<meta http-equiv="Content-Language" content="en">`))

	// Title
	if page.Title != "" {
		tags = append(tags, fmt.Sprintf(`<title>%s</title>`, html.EscapeString(page.Title)))
	}

	// Description
	if page.Description != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="description" content="%s">`, html.EscapeString(page.Description)))
	}

	// Keywords
	if page.Keywords != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="keywords" content="%s">`, html.EscapeString(page.Keywords)))
	}

	// Robots
	if page.Robots != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="robots" content="%s">`, html.EscapeString(page.Robots)))
	}

	// Canonical URL
	if page.Canonical != "" {
		tags = append(tags, fmt.Sprintf(`<link rel="canonical" href="%s">`, html.EscapeString(page.Canonical)))
	}

	// Open Graph tags
	if page.OGTitle != "" {
		tags = append(tags, fmt.Sprintf(`<meta property="og:title" content="%s">`, html.EscapeString(page.OGTitle)))
	} else if page.Title != "" {
		tags = append(tags, fmt.Sprintf(`<meta property="og:title" content="%s">`, html.EscapeString(page.Title)))
	}

	if page.OGDesc != "" {
		tags = append(tags, fmt.Sprintf(`<meta property="og:description" content="%s">`, html.EscapeString(page.OGDesc)))
	} else if page.Description != "" {
		tags = append(tags, fmt.Sprintf(`<meta property="og:description" content="%s">`, html.EscapeString(page.Description)))
	}

	if page.OGImage != "" {
		tags = append(tags, fmt.Sprintf(`<meta property="og:image" content="%s">`, html.EscapeString(page.OGImage)))
	}

	// Always include og:url and og:type
	if baseURL != "" && page.Route != "" {
		fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(page.Route, "/")
		tags = append(tags, fmt.Sprintf(`<meta property="og:url" content="%s">`, html.EscapeString(fullURL)))
	}

	pageType := "website"
	if page.Section == "blog" || page.Type == "post" {
		pageType = "article"
	}
	tags = append(tags, fmt.Sprintf(`<meta property="og:type" content="%s">`, pageType))

	// Twitter Card
	tags = append(tags, `<meta name="twitter:card" content="summary_large_image">`)
	if page.OGTitle != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="twitter:title" content="%s">`, html.EscapeString(page.OGTitle)))
	} else if page.Title != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="twitter:title" content="%s">`, html.EscapeString(page.Title)))
	}
	if page.OGDesc != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="twitter:description" content="%s">`, html.EscapeString(page.OGDesc)))
	} else if page.Description != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="twitter:description" content="%s">`, html.EscapeString(page.Description)))
	}
	if page.OGImage != "" {
		tags = append(tags, fmt.Sprintf(`<meta name="twitter:image" content="%s">`, html.EscapeString(page.OGImage)))
	}

	return strings.Join(tags, "\n")
}
