package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/italiatroller-1990/garterscopic/internal/build"
	"github.com/italiatroller-1990/garterscopic/internal/config"
	"github.com/italiatroller-1990/garterscopic/internal/server"
)

var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "build":
		runBuild()
	case "dev":
		runDev()
	case "clean":
		runClean()
	case "init":
		runInit()
	case "new":
		runNew()
	case "--help", "-h", "help":
		printHelp()
	case "--version", "-v", "version":
		fmt.Printf("garterscopic %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`Garterscopic - Declarative Lightweight Static Site Generator

Usage:
  garterscopic [command]

Commands:
  build           Build the site
  dev             Start the development server
  clean           Remove generated files
  init [name]     Create a new project
  new page <name>         Create a new page
  new post [section/]slug Create a new post (optionally in a section)

Options:
  --help, -h      Show this help
  --version, -v   Show version`)
}

func runBuild() {
	cfg, err := config.Load("site.yaml")
	if err != nil {
		printError("failed to load configuration", err)
		os.Exit(1)
	}

	fmt.Printf("Garterscopic %s\n\nBuilding site...\n\n", version)

	builder := build.New(cfg)
	if err := builder.Load(); err != nil {
		printError("failed to load site", err)
		os.Exit(1)
	}

	if buildErr := builder.Validate(); buildErr != nil {
		fmt.Fprintf(os.Stderr, "Validation failed:\n%s\n", buildErr.Error())
		os.Exit(1)
	}

	result, err := builder.Build()
	if err != nil {
		printError("build failed", err)
		os.Exit(1)
	}

	fmt.Println("✓ Loaded configuration")
	fmt.Printf("✓ Loaded %d components\n", result.Components)
	fmt.Printf("✓ Loaded %d layouts\n", result.Layouts)
	fmt.Printf("✓ Loaded %d page types\n", result.PageTypes)
	fmt.Printf("✓ Processed %d pages\n", result.PagesGenerated)
	fmt.Printf("✓ Copied %d assets\n", result.AssetsCopied)
	fmt.Printf("✓ Generated %d pages\n\n", result.PagesGenerated)
	fmt.Printf("Build complete.\n\nOutput: %s/\n", cfg.Build.Output)
}

func runDev() {
	cfg, err := config.Load("site.yaml")
	if err != nil {
		printError("failed to load configuration", err)
		os.Exit(1)
	}

	port := 8080
	if len(os.Args) > 2 && os.Args[2] == "--port" && len(os.Args) > 3 {
		fmt.Sscanf(os.Args[3], "%d", &port)
	}

	fmt.Printf("Garterscopic %s\n\nBuilding initial site...\n\n", version)

	builder := build.New(cfg)
	if err := builder.Load(); err != nil {
		printError("failed to load site", err)
		os.Exit(1)
	}

	result, err := builder.Build()
	if err != nil {
		printError("build failed", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Built %d pages\n\n", result.PagesGenerated)
	fmt.Printf("Starting development server on http://localhost:%d...\n", port)
	fmt.Println("Press Ctrl+C to stop.")

	srv := server.New(cfg, port, builder)
	if err := srv.Start(); err != nil {
		printError("server error", err)
		os.Exit(1)
	}
}

func runClean() {
	cfg, err := config.Load("site.yaml")
	if err != nil {
		printError("failed to load configuration", err)
		os.Exit(1)
	}

	outputDir := cfg.Build.Output
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		fmt.Println("Nothing to clean.")
		return
	}

	if err := os.RemoveAll(outputDir); err != nil {
		printError("failed to clean", err)
		os.Exit(1)
	}

	fmt.Printf("Removed: %s/\n", outputDir)
}

func runInit() {
	name := "my-site"
	if len(os.Args) > 2 {
		name = os.Args[2]
	}

	if err := initProject(name); err != nil {
		printError("failed to create project", err)
		os.Exit(1)
	}

	fmt.Printf("Created new project: %s/\n\n", name)
	fmt.Println("To get started:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  garterscopic build")
}

func initProject(name string) error {
	dir := name

	templates := map[string]string{
		"site.yaml": fmt.Sprintf(`name: %s
base_url: http://localhost:8080

build:
  source: .
  output: dist

default_layout: default
default_page_type: page

assets:
  directory: assets

styles:
  global:
    - styles/global.css

responsive:
  mobile: "768px"
  tablet: "1024px"
  desktop: "1200px"

links:
  - name: Home
    url: /
  - name: About
    url: /about/
`, strings.Title(name)),

		"components/definitions.yaml": `components:
  navbar:
    file: navbar.html
    style: styles/navbar.css
    options:
      logo:
        type: string
      links:
        type: links

  hero:
    file: hero.html
    style: styles/hero.css
    options:
      title:
        type: string
        required: true
      subtitle:
        type: string
      button:
        type: object

  footer:
    file: footer.html
    style: styles/footer.css
    options:
      copyright:
        type: string
`,

		"components/navbar.html": `<nav class="navbar">
    <a class="navbar-logo" href="/">{{ logo }}</a>
    <ul class="navbar-links">
        {{ links }}
    </ul>
</nav>
`,

		"components/hero.html": `<section class="hero">
    <h1 class="hero-title">{{ title }}</h1>
    <p class="hero-subtitle">{{ subtitle }}</p>
</section>
`,

		"components/footer.html": `<footer class="footer">
    <p class="footer-copyright">{{ copyright }}</p>
</footer>
`,

		"layouts/default.yaml": `name: default

components:
  - name: navbar
    options:
      logo: My Site
      from:
        links: site.links

  - name: content

  - name: footer
    options:
      copyright: "2024"
`,

		"page-types/page.yaml": `name: page

layout: default

fields:
  title:
    type: string
    required: true

  description:
    type: string

  draft:
    type: boolean
    default: false
`,

		"page-types/post.yaml": `name: post

layout: default

fields:
  title:
    type: string
    required: true

  description:
    type: string

  date:
    type: date
    required: true

  updated:
    type: date

  author:
    type: string

  tags:
    type: list

  section:
    type: string

  slug:
    type: string

  draft:
    type: boolean
    default: false
`,

		"pages/index.md": `---
type: page
title: Welcome
description: Welcome to my site
---

# Welcome

This is a Garterscopic website! 

## Getting Started

1. Edit the pages in the ` + "`" + `pages/` + "`" + ` directory
2. Modify components in ` + "`" + `components/` + "`" + `
3. Update the layout in ` + "`" + `layouts/` + "`" + `

Happy building!
`,

		"pages/about.md": `---
type: page
title: About
description: About this website
---

# About

This is the about page.

## What is Garterscopic?

Garterscopic is a static site generator that lets you build websites using:
- HTML components
- YAML configuration  
- Markdown content
- CSS styling

No Node.js required!
`,

		"styles/global.css": `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    line-height: 1.6;
    color: #333;
}

a {
    color: #0066cc;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
}
`,

		"styles/navbar.css": `.navbar {
    background: #333;
    color: white;
    padding: 1rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.navbar-logo {
    color: white;
    font-size: 1.5rem;
    font-weight: bold;
}

.navbar-links {
    list-style: none;
    display: flex;
    gap: 1rem;
}

.navbar-links a {
    color: white;
}
`,

		"styles/hero.css": `.hero {
    padding: 4rem 2rem;
    text-align: center;
    background: #f5f5f5;
}

.hero-title {
    font-size: 3rem;
    margin-bottom: 1rem;
}

.hero-subtitle {
    font-size: 1.25rem;
    color: #666;
}
`,

		"styles/footer.css": `.footer {
    background: #333;
    color: white;
    padding: 2rem;
    text-align: center;
}

.footer-copyright {
    margin: 0;
}
`,

		"assets/js/main.js": `console.log("Garterscopic site loaded!");
`,

		"README.md": "# " + strings.Title(name) + "\n\nBuilt with Garterscopic.\n\n## Commands\n\n```bash\ngarterscopic build   # Build the site\ngarterscopic dev     # Start development server\ngarterscopic clean   # Remove generated files\n```\n",
	}

	for path, content := range templates {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func runNew() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: garterscopic new page <name>")
		fmt.Fprintln(os.Stderr, "       garterscopic new post [section/]slug")
		os.Exit(1)
	}

	pageType := os.Args[2]
	pageName := os.Args[3]

	if pageType != "page" && pageType != "post" {
		fmt.Fprintf(os.Stderr, "Unknown page type: %s (use 'page' or 'post')\n", pageType)
		os.Exit(1)
	}

	if pageType == "page" {
		createPage(pageName)
	} else {
		createPost(pageName)
	}
}

func createPage(name string) {
	dir := "pages"

	if err := os.MkdirAll(dir, 0755); err != nil {
		printError("failed to create directory", err)
		os.Exit(1)
	}

	filename := name + ".md"
	filePath := filepath.Join(dir, filename)

	if _, err := os.Stat(filePath); err == nil {
		fmt.Fprintf(os.Stderr, "File already exists: %s\n", filePath)
		os.Exit(1)
	}

	content := fmt.Sprintf(`---
type: page
title: %s
---

# %s

Your content here.
`, strings.Title(name), strings.Title(name))

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		printError("failed to create page", err)
		os.Exit(1)
	}

	fmt.Printf("Created: %s\n", filePath)
}

func createPost(name string) {
	parts := strings.SplitN(name, "/", 2)
	var section, slug string

	if len(parts) == 2 {
		section = parts[0]
		slug = parts[1]
	} else {
		slug = parts[0]
	}

	dir := "posts"
	if section != "" {
		dir = filepath.Join("posts", section)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		printError("failed to create directory", err)
		os.Exit(1)
	}

	filename := slug + ".md"
	filePath := filepath.Join(dir, filename)

	if _, err := os.Stat(filePath); err == nil {
		fmt.Fprintf(os.Stderr, "File already exists: %s\n", filePath)
		os.Exit(1)
	}

	content := fmt.Sprintf(`---
type: post
title: %s
date: %s
tags: []
---

# %s

Your content here.
`, strings.Title(slug), timeNow().Format("2006-01-02"), strings.Title(slug))

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		printError("failed to create post", err)
		os.Exit(1)
	}

	fmt.Printf("Created: %s\n", filePath)
}

func timeNow() time.Time {
	return time.Now()
}

func printError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%v\n", msg, err)
}
