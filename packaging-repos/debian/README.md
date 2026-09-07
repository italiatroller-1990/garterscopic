# Garterscopic Debian Packages

This repository contains the Debian/Ubuntu packages for [Garterscopic](https://github.com/garterscopic/garterscopic).

Packages are automatically built and published when new releases are made upstream.

## Usage

```bash
# Add the repository
echo "deb [trusted=yes] https://italiatroller-1990.github.io/garterscopic-debian stable main" | sudo tee /etc/apt/sources.list.d/garterscopic.list

# Update and install
sudo apt update
sudo apt install garterscopic
```

## Structure

```
.
├── .github/workflows/
│   └── build.yml       # Auto-builds on upstream release
├── README.md
├── debian/             # Package build scripts
├── dists/              # Repository metadata
├── pool/               # Built packages
└── Makefile            # Local build helpers
```

## How It Works

1. Upstream Garterscopic creates a new release
2. A webhook triggers this repo's CI
3. CI downloads the release artifacts
4. CI builds DEB packages
5. CI publishes to this repo's `dist/` branch via GitHub Pages

See [`.github/workflows/build.yml`](.github/workflows/build.yml) for details.
