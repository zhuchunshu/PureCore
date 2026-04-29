#!/usr/bin/env bash
# ============================================================
# PureCore — Interactive Project Initialization Wizard
# 
# This script transforms a freshly cloned PureCore template into
# your own project. It will:
#   - Ask for your project name and rename the Go module
#   - Update all import paths across the entire codebase
#   - Generate secure random passwords for database and JWT
#   - Ask for custom ports, API connectivity, and theme
#   - Remove the original Git history and initialize a fresh repo
#   - Optionally run the complete Docker production deployment
#
# Usage:
#   chmod +x scripts/setup.sh
#   ./scripts/setup.sh
# ============================================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

log()   { echo -e "${BLUE}[Setup]${NC} $1"; }
ok()    { echo -e "${GREEN}✓${NC} $1"; }
warn()  { echo -e "${YELLOW}⚠${NC} $1"; }
err()   { echo -e "${RED}✗${NC} $1"; }
header() { echo -e "\n${BOLD}${BLUE}═══ $1 ═══${NC}\n"; }

# ─── Welcome ────────────────────────────────────────────────
clear
echo ""
echo -e "${BOLD}${BLUE}  ╔═══════════════════════════════════════╗"
echo -e "  ║                                       ║"
echo -e "  ║   PureCore — Project Setup Wizard     ║"
echo -e "  ║   Full-Stack Go + Vue 3 Framework      ║"
echo -e "  ║                                       ║"
echo -e "  ╚═══════════════════════════════════════╝${NC}"
echo ""
log "This wizard will set up your new project from the PureCore template."
log "You'll be asked a few questions, then everything will be configured automatically."
echo ""

# ─── Project Name ──────────────────────────────────────────
header "1. Project Name"
log "Your Go module name (e.g., github.com/yourusername/myapp)"
log "This will update go.mod and all import paths in the codebase."
echo ""
read -p "Project module name [github.com/username/myapp]: " project_name
project_name="${project_name:-github.com/username/myapp}"

# Validate project name
if ! [[ "$project_name" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
  err "Invalid project name: $project_name"
  exit 1
fi
ok "Project name: $project_name"

# ─── Ports ─────────────────────────────────────────────────
header "2. Port Configuration"
log "Choose ports for your services. Press Enter to accept defaults."
echo ""

ask_port() {
  local prompt="$1"
  local default="$2"
  local value=""
  read -p "${prompt} [${default}]: " value
  value="${value:-$default}"
  if ! [[ "$value" =~ ^[0-9]+$ ]] || [ "$value" -lt 1 ] || [ "$value" -gt 65535 ]; then
    warn "Invalid port, using default: $default"
    value="$default"
  fi
  echo "$value"
}

frontend_port=$(ask_port "Frontend HTTP port" "9001")
backend_port=$(ask_port "Backend API port" "9002")

# ─── API Connectivity ──────────────────────────────────────
header "3. API Connectivity (Frontend → Backend)"
log "Configure how the frontend connects to the backend API."
log "In Docker, the default host is 'backend' (Docker service name)."
log "For local development, use 'localhost' or '127.0.0.1'."
echo ""

read -p "API protocol [http]: " api_protocol
api_protocol="${api_protocol:-http}"

read -p "API host [localhost]: " api_host
api_host="${api_host:-localhost}"

api_port="$backend_port"  # Same as backend port by default

# ─── Database ──────────────────────────────────────────────
header "4. Database Configuration"
log "PostgreSQL connection settings."
echo ""

read -p "Database host [localhost]: " db_host
db_host="${db_host:-localhost}"

read -p "Database user [postgres]: " db_user
db_user="${db_user:-postgres}"

# Generate a secure random password
db_password=$(openssl rand -base64 24 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32)
log "Generated secure database password: ${db_password:0:8}... (full password saved to .env)"
echo ""

read -p "Database name [purecore]: " db_name
db_name="${db_name:-purecore}"

# ─── JWT Secret ────────────────────────────────────────────
jwt_secret=$(openssl rand -base64 48 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 64)
log "Generated secure JWT secret (saved to .env)"

# ─── Admin Route ───────────────────────────────────────────
header "5. Admin Configuration"
read -p "Admin route prefix [control-panel]: " admin_prefix
admin_prefix="${admin_prefix:-control-panel}"

# ─── Theme ─────────────────────────────────────────────────
header "6. Theme"
log "Default DaisyUI theme. Users can override this in the browser."
log "Available: sunset, cyberpunk, dracula, nord, emerald, corporate, synthwave, retro,"
log "           garden, forest, aqua, pastel, cupcake, bumblebee, fantasy, valentine,"
log "           halloween, black, luxury, cmyk, autumn, business, acid, lemonade,"
log "           night, coffee, winter, dim, wireframe, lofi, caramellatte, abyss, silk"
echo ""
read -p "Default theme [sunset]: " theme
theme="${theme:-sunset}"

# ─── Summary ───────────────────────────────────────────────
header "7. Summary"
echo -e "  ${BOLD}Project:${NC}     $project_name"
echo -e "  ${BOLD}Frontend:${NC}    http://localhost:$frontend_port"
echo -e "  ${BOLD}Backend:${NC}     http://localhost:$backend_port"
echo -e "  ${BOLD}API:${NC}         $api_protocol://$api_host:$api_port"
echo -e "  ${BOLD}Database:${NC}    $db_host/$db_name (user: $db_user, internal only)"
echo -e "  ${BOLD}Admin:${NC}       /$admin_prefix"
echo -e "  ${BOLD}Theme:${NC}       $theme"
echo ""
read -p "Proceed with setup? [Y/n]: " confirm
if [[ ! "$confirm" =~ ^[Yy]?$ ]]; then
  log "Setup cancelled."
  exit 0
fi

# ─── Apply Configuration ───────────────────────────────────
header "8. Applying Configuration"

# Step 1: Create .env from .env.example 
log "Creating .env file..."
cat > .env << EOF
# $project_name Configuration
# Frontend
FRONTEND_PORT=$frontend_port

# Backend
BACKEND_PORT=$backend_port

# API 协议和主机 (前端连接后端使用)
VITE_API_PROTOCOL=$api_protocol
VITE_API_HOST=$api_host
VITE_API_PORT=$api_port

# Database
DB_HOST=$db_host
DB_USER=$db_user
DB_PASSWORD=$db_password
DB_NAME=$db_name
DB_SSLMODE=disable

# App
APP_ENV=local
APP_DEBUG=true
TZ=Asia/Shanghai

# Theme
# 优先级: 用户手动选择 > 环境变量 THEME / VITE_THEME > theme.config.json > 'sunset'
THEME=$theme
VITE_THEME=$theme

# Admin
ADMIN_ROUTE_PREFIX=$admin_prefix
VITE_ADMIN_ROUTE_PREFIX=$admin_prefix
JWT_SECRET=$jwt_secret
EOF
ok ".env created with your configuration"

# Step 2: Rename Go module
log "Updating Go module from 'purecore' to '$project_name'..."
if command -v sd &>/dev/null; then
  # sd is faster and handles edge cases better
  find . -type f \( -name "*.go" -o -name "go.mod" \) -not -path "./.git/*" -exec sd 'purecore' "$project_name" {} \;
else
  # Fallback to sed
  find . -type f \( -name "*.go" -o -name "go.mod" \) -not -path "./.git/*" -exec sed -i "s|purecore|$project_name|g" {} \;
fi
ok "Go module renamed to $project_name"

# Update Go module name in go.mod explicitly
if [ -f go.mod ]; then
  sed -i "s|^module .*|module $project_name|" go.mod
fi

# Step 3: Reinitialize Git
log "Initializing fresh Git repository..."
if [ -d .git ]; then
  rm -rf .git
fi
git init
git add -A
git commit -m "Initial commit from PureCore template

Project: $project_name
Frontend: http://localhost:$frontend_port
Backend: http://localhost:$backend_port
API: $api_protocol://$api_host:$api_port
Database: $db_host/$db_name (internal only)
Theme: $theme"
ok "Fresh Git repository initialized"

# Step 4: Download Go dependencies
log "Downloading Go dependencies..."
if command -v go &>/dev/null; then
  go mod tidy
  ok "Go dependencies downloaded"
else
  warn "Go not found — skipping dependency download. Run 'go mod tidy' manually."
fi

# Step 5: Build the project
log "Building the project..."
if command -v go &>/dev/null; then
  go build -o "$(basename "$project_name")" . 2>&1 || warn "Build failed — this is OK if you're not ready to compile yet"
  if [ -f "$(basename "$project_name")" ]; then
    ok "Project built successfully: $(basename "$project_name")"
  fi
fi

# ─── Done ──────────────────────────────────────────────────
header "✓ Setup Complete!"
echo -e "  ${BOLD}Your project is ready!${NC}"
echo ""
echo "  Quick start:"
echo "    cd $(basename "$(pwd)")"
echo "    # Edit .env to adjust settings if needed"
echo "    ./$(basename "$project_name") serve     # Start backend"
echo "    cd web && bun install && bun run dev   # Start frontend"
echo ""
echo "  One-click Docker deployment:"
echo "    chmod +x scripts/deploy.sh && ./scripts/deploy.sh"
echo ""

# Optionally offer to start Docker deployment
read -p "Would you like to deploy with Docker Compose now? [y/N]: " deploy_now
if [[ "$deploy_now" =~ ^[Yy]$ ]]; then
  log "Starting Docker deployment..."
  chmod +x scripts/deploy.sh
  ./scripts/deploy.sh
fi
