#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_DIR="${ROOT_DIR}/services/gallery-api"
BUILD_DIR="${API_DIR}/build"
BOOTSTRAP="${BUILD_DIR}/bootstrap"
PACKAGE="${BUILD_DIR}/gallery-api.zip"
GO_BIN="${GO_BIN:-go}"

mkdir -p "${BUILD_DIR}"

# Lambda's provided.al2023 runtime expects an executable named bootstrap. Build
# explicitly for ARM64 so the deployed artifact matches the Terraform setting,
# even when the developer is working from an x86_64 workstation.
(
  cd "${API_DIR}"
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 "${GO_BIN}" build \
    -tags lambda.norpc \
    -trimpath \
    -ldflags="-s -w" \
    -o "${BOOTSTRAP}" \
    ./cmd/api
)

rm -f "${PACKAGE}"
(
  cd "${BUILD_DIR}"
  zip -q "${PACKAGE}" bootstrap
)

printf 'Created %s\n' "${PACKAGE}"
