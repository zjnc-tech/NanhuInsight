#!/bin/bash

# 检查参数数量
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

VERSION=$1

# 构建 Docker 镜像
echo "Building Docker image with tag: inspection-go:${VERSION}"
docker build -t inspection-go:"${VERSION}" -f Dockerfile .

# 标记 Docker 镜像
echo "Tagging Docker image..."
docker tag inspection-go:"${VERSION}" cr.registry.res.cloud.zhejianglab.com/infrahi-installer/inspection_manager:"${VERSION}"

# 推送 Docker 镜像
echo "Pushing Docker image..."
docker push cr.registry.res.cloud.zhejianglab.com/infrahi-installer/inspection_manager:"${VERSION}"

echo "Script completed successfully."