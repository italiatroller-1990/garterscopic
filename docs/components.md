# Components

Components are ordinary HTML files with simple `{{ placeholder }}` syntax.

## Creating Components

Create `components/navbar.html`:

```html
<nav class="navbar">
    <a class="navbar-logo" href="/">{{ logo }}</a>
    <ul class="navbar-links">
        {{ links }}
    </ul>
</nav>
```

## Defining Components

Create `components/definitions.yaml`:

```yaml
components:
  navbar:
    file: navbar.html
    style: navbar.css
    options:
      logo:
        type: string
      links:
        type: list
```

## Supported Option Types

| Type | Description | Example |
|------|-------------|---------|
| `string` | Plain text (HTML-escaped) | `"Hello"` |
| `number` | Numeric values | `42`, `3.14` |
| `boolean` | true/false | `true` |
| `list` | List of values | `["a", "b"]` |
| `object` | Key-value pairs | `{text: "Click"}` |
| `date` | Date string | `"2024-01-01"` |
| `html` | Raw HTML (not escaped) | `"<b>Bold</b>"` |

## Using Components in Layouts

In `layouts/default.yaml`:

```yaml
name: default

components:
  - name: navbar
    options:
      logo: My Site
      links: "Home,About,Contact"

  - name: content

  - name: footer
    options:
      copyright: "2024"
```

## No Programming Logic

Components should only contain:
- Static HTML
- `{{ placeholder }}` tags for dynamic values
- No `{{ if }}`, `{{ for }}`, `{{ range }}`
- No filters or transforms
- No custom expressions

The goal is **HTML remains HTML**.
