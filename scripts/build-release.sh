#!/bin/bash

# Script to build binaries for multiple platforms locally
# This is useful for testing the build process before pushing to GitHub

set -e

BINARY_NAME="pgm"
VERSION="dev"
BUILD_DIR="dist"

# Create build directory
mkdir -p $BUILD_DIR

# Define platforms to build for
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

echo "Building $BINARY_NAME version $VERSION for multiple platforms..."

for platform in "${PLATFORMS[@]}"; do
  GOOS=$(echo $platform | cut -d'/' -f1)
  GOARCH=$(echo $platform | cut -d'/' -f2)
  
  echo "Building for $GOOS/$GOARCH..."
  
  # Set binary name with extension for Windows
  if [ "$GOOS" = "windows" ]; then
    BIN_NAME="${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}.exe"
  else
    BIN_NAME="${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}"
  fi
  
  # Build the binary
  env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
    -ldflags="-s -w -X github.com/quadgod/pgm/pkg/pgm.Version=${VERSION}" \
    -o "$BUILD_DIR/$BIN_NAME" ./cmd/pgm
  
  # Create archive
  if [ "$GOOS" = "windows" ]; then
    cd $BUILD_DIR
    zip "${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}.zip" "$BIN_NAME" > /dev/null
    rm "$BIN_NAME"
    cd ..
  else
    tar -czf "$BUILD_DIR/${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}.tar.gz" -C $BUILD_DIR "$BIN_NAME" > /dev/null
    rm "$BUILD_DIR/$BIN_NAME"
  fi
done

# Calculate checksums
cd $BUILD_DIR
sha256sum * > ../SHA256SUMS
cd ..

echo "Build completed! Archives are in the $BUILD_DIR directory."
echo "Checksums saved to SHA256SUMS"