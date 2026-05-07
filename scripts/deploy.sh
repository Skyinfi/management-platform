#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "==> Deploying App Management Platform to Linux Server"
echo "    Project: $PROJECT_DIR"

cd "$PROJECT_DIR"

if [ ! -f .env ]; then
    echo "==> Creating .env from .env.example"
    cp .env.example .env
    echo "    ⚠ Please edit .env and set JWT_SECRET and ADMIN_PASSWORD"
    echo "    Then run this script again."
    exit 1
fi

echo "==> Building and starting services..."
docker compose up -d --build

echo ""
echo "==> Waiting for services to be healthy..."
sleep 3

if docker compose ps | grep -q "running"; then
    echo "==> ✅ Deployment successful!"
    echo ""
    echo "    Frontend: http://$(hostname -I | awk '{print $1}')"
    echo "    Backend:  http://$(hostname -I | awk '{print $1}'):8080"
    echo "    Health:   http://$(hostname -I | awk '{print $1}'):8080/api/health"
    echo ""
    echo "    Username: admin"
    echo "    Password: (see .env ADMIN_PASSWORD)"
else
    echo "==> ❌ Some services failed to start. Check logs:"
    echo "    docker compose logs"
    exit 1
fi
