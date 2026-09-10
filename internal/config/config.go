package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

type SiteConfig struct {
	Name            string           `yaml:"name"`
	BaseURL         string           `yaml:"base_url"`
	Build           BuildConfig      `yaml:"build"`
	DefaultLayout   string           `yaml:"default_layout"`
	DefaultPageType string           `yaml:"default_page_type"`
	Assets          AssetsConfig     `yaml:"assets"`
	Styles          StylesConfig     `yaml:"styles"`
	Scripts         []ScriptConfig   `yaml:"scripts"`
	Responsive      ResponsiveConfig `yaml:"responsive"`
	Links           LinksConfig      `yaml:"links"`
	Favicon         string           `yaml:"favicon"`
	Icon            string           `yaml:"icon"`
	IconDir         string           `yaml:"icon_dir"`
	Language        string           `yaml:"language"`
	SEO             SEOConfig        `yaml:"seo"`
	Sitemap         SitemapConfig    `yaml:"sitemap"`
	Feed            FeedConfig       `yaml:"feed"`
	Taxonomy        TaxonomyConfig   `yaml:"taxonomy"`
	Pagination      PaginationConfig `yaml:"pagination"`
	Gallery         GalleryConfig    `yaml:"gallery"`

	// IconConfigured reports whether icon/favicon was set explicitly in the
	// YAML (as opposed to coming from defaults). Used by validation to avoid
	// warning about icons the user never asked for.
	IconConfigured bool `yaml:"-"`
}

func (c *SiteConfig) UnmarshalYAML(value *yaml.Node) error {
	type plain SiteConfig
	var raw plain
	if err := value.Decode(&raw); err != nil {
		return err
	}
	*c = SiteConfig(raw)

	for i := 0; i+1 < len(value.Content); i += 2 {
		keyNode, valNode := value.Content[i], value.Content[i+1]
		if keyNode.Kind == yaml.ScalarNode && valNode != nil && valNode.Tag != "!!null" {
			if keyNode.Value == "icon" || keyNode.Value == "favicon" {
				c.IconConfigured = true
			}
		}
	}
	return nil
}

type ResponsiveConfig struct {
	Mobile  string `yaml:"mobile"`
	Tablet  string `yaml:"tablet"`
	Desktop string `yaml:"desktop"`
}

// LinksConfig accepts either the legacy list form
//
//	links:
//	  - name: Home
//	    url: /
//
// or the map form with fail_on_broken:
//
//	links:
//	  fail_on_broken: true
//	  entries:
//	    - name: Home
//	      url: /
type LinksConfig struct {
	Entries      []LinkConfig
	FailOnBroken bool
}

func (l *LinksConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var list []LinkConfig
	if err := unmarshal(&list); err == nil {
		l.Entries = list
		return nil
	}

	var raw struct {
		FailOnBroken bool         `yaml:"fail_on_broken"`
		Entries      []LinkConfig `yaml:"entries"`
	}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	l.FailOnBroken = raw.FailOnBroken
	l.Entries = raw.Entries
	return nil
}

type LinkConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type SEOConfig struct {
	Author      string `yaml:"author"`
	Keywords    string `yaml:"keywords"`
	ThemeColor  string `yaml:"theme_color"`
	SitemapFile string `yaml:"sitemap_file"`
	RobotsFile  string `yaml:"robots_file"`
}

type BuildConfig struct {
	Source string `yaml:"source"`
	Output string `yaml:"output"`
}

type AssetsConfig struct {
	Directory string `yaml:"directory"`
}

type StylesConfig struct {
	Global []string `yaml:"global"`
}

type ScriptConfig struct {
	Path string `yaml:"src"`
	Type string `yaml:"type"`
}

type SitemapConfig struct {
	Enabled bool `yaml:"enabled"`
}

type FeedConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
	Title   string `yaml:"title"`
}

type TaxonomyConfig struct {
	Tags       bool `yaml:"tags"`
	Categories bool `yaml:"categories"`
}

type PaginationConfig struct {
	Enabled bool `yaml:"enabled"`
	PerPage int  `yaml:"per_page"`
}

type GalleryConfig struct {
	Enabled bool `yaml:"enabled"`
}

func Load(path string) (*SiteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg SiteConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *SiteConfig) {
	if cfg.Build.Source == "" {
		cfg.Build.Source = "."
	}
	if cfg.Build.Output == "" {
		cfg.Build.Output = "dist"
	}
	if cfg.DefaultLayout == "" {
		cfg.DefaultLayout = "default"
	}
	if cfg.DefaultPageType == "" {
		cfg.DefaultPageType = "page"
	}
	if cfg.Assets.Directory == "" {
		cfg.Assets.Directory = "assets"
	}
	if cfg.Language == "" {
		cfg.Language = "en"
	}
	if cfg.SEO.SitemapFile == "" {
		cfg.SEO.SitemapFile = "sitemap.xml"
	}
	if cfg.SEO.RobotsFile == "" {
		cfg.SEO.RobotsFile = "robots.txt"
	}
	if cfg.Responsive.Mobile == "" {
		cfg.Responsive.Mobile = "768px"
	}
	if cfg.Responsive.Tablet == "" {
		cfg.Responsive.Tablet = "1024px"
	}
	if cfg.Responsive.Desktop == "" {
		cfg.Responsive.Desktop = "1200px"
	}
	if cfg.Icon == "" {
		cfg.Icon = "icon.png"
	}
	if cfg.IconDir == "" {
		cfg.IconDir = "assets/icon"
	}
	if cfg.Feed.Path == "" {
		cfg.Feed.Path = "/feed.xml"
	}
	if cfg.Feed.Title == "" {
		cfg.Feed.Title = cfg.Name
	}
	if cfg.Pagination.PerPage <= 0 {
		cfg.Pagination.PerPage = 10
	}
}

// ValidateResponsive checks that the responsive breakpoint configuration is
// valid: values are non-empty, valid CSS lengths, and ordered logically
// (mobile < tablet < desktop).
func (c *SiteConfig) ValidateResponsive() error {
	if c.Responsive.Mobile == "" {
		return fmt.Errorf("responsive.mobile is empty")
	}
	if c.Responsive.Tablet == "" {
		return fmt.Errorf("responsive.tablet is empty")
	}
	if c.Responsive.Desktop == "" {
		return fmt.Errorf("responsive.desktop is empty")
	}

	for _, field := range []struct {
		name  string
		value string
	}{
		{"mobile", c.Responsive.Mobile},
		{"tablet", c.Responsive.Tablet},
		{"desktop", c.Responsive.Desktop},
	} {
		if !isValidCSSLength(field.value) {
			return fmt.Errorf("responsive.%s: invalid CSS length %q (expected a value like \"768px\")", field.name, field.value)
		}
	}

	mobile := parseCSSLength(c.Responsive.Mobile)
	tablet := parseCSSLength(c.Responsive.Tablet)
	desktop := parseCSSLength(c.Responsive.Desktop)

	if mobile >= tablet {
		return fmt.Errorf("responsive: breakpoints must be ordered mobile < tablet < desktop, but mobile (%s) >= tablet (%s)", c.Responsive.Mobile, c.Responsive.Tablet)
	}
	if tablet >= desktop {
		return fmt.Errorf("responsive: breakpoints must be ordered mobile < tablet < desktop, but tablet (%s) >= desktop (%s)", c.Responsive.Tablet, c.Responsive.Desktop)
	}

	return nil
}

// cssLengthRe matches a numeric value followed by a CSS unit.
var cssLengthRe = regexp.MustCompile(`^\s*(\d+(?:\.\d+)?)\s*(px|rem|em|vw|vh|%)\s*$`)

// isValidCSSLength reports whether s is a valid CSS length value (e.g. "768px").
func isValidCSSLength(s string) bool {
	return cssLengthRe.MatchString(s)
}

// parseCSSLength extracts the numeric portion of a CSS length for comparison.
func parseCSSLength(s string) float64 {
	m := cssLengthRe.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	v, _ := strconv.ParseFloat(m[1], 64)
	return v
}

func (c *SiteConfig) SourcePath(parts ...string) string {
	return filepath.Join(append([]string{c.Build.Source}, parts...)...)
}

func (c *SiteConfig) OutputPath(parts ...string) string {
	return filepath.Join(append([]string{c.Build.Output}, parts...)...)
}
