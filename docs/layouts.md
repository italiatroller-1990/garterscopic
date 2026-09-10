# Layouts

Layouts define which components compose a page and in what order.

## Layout Files

Layouts are YAML files in the `layouts/` directory:

```yaml
name: default

components:
  - name: navbar
    position: top
    options:
      logo: My Site
      from:
        links: site.links

  - name: content

  - name: footer
    position: bottom
    options:
      copyright: "© 2024"
```

## Layout Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Layout identifier |
| `components` | list | No | Component instances |
| `inherit` | string | No | Parent layout to inherit from |

## Component Instances

Each component in the list is an instance with these fields:

```yaml
components:
  - name: navbar           # Required: component name from definitions.yaml
    position: top          # Optional: override position (top, bottom, left, right, center)
    options:               # Optional: component option values
      logo: My Site
      from:
        links: site.links
    slots:                 # Optional: slot content injection
      default: <p>Fallback content</p>
```

### Shorthand Form

Components can be specified as a plain string when no options or position override are needed:

```yaml
components:
  - name: content
```

## Positions

Components are rendered in a fixed position order:

```
top → left → center → right → bottom
```

### Supported Positions

| Position | Description |
|----------|-------------|
| `top` | Header/navigation region |
| `left` | Sidebar/aside region |
| `center` | Main content region (default) |
| `right` | Secondary sidebar region |
| `bottom` | Footer region |

Position is a **logical layout region**, not CSS styling. Your stylesheets must define how each position renders visually (CSS Grid, Flexbox, etc.).

Non-center positions are wrapped in `<div data-garterscopic-region="position">` elements for CSS targeting.

### Position Resolution

Position is resolved in this order:

1. Instance-level position (in layout YAML)
2. Component definition default position
3. Default: `center`

## Layout Inheritance

Layouts can inherit from a parent layout using the `inherit` field:

```yaml
# layouts/base.yaml
name: base

components:
  - name: navbar
    position: top
  - name: footer
    position: bottom
```

```yaml
# layouts/post.yaml
name: post
inherit: base.yaml

components:
  - name: navbar
    position: top
  - name: post-header
    position: center
  - name: content
    position: center
  - name: footer
    position: bottom
```

The child layout's `components` replace the parent's. The `inherit` field triggers resolution of the parent chain at load time, with circular dependency detection.

## Built-in Components

Two special component names have built-in behavior:

### `content`

Renders the page's Markdown body as HTML:

```yaml
components:
  - name: content
```

### `post-list`

Renders a list of posts from the page's `posts` metadata. Used automatically on section, tag, category, and paginated listing pages:

```yaml
components:
  - name: post-list
```

This generates a `<ul class="post-list">` with linked titles and dates. When pagination is enabled, it includes a `<nav class="pagination">` with Previous/Next links.

### `frontmatter-values`

Renders a definition list of resolved binding values:

```yaml
components:
  - name: frontmatter-values
    options:
      values:
        - frontmatter.title
        - frontmatter.date
        - site.name
```

Produces:

```html
<dl class="frontmatter-values">
  <dt>title</dt><dd>My Page</dd>
  <dt>date</dt><dd>2024-01-15</dd>
  <dt>name</dt><dd>My Site</dd>
</dl>
```

## Slots

Components can contain `<slot />` tags for content injection. Slots are defined in the component HTML and filled via the `slots` option in layouts.

### Default Slot

```html
<!-- components/card.html -->
<article class="card">
  <h2>{{ title }}</h2>
  <slot />
</article>
```

```yaml
# In layout
- name: card
  options:
    title: Featured
    slots:
      default: <p>This is the card body content.</p>
```

### Named Slots

```html
<!-- components/page.html -->
<header><slot name="header" /></header>
<main><slot /></main>
<footer><slot name="footer" /></footer>
```

```yaml
- name: page
  options:
    slots:
      header: <h1>Page Title</h1>
      default: <p>Main content here</p>
      footer: <p>Page footer</p>
```

Slots without provided content are stripped from the output.
