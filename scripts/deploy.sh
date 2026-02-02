#!/bin/bash
set -euo pipefail

# mohak.sh Deployment Script
# Usage: ./scripts/deploy.sh [environment]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ENV="${1:-production}"

echo "═══════════════════════════════════════════════"
echo "  mohak.sh Deployment - $ENV"
echo "═══════════════════════════════════════════════"

cd "$PROJECT_DIR"

# Setup .env if it doesn't exist
if [[ ! -f "$PROJECT_DIR/.env" ]]; then
    if [[ -f "$PROJECT_DIR/.env.example" ]]; then
        echo "▶ Creating .env from .env.example..."
        cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
        echo "⚠  Please edit .env with your API keys and run again"
        exit 1
    else
        echo "Error: .env.example not found"
        exit 1
    fi
fi

# Load environment
if [[ -f "$PROJECT_DIR/.env.$ENV" ]]; then
    set -a
    source "$PROJECT_DIR/.env.$ENV"
    set +a
elif [[ -f "$PROJECT_DIR/.env" ]]; then
    set -a
    source "$PROJECT_DIR/.env"
    set +a
fi

# Validate required variables
if [[ -z "${AI_GATEWAY_API_KEY:-}" ]] || [[ "${AI_GATEWAY_API_KEY}" == "your_api_key_here" ]]; then
    echo "Error: AI_GATEWAY_API_KEY is required"
    echo "Please edit .env with your API key"
    exit 1
fi

echo ""
echo "▶ Building images..."
docker compose build --no-cache

echo ""
echo "▶ Stopping existing containers..."
docker compose down --remove-orphans || true

echo ""
echo "▶ Starting services..."
docker compose up -d

echo ""
echo "▶ Waiting for services to be healthy..."
sleep 10

echo ""
echo "▶ Checking health..."
docker compose ps

# Health check
if curl -sf http://localhost:3001/health > /dev/null 2>&1; then
    echo "✓ AI Gateway is healthy"
else
    echo "✗ AI Gateway health check failed"
    docker compose logs ai-gateway --tail 50
    exit 1
fi

if nc -z localhost "${SSH_PORT:-2222}" 2>/dev/null; then
    echo "✓ TUI Server is healthy"
else
    echo "✗ TUI Server health check failed"
    docker compose logs tui-server --tail 50
    exit 1
fi

echo ""
echo "═══════════════════════════════════════════════"
echo "  Deployment Complete!"
echo "  SSH: ssh -p ${SSH_PORT:-2222} localhost"
echo "═══════════════════════════════════════════════"
