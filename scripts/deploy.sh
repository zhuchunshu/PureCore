#!/usr/bin/env bash
# ============================================================
# PureCore — One-Click Production Deployment Script
# 
# Deploys the full stack (PostgreSQL + Go backend + Bun SSR frontend)
# using Docker Compose. Supports custom ports, API connectivity,
# and theme configuration via environment variables.
#
# Usage:
#   chmod +x scripts/deploy.sh
#   ./scripts/deploy.sh              # Build & start
#   ./scripts/deploy.sh --build-only  # Build images only
#   ./scripts/deploy.sh --start-only  # Start existing containers
#   ./scripts/deploy.sh --down        # Stop all services
#   ./scripts/deploy.sh --status      # Show service status
#
# Configuration:
#   All settings are read from .env at the project root.
#   See .env.example for available options.
#
# Prerequisites:
#   - Docker Engine 24+ and Docker Compose v2
#   - .env file with production secrets
# ============================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log()  { echo -e "${BLUE}[PureCore]${NC} $1"; }
ok()   { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
err()  { echo -e "${RED}✗${NC} $1"; }

# ─── Load existing .env ────────────────────────────────────
load_existing_env() {
  if [ -f .env ]; then
    set -a
    source .env
    set +a
  fi
}

ask_port() {
  local prompt="$1"
  local default="$2"
  local var_name="$3"
  local value=""

  read -p "${prompt} [${default}]: " value
  value="${value:-$default}"

  if ! [[ "$value" =~ ^[0-9]+$ ]] || [ "$value" -lt 1 ] || [ "$value" -gt 65535 ]; then
    warn "Invalid port: $value. Using default: $default"
    value="$default"
  fi

  export "$var_name=$value"
  printf -v "$var_name" '%s' "$value"
}

# ─── Check prerequisites ───────────────────────────────────
check_prereqs() {
  if ! command -v docker &>/dev/null; then
    err "Docker is not installed. Please install Docker 24+ first."
    exit 1
  fi

  if ! docker compose version &>/dev/null; then
    err "Docker Compose v2 is not available. Please upgrade."
    exit 1
  fi

  load_existing_env

  if [ ! -f .env ]; then
    warn ".env file not found. Copying from .env.example..."
    if [ -f .env.example ]; then
      cp .env.example .env
      ok ".env created from .env.example — edit it with your production secrets"
      echo ""
      warn "IMPORTANT: Edit .env with your own passwords and secrets before proceeding!"
      echo "   DB_PASSWORD, JWT_SECRET must be changed from defaults"
      echo ""
      read -p "Press Enter to continue after editing .env, or Ctrl+C to abort... "
      load_existing_env
    else
      err ".env.example not found either. Please create .env with your configuration."
      exit 1
    fi
  fi

  # Interactive port configuration
  echo ""
  log "Port Configuration"
  log "─────────────────────"
  ask_port "Frontend HTTP port" "${FRONTEND_PORT:-9001}" "FRONTEND_PORT"
  ask_port "Backend API port" "${BACKEND_PORT:-9002}" "BACKEND_PORT"
  echo ""

  # Write the ports back to .env (add if missing, update if exists)
  if [ -f .env ]; then
    for var in FRONTEND_PORT BACKEND_PORT; do
      value="${!var}"
      if grep -q "^${var}=" .env 2>/dev/null; then
        sed -i "s/^${var}=.*/${var}=$value/" .env
      else
        echo "${var}=$value" >> .env
      fi
    done
  fi

  # Verify required variables are set with non-default values
  local required_vars=("DB_PASSWORD" "JWT_SECRET")
  local missing=()
  for var in "${required_vars[@]}"; do
    if ! grep -q "^${var}=" .env 2>/dev/null || \
       grep -q "^${var}=your_password_here$" .env 2>/dev/null || \
       grep -q "^${var}=your-jwt-secret-here$" .env 2>/dev/null; then
      missing+=("$var")
    fi
  done
  if [ ${#missing[@]} -gt 0 ]; then
    err "The following required variables are still using default values in .env:"
    for var in "${missing[@]}"; do
      echo "   - $var"
    done
    echo ""
    err "Please update them with secure values before deploying."
    exit 1
  fi

  ok "Prerequisites check passed"
}

# ─── Build ─────────────────────────────────────────────────
build_images() {
  log "Building Docker images..."
  docker compose build --pull
  ok "Docker images built successfully"
}

# ─── Start ─────────────────────────────────────────────────
start_services() {
  log "Starting PureCore services..."
  docker compose up -d

  echo ""
  log "Waiting for services to be healthy..."
  sleep 3

  local max_attempts=30
  local attempt=1
  while [ $attempt -le $max_attempts ]; do
    local backend_healthy=$(docker inspect --format='{{.State.Health.Status}}' purecore-backend 2>/dev/null || echo "unknown")
    local frontend_healthy=$(docker inspect --format='{{.State.Health.Status}}' purecore-frontend 2>/dev/null || echo "unknown")
    local db_healthy=$(docker inspect --format='{{.State.Health.Status}}' purecore-db 2>/dev/null || echo "unknown")

    if [ "$backend_healthy" = "healthy" ] && [ "$frontend_healthy" = "healthy" ] && [ "$db_healthy" = "healthy" ]; then
      ok "All services are healthy"
      break
    fi

    if [ "$attempt" -eq 1 ]; then
      echo "Waiting for services... (backend: $backend_healthy, frontend: $frontend_healthy, db: $db_healthy)"
    fi

    sleep 2
    attempt=$((attempt + 1))
  done

  if [ $attempt -gt $max_attempts ]; then
    warn "Services may still be starting. Check logs with: docker compose logs -f"
  fi

  local admin_path="${ADMIN_ROUTE_PREFIX:-control-panel}"
  local fe_port="${FRONTEND_PORT:-9001}"
  local be_port="${BACKEND_PORT:-9002}"

  echo ""
  log "PureCore is running!"
  echo "  Frontend  → http://localhost:${fe_port}"
  echo "  Backend   → http://localhost:${be_port} (API)"
  echo ""
  echo "  API connectivity:"
  echo "    Protocol: ${VITE_API_PROTOCOL:-http}"
  echo "    Host:     ${VITE_API_HOST:-backend}"
  echo "    Port:     ${VITE_API_PORT:-9002}"
  echo ""
  echo "  Theme: ${THEME:-sunset}"
  echo ""
  echo "  First-time setup:"
  echo "    Register an admin at http://localhost:${fe_port}/${admin_path}/register"
}

# ─── Stop ──────────────────────────────────────────────────
stop_services() {
  log "Stopping PureCore services..."
  docker compose down
  ok "All services stopped"
}

# ─── Status ────────────────────────────────────────────────
show_status() {
  echo ""
  docker compose ps
  echo ""
  log "Service URLs (from .env):"
  echo "  Frontend → http://localhost:${FRONTEND_PORT:-9001}"
  echo "  Backend  → http://localhost:${BACKEND_PORT:-9002}"
  echo ""
  log "API connectivity:"
  echo "  ${VITE_API_PROTOCOL:-http}://${VITE_API_HOST:-backend}:${VITE_API_PORT:-9002}"
}

# ─── Main ──────────────────────────────────────────────────
main() {
  case "${1:-}" in
    --build-only)
      check_prereqs
      build_images
      ;;
    --start-only)
      check_prereqs
      start_services
      ;;
    --down)
      stop_services
      ;;
    --status)
      show_status
      ;;
    *)
      # Default: full deployment
      check_prereqs
      build_images
      start_services
      ;;
  esac
}

main "$@"
