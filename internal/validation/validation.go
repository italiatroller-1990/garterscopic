package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/italiatroller-1990/garterscopic/internal/components"
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
	Errors []*Error
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

type Validator struct {
	Config     *config.SiteConfig
	Components map[string]components.Definition
	Layouts    map[string]layouts.Layout
	PageTypes  map[string]pagetypes.PageType
	Pages      []pages.Page
}

func New(cfg *config.SiteConfig) *Validator {
	return &Validator{Config: cfg}
}

func (v *Validator) ValidateAll(comps map[string]components.Definition, layos map[string]layouts.Layout, pts map[string]pagetypes.PageType, pgs []pages.Page) *BuildError {
	buildErr := &BuildError{}

	v.validateConfig(buildErr)
	v.validateComponentFiles(comps, buildErr)
	v.validateLayouts(layos, comps, buildErr)
	v.validatePageTypes(pts, layos, buildErr)
	v.validatePages(pgs, pts, buildErr)
	v.validateRoutes(pgs, buildErr)

	if len(buildErr.Errors) == 0 {
		return nil
	}
	return buildErr
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

func (v *Validator) validateLayouts(layos map[string]layouts.Layout, comps map[string]components.Definition, err *BuildError) {
	for name, layo := range layos {
		for _, inst := range layo.Components {
			if _, exists := comps[inst.Name]; !exists && inst.Name != "content" {
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
