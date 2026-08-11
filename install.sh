#!/usr/bin/env sh
# Installs exe from the latest GitHub release.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/prjvvl/exe/main/install.sh | sh

set -eu

OWNER="prjvvl"
REPO="exe"
INSTALL_DIR="${EXE_INSTALL_DIR:-$HOME/.local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  linux) os="linux" ;;
  darwin) os="darwin" ;;
  *)
    echo "Unsupported OS: $os" >&2
    exit 1
    ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

echo "Fetching latest release info for $OWNER/$REPO..."
latest_url="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
download_url="$(curl -fsSL "$latest_url" | grep "browser_download_url.*${os}_${arch}.tar.gz" | cut -d '"' -f 4)"

if [ -z "$download_url" ]; then
  echo "Could not find a release asset for ${os}_${arch}" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
tmp_dir="$(mktemp -d)"

echo "Downloading $download_url..."
curl -fsSL "$download_url" -o "$tmp_dir/exe.tar.gz"

echo "Extracting to $INSTALL_DIR..."
tar -xzf "$tmp_dir/exe.tar.gz" -C "$INSTALL_DIR" exe
rm -rf "$tmp_dir"
chmod +x "$INSTALL_DIR/exe"

hook_marker="# exe shell hook"
needs_path_export=0
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) needs_path_export=1 ;;
esac

case "${SHELL:-}" in
  */zsh) rc_file="$HOME/.zshrc"; hook_line='eval "$(exe init zsh)"' ;;
  */bash) rc_file="$HOME/.bashrc"; hook_line='eval "$(exe init bash)"' ;;
  *) rc_file="" ;;
esac

if [ -n "$rc_file" ]; then
  touch "$rc_file"
  if ! grep -qF "$hook_marker" "$rc_file" 2>/dev/null; then
    {
      echo ""
      echo "$hook_marker"
      if [ "$needs_path_export" -eq 1 ]; then
        echo "export PATH=\"$INSTALL_DIR:\$PATH\""
      fi
      echo "$hook_line"
    } >> "$rc_file"
    echo "Added the exe shell hook to $rc_file"
  else
    echo "exe shell hook already present in $rc_file"
  fi
else
  echo ""
  echo "Could not detect your shell (\$SHELL=${SHELL:-unset}) to install the hook automatically."
  echo "Add this to your shell profile manually:"
  echo ""
  if [ "$needs_path_export" -eq 1 ]; then
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
  fi
  echo "  bash (~/.bashrc):  eval \"\$(exe init bash)\""
  echo "  zsh  (~/.zshrc):   eval \"\$(exe init zsh)\""
fi

echo ""
echo "exe installed to $INSTALL_DIR"
echo "Restart your terminal to start using it."
