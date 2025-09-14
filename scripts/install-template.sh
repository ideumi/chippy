#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Verify files exist
[ -f "${SCRIPT_DIR}/installer" ] || { echo "Error: installer binary not found"; exit 1; }
[ -f "${SCRIPT_DIR}/content/chippy" ] || { echo "Error: chippy binary not found"; exit 1; }

# Run installer
chmod +x "${SCRIPT_DIR}/content/chippy"
"${SCRIPT_DIR}/content/chippy" "${SCRIPT_DIR}/installer"
