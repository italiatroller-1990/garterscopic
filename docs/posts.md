# Blog Posts

Posts are Markdown files in the `posts/` directory with YAML frontmatter. They support sections, tags, categories, and automatic listing pages.

## Creating Posts

```bash
garterscopic new post my-first-post
garterscopic new post blog/hello-world
garterscopic new post guides/getting-started
```

This creates a file with proper frontmatter:

```markdown
---
type: post
title: My First Post
date: 2024-01-15
tags: []
---

# My First Post

Your content here.
```

## Post Frontmatter

```markdown
---
type: post
title: My Post
date: 2024-01-15
description: A short description
author: Alice
tags:
  - go
  - web
categories:
  - tutorials
slug: my-post
draft: false
---

# My Post

Content in **Markdown**.
```

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Post title |
| `date` | date | Publication date (YYYY-MM-DD) |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Short description (used in SEO and listings) |
| `author` | string | Author name |
| `tags` | list | Tags for taxonomy pages |
| `categories` | list | Categories for taxonomy pages |
| `slug` | string | Custom slug (defaults to filename) |
| `draft` | boolean | Exclude from output (default: `false`) |
| `updated` | date | Last updated date |
| `canonical` | string | Canonical URL for SEO |
| `robots` | string | Robots meta tag value |
| `og_title` | string | Open Graph title override |
| `og_description` | string | Open Graph description override |
| `og_image` | string | Open Graph image URL |

## Sections

Posts are organized into sections by subdirectory:

```
posts/
├── blog/
│   ├── hello-world.md
│   └── getting-started.md
└── guides/
    ├── linux-basics.md
    └── css-basics.md
```

The section is detected from the directory structure:
- `posts/blog/hello-world.md` → section: `blog`
- `posts/guides/linux-basics.md` → section: `guides`
- `posts/standalone.md` → no section

### Auto-Generated Section Pages

For each section, an index page is auto-generated at `/{section}/`:

- `posts/blog/hello-world.md` and `posts/blog/getting-started.md` → auto-generated `/blog/`
- `posts/guides/linux-basics.md` → auto-generated `/guides/`

These pages use the default layout with a `post-list` component showing all posts in that section.

### Section Bindings

In layouts, section posts are available via:

```yaml
# All sections
from: sections.all

# Posts in a specific section
from: sections.blog
```

## Tags

Tags enable taxonomy pages. Add tags to post frontmatter:

```markdown
---
title: My Post
tags:
  - go
  - web
  - tutorial
---
```

### Auto-Generated Tag Pages

When `taxonomy.tags: true` is set in `site.yaml`:

- `/tags/` — index of all tags
- `/tags/go/` — all posts tagged "go"
- `/tags/web/` — all posts tagged "web"

### Tag Bindings

```yaml
# All tags
from: tags.all

# Posts with a specific tag
from: tags.go
```

## Categories

Categories work like tags but use a separate taxonomy:

```markdown
---
title: My Post
categories:
  - tutorials
  - reference
---
```

### Auto-Generated Category Pages

When `taxonomy.categories: true` is set in `site.yaml`:

- `/categories/` — index of all categories
- `/categories/tutorials/` — all posts in "tutorials"

## Pagination

When `pagination.enabled: true` is set in `site.yaml`, post listings are paginated:

- Global posts listing at `/posts/` with page 1
- Additional pages at `/posts/page/2/`, `/posts/page/3/`, etc.
- Section listings are also paginated

### Pagination Metadata

Each paginated page receives:

```yaml
pagination:
  page: 1          # Current page number
  pages: 5         # Total pages
  total: 42        # Total posts
  per_page: 10     # Posts per page
  has_previous: false
  has_next: true
  next: /posts/page/2/
  previous: ""     # Only on pages 2+
```

### Configuration

```yaml
# site.yaml
pagination:
  enabled: true
  per_page: 10
```

## Draft Posts

Set `draft: true` in frontmatter to exclude a post from output:

```markdown
---
title: Work in Progress
date: 2024-01-15
draft: true
---

This post is not published.
```

Draft posts are:
- Excluded from the output directory
- Excluded from sitemap.xml
- Excluded from RSS feed
- Excluded from tag/category listings
- Excluded from section listings

## Post Routing

Posts follow this routing pattern:

| Source File | Generated URL |
|-------------|---------------|
| `posts/hello.md` | `/hello/` |
| `posts/blog/hello.md` | `/blog/hello/` |
| `posts/guides/getting-started.md` | `/guides/getting-started/` |

## Post Order

Posts are sorted by date (newest first) in all listings.
