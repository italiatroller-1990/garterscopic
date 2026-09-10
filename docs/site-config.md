# Site Configuration

`site.yaml` defines your site's settings. All fields are optional with sensible defaults.

## Full Reference

```yaml
name: My Site
base_url: https://example.com

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

scripts:
  - src: /assets/js/main.js
  - src: /assets/js/app.js
    type: module

responsive:
  mobile: "768px"
  tablet: "1024px"
  desktop: "1200px"

links:
  - name: Home
    url: /
  - name: About
    url: /about/
  - name: Blog
    url: /blog/

favicon: favicon.png
icon: icon.png
icon_dir: assets/icon

language: en

seo:
  author: Alice
  keywords: go, static site, generator
  theme_color: "#333333"
  sitemap_file: sitemap.xml
  robots_file: robots.txt

sitemap:
  enabled: true

feed:
  enabled: true
  path: /feed.xml
  title: My Blog

taxonomy:
  tags: true
  categories: true

pagination:
  enabled: true
  per_page: 10

gallery:
  enabled: false
```

## Field Reference

### Basic Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | — | Site name (used in feeds and defaults) |
| `base_url` | string | — | Base URL for sitemap, feeds, and canonical URLs |
| `language` | string | `en` | HTML `lang` attribute |
| `default_layout` | string | `default` | Fallback layout when page type has none |
| `default_page_type` | string | `page` | Fallback page type when frontmatter has none |

### Build Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `build.source` | string | `.` | Source directory root |
| `build.output` | string | `dist` | Output directory |

### Assets

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `assets.directory` | string | `assets` | Static assets directory (copied verbatim to output) |

### Styles

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `styles.global` | list | — | Global CSS files included on every page |

Example:

```yaml
styles:
  global:
    - styles/global.css
    - styles/variables.css
```

### Scripts

Include JavaScript files in the HTML `<head>`:

```yaml
scripts:
  - src: /assets/js/main.js
  - src: /assets/js/app.js
    type: module
```

| Field | Type | Description |
|-------|------|-------------|
| `src` | string | Script path (relative to output root) |
| `type` | string | Script type (omit for classic, `module` for ES modules) |

### Responsive Breakpoints

Available as `site.mobile_width`, `site.tablet_width`, `site.desktop_width` bindings:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `responsive.mobile` | string | `768px` | Mobile breakpoint |
| `responsive.tablet` | string | `1024px` | Tablet breakpoint |
| `responsive.desktop` | string | `1200px` | Desktop breakpoint |

### Navigation Links

Available as `site.links` binding:

```yaml
# Simple list form
links:
  - name: Home
    url: /
  - name: About
    url: /about/

# Map form with broken link checking
links:
  fail_on_broken: true
  entries:
    - name: Home
      url: /
    - name: About
      url: /about/
```

### Favicon and Icon

```yaml
favicon: favicon.png       # Browser tab icon
icon: icon.png             # Mobile bookmark icon
icon_dir: assets/icon      # Directory containing icon files
```

### SEO

```yaml
seo:
  author: Alice                      # Default author for feeds
  keywords: go, static site          # Default keywords
  theme_color: "#333333"             # Theme color for mobile browsers
  sitemap_file: sitemap.xml          # Sitemap filename
  robots_file: robots.txt            # robots.txt filename
```

### Sitemap

```yaml
sitemap:
  enabled: true    # Generate sitemap.xml (default: false)
```

### RSS Feed

```yaml
feed:
  enabled: true                    # Generate feed.xml (default: false)
  path: /feed.xml                  # Feed output path
  title: My Blog                   # Feed title (defaults to site name)
```

### Taxonomy

```yaml
taxonomy:
  tags: true         # Generate /tags/ and /tags/{tag}/ pages (default: false)
  categories: true   # Generate /categories/ and /categories/{cat}/ pages (default: false)
```

### Pagination

```yaml
pagination:
  enabled: true      # Enable pagination for post listings (default: false)
  per_page: 10       # Posts per page (default: 10)
```

### Gallery

```yaml
gallery:
  enabled: true      # Generate /components/ gallery page (default: false)
```

The gallery page shows all component definitions with their options, types, and positions.

## Defaults Summary

| Setting | Default Value |
|---------|---------------|
| `build.source` | `.` |
| `build.output` | `dist` |
| `default_layout` | `default` |
| `default_page_type` | `page` |
| `assets.directory` | `assets` |
| `language` | `en` |
| `responsive.mobile` | `768px` |
| `responsive.tablet` | `1024px` |
| `responsive.desktop` | `1200px` |
| `icon` | `icon.png` |
| `icon_dir` | `assets/icon` |
| `feed.path` | `/feed.xml` |
| `feed.title` | site name |
| `pagination.per_page` | `10` |
