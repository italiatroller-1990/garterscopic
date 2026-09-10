# Page Types

Page types define validation schemas for different kinds of pages: required fields, field types, and default values.

## Page Type Files

Page types are YAML files in the `page-types/` directory:

```yaml
# page-types/page.yaml
name: page

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
```

## Page Type Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Page type identifier (matches `type:` in frontmatter) |
| `layout` | string | No | Default layout for pages of this type |
| `fields` | map | No | Field definitions with types, requirements, and defaults |

## Field Definitions

Each field in `fields` has:

```yaml
fields:
  field_name:
    type: string       # Required: field type
    required: false    # Optional: whether the field must be present
    default: value     # Optional: default value when field is missing
```

### Supported Field Types

| Type | YAML Value | Description |
|------|------------|-------------|
| `string` | `"text"` | Plain text (also accepts integers coerced to strings) |
| `number` | `42` or `3.14` | Integer or floating-point |
| `boolean` | `true` / `false` | True/false |
| `list` | `- item1` / `- item2` | YAML array |
| `object` | `key: value` | YAML mapping |
| `date` | `"2024-01-15"` | Date string or time.Time |
| `html` | `"<p>Raw HTML</p>"` | String (rendered without escaping) |

## How Page Types Work

### 1. Type Assignment

Pages declare their type in frontmatter:

```markdown
---
type: post
title: My Post
date: 2024-01-15
---
```

If no `type` is specified, the `default_page_type` from `site.yaml` is used (default: `page`).

### 2. Layout Resolution

The page type's `layout` field determines which layout is used. A page can override this in frontmatter:

```markdown
---
type: page
layout: wide
---
```

The override order is:

1. Frontmatter `layout:` field (highest priority)
2. Page type `layout:` field
3. `default_layout` from `site.yaml` (lowest priority)

### 3. Default Values

Missing fields are filled from the page type's defaults:

```yaml
# page-types/post.yaml
fields:
  draft:
    type: boolean
    default: false
  
  author:
    type: string
    default: "Anonymous"
```

A post without `draft` or `author` in frontmatter gets `false` and `"Anonymous"` respectively.

### 4. Validation

At build time, the builder validates:

- Required fields are present
- Field values match declared types
- Unknown fields produce warnings

Validation errors are reported with the field name and expected type.

## Example: Blog Post Type

```yaml
# page-types/post.yaml
name: post

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

  categories:
    type: list

  slug:
    type: string

  draft:
    type: boolean
    default: false
```

## Example: Documentation Page Type

```yaml
# page-types/doc.yaml
name: doc

layout: docs

fields:
  title:
    type: string
    required: true

  description:
    type: string

  sidebar:
    type: boolean
    default: true

  toc:
    type: boolean
    default: true

  weight:
    type: number
    default: 0
```

## Validation Without Building

Use `garterscopic check` to validate page types without producing output:

```bash
garterscopic check
```

This validates that all pages conform to their declared page type schemas.
