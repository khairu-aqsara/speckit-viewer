#!/bin/sh
# Installs the latest speckit release for the current OS/arch.
# Usage: curl -fsSL https://raw.githubusercontent.com/khairu-aqsara/speckit-viewer/main/install.sh | sh
set -e

REPO="khairu-aqsara/speckit-viewer"
BINARY="speckit"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

os=$(uname -s)
arch=$(uname -m)

case "$os" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  *)
    echo "speckit: unsupported OS: $os (download a release manually from https://github.com/${REPO}/releases)" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "speckit: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$version" ]; then
  echo "speckit: could not determine the latest release" >&2
  exit 1
fi

version_num="${version#v}"
archive="speckit_${version_num}_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${version}/${archive}"

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

echo "speckit: downloading ${version} for ${os}/${arch}..."
curl -fsSL "$url" -o "$tmp_dir/$archive"
tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"

if mkdir -p "$INSTALL_DIR" 2>/dev/null && [ -w "$INSTALL_DIR" ]; then
  mv "$tmp_dir/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "speckit: $INSTALL_DIR is not writable, using sudo..."
  sudo mkdir -p "$INSTALL_DIR"
  sudo mv "$tmp_dir/$BINARY" "$INSTALL_DIR/$BINARY"
fi
chmod +x "$INSTALL_DIR/$BINARY"

echo "speckit: installed to ${INSTALL_DIR}/${BINARY}"
"$INSTALL_DIR/$BINARY" --version
