#!/bin/bash

if [ -z "$1" ]; then
  echo "❌ version is necessary: ./push.sh v1.0.0"
  exit 1
fi

VERSION=$1
IMAGE_NAME=musebot

# Docker Hub 配置
DOCKER_HUB_USER=jackyin0822
DOCKER_HUB_REPO=${DOCKER_HUB_USER}/${IMAGE_NAME}

PLATFORMS="linux/amd64,linux/arm64"

echo "🚀 create multi-platform image..."
docker buildx build \
  --platform ${PLATFORMS} \
  -t ${DOCKER_HUB_REPO}:${VERSION} \
  -t ${DOCKER_HUB_REPO}:latest \
  -t crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot:latest \
  --push .


#docker buildx imagetools create \
#  --tag crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com/jackyin0822/musebot:latest \
#  jackyin0822/musebot:latest

echo "✅ success"
