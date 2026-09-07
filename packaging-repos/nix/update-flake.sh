#!/usr/bin/env bash
set -euo pipefail

# Update the Nix flake to track the latest upstream release
# This is run by CI on new releases

UPSTREAM_REPO="${UPSTREAM_REPO:-italiatroller-1990/garterscopic}"
VERSION="${1:-}"

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

echo "Updating flake for version $VERSION..."

# Get checksums from upstream release
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

curl -sL "https://github.com/${UPSTREAM_REPO}/releases/download/${VERSION}/SHA256SUMS" -o "$TMPDIR/SHA256SUMS"

# Extract version without 'v' prefix
CLEAN_VERSION="${VERSION#v}"

# Update flake.nix with new version and checksums
sed -i "s/version = \".*\";/version = \"$CLEAN_VERSION\";/" flake.nix

for entry in \
    "x86_64-linux:garterscopic-linux-amd64" \
    "aarch64-linux:garterscopic-linux-arm64" \
    "x86_64-darwin:garterscopic-darwin-amd64" \
    "aarch64-darwin:garterscopic-darwin-arm64"
do
    system="${entry%%:*}"
    binary="${entry##*:}"
    hash=$(grep "$binary" "$TMPDIR/SHA256SUMS" | awk '{print $1}')
    hash_hex=$(echo "$hash" | sed 's/^\(.\{8\}\).*\(.\{8\}\)$/sha256-\1\2/')
    sed -i "/$system = \"/s|sha256 = pkgs\.lib\.fakeSha256;|sha256 = \"$hash_hex\";|" flake.nix || true
done

# Run nix flake update to refresh lock file
nix flake update

echo "Flake updated for $VERSION"
echo "Don't forget to commit and push the changes!"
