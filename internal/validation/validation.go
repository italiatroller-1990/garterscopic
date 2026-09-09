package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/italiatroller-1990/garterscopic/internal/components"
	"github.com/italiatroller-1990/garterscopic/internal/markdown"
	"github.com/italiatroller-1990/garterscopic/internal/config"
	"github.com/italiatroller-1990/garterscopic/internal/layouts"
	"github.com/italiatroller-1990/garterscopic/internal/pages"
	"github.com/italiatroller-1990/garterscopic/internal/pagetypes"
	"github.com/italiatroller-1990/garterscopic/internal/routing"
)

type Error struct {
	Field   string
	Message string
	Path    string
	Hint    string
}

func (e *Error) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s\n\nFile: %s\n\nExpected:\n  %s\n\nReceived:\n  %s\n\nFix:\n  %s",
			e.Message, e.Field, e.Path, e.Message, e.Path, e.Hint)
	}
	return fmt.Sprintf("%s\n\nFile: %s\n\nFix:\n  %s", e.Message, e.Path, e.Hint)
}

type BuildError struct {
	Errors   []*Error
	Warnings []*Error
}

func (e *BuildError) Error() string {
	if len(e.Errors) == 0 {
		return "build failed"
	}
	return e.Errors[0].Error()
}

func (e *BuildError) Add(err *Error) {
	e.Errors = append(e.Errors, err)
}

func (e *BuildError) AddWarning(err *Error) {
	e.Warnings = append(e.Warnings, err)
}

// HasErrors reports whether any hard errors were found.
func (e *BuildError) HasErrors() bool { return len(e.Errors) > 0 }

// HasWarnings reports whether any warnings were found.
func (e *BuildError) HasWarnings() bool { return len(e.Warnings) > 0 }

type Validator struct {
	Config     *config.SiteConfig
	Components map[string]components.Definition
	Layouts    map[string]layouts.Layout
	PageTypes  map[string]pagetypes.PageType
	Pages      []pages.Page
	Posts      []pages.Page
}

func New(cfg *config.SiteConfig) *Validator {
	return &Validator{Config: cfg}
}

func (v *Validator) ValidateAll(comps map[string]components.Definition, layos map[string]layouts.Layout, pts map[string]pagetypes.PageType, pgs []pages.Page, posts []pages.Page) *BuildError {
	buildErr := &BuildError{}

	v.validateConfig(buildErr)
	v.validateComponentFiles(comps, buildErr)
	v.validateLayouts(layos, comps, buildErr)
	v.validatePageTypes(pts, layos, buildErr)
	v.validatePages(pgs, pts, buildErr)
	v.validatePages(posts, pts, buildErr)
	v.validateRoutes(append(pgs, posts...), buildErr)
	v.validateInstanceOptions(layos, comps, buildErr)
	v.validateGlobalStyles(buildErr)
	v.validateInternalLinks(append(pgs, posts...), buildErr)
	v.validateReferencedAssets(buildErr)

	if len(buildErr.Errors) == 0 && len(buildErr.Warnings) == 0 {
		return nil
	}
	return buildErr
}

// validateInstanceOptions checks component instances in layouts against the
// component definitions: unknown options, wrong types, missing required
// options and invalid positions.
func (v *Validator) validateInstanceOptions(layos map[string]layouts.Layout, comps map[string]components.Definition, buildErr *BuildError) {
	// Structural options that are not component options.
	structuralOptions := map[string]bool{"slots": true}

	for layoutName, layo := range layos {
		path := filepath.Join("layouts", layoutName+".yaml")

	for _, inst := range layo.Components {
		def, exists := comps[inst.Name]
		if !exists || inst.Name == "content" {
			continue
		}

		if inst.Position != "" && !layouts.ValidPosition(inst.Position) {
			buildErr.Add(&Error{
				Message: fmt.Sprintf("component '%s' in layout '%s': invalid position '%s' (must be: top, bottom, left, right, center)", inst.Name, layoutName, inst.Position),
				Path:    path,
				Hint:    "Use one of: top, bottom, left, right, center",
			})
		}

		// Filter structural options before type validation.
		opts := make(map[string]any, len(inst.Options))
		for k, v := range inst.Options {
			if !structuralOptions[k] {
				opts[k] = v
			}
		}

		for _, verr := range def.ValidateOptions(components.Instance{
			Name:     inst.Name,
			Options:  opts,
			Position: string(inst.Position),
		}) {
			buildErr.Add(&Error{
				Message: verr.Error(),
				Path:    path,
				Hint:    "Fix the component options in the layout to match components/definitions.yaml",
			})
		}
	}
	}
}

// validateGlobalStyles warns about configured global stylesheets that are
// missing on disk.
func (v *Validator) validateGlobalStyles(buildErr *BuildError) {
	for _, style := range v.Config.Styles.Global {
		path := v.Config.SourcePath(style)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			buildErr.Add(&Error{
				Message: fmt.Sprintf("global stylesheet '%s' not found", style),
				Path:    path,
				Hint:    "Create the file or remove it from styles.global in site.yaml",
			})
		}
	}
}

var hrefRegex = regexp.MustCompile(`href\s*=\s*"([^"]*)"`)

// validateInternalLinks checks that internal href targets resolve to known
// routes. External URLs, anchors, and special schemes are ignored.
// Broken links are warnings unless links.fail_on_broken is set.
func (v *Validator) validateInternalLinks(pgs []pages.Page, buildErr *BuildError) {
	if len(pgs) == 0 {
		return
	}

	known := make(map[string]bool)
	for _, pg := range pgs {
		if !pg.Draft {
			known[pg.Route] = true
		}
	}

	for _, pg := range pgs {
		if pg.Draft {
			continue
		}

		rendered, err := markdownRender(pg.Body)
		if err != nil {
			continue
		}

		// Scan both rendered markdown and the raw body: the markdown
		// renderer escapes hand-written HTML anchors, which must still be
		// checked.
		for _, match := range hrefRegex.FindAllStringSubmatch(rendered+pg.Body, -1) {
			href := match[1]
			if !isInternalLink(href) {
				continue
			}

			route := normalizeLinkTarget(href)
			if route == "/" || known[route] || known[route+"/"] {
				continue
			}

			linkErr := &Error{
				Message: fmt.Sprintf("links to missing page %s", route),
				Path:    pg.SourcePath,
				Hint:    "Create the target page or fix the link",
			}

			if v.Config.Links.FailOnBroken {
				buildErr.Add(linkErr)
			} else {
				buildErr.AddWarning(linkErr)
			}
		}
	}
}

// isInternalLink reports whether an href refers to a site-internal route
// (as opposed to external URLs, anchors, mail, etc.).
func isInternalLink(href string) bool {
	href = strings.TrimSpace(href)
	if href == "" || href == "#" {
		return false
	}
	if strings.HasPrefix(href, "#") {
		return false
	}
	for _, prefix := range []string{"http://", "https://", "//", "mailto:", "tel:", "ftp:", "data:", "javascript:"} {
		if strings.HasPrefix(strings.ToLower(href), prefix) {
			return false
		}
	}
	return strings.HasPrefix(href, "/")
}

// normalizeLinkTarget strips anchors, query strings, trailing index.html and
// ensures a trailing slash so targets match route naming.
func normalizeLinkTarget(href string) string {
	href = strings.SplitN(href, "#", 2)[0]
	href = strings.SplitN(href, "?", 2)[0]
	href = strings.TrimSuffix(href, "index.html")
	href = strings.TrimSuffix(href, ".html")
	if !strings.HasPrefix(href, "/") {
		href = "/" + href
	}
	if !strings.HasSuffix(href, "/") {
		href += "/"
	}
	return href
}

// validateReferencedAssets warns about asset references that cannot be
// resolved. Only icons the user explicitly configured are checked, so the
// built-in defaults never produce noise.
func (v *Validator) validateReferencedAssets(buildErr *BuildError) {
	// Favicon configured but missing.
	iconFile := v.Config.Icon
	if v.Config.Favicon != "" {
		iconFile = v.Config.Favicon
	}
	if iconFile != "" && v.Config.IconConfigured {
		iconPath := v.Config.SourcePath(v.Config.IconDir, iconFile)
		if _, err := os.Stat(iconPath); os.IsNotExist(err) {
			buildErr.AddWarning(&Error{
				Message: fmt.Sprintf("icon file '%s' not found in %s", iconFile, v.Config.IconDir),
				Path:    iconPath,
				Hint:    "Add the icon file or update icon/icon_dir in site.yaml",
			})
		}
	}
}

// markdownRender renders markdown to HTML for link extraction.
var markdownRender = func(content string) (string, error) {
	return markdown.Render(content)
}

func (v *Validator) validateConfig(err *BuildError) {
	if v.Config.Name == "" {
		err.Add(&Error{
			Message: "site name is required",
			Path:    "site.yaml",
			Hint:    "Add 'name: Your Site Name' to site.yaml",
		})
	}
}

func (v *Validator) validateComponentFiles(comps map[string]components.Definition, err *BuildError) {
	for name, comp := range comps {
		if comp.File == "" {
			err.Add(&Error{
				Message: fmt.Sprintf("component '%s' has no file specified", name),
				Path:    "components/definitions.yaml",
				Hint:    fmt.Sprintf("Add 'file: %s.html' to the component definition", name),
			})
			continue
		}

		path := v.Config.SourcePath("components", comp.File)
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			err.Add(&Error{
				Message: fmt.Sprintf("component file '%s' not found", comp.File),
				Path:    path,
				Hint:    fmt.Sprintf("Create the file 'components/%s'", comp.File),
			})
		}

		if comp.Style != "" {
			stylePath := v.Config.SourcePath(comp.Style)
			if _, statErr := os.Stat(stylePath); os.IsNotExist(statErr) {
				err.Add(&Error{
					Message: fmt.Sprintf("stylesheet '%s' not found", comp.Style),
					Path:    stylePath,
					Hint:    fmt.Sprintf("Create the file '%s'", comp.Style),
				})
			}
		}
	}
}

// builtInInstances can be used in layouts without a component definition.
var builtInInstances = map[string]bool{
	"content":            true,
	"post-list":          true,
	"frontmatter-values": true,
}

func (v *Validator) validateLayouts(layos map[string]layouts.Layout, comps map[string]components.Definition, err *BuildError) {
	for name, layo := range layos {
		for _, inst := range layo.Components {
			if _, exists := comps[inst.Name]; !exists && !builtInInstances[inst.Name] {
				err.Add(&Error{
					Message: fmt.Sprintf("unknown component '%s' in layout '%s'", inst.Name, name),
					Path:    filepath.Join("layouts", name+".yaml"),
					Hint:    "Define the component in components/definitions.yaml or remove it from the layout",
				})
			}
		}
	}
}

func (v *Validator) validatePageTypes(pts map[string]pagetypes.PageType, layos map[string]layouts.Layout, err *BuildError) {
	for name, pt := range pts {
		if pt.Layout != "" {
			if _, exists := layos[pt.Layout]; !exists {
				err.Add(&Error{
					Message: fmt.Sprintf("layout '%s' not found for page type '%s'", pt.Layout, name),
					Path:    filepath.Join("page-types", name+".yaml"),
					Hint:    "Use a layout that exists in the layouts/ directory",
				})
			}
		}
	}
}

func (v *Validator) validatePages(pgs []pages.Page, pts map[string]pagetypes.PageType, err *BuildError) {
	for _, pg := range pgs {
		pt, exists := pts[pg.Type]
		if !exists {
			err.Add(&Error{
				Message: fmt.Sprintf("unknown page type '%s'", pg.Type),
				Path:    pg.SourcePath,
				Hint:    "Define the page type in page-types/ or set 'type' in frontmatter",
			})
			continue
		}

		metadata := pt.ApplyDefaults(pg.Metadata)
		if validationErrs := pt.Validate(metadata); len(validationErrs) > 0 {
			for _, verr := range validationErrs {
				err.Add(&Error{
					Message: verr.Error(),
					Path:    pg.SourcePath,
					Hint:    "Fix the frontmatter to match the page type schema",
				})
			}
		}

		if pg.Type == "post" {
			v.validatePost(pg, err)
		}
	}
}

func (v *Validator) validatePost(pg pages.Page, err *BuildError) {
	if tags, ok := pg.Metadata["tags"]; ok {
		if _, ok := tags.([]any); !ok {
			err.Add(&Error{
				Message: "post tags must be a list",
				Path:    pg.SourcePath,
				Hint:    "Use 'tags: [tag1, tag2]' in frontmatter",
			})
		}
	}

	if date, ok := pg.Metadata["date"]; ok {
		if s, ok := date.(string); ok {
			if len(s) != 10 || s[4] != '-' || s[7] != '-' {
				err.Add(&Error{
					Message: fmt.Sprintf("invalid date format '%s'", s),
					Path:    pg.SourcePath,
					Hint:    "Use YYYY-MM-DD format (e.g., 2024-01-15)",
				})
			}
		}
	}
}

func (v *Validator) validateRoutes(pgs []pages.Page, err *BuildError) {
	type routeInfo struct {
		route string
		path  string
	}

	routes := make([]routeInfo, 0, len(pgs))
	for _, pg := range pgs {
		routes = append(routes, routeInfo{pg.Route, pg.SourcePath})
	}

	routeStrings := make([]string, len(routes))
	pathStrings := make([]string, len(routes))
	for i, r := range routes {
		routeStrings[i] = r.route
		pathStrings[i] = r.path
	}

	duplicates := routing.EnsureUniqueRoutes(func() []struct {
		Route string
		Path  string
	} {
		result := make([]struct {
			Route string
			Path  string
		}, len(routes))
		for i, r := range routes {
			result[i] = struct {
				Route string
				Path  string
			}{r.route, r.path}
		}
		return result
	}())

	for _, dup := range duplicates {
		err.Add(&Error{
			Message: "duplicate route detected",
			Path:    dup,
			Hint:    "Each page must have a unique route. Rename or move one of the files.",
		})
	}
}
