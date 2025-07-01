# Build stage
FROM golang:1.24 as builder

WORKDIR /app

# First copy only dependency files for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o telegram-deepseek-bot main.go

# Runtime stage
FROM debian:stable-slim

# Install certificates
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create necessary directories
RUN mkdir -p ./conf/i18n ./conf/mcp

# Copy only necessary files from builder
COPY --from=builder /app/telegram-deepseek-bot .
COPY --from=builder /app/conf/i18n/ ./conf/i18n/
COPY --from=builder /app/conf/mcp/ ./conf/mcp/

# (Optional) Create non-root user for security
RUN useradd -m appuser && \
    chown -R appuser:appuser /app
USER appuser

# Runtime command
ENTRYPOINT ["./telegram-deepseek-bot"]