#!/usr/bin/env bash
set -e

# Default package metadata
VERSION="${1:-1.0.0}"
REVISION="${2:-1}"
PACKAGE_NAME="nova"
ARCH="amd64"
MAINTAINER="Deepsayan Das <deepsayandas274@gmail.com>"
SHORT_DESC="GalactOS CLI tool for project development and container management."
LONG_DESC="Nova provides streamlined environment management and project orchestration for GalactOS."

BUILD_DIR="dist/${PACKAGE_NAME}_${VERSION}-${REVISION}_${ARCH}"
OUTPUT_DEB="${PACKAGE_NAME}_${VERSION}-${REVISION}_${ARCH}.deb"

echo "==> 1. Compiling Linux binary..."
GOOS=linux GOARCH=amd64 go build -o nova_bin main.go

echo "==> 2. Setting up directory layout..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/DEBIAN"
mkdir -p "$BUILD_DIR/usr/bin"

echo "==> 3. Copying binary..."
mv nova_bin "$BUILD_DIR/usr/bin/nova"
chmod 755 "$BUILD_DIR/usr/bin/nova"

echo "==> 4. Generating DEBIAN/control file..."
cat <<EOF > "$BUILD_DIR/DEBIAN/control"
Package: ${PACKAGE_NAME}
Version: ${VERSION}-${REVISION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${MAINTAINER}
Description: ${SHORT_DESC}
 ${LONG_DESC}
EOF

echo "==> 5. Building .deb package (using /tmp for WSL permission compliance)..."
TMP_BUILD="/tmp/${PACKAGE_NAME}_${VERSION}-${REVISION}_${ARCH}"
rm -rf "$TMP_BUILD"
cp -r "$BUILD_DIR" "$TMP_BUILD"
chmod 755 "$TMP_BUILD/DEBIAN"
chmod 644 "$TMP_BUILD/DEBIAN/control"

dpkg-deb --build --root-owner-group "$TMP_BUILD"
cp "${TMP_BUILD}.deb" "./${OUTPUT_DEB}"
rm -rf "$TMP_BUILD"

echo "==> SUCCESS! Built package: ./${OUTPUT_DEB}"
