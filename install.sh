#!/bin/sh
set -e

# xSeek CLI installer
# Usage: curl -fsSL https://cli.xseek.io/install.sh | sh

REPO="xseekio/xseek-cli"
INSTALL_DIR="/usr/local/bin"
BINARY="xseek"

# Detect OS and architecture
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)
      echo "Error: Unsupported OS: $OS"
      exit 1
      ;;
  esac

  case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
      echo "Error: Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac
}

# Get latest release version from GitHub
get_latest_version() {
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version"
    exit 1
  fi
}

main() {
  echo "Installing xSeek CLI..."
  echo ""

  detect_platform
  get_latest_version

  FILENAME="xseek_${VERSION}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"

  echo "  Platform:  ${OS}/${ARCH}"
  echo "  Version:   v${VERSION}"
  echo "  URL:       ${URL}"
  echo ""

  # Download to temp directory
  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  echo "Downloading..."
  curl -fsSL "$URL" -o "${TMP_DIR}/${FILENAME}"

  echo "Extracting..."
  tar -xzf "${TMP_DIR}/${FILENAME}" -C "$TMP_DIR"

  echo "Installing to ${INSTALL_DIR}/${BINARY}..."
  if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  else
    sudo mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  fi
  chmod +x "${INSTALL_DIR}/${BINARY}"

  echo ""
  echo "✓ xSeek CLI v${VERSION} installed successfully!"
  echo ""
  echo "Get started:"
  echo "  export XSEEK_API_KEY=your_api_key"
  echo "  xseek scan robots yoursite.com"
  echo ""
  echo "Get your API key at: https://www.xseek.io/dashboard/api-keys"
}

main
