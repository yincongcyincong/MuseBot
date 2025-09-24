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

ALIYUN_REGISTRY=crpi-i1dsvpjijxpgjgbv.cn-hangzhou.personal.cr.aliyuncs.com
ALIYUN_NAMESPACE=haha03942007
ALIYUN_REPO=${ALIYUN_REGISTRY}/${ALIYUN_NAMESPACE}/${IMAGE_NAME}

echo "🚀 create image..."
docker build -t ${IMAGE_NAME}:latest .

echo "📦 push Docker Hub..."
docker tag ${IMAGE_NAME}:latest ${DOCKER_HUB_REPO}:${VERSION}
docker push ${DOCKER_HUB_REPO}:${VERSION}

docker tag ${IMAGE_NAME}:latest ${DOCKER_HUB_REPO}:latest
docker push ${DOCKER_HUB_REPO}:latest

echo "📦 push ali ACR..."
docker tag ${IMAGE_NAME}:latest ${ALIYUN_REPO}:${VERSION}
docker push ${ALIYUN_REPO}:${VERSION}

docker tag ${IMAGE_NAME}:latest ${ALIYUN_REPO}:latest
docker push ${ALIYUN_REPO}:latest

echo "✅ success"
