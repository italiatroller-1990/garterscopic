package config

import (
	"os"
	"path/filepath"

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
	Links           []LinkConfig     `yaml:"links"`
	Favicon         string           `yaml:"favicon"`
	Language        string           `yaml:"language"`
	SEO             SEOConfig        `yaml:"seo"`
}

type ResponsiveConfig struct {
	Mobile  string `yaml:"mobile"`
	Tablet  string `yaml:"tablet"`
	Desktop string `yaml:"desktop"`
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
}

func (c *SiteConfig) SourcePath(parts ...string) string {
	return filepath.Join(append([]string{c.Build.Source}, parts...)...)
}

func (c *SiteConfig) OutputPath(parts ...string) string {
	return filepath.Join(append([]string{c.Build.Output}, parts...)...)
}
