#!/usr/bin/env bash
# Install DevForge (forge) from GitHub Releases.
# Usage: curl -sSL https://raw.githubusercontent.com/devforge/devforge/main/scripts/install.sh | bash

set -e

REPO="devforge/devforge"
BASE_URL="https://github.com/${REPO}/releases/latest/download"
BINARY_NAME="forge"

# Detect OS: Linux → linux, Darwin → darwin, Windows → windows
OS="$(uname -s)"
case "${OS}" in
  Linux)   OS=linux ;;
  Darwin)  OS=darwin ;;
  MINGW*|MSYS*|CYGWIN*) OS=windows ;;
  *)
    echo "Unsupported OS: ${OS}" >&2
    exit 1
    ;;
esac

# Detect architecture: x86_64 → amd64, arm64/aarch64 → arm64
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *)
    echo "Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

# Windows arm64 not published; only amd64 is available for Windows
if [ "${OS}" = "windows" ] && [ "${ARCH}" = "arm64" ]; then
  echo "Windows arm64 is not supported. Use Windows amd64 or install manually." >&2
  exit 1
fi

# Artifact filename: Windows uses .exe suffix
SUFFIX=""
[ "${OS}" = "windows" ] && SUFFIX=".exe"
FILE="${BINARY_NAME}-${OS}-${ARCH}${SUFFIX}"
URL="${BASE_URL}/${FILE}"

echo "Installing ${BINARY_NAME} (${OS}-${ARCH}) from ${URL}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required. Install curl and try again." >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT
TMP_FILE="${TMP_DIR}/${FILE}"
curl -fsSL "${URL}" -o "${TMP_FILE}"
chmod +x "${TMP_FILE}"

# Install destination
if [ "${OS}" = "windows" ]; then
  INSTALL_DIR="${HOME}/bin"
  INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}${SUFFIX}"
  mkdir -p "${INSTALL_DIR}"
  mv "${TMP_FILE}" "${INSTALL_PATH}"
  echo "Installed to ${INSTALL_PATH}"
  echo "Add ${INSTALL_DIR} to your PATH if it is not already."
else
  INSTALL_PATH="/usr/local/bin/${BINARY_NAME}"
  if [ -w /usr/local/bin ]; then
    mv "${TMP_FILE}" "${INSTALL_PATH}"
    echo "Installed to ${INSTALL_PATH}"
  else
    echo "Writing to ${INSTALL_PATH} requires sudo."
    sudo mv "${TMP_FILE}" "${INSTALL_PATH}"
    echo "Installed to ${INSTALL_PATH}"
  fi
fi

echo "Run: ${BINARY_NAME} --help"
