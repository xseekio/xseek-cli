#!/bin/sh
set -e

# xSeek CLI installer
# Usage: curl -fsSL https://cli.xseek.io/install.sh | sh

REPO="xseekio/xseek-cli"
XSEEK_HOME="${HOME}/.xseek"
INSTALL_DIR="${XSEEK_HOME}/bin"
BINARY="xseek"

# Detect OS and architecture
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
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

  # Set binary name and archive format based on OS
  if [ "$OS" = "windows" ]; then
    BINARY="xseek.exe"
    ARCHIVE_EXT="zip"
  else
    ARCHIVE_EXT="tar.gz"
  fi
}

# Get latest release version from GitHub
get_latest_version() {
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version"
    exit 1
  fi
}

# Add ~/.xseek/bin to PATH in shell profile
add_to_path() {
  if [ "$OS" = "windows" ]; then
    # On Git Bash, add to .bashrc
    PROFILE="${HOME}/.bashrc"
  else
    SHELL_NAME="$(basename "$SHELL")"
    case "$SHELL_NAME" in
      zsh)  PROFILE="${HOME}/.zshrc" ;;
      bash) PROFILE="${HOME}/.bashrc" ;;
      *)    PROFILE="${HOME}/.profile" ;;
    esac
  fi

  if [ -f "$PROFILE" ] && grep -q "${INSTALL_DIR}" "$PROFILE" 2>/dev/null; then
    return 0
  fi

  echo "" >> "$PROFILE"
  echo "# xSeek CLI" >> "$PROFILE"
  echo "export PATH=\"${INSTALL_DIR}:\$PATH\"" >> "$PROFILE"
  echo "  Added ${INSTALL_DIR} to PATH in ${PROFILE}"
}

main() {
  echo "Installing xSeek CLI..."
  echo ""

  detect_platform
  get_latest_version

  FILENAME="xseek_${VERSION}_${OS}_${ARCH}.${ARCHIVE_EXT}"
  URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"

  echo "  Platform:  ${OS}/${ARCH}"
  echo "  Version:   v${VERSION}"
  echo "  Install:   ${INSTALL_DIR}/${BINARY}"
  echo ""

  # Create ~/.xseek/bin
  mkdir -p "${INSTALL_DIR}"

  # Download to temp directory
  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  echo "Downloading..."
  curl -fsSL "$URL" -o "${TMP_DIR}/${FILENAME}"

  echo "Extracting..."
  if [ "$ARCHIVE_EXT" = "zip" ]; then
    unzip -q "${TMP_DIR}/${FILENAME}" -d "$TMP_DIR"
  else
    tar -xzf "${TMP_DIR}/${FILENAME}" -C "$TMP_DIR"
  fi

  echo "Installing..."
  mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"

  # Add to PATH if not already there
  add_to_path

  echo ""
  echo "  xSeek CLI v${VERSION} installed successfully!"
  echo ""
  echo "  Restart your terminal or run:"
  echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
  echo ""
  echo "  Then authenticate:"
  echo "    xseek login YOUR_API_KEY"
  echo ""
  echo "  Get your API key at: https://www.xseek.io/dashboard/api-keys"
}

main
