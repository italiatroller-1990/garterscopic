# Project Structure

A Garterscopic project has a simple, predictable structure:

```
site/
├── site.yaml              # Required: Site configuration
├── components/            # Optional: HTML component files
│   ├── definitions.yaml   # Required if using components
│   ├── navbar.html
│   ├── hero.html
│   └── footer.html
├── layouts/               # Optional: Page layouts
│   └── default.yaml
├── page-types/            # Optional: Page type definitions
│   └── page.yaml
├── pages/                 # Required: Content pages
│   ├── index.md
│   └── about.md
├── posts/                 # Optional: Blog posts
│   ├── blog/
│   │   ├── first-post.md
│   │   └── second-post.md
│   └── guides/
│       └── getting-started.md
├── styles/                # Optional: CSS stylesheets
│   ├── global.css
│   ├── navbar.css
│   └── hero.css
└── assets/                # Optional: Static assets
    ├── images/
    ├── fonts/
    └── js/
        └── main.js
```

## Directory Purposes

### `site.yaml`
The root configuration file. Defines site name, URLs, build settings, and global options.

### `components/`
HTML component files. Components are ordinary HTML with `{{ placeholder }}` syntax for dynamic values.

### `layouts/`
YAML files that define how components compose into pages.

### `page-types/`
YAML files that define validation rules and defaults for different page types.

### `pages/`
Markdown files with YAML frontmatter. The filesystem structure maps to URL routes.

### `posts/`
Blog post content. Supports sections (e.g., `posts/blog/`, `posts/guides/`). Created with `garterscopic new post`.

### `styles/`
CSS stylesheets. Can be global or component-specific.

### `assets/`
Static files (images, fonts, JavaScript). Copied verbatim to the output.

## Routing

| Source File | Generated URL |
|-------------|---------------|
| `pages/index.md` | `/` |
| `pages/about.md` | `/about/` |
| `posts/blog/hello.md` | `/posts/blog/hello/` |
| `posts/guides/getting-started.md` | `/posts/guides/getting-started/` |

## No Magic

Everything is a plain file. There are no generated entrypoints, no hidden configuration, no proprietary conventions beyond the structure above.
