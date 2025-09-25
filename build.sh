#!/bin/bash

set -e

VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COSIGN_PASSWORD=${COSIGN_PASSWORD:-""}
COMMIT_SHA=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

CLEAN_VERSION=$(echo "$VERSION" | sed 's/-dirty$//')

mkdir -p dist

TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

LDFLAGS="-X github.com/schnauzersoft/ami-util/cmd.Version=${VERSION} -X github.com/schnauzersoft/ami-util/cmd.CommitSHA=${COMMIT_SHA} -X github.com/schnauzersoft/ami-util/cmd.BuildTime=${BUILD_TIME}"

echo "Building ami-util version ${VERSION}"
echo "Commit: ${COMMIT_SHA}"
echo "Build time: ${BUILD_TIME}"
echo "Clean version: ${CLEAN_VERSION}"
echo "Targets: ${#TARGETS[@]}"
if [ -n "$COSIGN_PASSWORD" ]; then
    echo "Cosign signing: enabled"
else
    echo "Cosign signing: disabled"
fi
echo ""

for target in "${TARGETS[@]}"; do
    IFS='/' read -r os arch <<< "$target"
    
    echo "Building for ${os}/${arch}..."
    
    binary_name="ami-util-${CLEAN_VERSION}"
    
    export GOOS="$os"
    export GOARCH="$arch"
    
    go build -ldflags "$LDFLAGS" -o "dist/${binary_name}" main.go
    
    release_dir="dist/ami-util-${CLEAN_VERSION}-${os}-${arch}"
    mkdir -p "$release_dir"
    
    cp "dist/${binary_name}" "$release_dir/"
    
    if [ -f "README.md" ]; then
        cp "README.md" "$release_dir/"
    fi
    
    if [ -f "LICENSE" ]; then
        cp "LICENSE" "$release_dir/"
    fi
    
    cd dist
    tar -czf "ami-util-${CLEAN_VERSION}-${os}-${arch}.tar.gz" "ami-util-${CLEAN_VERSION}-${os}-${arch}/"
    cd ..
    
    if [ -n "$COSIGN_PASSWORD" ]; then
        echo "  Signing with cosign..."
        echo "$COSIGN_PASSWORD" | cosign sign-blob --yes --key /tmp/cosign.key --output-signature "dist/ami-util-${CLEAN_VERSION}-${os}-${arch}.tar.gz.sig" "dist/ami-util-${CLEAN_VERSION}-${os}-${arch}.tar.gz"
        if [ $? -eq 0 ]; then
            echo "  ✓ Signed: ami-util-${CLEAN_VERSION}-${os}-${arch}.tar.gz"
        else
            echo "  ✗ Failed to sign: ami-util-${CLEAN_VERSION}-${os}-${arch}.tar.gz"
        fi
    fi
    
    rm -rf "$release_dir"
    
    size=$(stat -f%z "dist/ami-util-${CLEAN_VERSION}" 2>/dev/null || stat -c%s "dist/ami-util-${CLEAN_VERSION}" 2>/dev/null || echo "unknown")
    
    echo "  ✓ Built: ami-util-${CLEAN_VERSION}-${os}-${arch}.tar.gz (${size} bytes)"
done

echo ""
echo "Build complete! Release files created in dist/ directory:"
ls -la dist/

echo ""
echo "To test a binary:"
echo "  tar -xzf dist/ami-util-${CLEAN_VERSION}-linux-amd64.tar.gz"
echo "  ./ami-util-${CLEAN_VERSION}-linux-amd64/ami-util-${CLEAN_VERSION} version"

if [ -n "$COSIGN_PASSWORD" ]; then
    echo ""
    echo "To verify signatures:"
    echo "  cosign verify-blob --key cosign.pub --signature dist/ami-util-${CLEAN_VERSION}-linux-amd64.tar.gz.sig dist/ami-util-${CLEAN_VERSION}-linux-amd64.tar.gz"
fi

rm dist/"ami-util-${CLEAN_VERSION}"
