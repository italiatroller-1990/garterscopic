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

### Frontmatter Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Page type (matches `page-types/*.yaml` name) |
| `layout` | string | Override the layout for this page |
| `title` | string | Page title (used in SEO meta tags) |
| `description` | string | Page description (used in SEO meta tags) |
| `draft` | boolean | Set `true` to exclude from output |

### Draft Pages

Set `draft: true` in frontmatter to exclude a page from the build:

```markdown
---
type: page
title: Coming Soon
draft: true
---

This page will not appear in the output.
```

## Building

```bash
garterscopic build
```

Output goes to the `dist/` directory.

## Creating Posts

```bash
garterscopic new post my-first-post
garterscopic new post blog/hello-world
```

Posts are organized into sections by subdirectory (`posts/blog/`, `posts/guides/`). See [Posts](posts.md) for full details on sections, tags, categories, and pagination.

## Development Server

```bash
garterscopic dev --port 3000
```

The server watches for changes and rebuilds automatically. It supports:

- File watching with 100ms debounce
- Automatic rebuild on content, component, or layout changes
- Server-Sent Events (SSE) for browser live reload

## Validation

Validate your project without building:

```bash
garterscopic check
garterscopic check --verbose
```

This checks component files, layout schemas, page type validation, required fields, and internal link targets.
