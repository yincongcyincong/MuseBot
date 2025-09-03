#!/bin/bash

set -e

# Clean up old files
rm -rf ./output ./release
mkdir -p ./output ./release

# Check if xgo is installed
if ! command -v xgo &> /dev/null; then
    echo "Installing xgo..."
    go install src.techknowlogick.com/xgo@latest
fi

# Build the admin binary (locally for the specified platform)
build_admin_local() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    local admin_output="admin-${os}-${arch}${ext}"

    echo "=============================="
    echo "Building admin [$os/$arch] using go build..."
    echo "=============================="
    xgo -out "$admin_output" -targets="$os/$arch" --hooksdir=./admin/shell ./
}

# Build main binary + package everything
compile_and_package() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    echo "=============================="
    echo "Building MuseBot [$os/$arch] using xgo..."
    echo "=============================="

    # Build the main bot binary
    xgo -out MuseBot -targets="$os/$arch" .

    # Build admin binary
    build_admin_local $os $arch

    local bot_binary="MuseBot-${os}-${arch}${ext}"
    local admin_binary="MuseBot-admin-${os}-${arch}${ext}"
    local release_name="MuseBot-${os}-${arch}.tar.gz"

    # Move compiled binaries to output
    mv ./MuseBot-${os}* ./output/${bot_binary}
    mv ./admin-${os}* ./output/${admin_binary}

    # Copy config files
    mkdir -p ./output/conf/
    cp -r ./conf/i18n ./output/conf/
    cp -r ./conf/mcp ./output/conf/
    cp -r ./conf/chat ./output/conf/
    mkdir -p ./output/data/

    # Copy admin UI files
    mkdir -p ./output/adminui/
    cp -r ./admin/adminui/* ./output/adminui/

    # Package everything into a tarball
    tar zcf "release/${release_name}" -C ./output .

    # Clean up intermediate files
    rm -rf ./output/* ./github.com/*
}

# Platforms to compile (uncomment Windows if needed)
compile_and_package linux amd64
compile_and_package darwin amd64
compile_and_package darwin arm64
compile_and_package windows amd64

# Final cleanup
rm -rf ./output
echo "✅ Compilation and packaging complete. Output is in ./release"
