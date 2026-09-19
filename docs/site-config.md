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
  viewport: "device-width"

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

### Responsive

The `responsive` section is the single source of truth for responsive
behavior. `site.yaml` values flow into the generated stylesheet and the
viewport meta tag — do not duplicate them with hard-coded pixel values in
your own CSS (later rules override the generated ones).

Available as `site.mobile_width`, `site.tablet_width`, `site.desktop_width` bindings:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `responsive.mobile` | string | `768px` | Max viewport width for mobile layout (`<= mobile` is mobile) |
| `responsive.tablet` | string | `1024px` | Max viewport width for tablet layout (`<= tablet` is tablet, wider is desktop) |
| `responsive.desktop` | string | `1200px` | Max content/container width (`.container` cap). Not a breakpoint |
| `responsive.viewport` | string | `device-width` | Viewport width: `"device-width"`, a number (`"1200"`), or a complete value (used verbatim) |

With the defaults above, viewports behave as `> 1024px` desktop,
`769–1024px` tablet, `<= 768px` mobile, with content capped at `1200px`.
`desktop` never becomes an `@media` breakpoint: at a `768px` viewport the
container (`width: 100%; max-width: 1200px`) simply shrinks to fit.

`viewport` renders into every page as:

```html
<meta name="viewport" content="width=device-width, initial-scale=1">
```

A complete value such as `"width=device-width, initial-scale=1"` is used
verbatim and never duplicated into `width=width=...`.

Omitting `responsive` entirely keeps working: the defaults above apply.
Breakpoints must be valid CSS lengths in `mobile < tablet` order, otherwise
the build fails with a validation error.

Previewing: `garterscopic dev` serves the generated site as-is, so what you
see is real responsive behavior — resize the browser window to the
configured widths (the server logs them on startup). Editing `site.yaml`
while `dev` runs reloads the config and rebuilds automatically.

### Navigation Links

Available as `site.links` binding:

```yaml
# Simple list form
links:
  - name: Home
    url: /
  - name: About
    url: /about/

# Map form with broken link checking and active link highlighting
links:
  fail_on_broken: true
  highlight_active: true
  highlight_color: "#ffffff"
  entries:
    - name: Home
      url: /
    - name: About
      url: /about/
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `links.fail_on_broken` | bool | `false` | Fail the build on links to missing pages (otherwise a warning) |
| `links.highlight_active` | bool | `false` | Mark the current page's nav link with `class="active"` and `aria-current="page"` |
| `links.highlight_color` | string | — | Color of the active nav link: hex (`"#ffffff"`, `"#fff"`) or rgb triplet (`"12, 69, 11"`). Needs `highlight_active: true` |

When `highlight_active` is on, the link whose URL matches the current page's
route renders as `<a href="..." class="active" aria-current="page">`.
Without `highlight_color`, style it with one rule:

```css
.navbar-links a.active {
    text-decoration: underline;
}
```

With `highlight_color` set, the active link also gets an inline
`style="color: ...;"` (a bare `"12, 69, 11"` triplet becomes
`rgb(12, 69, 11)`). Anything else is a build error; setting the color
without `highlight_active: true` is a warning since it has no effect.

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
| `responsive.viewport` | `device-width` |
| `icon` | `icon.png` |
| `icon_dir` | `assets/icon` |
| `feed.path` | `/feed.xml` |
| `feed.title` | site name |
| `pagination.per_page` | `10` |
