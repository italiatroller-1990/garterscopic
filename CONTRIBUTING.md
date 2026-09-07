# Contributing to Garterscopic

## Development

### Prerequisites

- Go 1.21 or later
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/garterscopic/garterscopic
cd garterscopic

# Build (no CGO required)
CGO_ENABLED=0 go build -o garterscopic ./cmd/garterscopic

# Run tests
go test ./...
```

### Running the Example Site

```bash
# Build the binary
CGO_ENABLED=0 go build -o garterscopic ./cmd/garterscopic

# Test with the example project
cd example
../garterscopic build
```

## Release Process

### Versioning

Garterscopic uses semantic versioning (SemVer):
- `vMAJOR.MINOR.PATCH` (e.g., `v1.2.0`)

### Creating a Release

1. **Ensure tests pass:**
   ```bash
   go test ./...
   CGO_ENABLED=0 go build -o garterscopic ./cmd/garterscopic
   ```

2. **Update version if needed** (currently hardcoded in `cmd/garterscopic/main.go`)

3. **Create and push a tag:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. **GitHub Actions handles the rest:**
   - Runs CI tests
   - Builds binaries for all platforms
   - Generates checksums
   - Creates GitHub release
   - Builds distro packages (DEB, RPM, Arch)

### Release Platforms

| Platform | Architectures |
|----------|---------------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |
| FreeBSD | amd64, arm64 |

Windows is **not** officially distributed.

### Build Matrix

The release workflow builds for these targets:
- `garterscopic-linux-amd64`
- `garterscopic-linux-arm64`
- `garterscopic-darwin-amd64`
- `garterscopic-darwin-arm64`
- `garterscopic-freebsd-amd64`
- `garterscopic-freebsd-arm64`

## Distro Packaging

### Debian/Ubuntu (DEB)

```bash
./packaging/deb/build.sh
# Creates: garterscopic_${VERSION}_${ARCH}.deb
```

### Fedora/RHEL (RPM)

```bash
./packaging/rpm/build.sh
# Creates: garterscopic-${VERSION}-1.${ARCH}.rpm
```

### Arch Linux

Use the `PKGBUILD` in `packaging/arch/`.

### NixOS

Use the `default.nix` in `packaging/nix/`.

## CI/CD Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        GitHub Actions                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  push tag ─────► Release Workflow                           │
│       │                                                         │
│       │         ┌──────────────────────────────────┐       │
│       │         │  1. CI (test, format, vet)       │       │
│       │         └──────────────────────────────────┘       │
│       │                      │                              │
│       │                      ▼                              │
│       │         ┌──────────────────────────────────┐       │
│       │         │  2. Build Matrix (6 targets)    │       │
│       │         └──────────────────────────────────┘       │
│       │                      │                              │
│       │                      ▼                              │
│       │         ┌──────────────────────────────────┐       │
│       │         │  3. Checksums (SHA256SUMS)       │       │
│       │         └──────────────────────────────────┘       │
│       │                      │                              │
│       │                      ▼                              │
│       │         ┌──────────────────────────────────┐       │
│       │         │  4. Create GitHub Release         │       │
│       │         └──────────────────────────────────┘       │
│       │                      │                              │
│       │                      ▼                              │
│       │         ┌──────────────────────────────────┐       │
│       │         │  5. Distro Packages (deb/rpm/arch)│       │
│       │         └──────────────────────────────────┘       │
│       │                      │                              │
│       └──────────────────────┴──────────────────────────────┘
```

## Dependency Policy

- **Runtime**: Zero external dependencies
- **Build time**: Pure Go modules only
- **CGO**: Disabled (`CGO_ENABLED=0`)
- **No native extensions**: All dependencies must be pure Go

## Code Style

- Run `go fmt ./...` before committing
- Run `go vet ./...` to check for issues
- All public functions should have documentation comments

## Testing

```bash
# Unit tests
go test ./...

# With race detection
go test -race ./...

# Verbose output
go test -v ./...
```

## Project Structure

```
garterscopic/
├── .github/workflows/     # GitHub Actions
│   ├── ci.yml            # Continuous integration
│   ├── release.yml       # Release automation
│   └── packages.yml      # Distro packaging
├── cmd/garterscopic/     # CLI entry point
├── internal/             # Core packages
│   ├── build/           # Site builder
│   ├── components/      # HTML components
│   ├── config/          # Configuration
│   ├── layouts/          # Page layouts
│   ├── markdown/         # Markdown rendering
│   ├── pages/            # Page handling
│   ├── renderer/          # HTML rendering
│   ├── routing/          # URL routing
│   ├── server/           # Dev server
│   └── validation/       # Input validation
├── packaging/            # Distro packaging
│   ├── deb/             # Debian packages
│   ├── rpm/             # Fedora/RHEL packages
│   ├── arch/            # Arch Linux packages
│   └── nix/             # NixOS packages
└── docs/                # Documentation
```
