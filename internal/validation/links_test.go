package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/components"
	"github.com/italiatroller-1990/garterscopic/internal/config"
	"github.com/italiatroller-1990/garterscopic/internal/layouts"
	"github.com/italiatroller-1990/garterscopic/internal/pages"
	"github.com/italiatroller-1990/garterscopic/internal/pagetypes"
)

func TestIsInternalLink(t *testing.T) {
	tests := []struct {
		href string
		want bool
	}{
		{"/about/", true},
		{"/blog/foo/", true},
		{"http://example.com", false},
		{"https://example.com/x", false},
		{"//cdn.example.com/x", false},
		{"mailto:a@b.c", false},
		{"tel:+123", false},
		{"javascript:void(0)", false},
		{"data:image/png;base64,xxx", false},
		{"#section", false},
		{"", false},
		{"#", false},
	}

	for _, tt := range tests {
		if got := isInternalLink(tt.href); got != tt.want {
			t.Errorf("isInternalLink(%q) = %v, want %v", tt.href, got, tt.want)
		}
	}
}

func TestNormalizeLinkTarget(t *testing.T) {
	tests := []struct {
		href string
		want string
	}{
		{"/about/", "/about/"},
		{"/about", "/about/"},
		{"/about/index.html", "/about/"},
		{"/about.html", "/about/"},
		{"/about/#section", "/about/"},
		{"/about/?x=1", "/about/"},
		{"/", "/"},
	}

	for _, tt := range tests {
		if got := normalizeLinkTarget(tt.href); got != tt.want {
			t.Errorf("normalizeLinkTarget(%q) = %q, want %q", tt.href, got, tt.want)
		}
	}
}

func TestValidateInternalLinksWarning(t *testing.T) {
	cfg := &config.SiteConfig{
		Name:  "Test",
		Build: config.BuildConfig{Source: ".", Output: "dist"},
	}

	pgs := []pages.Page{
		{Route: "/", Type: "page", Body: "[home](/) and [missing](/missing/)"},
		{Route: "/blog/foo/", Type: "post", Body: `<a href="/projects/">projects</a>`},
	}

	buildErr := &BuildError{}
	v := New(cfg)
	v.validateInternalLinks(pgs, buildErr)

	if !buildErr.HasWarnings() {
		t.Fatal("expected broken link warnings")
	}
	if len(buildErr.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(buildErr.Warnings))
	}
	if !strings.Contains(buildErr.Warnings[0].Message, "/missing/") {
		t.Errorf("unexpected warning: %s", buildErr.Warnings[0].Message)
	}
	if buildErr.HasErrors() {
		t.Error("warnings must not be errors by default")
	}
}

func TestValidateInternalLinksFailOnBroken(t *testing.T) {
	cfg := &config.SiteConfig{
		Name:  "Test",
		Build: config.BuildConfig{Source: ".", Output: "dist"},
	}
	cfg.Links.FailOnBroken = true

	pgs := []pages.Page{
		{Route: "/", Type: "page", Body: "[missing](/missing/)"},
	}

	buildErr := &BuildError{}
	v := New(cfg)
	v.validateInternalLinks(pgs, buildErr)

	if !buildErr.HasErrors() {
		t.Fatal("expected broken link error when fail_on_broken is set")
	}
}

func TestValidateInternalLinksSkipsKnownAndExternal(t *testing.T) {
	cfg := &config.SiteConfig{
		Name:  "Test",
		Build: config.BuildConfig{Source: ".", Output: "dist"},
	}

	pgs := []pages.Page{
		{
			Route: "/",
			Type:  "page",
			Body:  "[a](/) [b](/about/) [c](https://example.com) [d](#top)",
		},
		{Route: "/about/", Type: "page"},
	}

	buildErr := &BuildError{}
	v := New(cfg)
	v.validateInternalLinks(pgs, buildErr)

	if buildErr.HasWarnings() || buildErr.HasErrors() {
		t.Errorf("expected no findings, got warnings=%d errors=%d", len(buildErr.Warnings), len(buildErr.Errors))
	}
}

func TestValidateAllNoFindings(t *testing.T) {
	sourceDir := t.TempDir()
	writeFile := func(rel, content string) {
		t.Helper()
		path := filepath.Join(sourceDir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeFile("components/card.html", `<div class="card">{{ title }}</div>`)
	writeFile("styles/card.css", ".card { color: red; }")

	cfg := &config.SiteConfig{
		Name:  "Test",
		Build: config.BuildConfig{Source: sourceDir, Output: "dist"},
	}

	comps := map[string]components.Definition{
		"card": {Name: "card", File: "card.html", Style: "styles/card.css"},
	}
	layos := map[string]layouts.Layout{
		"default": {Name: "default", Components: []layouts.ComponentInstance{
			{Name: "card", Position: layouts.PositionTop},
			{Name: "content"},
		}},
	}
	pts := map[string]pagetypes.PageType{
		"page": {Name: "page", Layout: "default"},
	}
	pgs := []pages.Page{
		{Route: "/", SourcePath: "pages/index.md", Type: "page", Metadata: map[string]any{"title": "Home"}},
	}

	buildErr := New(cfg).ValidateAll(comps, layos, pts, pgs, nil)
	if buildErr != nil {
		for _, e := range buildErr.Errors {
			t.Logf("ERR: %s | %s", e.Message, e.Path)
		}
		t.Fatalf("expected no findings, got errors=%d warnings=%d", len(buildErr.Errors), len(buildErr.Warnings))
	}
}

func TestValidateInstanceOptions(t *testing.T) {
	cfg := &config.SiteConfig{
		Name:  "Test",
		Build: config.BuildConfig{Source: ".", Output: "dist"},
	}

	comps := map[string]components.Definition{
		"alert": {
			Name: "alert",
			Options: map[string]components.OptionDefinition{
				"message": {Type: "string", Required: true},
				"level":   {Type: "string"},
			},
		},
	}
	layos := map[string]layouts.Layout{
		"default": {Name: "default", Components: []layouts.ComponentInstance{
			{Name: "alert", Position: layouts.Position("sideways"), Options: map[string]any{
				"message": 123,   // wrong type (allowed: int->string coercion) - use unknown opt below
				"color":   "red", // unknown option
			}},
			{Name: "alert", Options: map[string]any{}}, // missing required
		}},
	}

	buildErr := &BuildError{}
	v := New(cfg)
	v.validateInstanceOptions(layos, comps, buildErr)

	if !buildErr.HasErrors() {
		t.Fatal("expected instance option errors")
	}

	found := map[string]bool{}
	for _, e := range buildErr.Errors {
		switch {
		case strings.Contains(e.Message, "invalid position"):
			found["position"] = true
		case strings.Contains(e.Message, "unknown option 'color'"):
			found["unknown"] = true
		case strings.Contains(e.Message, "required option 'message' is missing"):
			found["required"] = true
		}
	}
	for _, key := range []string{"position", "unknown", "required"} {
		if !found[key] {
			t.Errorf("expected %q error, got: %v", key, buildErr.Errors)
		}
	}
}

func TestValidateGlobalStyles(t *testing.T) {
	cfg := &config.SiteConfig{
		Name:  "Test",
		Build: config.BuildConfig{Source: t.TempDir(), Output: "dist"},
	}
	cfg.Styles.Global = []string{"styles/missing.css"}

	buildErr := &BuildError{}
	v := New(cfg)
	v.validateGlobalStyles(buildErr)

	if !buildErr.HasErrors() {
		t.Fatal("expected missing global stylesheet error")
	}
}
