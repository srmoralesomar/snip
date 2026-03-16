#!/bin/sh
set -e

REPO="srmoralesomar/snip"

echo "Installing snip..."

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "${OS}" in
    Linux*)     OS_NAME="Linux" ;;
    Darwin*)    OS_NAME="Darwin" ;;
    *)          echo "Unsupported OS: ${OS}"; exit 1 ;;
esac

case "${ARCH}" in
    x86_64)     ARCH_NAME="x86_64" ;;
    amd64)      ARCH_NAME="x86_64" ;;
    arm64)      ARCH_NAME="arm64" ;;
    aarch64)    ARCH_NAME="arm64" ;;
    *)          echo "Unsupported architecture: ${ARCH}"; exit 1 ;;
esac

# Fetch latest release data
RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
echo "Fetching latest release from GitHub..."
DOWNLOAD_URL=$(curl -sL "$RELEASE_URL" | grep "browser_download_url" | grep "${OS_NAME}_${ARCH_NAME}" | grep "\.tar\.gz" | cut -d '"' -f 4 | head -n 1)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Could not find a pre-compiled binary for ${OS_NAME} ${ARCH_NAME}."
    echo "Please check the releases page: https://github.com/${REPO}/releases"
    echo "Alternatively, you can build from source using Go."
    exit 1
fi

echo "Downloading ${DOWNLOAD_URL}..."
TMP_DIR=$(mktemp -d)
curl -sL -o "${TMP_DIR}/snip.tar.gz" "${DOWNLOAD_URL}"

echo "Extracting..."
tar -xzf "${TMP_DIR}/snip.tar.gz" -C "${TMP_DIR}"

# Install
INSTALL_DIR="/usr/local/bin"
echo "Installing to ${INSTALL_DIR} (may require sudo password)..."
sudo mv "${TMP_DIR}/snip" "${INSTALL_DIR}/snip"
sudo chmod +x "${INSTALL_DIR}/snip"

# Cleanup
rm -rf "${TMP_DIR}"

echo "snip was installed successfully to ${INSTALL_DIR}/snip!"
echo "Run 'snip --help' to get started."
