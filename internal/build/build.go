package build

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/italiatroller-1990/garterscopic/internal/assets"
	"github.com/italiatroller-1990/garterscopic/internal/binding"
	"github.com/italiatroller-1990/garterscopic/internal/components"
	"github.com/italiatroller-1990/garterscopic/internal/config"
	"github.com/italiatroller-1990/garterscopic/internal/layouts"
	"github.com/italiatroller-1990/garterscopic/internal/markdown"
	"github.com/italiatroller-1990/garterscopic/internal/pages"
	"github.com/italiatroller-1990/garterscopic/internal/pagetypes"
	"github.com/italiatroller-1990/garterscopic/internal/renderer"
	"github.com/italiatroller-1990/garterscopic/internal/routing"
	"github.com/italiatroller-1990/garterscopic/internal/styles"
	"github.com/italiatroller-1990/garterscopic/internal/validation"
)

type BuildResult struct {
	PagesGenerated int
	AssetsCopied   int
	Components     int
	Layouts        int
	PageTypes      int
}

func (r *BuildResult) GetPagesGenerated() int {
	return r.PagesGenerated
}

type Builder struct {
	Config     *config.SiteConfig
	Components map[string]components.Definition
	Layouts    map[string]layouts.Layout
	PageTypes  map[string]pagetypes.PageType
	Pages      []pages.Page
	Renderer   *renderer.Renderer
}

func New(cfg *config.SiteConfig) *Builder {
	return &Builder{Config: cfg}
}

func (b *Builder) Load() error {
	var err error

	b.Components, err = components.LoadDefinitions(b.Config)
	if err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	b.Layouts, err = layouts.LoadLayouts(b.Config)
	if err != nil {
		return fmt.Errorf("failed to load layouts: %w", err)
	}

	b.PageTypes, err = pagetypes.LoadPageTypes(b.Config)
	if err != nil {
		return fmt.Errorf("failed to load page types: %w", err)
	}

	b.Pages, err = pages.ParsePages(b.Config)
	if err != nil {
		return fmt.Errorf("failed to load pages: %w", err)
	}

	b.Renderer = renderer.New(b.Components)

	return nil
}

func (b *Builder) Validate() *validation.BuildError {
	v := validation.New(b.Config)
	return v.ValidateAll(b.Components, b.Layouts, b.PageTypes, b.Pages)
}

func (b *Builder) Build() (*BuildResult, error) {
	result := &BuildResult{
		Components: len(b.Components),
		Layouts:    len(b.Layouts),
		PageTypes:  len(b.PageTypes),
	}

	tmpDir := b.Config.Build.Output + ".tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	defer os.RemoveAll(tmpDir)

	sort.Slice(b.Pages, func(i, j int) bool {
		return b.Pages[i].Route < b.Pages[j].Route
	})

	for _, page := range b.Pages {
		html, err := b.renderPage(page)
		if err != nil {
			return nil, fmt.Errorf("failed to render page %s: %w", page.SourcePath, err)
		}

		outputPath := filepath.Join(tmpDir, routing.OutputPath(page.Route))
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for page: %w", err)
		}

		if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
			return nil, fmt.Errorf("failed to write page: %w", err)
		}

		result.PagesGenerated++
	}

	usedStyles := b.collectUsedStyles()
	combinedCSS, err := styles.LoadStyles(usedStyles, b.Config.Build.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to load styles: %w", err)
	}

	cssPath := filepath.Join(tmpDir, "styles.css")
	if len(combinedCSS) > 0 {
		if err := os.MkdirAll(filepath.Dir(cssPath), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(cssPath, combinedCSS, 0644); err != nil {
			return nil, fmt.Errorf("failed to write styles: %w", err)
		}
	}

	assetCount, err := assets.CopyAssets(b.Config.Build.Source, b.Config.Assets.Directory, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to copy assets: %w", err)
	}
	result.AssetsCopied = assetCount

	if err := os.RemoveAll(b.Config.Build.Output); err != nil {
		return nil, fmt.Errorf("failed to remove old output: %w", err)
	}

	if err := os.Rename(tmpDir, b.Config.Build.Output); err != nil {
		return nil, fmt.Errorf("failed to move new output: %w", err)
	}

	return result, nil
}

func (b *Builder) renderPage(page pages.Page) (string, error) {
	pageType, exists := b.PageTypes[page.Type]
	if !exists {
		pageType = pagetypes.PageType{Name: page.Type}
	}

	metadata := pageType.ApplyDefaults(page.Metadata)

	layoutName := pageType.Layout
	if layoutName == "" {
		layoutName = b.Config.DefaultLayout
	}

	layout, exists := b.Layouts[layoutName]
	if !exists {
		return "", fmt.Errorf("layout '%s' not found", layoutName)
	}

	htmlBody, err := b.renderLayout(layout, page, metadata)
	if err != nil {
		return "", err
	}

	return b.wrapInDocument(htmlBody, page, metadata), nil
}

func (b *Builder) renderLayout(layout layouts.Layout, page pages.Page, metadata map[string]any) (string, error) {
	resolver := binding.NewResolver(
		metadata,
		binding.PageInfo{
			Route: page.Route,
			URL:   page.Route,
			Type:  page.Type,
		},
		binding.SiteInfo{
			Name:    b.Config.Name,
			BaseURL: b.Config.BaseURL,
		},
	)

	groups := layout.GroupByPosition()

	positions := []layouts.Position{
		layouts.PositionTop,
		layouts.PositionLeft,
		layouts.PositionCenter,
		layouts.PositionRight,
		layouts.PositionBottom,
	}

	var body strings.Builder

	for _, pos := range positions {
		instances, ok := groups[pos]
		if !ok {
			continue
		}

		if pos != layouts.PositionCenter {
			fmt.Fprintf(&body, `<div data-garterscopic-region="%s">`+"\n", pos)
		}

		for _, inst := range instances {
			rendered, err := b.renderInstance(inst, page, metadata, resolver)
			if err != nil {
				return "", err
			}
			body.WriteString(rendered)
			body.WriteString("\n")
		}

		if pos != layouts.PositionCenter {
			body.WriteString(`</div>`)
			body.WriteString("\n")
		}
	}

	return body.String(), nil
}

func (b *Builder) renderInstance(inst layouts.ComponentInstance, page pages.Page, metadata map[string]any, resolver *binding.Resolver) (string, error) {
	inst = layouts.NormalizeInstance(inst)

	if inst.Name == "content" {
		renderedBody, err := markdown.Render(page.Body)
		if err != nil {
			return "", fmt.Errorf("failed to render markdown: %w", err)
		}
		return renderedBody, nil
	}

	if inst.Name == "frontmatter-values" {
		valuesAny, _ := inst.Options["values"].([]any)
		var result strings.Builder
		result.WriteString(`<dl class="frontmatter-values">` + "\n")
		for _, v := range valuesAny {
			vs, ok := v.(string)
			if !ok {
				continue
			}
			resolved, err := resolver.Resolve(vs)
			if err != nil {
				continue
			}
			parts := strings.Split(vs, ".")
			name := parts[len(parts)-1]
			fmt.Fprintf(&result, "  <dt>%s</dt><dd>%v</dd>\n", name, resolved)
		}
		result.WriteString("</dl>")
		return result.String(), nil
	}

	comp, exists := b.Components[inst.Name]
	if !exists {
		return "", nil
	}

	compHTML, err := components.LoadComponentHTML(b.Config, comp.File)
	if err != nil {
		return "", err
	}

	options := inst.Options
	if options == nil {
		options = make(map[string]any)
	}

	options, err = binding.ApplyBindings(options, resolver)
	if err != nil {
		return "", err
	}

	for k, v := range metadata {
		if _, exists := options[k]; !exists {
			options[k] = v
		}
	}

	rendered, err := b.Renderer.RenderComponentHTML(inst.Name, compHTML, options)
	if err != nil {
		return "", fmt.Errorf("failed to render component %s: %w", inst.Name, err)
	}

	return rendered, nil
}

func (b *Builder) wrapInDocument(body string, page pages.Page, metadata map[string]any) string {
	title := ""
	if t, ok := metadata["title"].(string); ok {
		title = t
	}

	description := ""
	if d, ok := metadata["description"].(string); ok {
		description = d
	}

	lang := b.Config.Language
	if l, ok := metadata["language"].(string); ok {
		lang = l
	}

	scripts := b.collectScripts()

	var stylesTag string
	if _, err := os.Stat(filepath.Join(b.Config.Build.Output, "styles.css")); err == nil {
		stylesTag = `<link rel="stylesheet" href="/styles.css">`
	}

	var scriptsTags []string
	for _, script := range scripts {
		if script.Type == "module" {
			scriptsTags = append(scriptsTags, fmt.Sprintf(`<script type="module" src="%s"></script>`, script.Path))
		} else {
			scriptsTags = append(scriptsTags, fmt.Sprintf(`<script src="%s"></script>`, script.Path))
		}
	}

	metaDescription := ""
	if description != "" {
		metaDescription = fmt.Sprintf(`    <meta name="description" content="%s">`, escapeHTML(description))
	}

	return fmt.Sprintf(`<!doctype html>
<html lang="%s">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>%s</title>
%s
    %s
    %s
</head>
<body>
%s
</body>
</html>`, lang, escapeHTML(title), metaDescription, stylesTag, strings.Join(scriptsTags, "\n    "), body)
}

func (b *Builder) collectUsedStyles() []string {
	styleSet := make(map[string]bool)

	for _, style := range b.Config.Styles.Global {
		styleSet[style] = true
	}

	for _, page := range b.Pages {
		pageType, exists := b.PageTypes[page.Type]
		if !exists {
			continue
		}

		layoutName := pageType.Layout
		if layoutName == "" {
			layoutName = b.Config.DefaultLayout
		}

		layout, exists := b.Layouts[layoutName]
		if !exists {
			continue
		}

		for _, inst := range layout.Components {
			if comp, exists := b.Components[inst.Name]; exists && comp.Style != "" {
				styleSet[comp.Style] = true
			}
		}
	}

	var result []string
	for style := range styleSet {
		result = append(result, style)
	}
	sort.Strings(result)
	return result
}

func (b *Builder) collectScripts() []config.ScriptConfig {
	var scripts []config.ScriptConfig
	scripts = append(scripts, b.Config.Scripts...)
	return scripts
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `'`, "&#39;")
	return s
}
