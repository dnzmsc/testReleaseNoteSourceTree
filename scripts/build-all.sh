#!/bin/bash
# build-all.sh
# Cross-compile the release-notes binary for all platforms.
#
# Fyne requires CGO, so cross-compilation needs a C cross-compiler
# for each target platform. For simplicity, build natively on each
# target OS, or use this script on the current platform only.
#
# Usage:
#   ./scripts/build-all.sh [version]
#   ./scripts/build-all.sh 1.0.0

set -e

VERSION=${1:-dev}
OUTPUT_DIR="build"

echo "📦 Building Release Notes Tool - Version $VERSION"
echo "=================================================="

mkdir -p "$OUTPUT_DIR"

# Detect current platform
CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m)

case "$CURRENT_ARCH" in
    x86_64)  CURRENT_ARCH="amd64" ;;
    arm64)   CURRENT_ARCH="arm64" ;;
    aarch64) CURRENT_ARCH="arm64" ;;
esac

case "$CURRENT_OS" in
    darwin) OS_LABEL="macos" ;;
    linux)  OS_LABEL="linux" ;;
    *)      OS_LABEL="$CURRENT_OS" ;;
esac

OUTPUT_NAME="release-notes-${OS_LABEL}-${CURRENT_ARCH}"

echo ""
echo "🔨 Building for $CURRENT_OS/$CURRENT_ARCH..."

CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.Version=$VERSION" \
    -o "$OUTPUT_DIR/$OUTPUT_NAME" \
    .

echo "✓ Built: $OUTPUT_DIR/$OUTPUT_NAME"

# Create archive
if command -v tar &> /dev/null; then
    tar -czf "$OUTPUT_DIR/$OUTPUT_NAME.tar.gz" -C "$OUTPUT_DIR" "$OUTPUT_NAME"
    echo "✓ Archive: $OUTPUT_DIR/$OUTPUT_NAME.tar.gz"
fi

echo ""
echo "✅ Build complete!"
echo "📂 Output: $OUTPUT_DIR/"
echo ""
echo "Nota: Fyne richiede CGO, quindi la cross-compilation necessita"
echo "di un cross-compiler C. Per build multi-piattaforma, compila"
echo "nativamente su ogni sistema operativo target."
