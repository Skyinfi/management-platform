#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

echo "==> Checking for updates..."

if command -v git &>/dev/null && [ -d .git ]; then
    echo "==> Pulling latest code..."
    git pull origin main || true
fi

echo "==> Rebuilding and restarting services..."
docker compose up -d --build

echo "==> Cleaning up old images..."
docker image prune -f

echo "==> Done!"
docker compose ps
