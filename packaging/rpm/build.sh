#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-}"
ARCH="${ARCH:-x86_64}"
OUTPUT="${OUTPUT:-.}"

if [ -z "$VERSION" ]; then
    echo "VERSION must be set"
    exit 1
fi

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

mkdir -p "$TEMP_DIR/rpmbuild/BUILD"
mkdir -p "$TEMP_DIR/rpmbuild/RPMS"
mkdir -p "$TEMP_DIR/rpmbuild/SOURCES"
mkdir -p "$TEMP_DIR/rpmbuild/SPECS"
mkdir -p "$TEMP_DIR/rpmbuild/SRPMS"

ARCH_MAP=""
case "$ARCH" in
    amd64|x86_64)
        ARCH_MAP="x86_64"
        ;;
    arm64|aarch64)
        ARCH_MAP="aarch64"
        ;;
    *)
        ARCH_MAP="$ARCH"
        ;;
esac

SPEC_FILE="$TEMP_DIR/rpmbuild/SPECS/garterscopic.spec"
sed "s/%{VERSION}/$VERSION/g; s/%{ARCH}/$ARCH_MAP/g" \
    "$(dirname "$0")/garterscopic.spec" > "$SPEC_FILE"

cp "$OUTPUT/garterscopic" "$TEMP_DIR/rpmbuild/"
cp -r "$(dirname "$0")/../.." "$TEMP_DIR/rpmbuild/BUILD/garterscopic-$VERSION" 2>/dev/null || true

rpmbuild --define "_topdir $TEMP_DIR/rpmbuild" \
    --define "dist ." \
    --target "$ARCH_MAP-linux" \
    -ba "$SPEC_FILE"

find "$TEMP_DIR/rpmbuild/RPMS" -name "*.rpm" -exec cp -v {} "$OUTPUT/" \;
find "$TEMP_DIR/rpmbuild/SRPMS" -name "*.rpm" -exec cp -v {} "$OUTPUT/" \;

echo "Created RPM packages in $OUTPUT"
