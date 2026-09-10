# Garterscopic

**D.L.S.S.G. — Declarative Lightweight Static Site Generator**

> *Easier than Gatsby, harder than HTML, more flexible than all.*

Garterscopic is a static site generator that builds websites from HTML components, YAML declarations, Markdown content, CSS styling, and optional JavaScript. It compiles to a single native binary with zero runtime dependencies.

## Quick Start

```bash
garterscopic init my-site
cd my-site
garterscopic build
garterscopic dev
```

## Features

- **HTML Components** — Plain HTML with `{{ placeholder }}` syntax and typed options
- **YAML Layouts** — Declare component composition with position-based regions
- **Page Type Schemas** — Validate frontmatter with required fields, types, and defaults
- **Blog Posts** — Sections, tags, categories, and paginated listings
- **SEO** — Meta tags, Open Graph, Twitter Cards, sitemap, robots.txt, RSS feed
- **Slots** — Named and default slot content injection into components
- **Data Binding** — Declarative binding from frontmatter, page, site, posts, sections, and tags
- **Draft Support** — Exclude pages from output with `draft: true`
- **Live Reload** — Development server with automatic rebuild and browser refresh
- **Incremental Builds** — Dependency graph for targeted rebuilds
- **Validation** — Pre-build checks for types, required fields, broken links
- **Zero Dependencies** — Single binary, no Node.js, no npm

## Documentation

- [Getting Started](getting-started.md) — Installation, philosophy, quick start
- [Site Configuration](site-config.md) — Full `site.yaml` reference
- [Project Structure](project-structure.md) — Directory layout and routing
- [Components](components.md) — HTML components, options, slots, data binding
- [Layouts](layouts.md) — Component composition, positions, inheritance
- [Page Types](page-types.md) — Schemas, validation, required fields
- [Posts](posts.md) — Blog posts, sections, tags, categories, pagination
- [SEO and Standards](seo.md) — Meta tags, sitemap, robots.txt, RSS feed
- [CLI Commands](commands.md) — All commands and flags
- [Architecture](architecture/performance.md) — Performance strategy and benchmarks
- [Packaging](packaging.md) — Release workflow and distribution

## Commands

| Command | Description |
|---------|-------------|
| `garterscopic init [name]` | Create a new project |
| `garterscopic build` | Build the site |
| `garterscopic dev` | Start development server |
| `garterscopic check` | Validate the project |
| `garterscopic graph` | Show page/component structure |
| `garterscopic clean` | Remove generated files |
| `garterscopic new page <name>` | Create a new page |
| `garterscopic new post [section/]slug` | Create a new post |

## Supported Platforms

| Platform | Architectures |
|----------|---------------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |
| FreeBSD | amd64, arm64 |

## License

Apache-2.0
