#!/bin/bash

set -e

# Clean up old files
rm -rf ./output ./release
mkdir -p ./output ./release

# Build the admin binary (locally for the specified platform)
build_admin_local() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    local output_name="MuseBotAdmin"
    echo "=============================="
    echo "Building admin [$os/$arch] using go build..."
    echo "=============================="

    GOOS=$os GOARCH=$arch CGO_ENABLED=1 go build -o "./output/${output_name}" ./admin
}

# Build main binary + package everything
compile_and_package_local() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    echo "=============================="
    echo "Building MuseBot [$os/$arch] using go build..."
    echo "=============================="

    local bot_output="MuseBot"

    # Build main bot binary
    GOOS=$os GOARCH=$arch CGO_ENABLED=1 go build -o "./output/${bot_output}" ./

    # Build admin binary
    build_admin_local $os $arch

    # Copy config files
    mkdir -p ./output/conf/
    cp -r ./conf/i18n ./output/conf/
    cp -r ./conf/mcp ./output/conf/
    cp -r ./conf/img ./output/conf/
    mkdir -p ./output/data/

    # Copy admin UI files
    mkdir -p ./output/adminui/
    cp -r ./admin/adminui/* ./output/adminui/

    # Package everything into a tarball
    local release_name="MuseBot-${os}-${arch}.tar.gz"
    tar zcf "./release/${release_name}" -C ./output .

    echo "✅ Packaged ${release_name}"
}

# Platforms to compile
#compile_and_package linux amd64
#compile_and_package windows amd64
compile_and_package_local darwin amd64
compile_and_package_local darwin arm64

# Final cleanup
rm -rf ./output
echo "✅ Compilation and packaging complete. Output is in ./release"
