#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-}"

if [ -z "$VERSION" ]; then
    echo "VERSION must be set"
    exit 1
fi

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Create tarball of source
tar czf "$TEMP_DIR/garterscopic-$VERSION.tar.gz" \
    --exclude='.git' \
    --exclude='site' \
    --exclude='dist' \
    --exclude='.github' \
    .

# Build with Nix
cd "$TEMP_DIR"
nix flake init --template nixos-shell
echo "Source tarball created at: $TEMP_DIR/garterscopic-$VERSION.tar.gz"
echo "For Nix packaging, use the release tarball with the flake.nix in packaging-repos/nix/"
