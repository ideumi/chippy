#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Verify installer exists
[ -f "${SCRIPT_DIR}/installer" ] || { echo "Error: installer binary not found"; exit 1; }

# Run installer in uninstall mode
"${SCRIPT_DIR}/installer" --uninstall