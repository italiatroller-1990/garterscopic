# SEO and Standards

Garterscopic generates SEO-friendly meta tags, sitemaps, robots.txt, and RSS feeds automatically.

## SEO Meta Tags

Every page receives meta tags in `<head>`:

```html
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="Content-Language" content="en">
<title>Page Title</title>
<meta name="description" content="Page description">
<meta name="keywords" content="keyword1, keyword2">
<meta name="robots" content="index, follow">
<link rel="canonical" href="https://example.com/page/">

<!-- Open Graph -->
<meta property="og:title" content="Page Title">
<meta property="og:description" content="Page description">
<meta property="og:image" content="https://example.com/image.jpg">
<meta property="og:url" content="https://example.com/page/">
<meta property="og:type" content="website">

<!-- Twitter Card -->
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="Page Title">
<meta name="twitter:description" content="Page description">
<meta name="twitter:image" content="https://example.com/image.jpg">
```

### SEO Frontmatter Fields

Set these in page/post frontmatter:

```markdown
---
title: My Page
description: A description for search engines
keywords: go, static site, generator
canonical: https://example.com/my-page/
robots: index, follow
og_title: Custom OG Title
og_description: Custom OG Description
og_image: https://example.com/og-image.jpg
---
```

| Field | HTML Output | Fallback |
|-------|-------------|----------|
| `title` | `<title>` and `og:title` | — |
| `description` | `meta description` and `og:description` | — |
| `keywords` | `meta keywords` | — |
| `robots` | `meta robots` | — |
| `canonical` | `link rel="canonical"` | — |
| `og_title` | `og:title` | Falls back to `title` |
| `og_description` | `og:description` | Falls back to `description` |
| `og_image` | `og:image` and `twitter:image` | — |

### Open Graph Type

- Posts in the `blog` section or with `type: post` → `og:type: article`
- All other pages → `og:type: website`

## Sitemap

Generate `sitemap.xml` automatically:

```yaml
# site.yaml
sitemap:
  enabled: true
```

### Sitemap Priorities

| Page Type | Priority | Change Frequency |
|-----------|----------|------------------|
| Homepage (`/`) | 1.0 | daily |
| Blog posts | 0.9 | weekly |
| Other pages | 0.8 | weekly |

### Sitemap Behavior

- Draft pages are excluded
- Pages with `canonical` use that URL
- Last modified uses page date or `updated` frontmatter field

### Custom Sitemap Filename

```yaml
seo:
  sitemap_file: sitemap.xml
```

## robots.txt

Generate `robots.txt` automatically with every build:

```
# Garterscopic robots.txt
# Generated automatically

User-agent: *
Allow: /
Disallow: /dist/
Disallow: /*.tmp

Sitemap: https://example.com/sitemap.xml
```

The sitemap reference is included when `sitemap.enabled: true`.

### Custom robots.txt Filename

```yaml
seo:
  robots_file: robots.txt
```

## RSS Feed

Generate an RSS 2.0 feed from posts:

```yaml
# site.yaml
feed:
  enabled: true
  path: /feed.xml
  title: My Blog
```

### Feed Configuration

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `false` | Enable feed generation |
| `path` | `/feed.xml` | Output path for the feed |
| `title` | site name | Feed title |

### Feed Content

- Only published posts (drafts excluded)
- Sorted by date (newest first)
- Includes title, link, publication date, description, and author
- Dates are formatted as RFC 1123Z (RSS 2.0 requirement)

### Feed Output

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
  <title>My Blog</title>
  <link>https://example.com/</link>
  <description>My Blog</description>
  <lastBuildDate>Mon, 15 Jan 2024 00:00:00 +0000</lastBuildDate>
  <item>
    <title>My Post</title>
    <link>https://example.com/blog/my-post/</link>
    <pubDate>Mon, 15 Jan 2024 00:00:00 +0000</pubDate>
    <description>A short description</description>
    <author>Alice</author>
  </item>
</channel>
</rss>
```

## Global SEO Settings

Set site-wide SEO defaults in `site.yaml`:

```yaml
seo:
  author: Alice
  keywords: go, static site, generator
  theme_color: "#333333"
  sitemap_file: sitemap.xml
  robots_file: robots.txt
```

## Internal Link Checking

Validate that all internal links point to existing pages:

```yaml
# site.yaml
links:
  fail_on_broken: true
  entries:
    - name: Home
      url: /
    - name: About
      url: /about/
```

When `fail_on_broken: true`, the build fails if any link target does not correspond to an existing page route. Use `garterscopic check --verbose` to see detailed reports of broken links.

## Favicon and Icon

Configure favicon and icon for browser tabs and mobile bookmarks:

```yaml
favicon: favicon.png
icon: icon.png
icon_dir: assets/icon
```

The favicon is rendered as:

```html
<link rel="icon" href="/assets/icon/icon.png">
```

Place your icon files in the `assets/icon/` directory.
