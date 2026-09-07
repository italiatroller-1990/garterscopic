# Garterscopic Arch Linux Packages

This repository contains the [Arch Linux](https://archlinux.org/) packages for [Garterscopic](https://github.com/garterscopic/garterscopic).

Packages are automatically built and published when new releases are made upstream.

## Usage

### From AUR (recommended)

```bash
# Using yay
yay -S garterscopic

# Using paru
paru -S garterscopic
```

### From this repo's package mirror

Add to `/etc/pacman.conf`:

```ini
[garterscopic]
SigLevel = Optional TrustAll
Server = https://italiatroller-1990.github.io/garterscopic-arch/$arch
```

Then:

```bash
sudo pacman -Sy garterscopic
```

## Structure

```
.
├── .github/workflows/
│   └── build.yml       # Auto-builds on upstream release
├── README.md
├── PKGBUILD            # Build script
├── garterscopic.install # Install hooks
├── dist/               # Built packages (auto-generated)
└── makepkg.conf        # Build config
```

## How It Works

1. Upstream Garterscopic creates a new release
2. A webhook triggers this repo's CI
3. CI downloads the release artifacts
4. CI builds packages with `makepkg`
5. CI publishes to this repo's `dist/` branch via GitHub Pages
