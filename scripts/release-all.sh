#!/bin/sh
set -e

( TARGET_GOARCH=amd64  TARGET_ARCH=x86_64  ../scripts/make-release.sh )
( TARGET_GOARCH=arm64  TARGET_ARCH=aarch64 ../scripts/make-release.sh )
