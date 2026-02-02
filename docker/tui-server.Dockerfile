# TUI Server Dockerfile
# Build context should be the repo root

# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY apps/tui-server/go.mod apps/tui-server/go.sum ./apps/tui-server/

# Download dependencies
WORKDIR /build/apps/tui-server
RUN go mod download

# Copy source code
WORKDIR /build
COPY apps/tui-server ./apps/tui-server
COPY packages/shared-content ./packages/shared-content

# Build binary
WORKDIR /build/apps/tui-server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/bin/tui-server .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata netcat-openbsd

# Create non-root user
RUN addgroup -g 1000 mohak && \
    adduser -u 1000 -G mohak -s /bin/sh -D mohak

# Create directories
RUN mkdir -p /app/.ssh /app/content && \
    chown -R mohak:mohak /app

# Copy binary
COPY --from=builder /build/bin/tui-server /app/tui-server

# Copy content
COPY --from=builder /build/packages/shared-content /app/content

# Set ownership
RUN chown -R mohak:mohak /app

USER mohak

# Environment
ENV SSH_HOST=0.0.0.0
ENV SSH_PORT=2222
ENV CONTENT_PATH=/app/content
ENV LOG_FORMAT=json

EXPOSE 2222

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD nc -z localhost 2222 || exit 1

CMD ["/app/tui-server"]
