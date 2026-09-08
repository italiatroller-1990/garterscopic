#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-}"
ARCH="${ARCH:-x86_64}"
OUTPUT="${OUTPUT:-.}"

if [ -z "$VERSION" ]; then
    echo "VERSION must be set"
    exit 1
fi

# Strip leading v if present
VERSION="${VERSION#v}"

# Map architecture names
RPM_ARCH="$ARCH"
case "$ARCH" in
    amd64|x86_64)
        RPM_ARCH="x86_64"
        ;;
    arm64|aarch64)
        RPM_ARCH="aarch64"
        ;;
    *)
        RPM_ARCH="$ARCH"
        ;;
esac

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

mkdir -p "$TEMP_DIR/rpmbuild/BUILD"
mkdir -p "$TEMP_DIR/rpmbuild/RPMS"
mkdir -p "$TEMP_DIR/rpmbuild/SOURCES"
mkdir -p "$TEMP_DIR/rpmbuild/SPECS"
mkdir -p "$TEMP_DIR/rpmbuild/SRPMS"

# Prepare sources
cp "$OUTPUT/garterscopic" "$TEMP_DIR/rpmbuild/"
cp LICENSE README.md "$TEMP_DIR/rpmbuild/"
cp -r packaging "$TEMP_DIR/rpmbuild/packaging"

# Create spec file from template
SPEC_FILE="$TEMP_DIR/rpmbuild/SPECS/garterscopic.spec"
sed "s/%{_version}/$VERSION/g; s/%{_arch}/$RPM_ARCH/g" \
    "$(dirname "$0")/garterscopic.spec" > "$SPEC_FILE"

rpmbuild --define "_topdir $TEMP_DIR/rpmbuild" \
    --define "_version $VERSION" \
    --define "_arch $RPM_ARCH" \
    --define "dist ." \
    --target "$RPM_ARCH-linux" \
    -bb "$SPEC_FILE"

find "$TEMP_DIR/rpmbuild/RPMS" -name "*.rpm" -exec cp -v {} "$OUTPUT/" \;

echo "Created RPM packages in $OUTPUT"
