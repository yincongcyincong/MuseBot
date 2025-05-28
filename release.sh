#!/bin/bash

# 清理旧文件
rm -rf ./output ./release
mkdir -p ./output ./release

# 检查是否安装xgo
if ! command -v xgo &> /dev/null; then
    echo "正在安装 xgo..."
    go install src.techknowlogick.com/xgo@latest
fi

# 编译函数
compile_and_package() {
    local os=$1
    local arch=$2
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"

    echo "正在编译 $os/$arch ..."

    # 使用xgo直接编译
    xgo -v -x \
        -targets="$os/$arch" \
        .

    # 打包
    local binary_name="telegram-deepseek-bot-${os}-${arch}${ext}"
    local release_name="telegram-deepseek-bot-${os}-${arch}.tar.gz"

    mv "./github.com/yincongcyincong/telegram-deepseek-bot-${os}"* "./output/$binary_name"
    tar zcfv "release/$release_name" -C ./output "$binary_name"
    rm -rf output github.com
}

# 开始编译
compile_and_package linux amd64
compile_and_package darwin amd64
compile_and_package darwin arm64
#compile_and_package windows amd64

# 清理临时文件
rm -rf ./output
echo "所有平台编译完成！结果保存在 ./release 目录"
