# AI Gateway Dockerfile
# Build context should be the repo root

FROM oven/bun:1.1-alpine AS builder

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
FROM oven/bun:1.1-alpine

WORKDIR /app

# Create non-root user
RUN addgroup -g 1000 mohak && \
    adduser -u 1000 -G mohak -s /bin/sh -D mohak

# Copy built files and dependencies
COPY --from=builder /app/apps/ai-gateway/dist ./
COPY --from=builder /app/apps/ai-gateway/package.json ./
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/packages ./packages

# Set ownership
RUN chown -R mohak:mohak /app

USER mohak

# Environment
ENV AI_GATEWAY_PORT=3001
ENV NODE_ENV=production
ENV LOG_FORMAT=json

EXPOSE 3001

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

CMD ["bun", "run", "index.js"]
