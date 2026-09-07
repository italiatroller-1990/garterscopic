# Components

Components are ordinary HTML files with simple `{{ placeholder }}` syntax.

**Core philosophy:** Components are HTML, not templates. YAML declares data and configuration. HTML defines presentation.

## Placeholder Syntax

Use simple placeholders in your HTML components:

```html
<h1>{{ title }}</h1>
<p>{{ description }}</p>
```

### Rules

- Placeholders use the syntax: `{{ name }}`
- Whitespace around identifiers is accepted: `{{ title }}`, `{{title}}`, `{{ title   }}`
- Placeholder names must be simple identifiers (alphanumeric and underscore)
- Do NOT use expressions, functions, filters, or transformations

### Unsupported Constructs

The component system deliberately does NOT support:

```
{{ if something }}     ✗ No conditionals
{{ for item in ... }}  ✗ No loops
{{ foo.bar() }}        ✗ No functions
{{ value | filter }}   ✗ No filters
{{ a + b }}            ✗ No expressions
{{ eval("code") }}     ✗ No evaluation
```

This constraint is intentional and preserves the principle that **HTML remains HTML.**

---

## Creating Components

Create a simple HTML file like `components/navbar.html`:

```html
<nav class="navbar">
    <a class="navbar-logo" href="/">{{ logo }}</a>
    <ul class="navbar-links">
        {{ links }}
    </ul>
</nav>
```

---

## Defining Components

Create `components/definitions.yaml` to declare your components:

```yaml
components:
  navbar:
    file: navbar.html
    style: navbar.css
    position: top
    
    options:
      logo:
        type: string
        required: true
      
      links:
        type: list
        required: true
```

### Component Definition Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | string | Yes | Path to component HTML file (relative to `components/`) |
| `style` | string | No | Path to component stylesheet (relative to `styles/`) |
| `position` | string | No | Default position: `top`, `bottom`, `left`, `right`, `center` |
| `options` | object | No | Map of option definitions |

---

## Supported Option Types

### string

Plain text value (HTML-escaped by default).

```yaml
options:
  title:
    type: string
    required: true
```

Usage:

```yaml
components:
  - name: hero
    options:
      title: "Welcome to my site"
```

### number

Numeric value (integer or floating-point).

```yaml
options:
  count:
    type: number
    default: 0
```

Usage:

```yaml
components:
  - name: pagination
    options:
      count: 10
      page_size: 5
```

### boolean

True/false value.

```yaml
options:
  enabled:
    type: boolean
    default: false
```

Usage:

```yaml
components:
  - name: feature
    options:
      enabled: true
```

### list

A list of values (YAML array).

```yaml
options:
  links:
    type: list
    required: true
```

Usage in layout:

```yaml
components:
  - name: navbar
    options:
      links:
        - Home
        - About
        - Contact
```

HTML component receives the list as-is. By default, list items are joined with commas and HTML-escaped.

### object

Key-value pairs (YAML mapping).

```yaml
options:
  metadata:
    type: object
```

Usage:

```yaml
components:
  - name: card
    options:
      metadata:
        author: Alice
        date: "2024-01-15"
        tags: important
```

### date

Date value (passed as a string).

```yaml
options:
  published:
    type: date
    required: true
```

Usage:

```yaml
components:
  - name: post-header
    options:
      published: "2024-01-15"
```

### html

Raw HTML (NOT escaped).

```yaml
options:
  body:
    type: html
```

Usage:

```yaml
components:
  - name: hero
    options:
      body: "<strong>Bold</strong> and <em>italic</em>"
```

**Important:** Use the `html` type only for trusted content. User input should use `string` type instead.

---

## Required Options

Mark options that must be provided:

```yaml
components:
  post-header:
    file: post-header.html
    options:
      title:
        type: string
        required: true
      
      date:
        type: date
        required: true
      
      excerpt:
        type: string
```

Missing required options will produce a clear validation error:

```
Error: invalid component option

Component: post-header
Option: title

Expected:
  string (required)

Received:
  missing

Fix:
  Add "title" to the component options.
```

---

## Default Values

Provide defaults for optional options:

```yaml
components:
  footer:
    file: footer.html
    options:
      copyright:
        type: string
        default: "© 2024"
      
      color:
        type: string
        default: "#333333"
```

If an option is not provided and has a default, the default is used.

---

## Using Components in Layouts

In `layouts/home.yaml`:

```yaml
name: home

components:
  - name: navbar
    position: top
    options:
      logo: My Site
      links:
        - Home
        - About
        - Blog

  - name: hero
    position: center
    options:
      title: Welcome
      subtitle: Build fast, static sites

  - name: footer
    position: bottom
    options:
      copyright: "© 2024 My Site"
```

### Instance Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Component name (from definitions) |
| `position` | string | Optional position override (see Component Position section) |
| `options` | object | Option values for this instance |
| `bindings` | object | Optional data bindings (see Data Binding section) |

---

## Component Position

Components can declare a logical position in the layout.

### Supported Positions

- `top`: Header/navigation region
- `bottom`: Footer region
- `left`: Sidebar/aside region
- `right`: Secondary sidebar/aside region
- `center`: Main content region (default)

### Position Resolution

Position is resolved in this order:

1. Instance-level position (in layout YAML)
2. Component definition default position
3. Default: `center`

Example:

```yaml
components:
  # Sidebar definition with default position
  sidebar:
    file: sidebar.html
    position: left

  # In layout, use default position (left)
  - name: sidebar
    options:
      title: Navigation

  # Or override with a different position
  - name: sidebar
    position: right
    options:
      title: Related Links
```

### Important

Position is a **logical layout region**, not CSS styling. The `position: left` value does NOT automatically apply `position: absolute; left: 0;` or similar CSS.

Your stylesheets must define how each position renders visually (via CSS Grid, Flexbox, CSS Positions, or other techniques).

---

## Data Binding

Extend component options with declarative data binding.

Supported data sources:

- `frontmatter.*` – Page frontmatter fields
- `page.*` – Page metadata (route, url, type, etc.)
- `site.*` – Site configuration (name, base_url, etc.)

### Example

In `layouts/post.yaml`:

```yaml
name: post

components:
  - name: post-header
    bindings:
      title:
        from: frontmatter.title
      
      date:
        from: frontmatter.published_date
      
      author:
        from: frontmatter.author
      
      url:
        from: page.url
      
      site_name:
        from: site.name
```

The component system will:

1. Read the declared data source
2. Look up the value in the appropriate context
3. Pass the resolved value to the component option
4. Validate the type matches the component definition

### Important

Data binding is declarative data plumbing, NOT a template language.

These are **supported:**

```yaml
bindings:
  title:
    from: frontmatter.title

  date:
    from: page.created_at

  site_name:
    from: site.name
```

These are **NOT supported:**

```yaml
bindings:
  title:
    from: frontmatter.title.toUpperCase()  ✗ No method calls

  display:
    from: site.name + "!"                  ✗ No expressions

  content:
    from: if(condition) { value1 } else { value2 }  ✗ No conditionals
```

---

## Component Styles

A component may reference a stylesheet:

```yaml
components:
  navbar:
    file: navbar.html
    style: navbar.css
```

### How Styles Are Handled

1. The builder detects which components are used on a page
2. It resolves their stylesheets
3. It includes them exactly once per build
4. It preserves deterministic ordering (alphabetical by path)

### Global Styles

For site-wide styling, use `styles/global.css`. It's included on all pages.

### Important

- Do NOT use CSS-in-JS, Tailwind, Sass, CSS Modules, or other build-time CSS transformations (unless you add them separately)
- Stylesheets are copied as-is to the output
- The builder does NOT minify or transform CSS

---

## Component Rendering Model

The component system follows this conceptual pipeline:

```
YAML layout declaration
        ↓
Component definition lookup
        ↓
Component instance creation
        ↓
Data binding resolution
        ↓
Type validation
        ↓
HTML placeholder substitution
        ↓
HTML output
```

Each step is independent:

- **YAML** declares which components to use and how to configure them
- **Definitions** describe the component's structure and expected options
- **Instance** is a specific use of the component in a layout
- **Data binding** resolves external data sources to component options
- **Validation** ensures types and required fields are correct
- **Substitution** replaces `{{ placeholder }}` with values
- **Output** is the final HTML

The HTML component itself knows nothing about data sources or the binding system. It only contains static HTML with placeholders.

---

## Security

Component rendering is safe by default.

### HTML Escaping

All values are HTML-escaped **by default**:

```yaml
options:
  description: "<script>alert('xss')</script>"
```

Renders as:

```html
&lt;script&gt;alert('xss')&lt;/script&gt;
```

NOT:

```html
<script>alert('xss')</script>
```

### Raw HTML

Only the `html` type permits raw HTML:

```yaml
components:
  hero:
    file: hero.html
    options:
      body:
        type: html

# In layout:
- name: hero
  options:
    body: "<strong>Bold</strong>"
```

### Important Security Rules

- NEVER execute component contents
- NEVER execute YAML
- NEVER execute frontmatter
- NEVER evaluate placeholders as code
- NEVER use `eval`, `exec`, shell execution, or dynamic plugins for component rendering

User input should use `string` type (which escapes HTML). Use `html` type only for content you control.

---

## No Programming Logic

Components should only contain:

- Static HTML
- `{{ placeholder }}` tags for dynamic values
- No `{{ if }}`, `{{ for }}`, `{{ range }}`
- No filters or transforms
- No custom expressions

The goal is **HTML remains HTML.**

If you need conditional rendering or loops, use:

1. Multiple component instances with different options
2. Pre-computed values in frontmatter
3. Separate page types for different layouts
4. CSS for styling variations

---

## Validation

The builder validates components at build time.

### Validation Checks

- ✓ Component definition exists
- ✓ Component HTML file exists
- ✓ Component stylesheet (if referenced) exists
- ✓ Required options are provided
- ✓ Option values match declared types
- ✓ Unknown options are rejected
- ✓ Placeholder names are valid identifiers
- ✓ All placeholders are resolved to options
- ✓ Data bindings reference valid data sources
- ✓ Position values are valid

---

## Backwards Compatibility

Existing component projects remain valid.

The `{{ placeholder }}` syntax is unchanged. Existing component definitions continue to work.

New features like `position`, `bindings`, and `data binding` are optional and do not affect existing components.

If you have existing components without these features, they work exactly as before.
