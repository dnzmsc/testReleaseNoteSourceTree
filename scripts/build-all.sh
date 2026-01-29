#!/bin/bash
# build-all.sh
# Cross-compile for all platforms

set -e

VERSION=${1:-dev}
OUTPUT_DIR="build"

echo "📦 Building Release Notes Tool - Version $VERSION"
echo "=================================================="

mkdir -p "$OUTPUT_DIR"

# Array of platforms
platforms=(
  "darwin:amd64:macos-x86_64"
  "darwin:arm64:macos-arm64"
  "linux:amd64:linux-x86_64"
  "linux:arm64:linux-arm64"
  "windows:amd64:windows-x86_64"
)

for platform in "${platforms[@]}"; do
  IFS=":" read -r os arch osLabel <<< "$platform"
  
  echo ""
  echo "🔨 Building for $os/$arch..."
  
  output_name="release-notes-$osLabel"
  [ "$os" = "windows" ] && output_name="$output_name.exe"
  
  GOOS=$os GOARCH=$arch go build \
    -ldflags="-s -w -X main.Version=$VERSION" \
    -o "$OUTPUT_DIR/$output_name" \
    .
  
  # Create archive
  cd "$OUTPUT_DIR"
  if [ "$os" = "windows" ]; then
    zip -q "release-notes-$osLabel.zip" "$output_name"
    echo "✓ $osLabel → release-notes-$osLabel.zip"
  else
    tar -czf "release-notes-$osLabel.tar.gz" "$output_name"
    echo "✓ $osLabel → release-notes-$osLabel.tar.gz"
  fi
  cd ..
done

echo ""
echo "✅ Build complete!"
echo "📂 Artifacts in: $OUTPUT_DIR/"
echo ""
echo "To distribute:"
echo "  • Upload $OUTPUT_DIR/*.tar.gz and *.zip to GitHub Releases"
echo "  • Or copy individual binaries to system PATH"
