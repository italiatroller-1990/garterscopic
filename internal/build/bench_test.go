package build

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/garterscopic/garterscopic/internal/config"
)

func benchmarkSite(b *testing.B, pageCount int) {
	tmpdir, err := os.MkdirTemp("", "bench")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { os.RemoveAll(tmpdir) })

	createBenchSite(b, tmpdir, pageCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		os.RemoveAll(filepath.Join(tmpdir, "dist"))
		b.StartTimer()

		cfg := &config.SiteConfig{
			Name:            "Bench",
			BaseURL:         "http://localhost",
			Build:           config.BuildConfig{Source: tmpdir, Output: filepath.Join(tmpdir, "dist")},
			DefaultLayout:   "default",
			DefaultPageType: "page",
			Assets:          config.AssetsConfig{Directory: "assets"},
			Styles:          config.StylesConfig{Global: []string{"styles/global.css"}},
		}
		builder := New(cfg)
		if err := builder.Load(); err != nil {
			b.Fatal(err)
		}
		_, err := builder.Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild100(b *testing.B) {
	benchmarkSite(b, 100)
}

func BenchmarkBuild1000(b *testing.B) {
	benchmarkSite(b, 1000)
}

func BenchmarkBuild10000(b *testing.B) {
	benchmarkSite(b, 10000)
}

func createBenchSite(b *testing.B, tmpdir string, pageCount int) {
	dirs := []string{
		"components", "layouts", "page-types", "pages",
		"styles", "assets",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(tmpdir, d), 0755)
	}

	os.WriteFile(filepath.Join(tmpdir, "site.yaml"), []byte(`name: Bench
base_url: http://localhost
build:
  source: .
  output: dist
default_layout: default
default_page_type: page
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "components", "definitions.yaml"), []byte(`components:
  navbar:
    file: navbar.html
    style: navbar.css
    options:
      logo:
        type: string
  hero:
    file: hero.html
    style: hero.css
    options:
      title:
        type: string
        required: true
  footer:
    file: footer.html
    style: footer.css
    options:
      copyright:
        type: string
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "components", "navbar.html"), []byte(`<nav>{{ logo }}</nav>`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "components", "hero.html"), []byte(`<h1>{{ title }}</h1>`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "components", "footer.html"), []byte(`<footer>{{ copyright }}</footer>`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "layouts", "default.yaml"), []byte(`name: default
components:
  - name: navbar
    options:
      logo: Test
    position: top
  - name: hero
    options:
      title: Welcome
  - name: content
  - name: footer
    options:
      copyright: "2024"
    position: bottom
`), 0644)

	os.WriteFile(filepath.Join(tmpdir, "page-types", "page.yaml"), []byte(`name: page
layout: default
fields:
  title:
    type: string
    required: true
`), 0644)

	for i := 0; i < pageCount; i++ {
		content := fmt.Sprintf(`---
type: page
title: Page %d
---

# Page %d

This is page number %d.

## Section

Some content here with more text.

- Item 1
- Item 2
- Item 3

[Link](/about/)
`, i, i, i)

		filename := filepath.Join(tmpdir, "pages", fmt.Sprintf("page-%d.md", i))
		os.WriteFile(filename, []byte(content), 0644)
	}

	os.WriteFile(filepath.Join(tmpdir, "styles", "global.css"), []byte(`body { margin: 0; }`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "styles", "navbar.css"), []byte(`nav { color: blue; }`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "styles", "hero.css"), []byte(`h1 { font-size: 2em; }`), 0644)
	os.WriteFile(filepath.Join(tmpdir, "styles", "footer.css"), []byte(`footer { color: gray; }`), 0644)
}
