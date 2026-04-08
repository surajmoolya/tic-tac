#!/usr/bin/env bash
set -euo pipefail

FRONTEND_API_HOST="${FRONTEND_API_HOST:-127.0.0.1}"
FRONTEND_API_PORT="${FRONTEND_API_PORT:-7350}"
SKIP_FRONTEND_INSTALL="false"

for arg in "$@"; do
  case "$arg" in
    --skip-frontend-install)
      SKIP_FRONTEND_INSTALL="true"
      ;;
    --frontend-api-host=*)
      FRONTEND_API_HOST="${arg#*=}"
      ;;
    --frontend-api-port=*)
      FRONTEND_API_PORT="${arg#*=}"
      ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Building and starting backend containers (production-style)..."
docker compose up -d --build

cd "$SCRIPT_DIR/frontend"
if [[ "$SKIP_FRONTEND_INSTALL" != "true" ]]; then
  echo "Installing frontend dependencies..."
  npm install
fi

echo "Building frontend for production..."
VITE_NAKAMA_HOST="$FRONTEND_API_HOST" VITE_NAKAMA_PORT="$FRONTEND_API_PORT" npm run build

echo "Production build completed."
echo "Frontend artifacts: frontend/dist"
echo "Backend services are running in Docker."
