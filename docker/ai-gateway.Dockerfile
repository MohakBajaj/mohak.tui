# AI Gateway Dockerfile
# Build context should be the repo root

FROM oven/bun:1.3-alpine AS builder

WORKDIR /app

# Copy package files for dependency resolution
COPY package.json bun.lockb* ./
COPY apps/ai-gateway/package.json ./apps/ai-gateway/
COPY packages/shared-content/package.json ./packages/shared-content/

# Install all dependencies
RUN bun install --frozen-lockfile

# Copy source files
COPY apps/ai-gateway ./apps/ai-gateway
COPY packages/shared-content ./packages/shared-content

# Build (optional - Bun can run TypeScript directly)
WORKDIR /app/apps/ai-gateway
RUN bun build index.ts --outdir=./dist --target=bun

# Runtime stage
FROM oven/bun:1.3-alpine

WORKDIR /app

# Create non-root user (use different GID/UID to avoid conflicts)
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# Copy built files and dependencies
COPY --from=builder /app/apps/ai-gateway/dist ./
COPY --from=builder /app/apps/ai-gateway/package.json ./
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/packages ./packages

# Set ownership
RUN chown -R appuser:appgroup /app

USER appuser

# Environment
ENV AI_GATEWAY_PORT=3001
ENV NODE_ENV=production
ENV LOG_FORMAT=json
ENV CONTENT_PATH=/app/packages/shared-content

EXPOSE 3001

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

CMD ["bun", "run", "index.js"]
