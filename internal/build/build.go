// Package build orchestrates site generation: loading components, layouts,
// pages, and posts, then rendering them into a static HTML output.
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
	"github.com/italiatroller-1990/garterscopic/internal/feed"
	"github.com/italiatroller-1990/garterscopic/internal/graph"
	"github.com/italiatroller-1990/garterscopic/internal/incremental"
	"github.com/italiatroller-1990/garterscopic/internal/layouts"
	"github.com/italiatroller-1990/garterscopic/internal/markdown"
	"github.com/italiatroller-1990/garterscopic/internal/pages"
	"github.com/italiatroller-1990/garterscopic/internal/pagetypes"
	"github.com/italiatroller-1990/garterscopic/internal/renderer"
	"github.com/italiatroller-1990/garterscopic/internal/robots"
	"github.com/italiatroller-1990/garterscopic/internal/routing"
	"github.com/italiatroller-1990/garterscopic/internal/sitemap"
	"github.com/italiatroller-1990/garterscopic/internal/styles"
	"github.com/italiatroller-1990/garterscopic/internal/validation"
)

// BuildResult summarises the outcome of a build or partial rebuild.
type BuildResult struct {
	PagesGenerated int
	AssetsCopied   int
	Components     int
	Layouts        int
	PageTypes      int
	// RebuiltRoutes lists routes re-rendered by a partial rebuild.
	// Empty after a full build.
	RebuiltRoutes []string
}

func (r *BuildResult) GetPagesGenerated() int {
	return r.PagesGenerated
}

// Builder holds the loaded site state and renders pages to the output directory.
type Builder struct {
	Config     *config.SiteConfig
	Components map[string]components.Definition
	Layouts    map[string]layouts.Layout
	PageTypes  map[string]pagetypes.PageType
	Pages      []pages.Page
	Posts      []pages.Page
	Renderer   *renderer.Renderer
	Graph      *incremental.DependencyGraph

	// routesBuilt tracks which routes exist in the output after the last
	// full build; partial rebuilds fall back to a full build when the
	// route set changed (new/renamed/deleted pages).
	routesBuilt map[string]bool
}

// New creates a Builder for the given site configuration.
func New(cfg *config.SiteConfig) *Builder {
	return &Builder{
		Config: cfg,
		Graph:  incremental.NewDependencyGraph(),
	}
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

	b.Posts, err = pages.ParsePosts(b.Config)
	if err != nil {
		return fmt.Errorf("failed to load posts: %w", err)
	}

	b.Renderer = renderer.New(b.Components)

	return nil
}

func (b *Builder) Validate() *validation.BuildError {
	v := validation.New(b.Config)
	return v.ValidateAll(b.Components, b.Layouts, b.PageTypes, b.Pages, b.Posts)
}

// SiteGraph builds the resolved page -> layout/component/style structure
// for the `graph` command. Published and draft pages are both included;
// drafts are marked in their source path.
func (b *Builder) SiteGraph() *graph.SiteGraph {
	layoutFor := func(name string) (layouts.Layout, bool) {
		layo, ok := b.Layouts[name]
		return layo, ok
	}

	entries := make([]graph.Entry, 0, len(b.Pages)+len(b.Posts))

	addPage := func(p pages.Page) {
		layoutName := p.AutoLayout
		if layoutName == "" {
			if pt, ok := b.PageTypes[p.Type]; ok && pt.Layout != "" {
				layoutName = pt.Layout
			} else {
				layoutName = b.Config.DefaultLayout
			}
		}

		source := p.SourcePath
		if source == "" {
			source = "(generated)"
		}
		if p.Draft {
			source += " (draft)"
		}

		entries = append(entries, graph.Entry{
			Route:  p.Route,
			Source: source,
			Layout: layoutName,
		})
	}

	for _, p := range b.Pages {
		addPage(p)
	}
	for _, p := range b.Posts {
		addPage(p)
	}

	siteGraph := graph.Build(entries, b.Components, layoutFor)
	siteGraph.SortNodes()
	return siteGraph
}

func (b *Builder) Build() (*BuildResult, error) {
	b.Graph = incremental.NewDependencyGraph()

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

	usedStyles := b.collectUsedStyles()
	combinedCSS, err := styles.LoadStyles(usedStyles, b.Config.Build.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to load styles: %w", err)
	}

	allContent := append(b.Pages, b.Posts...)

	sort.Slice(allContent, func(i, j int) bool {
		return allContent[i].Route < allContent[j].Route
	})

	contentInfo := b.buildContentInfo()

	for _, page := range allContent {
		if page.Draft {
			continue
		}

		if err := b.writePageFile(tmpDir, page, len(combinedCSS) > 0, contentInfo); err != nil {
			return nil, err
		}
		result.PagesGenerated++
	}

	galleryPage := b.generateGalleryPage()
	if galleryPage != nil {
		if err := b.writePageFile(tmpDir, *galleryPage, len(combinedCSS) > 0, contentInfo); err != nil {
			return nil, err
		}
		result.PagesGenerated++
	}

	autoPages := b.generateAutoPages(contentInfo)
	for _, page := range autoPages {
		if err := b.writePageFile(tmpDir, page, len(combinedCSS) > 0, contentInfo); err != nil {
			return nil, err
		}
		result.PagesGenerated++
	}

	b.routesBuilt = make(map[string]bool)
	for _, page := range allContent {
		if !page.Draft {
			b.routesBuilt[page.Route] = true
		}
	}
	for _, page := range autoPages {
		b.routesBuilt[page.Route] = true
	}
	if galleryPage != nil {
		b.routesBuilt[galleryPage.Route] = true
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

	// Generate sitemap.xml (optional, on by default for compatibility)
	if b.Config.Sitemap.Enabled {
		allPages := append(b.Pages, b.Posts...)
		sitemapContent := sitemap.Generate(allPages, b.Config.BaseURL)
		sitemapPath := filepath.Join(tmpDir, b.Config.SEO.SitemapFile)
		if err := os.WriteFile(sitemapPath, []byte(sitemapContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write sitemap: %w", err)
		}
	}

	// Generate robots.txt
	sitemapRef := ""
	if b.Config.Sitemap.Enabled {
		sitemapRef = b.Config.SEO.SitemapFile
	}
	robotsContent := robots.Generate(b.Config.BaseURL, sitemapRef)
	robotsPath := filepath.Join(tmpDir, b.Config.SEO.RobotsFile)
	if err := os.WriteFile(robotsPath, []byte(robotsContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write robots.txt: %w", err)
	}

	// Generate feed.xml (opt-in)
	if b.Config.Feed.Enabled {
		if err := b.writeFeed(tmpDir); err != nil {
			return nil, err
		}
	}

	if err := os.RemoveAll(b.Config.Build.Output); err != nil {
		return nil, fmt.Errorf("failed to remove old output: %w", err)
	}

	if err := os.Rename(tmpDir, b.Config.Build.Output); err != nil {
		return nil, fmt.Errorf("failed to move new output: %w", err)
	}

	return result, nil
}

// writePageFile renders a single page and writes it into outputDir using
// the route's output path. Shared by full builds and partial rebuilds.
func (b *Builder) writePageFile(outputDir string, page pages.Page, hasStyles bool, contentInfo binding.ContentInfo) error {
	html, err := b.renderPage(page, hasStyles, contentInfo)
	if err != nil {
		if page.AutoGenerated {
			return fmt.Errorf("failed to render auto page %s: %w", page.Route, err)
		}
		return fmt.Errorf("failed to render page %s: %w", page.SourcePath, err)
	}

	outputPath := filepath.Join(outputDir, routing.OutputPath(page.Route))
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for page: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write page: %w", err)
	}

	return nil
}

// writeFeed generates the RSS feed from published posts.
func (b *Builder) writeFeed(outputDir string) error {
	posts := make([]feed.Post, 0, len(b.Posts))
	for _, p := range b.Posts {
		if p.Draft {
			continue
		}
		author, _ := p.Metadata["author"].(string)
		posts = append(posts, feed.Post{
			Title:       p.Title,
			Route:       p.Route,
			Date:        p.Date,
			Description: p.Description,
			Author:      author,
		})
	}

	content := feed.Generate(posts, b.Config.Feed.Title, b.Config.BaseURL)

	feedPath := strings.TrimPrefix(b.Config.Feed.Path, "/")
	fullPath := filepath.Join(outputDir, filepath.FromSlash(feedPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create feed directory: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write feed: %w", err)
	}
	return nil
}

func (b *Builder) renderPage(page pages.Page, hasStyles bool, contentInfo binding.ContentInfo) (string, error) {
	pageType, exists := b.PageTypes[page.Type]
	if !exists {
		pageType = pagetypes.PageType{Name: page.Type}
	}

	metadata := pageType.ApplyDefaults(page.Metadata)

	if page.AutoPageType != "" {
		if pt, ok := b.PageTypes[page.AutoPageType]; ok {
			pageType = pt
			metadata = pt.ApplyDefaults(metadata)
		}
	}

	layoutName := pageType.Layout
	if layoutName == "" {
		layoutName = b.Config.DefaultLayout
	}

	// Frontmatter layout override takes precedence over page type layout.
	if page.Layout != "" {
		layoutName = page.Layout
	}

	if page.AutoLayout != "" {
		layoutName = page.AutoLayout
	}

	// Handle the pagelist built-in layout: filter pages by directory and
	// expose them through metadata so the post-list component renders them.
	if layoutName == "pagelist" {
		b.applyPagelist(&page, metadata)
		// Use the default layout since there is no pagelist.yaml file.
		layoutName = b.Config.DefaultLayout
	}

	layout, exists := b.Layouts[layoutName]
	if !exists {
		return "", fmt.Errorf("layout '%s' not found", layoutName)
	}

	// Record dependencies for incremental rebuilds: page -> layout file,
	// page -> page-type file, page -> component/style files (below).
	b.Graph.AddDependency(page.Route, "layouts/"+layoutName+".yaml")
	b.Graph.AddDependency(page.Route, "page-types/"+page.Type+".yaml")
	b.Graph.AddDependency(page.Route, page.SourcePath)

	htmlBody, err := b.renderLayout(layout, page, metadata, contentInfo)
	if err != nil {
		return "", err
	}

	return b.wrapInDocument(htmlBody, page, metadata, hasStyles), nil
}

// hasLayout reports whether a named layout exists.
func (b *Builder) hasLayout(name string) bool {
	_, ok := b.Layouts[name]
	return ok
}

// applyPagelist filters all pages by the configured directory and stores
// the result in metadata["pagelist_pages"] so that post-list can render it.
func (b *Builder) applyPagelist(page *pages.Page, metadata map[string]any) {
	cfg := page.PagelistConfig
	if cfg == nil {
		cfg = &pages.PagelistConfig{}
	}

	// Default dir: the listing page's own directory.
	dir := cfg.Dir
	if dir == "" {
		pagesDir := b.Config.SourcePath("pages")
		rel, _ := filepath.Rel(pagesDir, page.SourcePath)
		dir = filepath.Dir(rel)
		if dir == "." {
			dir = ""
		}
	}

	pagesDir := b.Config.SourcePath("pages")
	filtered, err := pages.FilterPagesByDir(b.Pages, pagesDir, dir, page.SourcePath)
	if err != nil {
		metadata["pagelist_pages"] = []map[string]any{}
		return
	}

	// Sort.
	pages.SortPages(filtered, cfg.Sort, cfg.Order)

	// Limit.
	if cfg.Limit > 0 && len(filtered) > cfg.Limit {
		filtered = filtered[:cfg.Limit]
	}

	// Convert to data maps.
	data := make([]map[string]any, 0, len(filtered))
	for _, p := range filtered {
		data = append(data, pages.PagelistPageData(p))
	}

	metadata["pagelist_pages"] = data
	metadata["pagelist_dir"] = dir
}

func (b *Builder) renderLayout(layout layouts.Layout, page pages.Page, metadata map[string]any, contentInfo binding.ContentInfo) (string, error) {
	resolver := binding.NewResolverWithContent(
		metadata,
		binding.PageInfo{
			Route: page.Route,
			URL:   page.Route,
			Type:  page.Type,
		},
		binding.SiteInfo{
			Name:         b.Config.Name,
			BaseURL:      b.Config.BaseURL,
			MobileWidth:  b.Config.Responsive.Mobile,
			TabletWidth:  b.Config.Responsive.Tablet,
			DesktopWidth: b.Config.Responsive.Desktop,
			Links:        convertLinks(b.Config.Links.Entries),
		},
		contentInfo,
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

	if inst.Name == "post-list" {
		return b.renderPostList(metadata, resolver)
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

	// Record dependencies: page -> component HTML file and stylesheet.
	b.Graph.AddDependency(page.Route, "components/"+comp.File)
	if comp.Style != "" {
		b.Graph.AddDependency(page.Route, comp.Style)
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

	// Slots: pull out the 'slots' option so nested content can be injected.
	slots := renderer.ExtractSlots(options)
	if slots != nil {
		options["__slots__"] = slots
	}

	rendered, err := b.Renderer.RenderComponentHTML(inst.Name, compHTML, options)
	if err != nil {
		return "", fmt.Errorf("failed to render component %s: %w", inst.Name, err)
	}

	return rendered, nil
}

func (b *Builder) wrapInDocument(body string, page pages.Page, metadata map[string]any, hasStyles bool) string {
	lang := b.Config.Language
	if l, ok := metadata["language"].(string); ok {
		lang = l
	}

	scripts := b.collectScripts()

	var stylesTag string
	if hasStyles {
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

	var faviconTag string
	iconFile := b.Config.Icon
	if b.Config.Favicon != "" {
		iconFile = b.Config.Favicon
	}
	if iconFile != "" {
		faviconTag = fmt.Sprintf(`<link rel="icon" href="/%s/%s">`, b.Config.IconDir, iconFile)
	}

	// Generate SEO meta tags
	seoMetaTags := b.Renderer.RenderSEOMetaTags(&page, b.Config.BaseURL)

	return fmt.Sprintf(`<!doctype html>
<html lang="%s">
<head>
%s
    %s
    %s
    %s
</head>
<body>
%s
</body>
</html>`, lang, seoMetaTags, faviconTag, stylesTag, strings.Join(scriptsTags, "\n    "), body)
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

func (b *Builder) buildContentInfo() binding.ContentInfo {
	postsData := make([]map[string]any, 0, len(b.Posts))
	for _, p := range b.Posts {
		if p.Draft {
			continue
		}
		data := map[string]any{
			"title":       p.Metadata["title"],
			"description": p.Metadata["description"],
			"date":        p.Date,
			"tags":        p.Tags,
			"section":     p.Section,
			"slug":        p.Slug,
			"route":       p.Route,
			"url":         p.Route,
			"author":      p.Metadata["author"],
		}
		postsData = append(postsData, data)
	}

	pagesData := make([]map[string]any, 0, len(b.Pages))
	for _, p := range b.Pages {
		data := map[string]any{
			"title":       p.Metadata["title"],
			"description": p.Metadata["description"],
			"route":       p.Route,
			"url":         p.Route,
			"type":        p.Type,
		}
		pagesData = append(pagesData, data)
	}

	sections := pages.Sections(b.Posts)
	sectionPosts := make(map[string][]map[string]any)
	for _, section := range sections {
		posts := pages.PostsBySection(b.Posts, section)
		var data []map[string]any
		for _, p := range posts {
			if p.Draft {
				continue
			}
			data = append(data, map[string]any{
				"title":       p.Metadata["title"],
				"description": p.Metadata["description"],
				"date":        p.Date,
				"tags":        p.Tags,
				"section":     p.Section,
				"slug":        p.Slug,
				"route":       p.Route,
				"url":         p.Route,
				"author":      p.Metadata["author"],
			})
		}
		sectionPosts[section] = data
	}

	tags := pages.AllTags(b.Posts)
	tagIndex := make(map[string][]map[string]any)
	for _, tag := range tags {
		posts := pages.PostsByTag(b.Posts, tag)
		var data []map[string]any
		for _, p := range posts {
			if p.Draft {
				continue
			}
			data = append(data, map[string]any{
				"title":       p.Metadata["title"],
				"description": p.Metadata["description"],
				"date":        p.Date,
				"tags":        p.Tags,
				"section":     p.Section,
				"slug":        p.Slug,
				"route":       p.Route,
				"url":         p.Route,
				"author":      p.Metadata["author"],
			})
		}
		tagIndex[tag] = data
	}

	return binding.ContentInfo{
		Posts:        postsData,
		Pages:        pagesData,
		Sections:     sections,
		Tags:         tags,
		TagIndex:     tagIndex,
		SectionPosts: sectionPosts,
	}
}

func (b *Builder) generateAutoPages(contentInfo binding.ContentInfo) []pages.Page {
	var autoPages []pages.Page

	// appendListing emits either a paginated set (pages 1..N, page 1 at the
	// base route) or a single unpaginated page. Pagination replaces the base
	// listing so routes never collide.
	appendListing := func(baseRoute string, metadata map[string]any, postsData []map[string]any) {
		if paginated := b.paginatePosts(baseRoute, postsData, metadata); len(paginated) > 0 {
			autoPages = append(autoPages, paginated...)
			return
		}
		autoPages = append(autoPages, pages.Page{
			Route:         baseRoute,
			Type:          b.Config.DefaultPageType,
			Metadata:      metadata,
			AutoGenerated: true,
			AutoLayout:    b.Config.DefaultLayout,
			AutoPageType:  b.Config.DefaultPageType,
		})
	}

	for _, section := range contentInfo.Sections {
		sectionPosts := contentInfo.SectionPosts[section]
		metadata := map[string]any{
			"title":       section,
			"description": fmt.Sprintf("Posts in %s", section),
			"section":     section,
			"posts":       sectionPosts,
		}

		appendListing("/"+section+"/", metadata, sectionPosts)
	}

	// Taxonomy pages are generated only when configured.
	if b.Config.Taxonomy.Tags {
		for _, tag := range contentInfo.Tags {
			tagPosts := contentInfo.TagIndex[tag]
			metadata := map[string]any{
				"title":       fmt.Sprintf("Tag: %s", tag),
				"description": fmt.Sprintf("Posts tagged with %s", tag),
				"tag":         tag,
				"posts":       tagPosts,
			}

			appendListing("/tags/"+tag+"/", metadata, tagPosts)
		}

		if len(contentInfo.Tags) > 0 {
			metadata := map[string]any{
				"title":       "Tags",
				"description": "All tags",
				"tags":        contentInfo.Tags,
			}

			autoPages = append(autoPages, pages.Page{
				Route:         "/tags/",
				Type:          b.Config.DefaultPageType,
				Metadata:      metadata,
				AutoGenerated: true,
				AutoLayout:    b.Config.DefaultLayout,
				AutoPageType:  b.Config.DefaultPageType,
			})
		}
	}

	// Category taxonomy pages.
	if b.Config.Taxonomy.Categories {
		categories := b.categoriesIndex()
		for _, category := range categories.all() {
			categoryPosts := categories.byCategory(category)
			metadata := map[string]any{
				"title":       fmt.Sprintf("Category: %s", category),
				"description": fmt.Sprintf("Posts in category %s", category),
				"category":    category,
				"posts":       categoryPosts,
			}

			appendListing("/categories/"+category+"/", metadata, categoryPosts)
		}

		if len(categories.all()) > 0 {
			metadata := map[string]any{
				"title":       "Categories",
				"description": "All categories",
				"categories":  categories.all(),
			}

			autoPages = append(autoPages, pages.Page{
				Route:         "/categories/",
				Type:          b.Config.DefaultPageType,
				Metadata:      metadata,
				AutoGenerated: true,
				AutoLayout:    b.Config.DefaultLayout,
				AutoPageType:  b.Config.DefaultPageType,
			})
		}
	}

	// Global post listing with pagination.
	if b.Config.Pagination.Enabled {
		metadata := map[string]any{
			"title":       "Posts",
			"description": "All posts",
			"posts":       contentInfo.Posts,
		}
		autoPages = append(autoPages, b.paginatePosts("/posts/", contentInfo.Posts, metadata)...)
	}

	return autoPages
}

type categoryIndex struct {
	sorted []string
	index  map[string][]map[string]any
}

func (c *categoryIndex) all() []string { return c.sorted }

func (c *categoryIndex) byCategory(category string) []map[string]any {
	return c.index[category]
}

// categoriesIndex builds the category taxonomy from published posts.
func (b *Builder) categoriesIndex() *categoryIndex {
	index := make(map[string][]map[string]any)
	seen := make(map[string]bool)
	var sorted []string

	for _, p := range b.Posts {
		if p.Draft {
			continue
		}
		data := map[string]any{
			"title":       p.Metadata["title"],
			"description": p.Metadata["description"],
			"date":        p.Date,
			"tags":        p.Tags,
			"categories":  p.Categories,
			"section":     p.Section,
			"slug":        p.Slug,
			"route":       p.Route,
			"url":         p.Route,
			"author":      p.Metadata["author"],
		}
		for _, category := range p.Categories {
			if !seen[category] {
				seen[category] = true
				sorted = append(sorted, category)
			}
			index[category] = append(index[category], data)
		}
	}

	sort.Strings(sorted)
	return &categoryIndex{sorted: sorted, index: index}
}

// paginatePosts generates the paginated variants of a post listing.
// Page 1 lives at the base route; further pages at <base>/page/N/. Returns
// nil when pagination is disabled or there is nothing to list, so callers
// can fall back to the unpaginated page.
func (b *Builder) paginatePosts(baseRoute string, postsData []map[string]any, baseMetadata map[string]any) []pages.Page {
	if !b.Config.Pagination.Enabled || b.Config.Pagination.PerPage <= 0 || len(postsData) == 0 {
		return nil
	}

	perPage := b.Config.Pagination.PerPage
	totalPages := (len(postsData) + perPage - 1) / perPage

	var result []pages.Page
	base := strings.TrimSuffix(baseRoute, "/")

	for i := 1; i <= totalPages; i++ {
		start := (i - 1) * perPage
		end := start + perPage
		if end > len(postsData) {
			end = len(postsData)
		}

		pagePosts := postsData[start:end]

		paginationData := map[string]any{
			"page":         i,
			"pages":        totalPages,
			"total":        len(postsData),
			"per_page":     perPage,
			"has_previous": i > 1,
			"has_next":     i < totalPages,
		}
		if i > 1 {
			paginationData["previous"] = pageRouteFor(base, i-1)
		}
		if i < totalPages {
			paginationData["next"] = pageRouteFor(base, i+1)
		}

		metadata := make(map[string]any, len(baseMetadata)+2)
		for k, v := range baseMetadata {
			metadata[k] = v
		}
		metadata["posts"] = pagePosts
		metadata["pagination"] = paginationData
		if i > 1 {
			metadata["title"] = fmt.Sprintf("%v (page %d)", baseMetadata["title"], i)
		}

		route := pageRouteFor(base, i)
		result = append(result, pages.Page{
			Route:         route,
			Type:          b.Config.DefaultPageType,
			Metadata:      metadata,
			AutoGenerated: true,
			AutoLayout:    b.Config.DefaultLayout,
			AutoPageType:  b.Config.DefaultPageType,
		})
	}

	return result
}

func pageRouteFor(base string, pageNum int) string {
	if pageNum <= 1 {
		if base == "" {
			return "/"
		}
		return base + "/"
	}
	return base + "/page/" + fmt.Sprintf("%d", pageNum) + "/"
}

// generateGalleryPage returns the component gallery page when enabled.
func (b *Builder) generateGalleryPage() *pages.Page {
	if !b.Config.Gallery.Enabled {
		return nil
	}

	metadata := map[string]any{
		"title":       "Components",
		"description": "Component gallery",
		"components":  b.galleryComponents(),
	}

	return &pages.Page{
		Route:         "/components/",
		Type:          b.Config.DefaultPageType,
		Metadata:      metadata,
		AutoGenerated: true,
		AutoLayout:    b.Config.DefaultLayout,
		AutoPageType:  b.Config.DefaultPageType,
	}
}

// galleryComponents flattens component definitions with example values for
// templates to render.
func (b *Builder) galleryComponents() []map[string]any {
	names := make([]string, 0, len(b.Components))
	for name := range b.Components {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]map[string]any, 0, len(names))
	for _, name := range names {
		def := b.Components[name]

		options := make([]map[string]any, 0, len(def.Options))
		optNames := make([]string, 0, len(def.Options))
		for optName := range def.Options {
			optNames = append(optNames, optName)
		}
		sort.Strings(optNames)
		for _, optName := range optNames {
			optDef := def.Options[optName]
			options = append(options, map[string]any{
				"name":     optName,
				"type":     optDef.Type,
				"required": optDef.Required,
				"default":  optDef.Default,
			})
		}

		result = append(result, map[string]any{
			"name":     name,
			"file":     def.File,
			"style":    def.Style,
			"position": def.Position,
			"options":  options,
		})
	}
	return result
}

// renderPostList renders the page's "posts" metadata as a simple linked
// list. When "pagelist_pages" is present (from the pagelist built-in layout),
// it is used instead of "posts". It keeps auto-generated listing pages usable
// out of the box without forcing a visual design; users can build their own
// listing components from the posts binding instead.
func (b *Builder) renderPostList(metadata map[string]any, resolver *binding.Resolver) (string, error) {
	postsAny, _ := metadata["posts"].([]map[string]any)
	if len(postsAny) == 0 {
		postsAny, _ = metadata["pagelist_pages"].([]map[string]any)
	}
	if len(postsAny) == 0 {
		return "", nil
	}

	var result strings.Builder
	result.WriteString(`<ul class="post-list">` + "\n")
	for _, post := range postsAny {
		route, _ := post["route"].(string)
		title := fmt.Sprintf("%v", post["title"])
		date, _ := post["date"].(string)

		result.WriteString(`  <li><a href="` + escapeHTML(route) + `">` + escapeHTML(title) + `</a>`)
		if date != "" {
			result.WriteString(` <time datetime="` + escapeHTML(date) + `">` + escapeHTML(date) + `</time>`)
		}
		result.WriteString("</li>\n")
	}
	result.WriteString("</ul>")

	// Pagination footer when there is more than one page.
	if pag, ok := metadata["pagination"].(map[string]any); ok {
		pages, _ := pag["pages"].(int)
		if pages > 1 {
			page, _ := pag["page"].(int)
			result.WriteString(fmt.Sprintf(`
<nav class="pagination"><span>Page %d of %d</span>`, page, pages))
			if prev, ok := pag["previous"].(string); ok {
				result.WriteString(` <a href="` + escapeHTML(prev) + `" rel="prev">Previous</a>`)
			}
			if next, ok := pag["next"].(string); ok {
				result.WriteString(` <a href="` + escapeHTML(next) + `" rel="next">Next</a>`)
			}
			result.WriteString("</nav>")
		}
	}

	return result.String(), nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `'`, "&#39;")
	return s
}

func convertLinks(links []config.LinkConfig) []binding.LinkConfig {
	result := make([]binding.LinkConfig, len(links))
	for i, l := range links {
		result[i] = binding.LinkConfig{Name: l.Name, URL: l.URL}
	}
	return result
}

// RebuildRoutes re-renders the given routes directly into the output
// directory without touching anything else. It falls back to a full build
// when a route is unknown (e.g. a new page was added). Styles are not
// regenerated; the caller can use RegenerateStyles for global CSS changes.
func (b *Builder) RebuildRoutes(routes []string) (*BuildResult, error) {
	result := &BuildResult{}

	needsFull := b.routesBuilt == nil
	for _, route := range routes {
		if !b.routesBuilt[route] {
			needsFull = true
			break
		}
	}
	if needsFull {
		return b.Build()
	}

	usedStyles := b.collectUsedStyles()
	combinedCSS, err := styles.LoadStyles(usedStyles, b.Config.Build.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to load styles: %w", err)
	}
	hasStyles := len(combinedCSS) > 0
	contentInfo := b.buildContentInfo()

	byRoute := make(map[string]pages.Page)
	for _, page := range append(b.Pages, b.Posts...) {
		byRoute[page.Route] = page
	}

	for _, route := range routes {
		if !b.routesBuilt[route] {
			continue
		}
		page, ok := byRoute[route]
		if !ok {
			// Route belongs to an auto page; regenerate those instead.
			continue
		}
		if err := b.writePageFile(b.Config.Build.Output, page, hasStyles, contentInfo); err != nil {
			return nil, err
		}
		result.RebuiltRoutes = append(result.RebuiltRoutes, route)
		result.PagesGenerated++
	}

	return result, nil
}

// RegenerateStyles rewrites styles.css in the output directory.
func (b *Builder) RegenerateStyles() (int, error) {
	usedStyles := b.collectUsedStyles()
	combinedCSS, err := styles.LoadStyles(usedStyles, b.Config.Build.Source)
	if err != nil {
		return 0, fmt.Errorf("failed to load styles: %w", err)
	}
	if len(combinedCSS) == 0 {
		return 0, nil
	}
	cssPath := b.Config.OutputPath("styles.css")
	if err := os.MkdirAll(filepath.Dir(cssPath), 0755); err != nil {
		return 0, err
	}
	if err := os.WriteFile(cssPath, combinedCSS, 0644); err != nil {
		return 0, fmt.Errorf("failed to write styles: %w", err)
	}
	return len(combinedCSS), nil
}

// CopySingleAsset copies one asset file into the output directory.
func (b *Builder) CopySingleAsset(relPath string) error {
	src := filepath.Join(b.Config.Build.Source, b.Config.Assets.Directory, relPath)
	dest := filepath.Join(b.Config.Build.Output, b.Config.Assets.Directory, relPath)

	info, err := os.Stat(src)
	if err != nil {
		// File was removed: drop it from the output.
		if os.IsNotExist(err) {
			_ = os.Remove(dest)
			return nil
		}
		return err
	}
	if info.IsDir() {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read asset %s: %w", relPath, err)
	}
	return os.WriteFile(dest, data, 0644)
}

// PagesAffectedBy returns the routes that depend on the given source file
// (component HTML, stylesheet, layout, page, ...).
func (b *Builder) PagesAffectedBy(changedFile string) []string {
	return b.Graph.PagesAffectedBy(changedFile)
}
