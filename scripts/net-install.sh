#!/bin/sh
# curl -fsSL https://codeberg.org/ideumi/chip-go/raw/branch/main/scripts/net-install.sh | sh

set -e

REPO="ideumi/chip-go"
PREFIX="chiplang"

die() { echo "Error: $1" >&2; exit 1; }

command -v curl >/dev/null 2>&1 || die "curl is required"
command -v tar  >/dev/null 2>&1 || die "tar is required"

case "$(uname -m)" in
	x86_64)         ARCH="x86_64" ;;
	aarch64|arm64)  ARCH="aarch64" ;;
	*)              die "Unsupported architecture: $(uname -m)" ;;
esac

if [ -n "$TERMUX_VERSION" ]; then
	SUDO=""
elif [ "$(id -u)" -ne 0 ] && command -v sudo >/dev/null 2>&1; then
	SUDO="sudo"
else
	SUDO=""
fi

echo "Fetching latest release info..."
TAG=$(curl -fsSL "https://codeberg.org/api/v1/repos/${REPO}/releases" \
	| grep -o '"tag_name":"[^"]*"' | head -1 | cut -d'"' -f4)
[ -n "$TAG" ] || die "Could not determine latest version"
VERSION="${TAG#V-}"
echo "Latest: $TAG"

TARBALL="${PREFIX}-${VERSION}-linux-${ARCH}.tar.xz"
URL="https://codeberg.org/${REPO}/releases/download/${TAG}/${TARBALL}"

TMPDIR=$(mktemp -d)
trap 'rm -r "$TMPDIR"' EXIT

echo "Downloading $TARBALL..."
curl -fsSL -o "$TMPDIR/$TARBALL" "$URL" || die "Download failed: $URL"

echo "Extracting..."
tar -xf "$TMPDIR/$TARBALL" -C "$TMPDIR"
EXTRACTED="$TMPDIR/${PREFIX}-${VERSION}-linux-${ARCH}"
[ -d "$EXTRACTED" ] || die "Expected directory not found after extract: $EXTRACTED"

[ -f "$EXTRACTED/install.sh" ] || die "Bundled install.sh not found in tarball"

cd "$EXTRACTED"
$SUDO sh ./install.sh
