#!/usr/bin/env bash
set -euo pipefail

SKIP_FRONTEND_INSTALL="false"
if [[ "${1:-}" == "--skip-frontend-install" ]]; then
  SKIP_FRONTEND_INSTALL="true"
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Starting local backend (Docker)..."
docker compose up -d --build

echo "Preparing local frontend..."
cd "$SCRIPT_DIR/frontend"

if [[ "$SKIP_FRONTEND_INSTALL" != "true" ]]; then
  npm install
fi

echo "Launching frontend dev server on http://localhost:5173 ..."
npm run dev
