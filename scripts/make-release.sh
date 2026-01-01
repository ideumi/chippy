#!/bin/sh
set -e

# Configuration
RELEASE_NAME="chiplang"
VERSION="1.0.7"
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
RELEASE_DIR="${RELEASE_NAME}-${VERSION}-${OS}-${ARCH}"
ARCHIVE_NAME="${RELEASE_DIR}.tar.xz"

echo "Building release: ${ARCHIVE_NAME}"

# Helper functions
die() { echo "Error: $1" >&2; exit 1; }
run_cmd() { echo "$1"; eval "$1"; }

# Setup
[ -d "${RELEASE_DIR}" ] && rm -rf "${RELEASE_DIR}"
mkdir -p "${RELEASE_DIR}/content/lib" "${RELEASE_DIR}/content/doc"

# Build
[ -f "../Makefile" ] || die "Makefile not found."
run_cmd "cd .. && make build && cd rel"
[ -f "../chippy" ] || die "chippy binary not found after build"

# Copy files
cp ../chippy "${RELEASE_DIR}/content/"
[ -d "../lib" ] || die "lib directory not found"
cp ../lib/*.chh "${RELEASE_DIR}/content/lib/"
[ -d "../doc" ] && cp -r ../doc/* "${RELEASE_DIR}/content/doc/"

# Build installer
run_cmd "cd ../installer && ../chippy combine && cd ../rel"
[ -f "../installer/out/installer" ] || die "installer binary not found"
cp ../installer/out/installer "${RELEASE_DIR}/"

# Copy scripts and licenses
cp ../scripts/install-template.sh "${RELEASE_DIR}/install.sh"
cp ../scripts/uninstall-template.sh "${RELEASE_DIR}/uninstall.sh"
chmod +x "${RELEASE_DIR}/install.sh"
chmod +x "${RELEASE_DIR}/uninstall.sh"
[ -f "../LICENCE.txt" ] && cp ../LICENCE.txt "${RELEASE_DIR}/"
[ -f "../LICENCES_THIRDPARTY.txt" ] && cp ../LICENCES_THIRDPARTY.txt "${RELEASE_DIR}/"

# Package
echo "Creating archive..."
tar -cJf "${ARCHIVE_NAME}" --owner=0 --group=0 "${RELEASE_DIR}"
rm -rf "${RELEASE_DIR}"

# Results
if [ -f "${ARCHIVE_NAME}" ]; then
    SIZE=$(du -h "${ARCHIVE_NAME}" | cut -f1)
    echo "Created: ${ARCHIVE_NAME} (${SIZE})"
else
    die "Failed to create release package"
fi
