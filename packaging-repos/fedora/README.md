# Garterscopic Fedora Packages

This repository contains the RPM packages for [Garterscopic](https://github.com/garterscopic/garterscopic) for Fedora, RHEL, and CentOS.

Packages are automatically built and published when new releases are made upstream.

## Usage

```bash
# Add the repository
sudo dnf config-manager --add-repo https://italiatroller-1990.github.io/garterscopic-fedora/garterscopic.repo

# Install
sudo dnf install garterscopic
```

## Structure

```
.
├── .github/workflows/
│   └── build.yml       # Auto-builds on upstream release
├── README.md
├── SPECS/              # RPM spec files
├── SOURCES/            # Tarballs and patches
├── dist/               # Built packages (auto-generated)
└── Makefile            # Local build helpers
```

## How It Works

1. Upstream Garterscopic creates a new release
2. A webhook triggers this repo's CI
3. CI downloads the release artifacts
4. CI builds RPM packages for multiple Fedora/RHEL versions
5. CI publishes to this repo's `dist/` branch via GitHub Pages
