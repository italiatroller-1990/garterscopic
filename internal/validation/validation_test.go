package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/components"
	"github.com/italiatroller-1990/garterscopic/internal/config"
	"github.com/italiatroller-1990/garterscopic/internal/layouts"
	"github.com/italiatroller-1990/garterscopic/internal/pages"
	"github.com/italiatroller-1990/garterscopic/internal/pagetypes"
)

func TestValidatePostTags(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "validation_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	os.MkdirAll(filepath.Join(tmpdir, "page-types"), 0755)
	os.WriteFile(filepath.Join(tmpdir, "page-types", "post.yaml"), []byte(`name: post
layout: default
fields:
  title:
    type: string
    required: true
  date:
    type: date
    required: true
  tags:
    type: list
`), 0644)

	cfg := &config.SiteConfig{
		Name: "Test",
		Build: config.BuildConfig{
			Source: tmpdir,
			Output: "dist",
		},
		DefaultLayout:   "default",
		DefaultPageType: "page",
		Responsive: config.ResponsiveConfig{
			Mobile:  "768px",
			Tablet:  "1024px",
			Desktop: "1200px",
		},
	}

	tests := []struct {
		name    string
		tags    any
		wantErr bool
	}{
		{
			name:    "valid tags",
			tags:    []any{"linux", "tutorial"},
			wantErr: false,
		},
		{
			name:    "empty tags",
			tags:    []any{},
			wantErr: false,
		},
		{
			name:    "nil tags",
			tags:    nil,
			wantErr: false,
		},
		{
			name:    "invalid tags - string",
			tags:    "linux",
			wantErr: true,
		},
		{
			name:    "invalid tags - number",
			tags:    123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := map[string]any{
				"title": "Test Post",
				"date":  "2024-01-15",
			}
			if tt.tags != nil {
				metadata["tags"] = tt.tags
			}

			pg := pages.Page{
				Type:     "post",
				Metadata: metadata,
			}

			v := New(cfg)
			buildErr := &BuildError{}
			v.validatePost(pg, buildErr)

			if tt.wantErr && len(buildErr.Errors) == 0 {
				t.Error("expected validation error, got none")
			}
			if !tt.wantErr && len(buildErr.Errors) > 0 {
				t.Errorf("unexpected validation error: %v", buildErr.Errors[0])
			}
		})
	}
}

func TestValidatePostDate(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "validation_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	cfg := &config.SiteConfig{
		Name: "Test",
		Build: config.BuildConfig{
			Source: tmpdir,
			Output: "dist",
		},
		Responsive: config.ResponsiveConfig{
			Mobile:  "768px",
			Tablet:  "1024px",
			Desktop: "1200px",
		},
	}

	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{
			name:    "valid date",
			date:    "2024-01-15",
			wantErr: false,
		},
		{
			name:    "invalid date - no dashes",
			date:    "20240115",
			wantErr: true,
		},
		{
			name:    "invalid date - wrong format",
			date:    "01/15/2024",
			wantErr: true,
		},
		{
			name:    "invalid date - too short",
			date:    "2024-01",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := map[string]any{
				"title": "Test Post",
				"date":  tt.date,
			}

			pg := pages.Page{
				Type:     "post",
				Metadata: metadata,
			}

			v := New(cfg)
			buildErr := &BuildError{}
			v.validatePost(pg, buildErr)

			if tt.wantErr && len(buildErr.Errors) == 0 {
				t.Error("expected validation error, got none")
			}
			if !tt.wantErr && len(buildErr.Errors) > 0 {
				t.Errorf("unexpected validation error: %v", buildErr.Errors[0])
			}
		})
	}
}

func TestValidateAllWithPosts(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "validation_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	os.MkdirAll(filepath.Join(tmpdir, "layouts"), 0755)
	os.WriteFile(filepath.Join(tmpdir, "layouts", "default.yaml"), []byte(`name: default
components:
  - content
`), 0644)

	os.MkdirAll(filepath.Join(tmpdir, "page-types"), 0755)
	os.WriteFile(filepath.Join(tmpdir, "page-types", "page.yaml"), []byte(`name: page
layout: default
fields:
  title:
    type: string
    required: true
`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "page-types", "post.yaml"), []byte(`name: post
layout: default
fields:
  title:
    type: string
    required: true
  date:
    type: date
    required: true
`), 0644)

	cfg := &config.SiteConfig{
		Name: "Test",
		Build: config.BuildConfig{
			Source: tmpdir,
			Output: "dist",
		},
		Responsive: config.ResponsiveConfig{
			Mobile:  "768px",
			Tablet:  "1024px",
			Desktop: "1200px",
		},
	}

	pts, _ := pagetypes.LoadPageTypes(cfg)
	layos, _ := layouts.LoadLayouts(cfg)
	comps := map[string]components.Definition{}

	pgs := []pages.Page{
		{
			Type: "page",
			Metadata: map[string]any{
				"title": "About",
			},
			Route: "/about/",
		},
	}

	posts := []pages.Page{
		{
			Type: "post",
			Metadata: map[string]any{
				"title": "Blog Post",
				"date":  "2024-01-15",
			},
			Route:   "/blog/post/",
			Section: "blog",
		},
	}

	v := New(cfg)
	buildErr := v.ValidateAll(comps, layos, pts, pgs, posts)

	if buildErr != nil {
		t.Errorf("unexpected validation error: %v", buildErr)
	}
}

func TestValidateAllWithInvalidPost(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "validation_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	os.MkdirAll(filepath.Join(tmpdir, "layouts"), 0755)
	os.WriteFile(filepath.Join(tmpdir, "layouts", "default.yaml"), []byte(`name: default
components:
  - content
`), 0644)

	os.MkdirAll(filepath.Join(tmpdir, "page-types"), 0755)
	os.WriteFile(filepath.Join(tmpdir, "page-types", "post.yaml"), []byte(`name: post
layout: default
fields:
  title:
    type: string
    required: true
  date:
    type: date
    required: true
`), 0644)

	cfg := &config.SiteConfig{
		Name: "Test",
		Build: config.BuildConfig{
			Source: tmpdir,
			Output: "dist",
		},
		Responsive: config.ResponsiveConfig{
			Mobile:  "768px",
			Tablet:  "1024px",
			Desktop: "1200px",
		},
	}

	pts, _ := pagetypes.LoadPageTypes(cfg)
	layos, _ := layouts.LoadLayouts(cfg)
	comps := map[string]components.Definition{}

	posts := []pages.Page{
		{
			Type: "post",
			Metadata: map[string]any{
				"title": "Blog Post",
			},
			Route:      "/blog/post/",
			Section:    "blog",
			SourcePath: filepath.Join(tmpdir, "posts", "blog", "post.md"),
		},
	}

	v := New(cfg)
	buildErr := v.ValidateAll(comps, layos, pts, nil, posts)

	if buildErr == nil {
		t.Error("expected validation error for missing date")
	}
}

func TestValidateResponsiveConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.SiteConfig
		wantErr  bool
		errField string
	}{
		{
			name: "valid responsive config",
			cfg: config.SiteConfig{
				Name:          "Test",
				Responsive:    config.ResponsiveConfig{Mobile: "690px", Tablet: "820px", Desktop: "1000px"},
				DefaultLayout: "default",
			},
			wantErr: false,
		},
		{
			name: "invalid ordering",
			cfg: config.SiteConfig{
				Name:          "Test",
				Responsive:    config.ResponsiveConfig{Mobile: "900px", Tablet: "500px", Desktop: "1000px"},
				DefaultLayout: "default",
			},
			wantErr:  true,
			errField: "responsive",
		},
		{
			name: "invalid CSS length",
			cfg: config.SiteConfig{
				Name:          "Test",
				Responsive:    config.ResponsiveConfig{Mobile: "mobile", Tablet: "820px", Desktop: "1000px"},
				DefaultLayout: "default",
			},
			wantErr:  true,
			errField: "responsive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(&tt.cfg)
			buildErr := &BuildError{}
			v.validateResponsiveConfig(buildErr)

			if tt.wantErr && len(buildErr.Errors) == 0 {
				t.Error("expected validation error, got none")
			}
			if !tt.wantErr && len(buildErr.Errors) > 0 {
				t.Errorf("unexpected validation error: %v", buildErr.Errors[0])
			}
		})
	}
}
