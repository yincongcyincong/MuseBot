#!/bin/bash
set -e

if [ -z "$1" ]; then
  echo "❌ version is necessary: ./push.sh v1.0.0"
  exit 1
fi

VERSION=$1
IMAGE_NAME=musebot

# Docker Hub 配置
DOCKER_HUB_USER=jackyin0822
DOCKER_HUB_REPO=${DOCKER_HUB_USER}/${IMAGE_NAME}

# 阿里云配置
ALIYUN_REPO=crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot

PLATFORMS="linux/amd64,linux/arm64"
# 本机网络限制(2026-10):buildkit 无法直连 docker.io(基础镜像需先 docker pull 到本地);
# daemon/代理链路只稳定支持下载,大体积上传必超时 → 构建后本地导出,用 skopeo(宿主机进程)推送
# 基础镜像双平台都 pull 到本地后,可用 BUILD_CONTEXT_ARGS 完全绕开 registry:
#   BUILD_CONTEXT_ARGS="--build-context golang:1.26=docker-image://golang:1.26 \
#     --build-context mwader/static-ffmpeg:7.1=docker-image://mwader/static-ffmpeg:7.1 \
#     --build-context debian:stable-slim=docker-image://debian:stable-slim" ./docker.sh v2.0.1
TAR=/tmp/musebot-${VERSION}.tar

echo "🚀 create multi-platform image (local load)..."
docker buildx build \
  --platform ${PLATFORMS} \
  --provenance=false --sbom=false \
  ${BUILD_CONTEXT_ARGS:-} \
  -t ${DOCKER_HUB_REPO}:${VERSION} \
  -t ${DOCKER_HUB_REPO}:latest \
  -t ${ALIYUN_REPO}:latest \
  --load .

echo "🚀 export image to ${TAR}..."
docker save -o ${TAR} ${DOCKER_HUB_REPO}:${VERSION}

# 凭证从 Docker Desktop 钥匙串读取,不落盘不回显
read_cred() {
  echo "$1" | docker-credential-desktop get | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['Username']+':'+d['Secret'])"
}

echo "🚀 push to Aliyun (host direct)..."
skopeo copy --all --dest-creds "$(read_cred crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com)" \
  oci-archive:${TAR} docker://${ALIYUN_REPO}:latest

echo "🚀 push to Docker Hub (via local proxy)..."
HTTPS_PROXY=http://127.0.0.1:7892 HTTP_PROXY=http://127.0.0.1:7892 \
  skopeo copy --all --dest-creds "$(read_cred https://index.docker.io/v1/)" \
  oci-archive:${TAR} docker://docker.io/${DOCKER_HUB_REPO}:${VERSION}

echo "🚀 tag latest on Docker Hub..."
HTTPS_PROXY=http://127.0.0.1:7892 \
  docker buildx imagetools create -t ${DOCKER_HUB_REPO}:latest ${DOCKER_HUB_REPO}:${VERSION}

echo "✅ success"
