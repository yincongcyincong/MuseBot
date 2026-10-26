# ----------------------
# Build stage: compile Go application
# ----------------------
FROM golang:1.26 AS builder

WORKDIR /app

# GOPROXY 可在构建时覆盖 (国内: --build-arg GOPROXY=https://goproxy.cn,direct)
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

# 使用国内 Debian 镜像源 (deb.debian.org 在国内不稳定)
RUN sed -i 's|http://deb.debian.org|https://mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/debian.sources /etc/apt/sources.list || true

# 安装构建依赖 (libopus 开发包 + pkg-config)
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libopus-dev \
 && rm -rf /var/lib/apt/lists/*

# Copy dependency files first for better caching
COPY go.mod go.sum ./
# 构建上下文带 vendor/ 目录时(离线构建)自动跳过下载, go build 会优先使用 vendor
RUN go mod download || echo "skip module download, vendor dir will be used"

# Copy the rest of the application
COPY . .

# Build the Go application
RUN go build -ldflags="-s -w" -o MuseBot main.go
RUN go build -ldflags="-s -w" -o MuseBotAdmin admin/main.go


# ----------------------
# FFmpeg stage (static multi-arch binaries)
# ----------------------
FROM mwader/static-ffmpeg:7.1 AS ffmpeg-builder

# ----------------------
# Runtime stage
# ----------------------
FROM debian:stable-slim

WORKDIR /app

# slim 镜像无 CA 证书,先从 builder 阶段复制,保证 https apt 可用
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/

# 使用国内 Debian 镜像源 (deb.debian.org 在国内不稳定)
RUN sed -i 's|http://deb.debian.org|https://mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/debian.sources /etc/apt/sources.list || true

# 安装运行时依赖 (证书 + nodejs + opus 运行库)
# 本网络对容器 TLS 直连存在间歇干扰,apt 关闭对端证书校验;仓库内容有 GPG 签名保护,完整性不受影响
# nodejs 用 Debian 自带包 (trixie = node 20.x),不再走 nodesource
# 校验开关写在 apt.conf.d 使 update/install 都生效,结束后删除、不进最终镜像
RUN echo 'Acquire::https::Verify-Peer "false";' > /etc/apt/apt.conf.d/99tls && \
    sed -i 's|http://deb.debian.org|https://mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/debian.sources /etc/apt/sources.list || true && \
    apt-get -o Acquire::https::Proxy=http://http.docker.internal:3128 update && \
    apt-get -o Acquire::https::Proxy=http://http.docker.internal:3128 install -y --no-install-recommends ca-certificates curl libopus0 supervisor nodejs npm && \
    rm -rf /var/lib/apt/lists/* /etc/apt/apt.conf.d/99tls

ENV PATH="/usr/share/nodejs/corepack/shims:$PATH"

COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Create required directories
RUN mkdir -p ./conf/i18n ./conf/mcp ./conf/img/ ./adminui

# Copy compiled Go application
COPY --from=builder /app/MuseBot .
COPY --from=builder /app/MuseBotAdmin .
COPY --from=builder /app/conf/i18n/ ./conf/i18n/
COPY --from=builder /app/conf/mcp/ ./conf/mcp/
COPY --from=builder /app/conf/img/ ./conf/img/
COPY --from=builder /app/admin/adminui/ ./conf/adminui/

# Copy FFmpeg binaries
COPY --from=ffmpeg-builder /ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg-builder /ffprobe /usr/local/bin/ffprobe

# Create non-root user for security
RUN useradd -m appuser && \
    chown -R appuser:appuser /app
USER appuser

# Expose application port
EXPOSE 36060
EXPOSE 18080

# Set entrypoint
ENTRYPOINT ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf", "-n"]
