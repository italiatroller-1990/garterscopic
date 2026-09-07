# Garterscopic Packaging Infrastructure

This document describes the multi-repo packaging strategy for Garterscopic.

## Overview

Garterscopic distributes through multiple package repositories, each managed as a separate Git repository:

| Repository | Purpose | URL |
|-----------|---------|-----|
| [garterscopic](https://github.com/italiatroller-1990/garterscopic) | Main source code | Base |
| [garterscopic-debian](https://github.com/italiatroller-1990/garterscopic-debian) | Debian/Ubuntu packages | APT repo |
| [garterscopic-fedora](https://github.com/italiatroller-1990/garterscopic-fedora) | Fedora/RHEL packages | YUM/DNF repo |
| [garterscopic-arch](https://github.com/italiatroller-1990/garterscopic-arch) | Arch Linux packages | pacman repo + AUR |
| [garterscopic-nix](https://github.com/italiatroller-1990/garterscopic-nix) | Nix/NixOS packages | Flake input |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│         garterscopic (main repo)                            │
│                                                              │
│  • Source code                                              │
│  • Go binary                                                │
│  • CI/CD                                                    │
│  • GitHub releases                                          │
└──────────────┬──────────────────────────────────────────────┘
               │ GitHub Release Created
               │
       ┌───────┴────────┬─────────────┬────────────┐
       │                │             │            │
       ▼                ▼             ▼            ▼
┌─────────────┐  ┌─────────────┐  ┌──────────┐  ┌─────────┐
│  debian     │  │   fedora    │  │   arch   │  │   nix   │
│  repo       │  │   repo      │  │   repo   │  │  flake  │
│             │  │             │  │          │  │         │
│  • DEB      │  │  • RPM      │  │  • PKG   │  │  • Flake│
│  • APT      │  │  • YUM      │  │  • AUR   │  │  • Nix  │
└─────────────┘  └─────────────┘  └──────────┘  └─────────┘
       │                │             │            │
       └────────────────┴─────────────┴────────────┘
                       │
              ┌────────┴────────┐
              │  Users install  │
              │   garterscopic  │
              │   via native    │
              │   package mgr   │
              └─────────────────┘
```

## Workflow

### Release Process

1. **Tag a new version** in the main `garterscopic` repo:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions** in the main repo:
   - Runs tests
   - Builds binaries for 6 platforms
   - Creates GitHub release with artifacts
   - Sends `repository_dispatch` webhook to all packaging repos

3. **Each packaging repo's CI** automatically:
   - Detects the new release
   - Downloads the appropriate binary
   - Builds a native package
   - Publishes to its repository

### Installation by Users

```bash
# Debian/Ubuntu
echo "deb [trusted=yes] https://italiatroller-1990.github.io/garterscopic-debian stable main" | sudo tee /etc/apt/sources.list.d/garterscopic.list
sudo apt update && sudo apt install garterscopic

# Fedora/RHEL
sudo dnf config-manager --add-repo https://italiatroller-1990.github.io/garterscopic-fedora/garterscopic.repo
sudo dnf install garterscopic

# Arch Linux
yay -S garterscopic  # or use the GitHub Pages mirror

# Nix/NixOS
nix profile install github:italiatroller-1990/garterscopic-nix
```

## Why Multiple Repos?

1. **Isolation**: Package-specific tooling and tooling configs don't pollute the main repo
2. **Independent releases**: Packages can be rebuilt/patched without changing source
3. **Smaller CI**: Each repo's CI is focused and faster
4. **Community contributions**: Easier for distro packagers to contribute
5. **Mirroring**: Distribution can use existing mirror infrastructure

## Local Development

To test packaging locally:

```bash
# Test DEB build
cd ../garterscopic-debian
./build-deb.sh

# Test RPM build
cd ../garterscopic-fedora
make test

# Test Arch build
cd ../garterscopic-arch
makepkg

# Test Nix flake
cd ../garterscopic-nix
nix build
```

## Status

| Repo | Status | Published |
|------|--------|-----------|
| debian | ✅ Active | APT repo via GitHub Pages |
| fedora | ✅ Active | YUM repo via GitHub Pages |
| arch | ✅ Active | pacman repo + AUR |
| nix | ✅ Active | Flake input |
