# Garterscopic

**D.L.S.S.G. — Declarative Lightweight Static Site Generator**

> *Easier than Gatsby, harder than HTML, more flexible than all.*

## Philosophy

Garterscopic is a static site generator that lets you build websites using technologies you already know:

- **HTML** = Components
- **YAML** = Declarations/configuration
- **Markdown** = Content
- **CSS** = Styling
- **JavaScript** = Optional enhancement
- **Go** = Build system

## Zero-Dependency Distribution

Garterscopic is distributed as a **single native binary with no runtime dependencies**.

A user should be able to:

1. Download the binary
2. Make it executable
3. Run `garterscopic --help`

No Node.js, no npm, no Python, no Ruby, no Java, no Docker, no package manager at runtime.

### Supported Platforms

| Platform | Architectures | Status |
|----------|---------------|--------|
| Linux | amd64, arm64 | Officially distributed |
| macOS | amd64, arm64 | Officially distributed |
| FreeBSD | amd64, arm64 | Officially distributed |
| Windows | any | **Not officially distributed** |

Windows users may compile Garterscopic from source themselves.

## Installation

### Official Binaries

Download from the [latest release](https://github.com/italiatroller-1990/garterscopic/releases/latest):

```bash
# Linux amd64
curl -sL https://github.com/italiatroller-1990/garterscopic/releases/latest/download/garterscopic-linux-amd64 -o garterscopic
chmod +x garterscopic
sudo mv garterscopic /usr/local/bin/

# macOS arm64 (Apple Silicon)
curl -sL https://github.com/italiatroller-1990/garterscopic/releases/latest/download/garterscopic-darwin-arm64 -o garterscopic
chmod +x garterscopic
sudo mv garterscopic /usr/local/bin/
```

### Verify Installation

```bash
garterscopic --version
gartersopcic --help
```

### Package Managers

**Homebrew (macOS/Linux):**

```bash
brew install garterscopic/tap/garterscopic
```

**Linux (various):**

Packages for Debian, Fedora, and Arch Linux are built automatically via GitHub Actions.

### Source Build

```bash
git clone https://github.com/italiatroller-1990/garterscopic
cd garterscopic
CGO_ENABLED=0 go build -o garterscopic ./cmd/garterscopic
```

Requirements:
- Go 1.21+
- `CGO_ENABLED=0` (no cgo required)

## Quick Start

```bash
# Create a new project
garterscopic init my-site
cd my-site

# Build
garterscopic build

# Development server
garterscopic dev
```

## Features

- **Simple**: Plain HTML, YAML, Markdown, CSS
- **Fast**: Built in Go
- **Self-contained**: Single binary, no runtime dependencies
- **Flexible**: Components compose through YAML
- **Production-ready**: Atomic builds, deterministic output

## Project Structure

```
my-site/
├── site.yaml              # Configuration
├── components/            # HTML components
│   ├── definitions.yaml
│   ├── navbar.html
│   └── footer.html
├── layouts/               # Page layouts
├── page-types/            # Page schemas
├── pages/                 # Markdown content
├── styles/                # CSS
└── assets/                # Static files
```

## Commands

| Command | Description |
|---------|-------------|
| `garterscopic init name` | Create a new project |
| `garterscopic build` | Build the site |
| `garterscopic dev` | Start development server |
| `garterscopic clean` | Remove generated files |
| `garterscopic new page name` | Create a new page |
| `garterscopic new post name` | Create a new post |

## JavaScript

JavaScript is completely optional. Place files in `assets/js/` and they copy to the output. No bundlers, no npm.

## Architecture

```
                    GARTERSCOPIC

                 ┌───────────────┐
                 │      YAML     │
                 │ declarations  │
                 └───────┬───────┘
                         │
                         ▼
┌──────────┐       ┌───────────────┐       ┌───────────┐
│ Markdown │ ────► │  Garterscopic │ ◄──── │   HTML    │
└──────────┘       │      Go       │       └───────────┘
                   └───────┬───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
           HTML           CSS       JavaScript
              │            │            │
              └────────────┼────────────┘
                           ▼
                       static site
```

## Development

```bash
# Clone and build
git clone https://github.com/italiatroller-1990/garterscopic
cd garterscopic
CGO_ENABLED=0 go build -o garterscopic ./cmd/garterscopic

# Run tests
go test ./...

# Build release for current platform
CGO_ENABLED=0 go build -ldflags="-s -w" -o garterscopic ./cmd/garterscopic
```

## Release Process

Releases are automated via GitHub Actions. To create a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow will:
1. Run CI tests
2. Build binaries for all platforms
3. Generate checksums
4. Create GitHub release
5. Build distro packages (deb, rpm, arch)

## License

MIT
