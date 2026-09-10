# CLI Commands

Garterscopic provides the following commands:

## `garterscopic init [name]`

Create a new project with all boilerplate files.

```bash
garterscopic init my-site
```

Creates a complete project structure:

```
my-site/
├── site.yaml
├── components/
│   ├── definitions.yaml
│   ├── navbar.html
│   ├── hero.html
│   └── footer.html
├── layouts/
│   └── default.yaml
├── page-types/
│   ├── page.yaml
│   └── post.yaml
├── pages/
│   ├── index.md
│   └── about.md
├── styles/
│   ├── global.css
│   ├── navbar.css
│   ├── hero.css
│   └── footer.css
└── assets/
    └── js/
        └── main.js
```

Project names must not contain path separators (`/`, `\`) or `..`.

## `garterscopic build`

Build the site from source.

```bash
garterscopic build
```

The build process:

1. Loads configuration from `site.yaml`
2. Loads components, layouts, and page types
3. Parses pages and posts from Markdown
4. Validates the project
5. Renders pages to HTML
6. Generates auto-pages (section listings, tag pages, pagination)
7. Generates sitemap.xml, robots.txt, and feed.xml (if enabled)
8. Combines stylesheets into `styles.css`
9. Copies assets to output
10. Atomically writes to the output directory

Output goes to the `dist/` directory (configurable in `site.yaml`).

## `garterscopic dev`

Start a development server with live reload.

```bash
garterscopic dev --port 3000
```

The dev server:

- Builds the site initially
- Serves files from the output directory
- Watches for file changes (`.yaml`, `.md`, `.html`, `.css`, `.js`, images, fonts)
- Rebuilds automatically on changes (100ms debounce)
- Sends browser refresh via Server-Sent Events (SSE)
- Skips `.git/`, `.tmp/`, and `dist/` directories

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | HTTP server port |

## `garterscopic check`

Validate the project without building.

```bash
garterscopic check
garterscopic check --verbose
```

Performs pre-build validation:

- Configuration loads correctly
- Component HTML files exist
- Component stylesheets exist
- Layout files parse correctly
- Page type schemas are valid
- Page metadata conforms to page types
- Required component options are provided
- Option types match definitions
- Internal links resolve to existing pages (when `fail_on_broken: true`)
- Global stylesheets exist
- Favicon/icon files exist

### Flags

| Flag | Description |
|------|-------------|
| `--verbose` | Show file path and fix hint for every finding |

### Output

```
Garterscopic check

✓ Configuration
✓ Pages
✓ Components
✓ Styles
✓ Assets

No problems found.
```

Or with errors:

```
error: required field 'title' is missing
  File: pages/about.md
  Fix:  Add "title" to the frontmatter.
```

Exit status is non-zero when errors are found. Warnings alone do not cause failure (unless `links.fail_on_broken` promotes them).

## `garterscopic graph`

Show the resolved page/component structure.

```bash
garterscopic graph
garterscopic graph --json
```

Outputs a tree showing each page's route, source, layout, components, and styles.

### Text Output

```
/pages/index.md → /
  layout: default
  components: navbar, content, footer
  styles: global.css, navbar.css, footer.css

/posts/blog/hello-world.md → /blog/hello-world/
  layout: default
  components: navbar, content, footer
  styles: global.css, navbar.css, footer.css
```

### JSON Output

```json
[
  {
    "route": "/",
    "source": "pages/index.md",
    "layout": "default",
    "components": ["navbar", "content", "footer"],
    "styles": ["global.css", "navbar.css", "footer.css"]
  }
]
```

Draft pages are marked in the source path.

### Flags

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON instead of text tree |

## `garterscopic new page <name>`

Create a new page.

```bash
garterscopic new page about
garterscopic new page contact
```

Creates `pages/about.md` with frontmatter:

```markdown
---
type: page
title: About
---

# About

Your content here.
```

Names must not contain path separators or special characters (`/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`).

## `garterscopic new post [section/]slug`

Create a new post, optionally in a section.

```bash
garterscopic new post hello-world
garterscopic new post blog/getting-started
garterscopic new post guides/linux-basics
```

Creates `posts/blog/getting-started.md` with frontmatter:

```markdown
---
type: post
title: Getting Started
date: 2024-01-15
tags: []
---

# Getting Started

Your content here.
```

## `garterscopic clean`

Remove the output directory.

```bash
garterscopic clean
```

Removes the `dist/` directory (or whatever `build.output` is set to in `site.yaml`).

## `garterscopic --version`

Print the version number.

```bash
garterscopic --version
# garterscopic 0.1.0
```

## `garterscopic --help`

Print help text with all commands and flags.

```bash
garterscopic --help
```
