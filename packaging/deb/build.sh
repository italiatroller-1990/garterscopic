#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-}"
ARCH="${ARCH:-amd64}"
OUTPUT="${OUTPUT:-.}"

if [ -z "$VERSION" ]; then
    echo "VERSION must be set"
    exit 1
fi

# Strip leading v if present
VERSION="${VERSION#v}"

# Map architecture names
DEB_ARCH="$ARCH"
case "$ARCH" in
    amd64|x86_64)
        DEB_ARCH="amd64"
        ;;
    arm64|aarch64)
        DEB_ARCH="arm64"
        ;;
    *)
        DEB_ARCH="$ARCH"
        ;;
esac

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

PKG_DIR="$TEMP_DIR/package"
mkdir -p "$PKG_DIR/DEBIAN"
mkdir -p "$PKG_DIR/usr/local/bin"
mkdir -p "$PKG_DIR/usr/share/doc/garterscopic"
mkdir -p "$PKG_DIR/usr/share/bash-completion/completions"
mkdir -p "$PKG_DIR/usr/share/zsh/site-functions"

cat > "$PKG_DIR/DEBIAN/control" << EOF
Package: garterscopic
Version: ${VERSION}
Section: web
Priority: optional
Maintainer: Garterscopic Team <italiatroller@protonmail.com>
Homepage: https://github.com/italiatroller-1990/garterscopic
Description: Declarative Lightweight Static Site Generator
 Garterscopic is a simple, fast static site generator that lets you
 build websites using HTML, YAML, Markdown, and CSS - no Node.js required.
Architecture: ${DEB_ARCH}
Depends: bash (>= 4.0)
License: Apache-2.0
EOF

cat > "$PKG_DIR/DEBIAN/postinst" << 'POSTINST'
#!/bin/bash
set -e
if [ "$1" = "configure" ]; then
    if [ -x /usr/bin/update-alternatives ]; then
        update-alternatives --install /usr/bin/garterscopic garterscopic /usr/local/bin/garterscopic 100 2>/dev/null || true
    fi
fi
POSTINST
chmod +x "$PKG_DIR/DEBIAN/postinst"

cat > "$PKG_DIR/DEBIAN/prerm" << 'PRERM'
#!/bin/bash
set -e
if [ "$1" = "remove" ]; then
    if [ -x /usr/bin/update-alternatives ]; then
        update-alternatives --remove garterscopic /usr/local/bin/garterscopic 2>/dev/null || true
    fi
fi
PRERM
chmod +x "$PKG_DIR/DEBIAN/prerm"

cat > "$PKG_DIR/usr/share/doc/garterscopic/copyright" << EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: garterscopic
Upstream-Contact: Garterscopic Team <italiatroller@protonmail.com>
Source: https://github.com/italiatroller-1990/garterscopic

Files: *
Copyright: 2024-2026 Garterscopic Contributors
License: Apache-2.0

License: Apache-2.0
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
 .
     http://www.apache.org/licenses/LICENSE-2.0
 .
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
EOF

cat > "$PKG_DIR/usr/share/doc/garterscopic/changelog" << EOF
garterscopic (${VERSION}) stable; urgency=low

  * See https://github.com/italiatroller-1990/garterscopic/releases

  -- Garterscopic Team <italiatroller@protonmail.com>  $(date -R)
EOF
gzip -n "$PKG_DIR/usr/share/doc/garterscopic/changelog"

mkdir -p "$PKG_DIR/usr/share/doc/garterscopic/completions/bash"
cat > "$PKG_DIR/usr/share/doc/garterscopic/completions/bash/garterscopic" << 'COMPLETION'
_garterscopic()
{
    local cur prev words cword
    _init_completion || return

    case $prev in
        --help|--version)
            return
            ;;
        --port)
            return
            ;;
        init|new)
            COMPREPLY=($(compgen -W "page post" -- "$cur"))
            return
            ;;
    esac

    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "init build dev clean new --help --version" -- "$cur"))
    fi
} && complete -F _garterscopic garterscopic
COMPLETION

cp "$OUTPUT/garterscopic" "$PKG_DIR/usr/local/bin/garterscopic"
chmod 755 "$PKG_DIR/usr/local/bin/garterscopic"

dpkg-deb --build "$PKG_DIR" "$OUTPUT/garterscopic_${VERSION}_${DEB_ARCH}.deb"
echo "Created: $OUTPUT/garterscopic_${VERSION}_${DEB_ARCH}.deb"
