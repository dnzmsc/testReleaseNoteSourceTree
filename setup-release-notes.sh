#!/bin/bash
# setup-release-notes.sh
# One-time setup script for developers
# This script installs the release-notes hook globally for all Git repositories

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOL_PATH="$SCRIPT_DIR/release-notes"

echo "📝 Release Notes Hook Setup"
echo "=========================="

# Check if binary exists
if [ ! -f "$TOOL_PATH" ]; then
  echo "❌ Error: release-notes binary not found at $TOOL_PATH"
  echo "Please build it first: cd $SCRIPT_DIR && go build -o release-notes ."
  exit 1
fi

# Determine OS
OS=$(uname -s)
case "$OS" in
  Linux*)   SCRIPT_EXT="" ;;
  Darwin*)  SCRIPT_EXT="" ;;
  MINGW*)   SCRIPT_EXT=".bat" ;;
  MSYS*)    SCRIPT_EXT=".bat" ;;
  *)        SCRIPT_EXT="" ;;
esac

# Setup Git template directory (cross-platform)
if [[ "$OS" == "MINGW"* ]] || [[ "$OS" == "MSYS"* ]]; then
  # Windows
  GIT_TEMPLATE_DIR="$APPDATA\Git\templates\hooks"
else
  # macOS / Linux
  GIT_TEMPLATE_DIR="$HOME/.git-templates/hooks"
fi

echo "Setting up Git template at: $GIT_TEMPLATE_DIR"
mkdir -p "$GIT_TEMPLATE_DIR"

# Copy hooks to template directory
echo "Installing hooks..."
cp "$SCRIPT_DIR/.git-hooks/prepare-commit-msg.new" "$GIT_TEMPLATE_DIR/prepare-commit-msg"
cp "$SCRIPT_DIR/.git-hooks/pre-commit.new" "$GIT_TEMPLATE_DIR/pre-commit"
chmod +x "$GIT_TEMPLATE_DIR/prepare-commit-msg"
chmod +x "$GIT_TEMPLATE_DIR/pre-commit"

# Copy binary to a location in PATH or use absolute reference
if [[ "$OS" == "MINGW"* ]] || [[ "$OS" == "MSYS"* ]]; then
  # Windows: copy to Local AppData or Git bin
  DEST="$APPDATA\release-notes.exe"
  cp "$TOOL_PATH.exe" "$DEST" 2>/dev/null || cp "$TOOL_PATH" "$DEST"
  echo "Binary installed at: $DEST"
  echo "⚠️  Make sure this directory is in your PATH"
else
  # macOS / Linux: copy to /usr/local/bin or ~/.local/bin
  if [ -w "/usr/local/bin" ]; then
    sudo cp "$TOOL_PATH" "/usr/local/bin/release-notes"
    echo "Binary installed at: /usr/local/bin/release-notes"
  else
    mkdir -p "$HOME/.local/bin"
    cp "$TOOL_PATH" "$HOME/.local/bin/release-notes"
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
      echo ""
      echo "⚠️  Add ~/.local/bin to your PATH by adding this to ~/.bashrc or ~/.zshrc:"
      echo "export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
  fi
fi

# Configure Git to use template directory
git config --global init.templatedir "$GIT_TEMPLATE_DIR"

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. For NEW repositories:"
echo "   git init"
echo ""
echo "2. For EXISTING repositories (run once in each repo):"
echo "   rm -rf .git/hooks && git init"
echo "   OR manually copy hooks:"
echo "   cp -r $GIT_TEMPLATE_DIR/* your-repo/.git/hooks/"
echo ""
echo "Test it:"
echo "   cd your-repo && git commit --allow-empty -m 'test'"
