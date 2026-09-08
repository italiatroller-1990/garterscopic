# Getting Started with Garterscopic

**Garterscopic** (D.L.S.S.G. — Declarative Lightweight Static Site Generator) is a static site generator that lets you build websites using technologies you already know.

## Philosophy

- **HTML** = Components
- **YAML** = Declarations/configuration
- **Markdown** = Content
- **CSS** = Styling
- **JavaScript** = Optional enhancement
- **Go** = Build system

No Node.js required. No npm. No proprietary template language.

## Quick Start

```bash
# Install (when available)
go install github.com/italiatroller-1990/garterscopic/cmd/garterscopic@latest

# Create a new project
garterscopic init my-site
cd my-site

# Build the site
garterscopic build

# Start development server
garterscopic dev
```

## Project Structure

```
my-site/
├── site.yaml              # Site configuration
├── components/            # HTML components
│   ├── definitions.yaml   # Component definitions
│   ├── navbar.html
│   └── footer.html
├── layouts/               # Page layouts
│   └── default.yaml
├── page-types/            # Page type definitions
│   ├── page.yaml
│   └── post.yaml
├── pages/                 # Markdown content
│   ├── index.md
│   └── about.md
├── posts/                 # Blog posts (optional)
│   └── blog/
│       └── hello.md
├── styles/                # CSS stylesheets
│   └── global.css
└── assets/                # Static assets
    └── js/
        └── main.js
```

## Site Configuration

`site.yaml` defines your site's settings:

```yaml
name: My Site
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
```

Links defined in `site.yaml` are available as `site.links` bindings in layouts.

## Creating Pages

```markdown
---
type: page
title: About
description: About my website
---

# About Page

This is the about page content in **Markdown**.
```

## Building

```bash
garterscopic build
```

Output goes to the `dist/` directory.

## Development Server

```bash
garterscopic dev --port 3000
```

The server watches for changes and rebuilds automatically.
