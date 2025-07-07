#!/bin/bash

set -e

# 清理旧文件
rm -rf ./output ./release
mkdir -p ./output ./release

# 检查是否安装 xgo
if ! command -v xgo &> /dev/null; then
    echo "正在安装 xgo..."
    go install src.techknowlogick.com/xgo@latest
fi

# 编译 admin（本地平台）
build_admin_local() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    local admin_output="admin-${os}-${arch}${ext}"

    echo "=============================="
    echo "使用 go build 编译 admin [$os/$arch] ..."
    echo "=============================="
    xgo -out "$admin_output" -targets="$os/$arch" ./admin
}

# 编译主程序 + 打包
compile_and_package() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    echo "=============================="
    echo "使用 xgo 编译 telegram-deepseek-bot [$os/$arch] ..."
    echo "=============================="

    # 编译主程序
    xgo -out telegram-deepseek-bot -targets="$os/$arch" .

    local bot_binary="telegram-deepseek-bot-${os}-${arch}${ext}"
    local admin_binary="admin-${os}-${arch}${ext}"
    local release_name="telegram-deepseek-bot-${os}-${arch}.tar.gz"

    # 移动 bot 二进制
    mv ./telegram-deepseek-bot-${os}* ./output/${bot_binary}
    mv ./admin-${os}* ./output/${admin_binary}

    # 编译 admin (本地编译)
    build_admin_local $os $arch

    # 拷贝配置
    mkdir -p ./output/conf/
    cp -r ./conf/i18n ./output/conf/
    cp -r ./conf/mcp ./output/conf/

    # 拷贝 adminui
    mkdir -p ./output/adminui/
    cp -r ./admin/adminui/* ./output/adminui/

    # 打包
    tar zcf "release/${release_name}" -C ./output .

    # 清理中间产物
    rm -rf ./output/* ./github.com/*
}

# 编译平台（如需 Windows 可解开）
compile_and_package linux amd64
compile_and_package darwin amd64
compile_and_package darwin arm64
# compile_and_package windows amd64

# 清理
rm -rf ./output
echo "✅ 所有平台编译完成，打包输出在 ./release"
