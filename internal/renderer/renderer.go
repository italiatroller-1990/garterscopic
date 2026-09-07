package renderer

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/italiatroller-1990/garterscopic/internal/components"
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
