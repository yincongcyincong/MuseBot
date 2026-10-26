#!/bin/bash

set -e

# MuseBot release builder (merges the old release.sh + release_xgo.sh).
#
# darwin targets build locally with go (CGO works natively on macOS);
# linux/windows targets cross-compile through xgo (docker, handles CGO).
# Every target is packaged into ./release as MuseBot-<os>-<arch>.tar.gz.
#
# Override the target lists to build a subset, e.g.:
#   XGO_TARGETS="linux/amd64" LOCAL_TARGETS="" ./release.sh
# (an explicitly empty list skips that group)

XGO_TARGETS="${XGO_TARGETS-linux/amd64 windows/amd64}"
LOCAL_TARGETS="${LOCAL_TARGETS-darwin/amd64 darwin/arm64}"

# xgo always talks to /var/run/docker.sock, which doesn't exist for
# Docker Desktop / colima / OrbStack setups. Point it at the socket of
# the currently active docker context instead.
setup_docker_host() {
    if [ -z "$DOCKER_HOST" ] && [ ! -S /var/run/docker.sock ]; then
        detected_host=$(docker context inspect --format '{{.Endpoints.docker.Host}}' 2>/dev/null || true)
        if [ -n "$detected_host" ]; then
            echo "Using Docker socket from current context: $detected_host"
            export DOCKER_HOST="$detected_host"
        fi
    fi
}

ensure_xgo() {
    if ! command -v xgo &> /dev/null; then
        echo "Installing xgo..."
        go install src.techknowlogick.com/xgo@latest
    fi
}

# Copy the shared runtime payload into ./output and package the tarball
package_release() {
    local os=$1
    local arch=$2

    # Copy config files
    mkdir -p ./output/conf/
    cp -r ./conf/i18n ./output/conf/
    cp -r ./conf/mcp ./output/conf/
    cp -r ./conf/examples ./output/conf/
    # conf/img is a runtime directory (image_path default), create it empty;
    # conf/skills (skill_path default) was removed from the repo — same treatment
    mkdir -p ./output/conf/img
    mkdir -p ./output/conf/skills
    mkdir -p ./output/data/

    # Copy admin UI files
    mkdir -p ./output/adminui/
    cp -r ./admin/adminui/* ./output/adminui/

    # Package everything into a tarball
    local release_name="MuseBot-${os}-${arch}.tar.gz"
    tar zcf "./release/${release_name}" -C ./output .
    echo "✅ Packaged ${release_name}"

    # Reset the staging dir for the next target
    rm -rf ./output/*
}

compile_local() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    echo "=============================="
    echo "Building MuseBot + admin [$os/$arch] using go build..."
    echo "=============================="

    GOOS=$os GOARCH=$arch CGO_ENABLED=1 go build -o "./output/MuseBot${ext}" ./
    GOOS=$os GOARCH=$arch CGO_ENABLED=1 go build -o "./output/MuseBotAdmin${ext}" ./admin

    package_release $os $arch
}

compile_xgo() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    echo "=============================="
    echo "Building MuseBot + admin [$os/$arch] using xgo..."
    echo "=============================="

    # Build the main bot binary and the admin binary
    xgo -out MuseBot -targets="$os/$arch" .
    xgo -out MuseBotAdmin -targets="$os/$arch" ./admin

    # Move compiled binaries to output (xgo names include the os and arch)
    mv ./MuseBot-${os}* ./output/MuseBot${ext}
    mv ./MuseBotAdmin-${os}* ./output/MuseBotAdmin${ext}

    package_release $os $arch

    # Clean up xgo intermediate files (dropped under ./github.com/)
    rm -rf ./github.com
}

# Clean up old files
rm -rf ./output ./release
mkdir -p ./output ./release

for target in $LOCAL_TARGETS; do
    compile_local "${target%/*}" "${target#*/}"
done

if [ -n "$XGO_TARGETS" ]; then
    setup_docker_host
    ensure_xgo
    for target in $XGO_TARGETS; do
        compile_xgo "${target%/*}" "${target#*/}"
    done
fi

# Final cleanup
rm -rf ./output
echo "✅ Compilation and packaging complete. Output is in ./release"
