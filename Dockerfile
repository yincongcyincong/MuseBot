# ----------------------
# Build stage: compile Go application
# ----------------------
FROM golang:1.24 as builder

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the Go application
RUN go build -ldflags="-s -w" -o MuseBot main.go


# ----------------------
# FFmpeg download stage
# ----------------------
FROM debian:stable-slim as ffmpeg-builder

WORKDIR /tmp

# Install required tools (wget + xz-utils for tar.xz extraction)
RUN apt-get update && \
    apt-get install -y --no-install-recommends wget ca-certificates xz-utils && \
    rm -rf /var/lib/apt/lists/*

# Download and extract static FFmpeg binaries
RUN wget https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz && \
    mkdir -p /tmp/ffmpeg && \
    tar -xJf ffmpeg-release-amd64-static.tar.xz -C /tmp/ffmpeg && \
    cp /tmp/ffmpeg/ffmpeg-*-amd64-static/ffmpeg /usr/local/bin/ && \
    cp /tmp/ffmpeg/ffmpeg-*-amd64-static/ffprobe /usr/local/bin/ && \
    chmod +x /usr/local/bin/ffmpeg /usr/local/bin/ffprobe && \
    rm -rf /tmp/ffmpeg

# ----------------------
# Runtime stage
# ----------------------
FROM debian:stable-slim

WORKDIR /app

# Install certificates
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y --no-install-recommends nodejs && \
    rm -rf /var/lib/apt/lists/*

# Create required directories
RUN mkdir -p ./conf/i18n ./conf/mcp

# Copy compiled Go application
COPY --from=builder /app/MuseBot .
COPY --from=builder /app/conf/i18n/ ./conf/i18n/
COPY --from=builder /app/conf/mcp/ ./conf/mcp/

# Copy FFmpeg binaries
COPY --from=ffmpeg-builder /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg-builder /usr/local/bin/ffprobe /usr/local/bin/ffprobe

# Create non-root user for security
RUN useradd -m appuser && \
    chown -R appuser:appuser /app
USER appuser

# Expose application port
EXPOSE 36060

# Set entrypoint
ENTRYPOINT ["./MuseBot"]
