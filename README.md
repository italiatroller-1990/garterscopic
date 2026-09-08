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

# Linux arm64
curl -sL https://github.com/italiatroller-1990/garterscopic/releases/latest/download/garterscopic-linux-arm64 -o garterscopic
chmod +x garterscopic
sudo mv garterscopic /usr/local/bin/

# macOS amd64 (Intel)
curl -sL https://github.com/italiatroller-1990/garterscopic/releases/latest/download/garterscopic-darwin-amd64 -o garterscopic
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
garterscopic --help
```

### Package Managers

**Go:**

```bash
go install github.com/italiatroller-1990/garterscopic/cmd/garterscopic@latest
```

**Fedora / RHEL / Rocky / AlmaLinux:**

```bash
# Download the RPM from GitHub Releases
sudo dnf install ./garterscopic-<version>-1.x86_64.rpm
```

**Debian / Ubuntu:**

```bash
# Download the DEB from GitHub Releases
sudo apt install ./garterscopic_<version>_amd64.deb
```

**Arch Linux:**

```bash
# Using makepkg from the AUR or GitHub Releases
makepkg -si garterscopic-<version>-1-x86_64.pkg.tar.zst
```

**Nix:**

```bash
nix run github:italiatroller-1990/garterscopic
```

**Homebrew (macOS/Linux):**

```bash
brew install garterscopic/tap/garterscopic
```

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
├── posts/                 # Blog posts (optional)
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
1. Run CI tests (formatting, vet, unit tests, race tests, and integration tests)
2. Build binaries for all platforms (Linux, macOS, FreeBSD)
3. Generate SHA-256 checksums
4. Create GitHub release with binaries and source archive
5. Build distro packages (DEB, RPM, Arch, Nix) as GitHub Actions artifacts
6. Attach the package artifacts and package checksums to the GitHub release
7. Attest binary build provenance
8. Notify external packaging repos (Debian, Fedora, Arch, Nix)

### Supported Distributions

| Distribution | Package Type | Architectures |
|-------------|-------------|---------------|
| Fedora 38+ | RPM | x86_64, aarch64 |
| RHEL 8+ / Rocky / Alma | RPM | x86_64, aarch64 |
| openSUSE Tumbleweed | RPM | x86_64, aarch64 |
| Debian 12+ | DEB | amd64, arm64 |
| Ubuntu 22.04+ | DEB | amd64, arm64 |
| Arch Linux | PKG | x86_64, aarch64 |
| NixOS | Flake | x86_64, aarch64 |

### Go Requirements

- Go 1.21+
- `CGO_ENABLED=0` (no cgo required)

### How to Download Actions Artifacts

Pull request and branch builds expose short-lived verification artifacts on the [Actions tab](https://github.com/italiatroller-1990/garterscopic/actions). Release binaries and packages are attached to the release itself and should be downloaded from the [Releases page](https://github.com/italiatroller-1990/garterscopic/releases).

See [Packaging and Releases](docs/packaging.md) for the tag workflow, checksum verification, required secrets, and local package checks.

### How to Download Releases

Visit the [Releases page](https://github.com/italiatroller-1990/garterscopic/releases) to download binaries and packages.

## License

Apache-2.0
