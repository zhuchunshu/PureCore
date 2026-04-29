#!/usr/bin/env bash
# ============================================================
#  ██████╗ ██╗   ██╗██████╗ ███████╗ ██████╗ ██████╗ ██████╗ ███████╗
#  ██╔══██╗██║   ██║██╔══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝
#  ██████╔╝██║   ██║██████╔╝█████╗  ██║     ██║   ██║██████╔╝█████╗
#  ██╔═══╝ ██║   ██║██╔══██╗██╔══╝  ██║     ██║   ██║██╔══██╗██╔══╝
#  ██║     ╚██████╔╝██║  ██║███████╗╚██████╗╚██████╔╝██║  ██║███████╗
#  ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝
#
#  PureCore — One-Click Production Deployment
#  PureCore — 一键生产环境部署
#
#  This script automates the entire deployment pipeline:
#  此脚本自动化整个部署流程：
#    1. Docker environment detection & installation
#    2. ghcr.io authentication
#    3. Pull images & start the full stack
#
#  Usage / 用法:
#    ./scripts/deploy.sh              # Full interactive deployment
#    ./scripts/deploy.sh --down       # Stop all services
#    ./scripts/deploy.sh --status     # View service health
#    ./scripts/deploy.sh --lang zh    # Force Chinese
#    ./scripts/deploy.sh --lang en    # Force English
# ============================================================

set -euo pipefail

# ─── Global state ──────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_DIR/.env"
ENV_EXAMPLE="$PROJECT_DIR/.env.example"
COMPOSE_FILE="$PROJECT_DIR/docker-compose.yml"
LANG="auto"          # auto / en / zh
DOCKER_COMPOSE_CMD="" # will be set after detection
DOCKER_INSTALLED=false

# ─── Color palette (cyber / tech aesthetic) ────────────────
C_RESET='\033[0m'
C_BOLD='\033[1m'
C_DIM='\033[2m'
C_CYAN='\033[0;36m'
C_GREEN='\033[0;32m'
C_YELLOW='\033[1;33m'
C_RED='\033[0;31m'
C_MAGENTA='\033[0;35m'
C_BLUE='\033[0;34m'
C_BG_BLACK='\033[40m'
C_BRIGHT_CYAN='\033[1;36m'
C_BRIGHT_GREEN='\033[1;32m'
C_BRIGHT_RED='\033[1;31m'

# ─── Multi-language message catalog ────────────────────────
# Usage: t "key"  →  outputs translated string based on $LANG
declare -A MSG_EN MSG_ZH

# --- English ---
MSG_EN[title]="PureCore Deployment"
MSG_EN[title_sub]="One-Click Production Stack"
MSG_EN[divider]="━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
MSG_EN[step_detect]="Detecting environment..."
MSG_EN[step_config]="Configuring deployment..."
MSG_EN[step_pull]="Pulling images from registry..."
MSG_EN[step_start]="Starting services..."
MSG_EN[step_done]="Deployment complete!"
MSG_EN[docker_not_found]="Docker is not installed on this system."
MSG_EN[ask_install_docker]="Would you like to install Docker now? [y/N]: "
MSG_EN[installing_docker]="Installing Docker via official script..."
MSG_EN[docker_ok]="Docker is ready."
MSG_EN[docker_compose_v2]="Docker Compose v2 (plugin) detected."
MSG_EN[docker_compose_v1]="Docker Compose v1 (standalone) detected."
MSG_EN[docker_compose_none]="Docker Compose not found. Installing Docker will include it."
MSG_EN[checking_env]="Checking environment configuration..."
MSG_EN[env_missing]=".env file not found. Creating from .env.example..."
MSG_EN[env_created]=".env file created. Please review and edit it if needed."
MSG_EN[env_press_enter]="Press Enter to continue after reviewing .env..."
MSG_EN[port_config]="Port Configuration"
MSG_EN[ask_frontend_port]="Frontend HTTP port"
MSG_EN[ask_backend_port]="Backend API port"
MSG_EN[port_sync_warn]="WARNING: VITE_API_PORT must equal BACKEND_PORT for the app to work."
MSG_EN[port_sync_fixed]="VITE_API_PORT has been synced to BACKEND_PORT (%s)."
MSG_EN[secret_warn]="The following settings still use default values. Please update them:"
MSG_EN[ghcr_login]="Authenticating with GitHub Container Registry..."
MSG_EN[ghcr_ok]="ghcr.io accessible."
MSG_EN[ghcr_fail]="Cannot access ghcr.io. Log in manually:"
MSG_EN[ghcr_cmd_hint]="  echo \$CR_PAT | docker login ghcr.io -u YOUR_USERNAME --password-stdin"
MSG_EN[pulling]="Pulling PureCore images..."
MSG_EN[pull_done]="Images pulled successfully."
MSG_EN[starting]="Starting PureCore stack..."
MSG_EN[wait_healthy]="Waiting for services to become healthy..."
MSG_EN[service_status]="Service status:"
MSG_EN[all_healthy]="All services are healthy ✓"
MSG_EN[still_starting]="Services are still starting up. Check logs with:"
MSG_EN[health_timeout]="Health check timed out. Services may still be initializing."
MSG_EN[compose_up_failed]="docker compose up returned an error."
MSG_EN[showing_container_logs]="Showing logs for the failing container:"
MSG_EN[container_unhealthy]="Container '%s' is unhealthy. Logs:"
MSG_EN[url_frontend]="Frontend"
MSG_EN[url_backend]="Backend (API)"
MSG_EN[url_admin]="Admin panel"
MSG_EN[first_time]="First-time setup: register an admin at"
MSG_EN[stopping]="Stopping all services..."
MSG_EN[stopped]="All services stopped."
MSG_EN[status_title]="PureCore Service Status"
MSG_EN[error_docker_install]="Docker installation failed or was cancelled."
MSG_EN[error_compose]="Docker Compose command failed."
MSG_EN[error_original_output]="Original error output:"
MSG_EN[lang_selected]="Language: English"
MSG_EN[select_lang]="Select Language / 选择语言"
MSG_EN[lang_option_en]="English"
MSG_EN[lang_option_zh]="中文 (Chinese)"
MSG_EN[version_config]="Version Configuration"
MSG_EN[ask_version]="Image version to deploy"
MSG_EN[deploy_aborted]="Deployment aborted."
MSG_EN[compose_cmd_not_found]="Could not find docker compose or docker-compose command."
MSG_EN[install_success]="Docker installed successfully! You may need to log out and back in."
MSG_EN[press_enter_continue]="Press Enter to continue..."

# --- 中文 ---
MSG_ZH[title]="PureCore 部署工具"
MSG_ZH[title_sub]="一键生产环境部署"
MSG_ZH[divider]="━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
MSG_ZH[step_detect]="正在检测环境..."
MSG_ZH[step_config]="正在配置部署参数..."
MSG_ZH[step_pull]="正在从镜像仓库拉取..."
MSG_ZH[step_start]="正在启动服务..."
MSG_ZH[step_done]="部署完成！"
MSG_ZH[docker_not_found]="当前系统未安装 Docker。"
MSG_ZH[ask_install_docker]="是否现在安装 Docker？[y/N]: "
MSG_ZH[installing_docker]="正在通过官方脚本安装 Docker..."
MSG_ZH[docker_ok]="Docker 已就绪。"
MSG_ZH[docker_compose_v2]="检测到 Docker Compose v2 (插件模式)。"
MSG_ZH[docker_compose_v1]="检测到 Docker Compose v1 (独立模式)。"
MSG_ZH[docker_compose_none]="未检测到 Docker Compose，安装 Docker 时将一并安装。"
MSG_ZH[checking_env]="正在检查环境配置..."
MSG_ZH[env_missing]="未找到 .env 文件，正在从 .env.example 创建..."
MSG_ZH[env_created]="已创建 .env 文件，请检查并根据需要修改。"
MSG_ZH[env_press_enter]="检查完 .env 后按 Enter 继续..."
MSG_ZH[port_config]="端口配置"
MSG_ZH[ask_frontend_port]="前端 HTTP 端口"
MSG_ZH[ask_backend_port]="后端 API 端口"
MSG_ZH[port_sync_warn]="⚠ 注意：VITE_API_PORT 必须与 BACKEND_PORT 保持一致，否则程序将无法运行。"
MSG_ZH[port_sync_fixed]="已将 VITE_API_PORT 同步为 BACKEND_PORT (%s)。"
MSG_ZH[secret_warn]="以下配置仍在使用默认值，请务必修改："
MSG_ZH[ghcr_login]="正在验证 GitHub Container Registry 访问..."
MSG_ZH[ghcr_ok]="ghcr.io 可正常访问。"
MSG_ZH[ghcr_fail]="无法访问 ghcr.io，请手动登录："
MSG_ZH[ghcr_cmd_hint]="  echo \$CR_PAT | docker login ghcr.io -u 你的用户名 --password-stdin"
MSG_ZH[pulling]="正在拉取 PureCore 镜像..."
MSG_ZH[pull_done]="镜像拉取成功。"
MSG_ZH[starting]="正在启动 PureCore 服务栈..."
MSG_ZH[wait_healthy]="等待服务健康检查通过..."
MSG_ZH[service_status]="服务状态："
MSG_ZH[all_healthy]="所有服务已健康运行 ✓"
MSG_ZH[still_starting]="服务仍在启动中，查看日志："
MSG_ZH[health_timeout]="健康检查超时，服务可能仍在初始化中。"
MSG_ZH[compose_up_failed]="docker compose up 执行失败。"
MSG_ZH[showing_container_logs]="以下为异常容器的日志："
MSG_ZH[container_unhealthy]="容器 '%s' 处于不健康状态，日志如下："
MSG_ZH[url_frontend]="前端页面"
MSG_ZH[url_backend]="后端 API"
MSG_ZH[url_admin]="管理后台"
MSG_ZH[first_time]="首次使用请访问以下地址注册管理员："
MSG_ZH[stopping]="正在停止所有服务..."
MSG_ZH[stopped]="所有服务已停止。"
MSG_ZH[status_title]="PureCore 服务状态"
MSG_ZH[error_docker_install]="Docker 安装失败或被取消。"
MSG_ZH[error_compose]="Docker Compose 命令执行失败。"
MSG_ZH[error_original_output]="原始错误输出："
MSG_ZH[lang_selected]="语言：中文"
MSG_ZH[select_lang]="Select Language / 选择语言"
MSG_ZH[lang_option_en]="English"
MSG_ZH[lang_option_zh]="中文 (Chinese)"
MSG_ZH[version_config]="版本配置"
MSG_ZH[ask_version]="要部署的镜像版本"
MSG_ZH[deploy_aborted]="部署已取消。"
MSG_ZH[compose_cmd_not_found]="无法找到 docker compose 或 docker-compose 命令。"
MSG_ZH[install_success]="Docker 安装成功！你可能需要重新登录终端使其生效。"
MSG_ZH[press_enter_continue]="按 Enter 继续..."

# ─── Helper: translate ─────────────────────────────────────
t() {
  local key="$1"
  local msg=""
  if [ "$LANG" = "zh" ]; then
    msg="${MSG_ZH[$key]:-${MSG_EN[$key]:-$key}}"
  else
    msg="${MSG_EN[$key]:-$key}"
  fi
  # Replace %s placeholders with additional args
  shift 2>/dev/null || true
  if [ $# -gt 0 ]; then
    printf "$msg" "$@"
  else
    echo "$msg"
  fi
}

# ─── UI helpers ─────────────────────────────────────────────
header() {
  clear 2>/dev/null || true
  echo -e "${C_BRIGHT_CYAN}${C_BOLD}"
  echo "  ██████╗ ██╗   ██╗██████╗ ███████╗ ██████╗ ██████╗ ██████╗ ███████╗ "
  echo "  ██╔══██╗██║   ██║██╔══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝ "
  echo "  ██████╔╝██║   ██║██████╔╝█████╗  ██║     ██║   ██║██████╔╝█████╗   "
  echo "  ██╔═══╝ ██║   ██║██╔══██╗██╔══╝  ██║     ██║   ██║██╔══██╗██╔══╝   "
  echo "  ██║     ╚██████╔╝██║  ██║███████╗╚██████╗╚██████╔╝██║  ██║███████╗ "
  echo "  ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝ "
  echo -e "${C_RESET}"
  echo -e "  ${C_MAGENTA}$(t title)${C_RESET} ${C_DIM}— $(t title_sub)${C_RESET}"
  echo -e "${C_DIM}  $(t divider)${C_RESET}"
  echo ""
}

step_banner() {
  echo ""
  echo -e " ${C_BRIGHT_CYAN}▸${C_RESET} ${C_BOLD}$1${C_RESET}"
  echo -e " ${C_DIM}────────────────────────────────────────────${C_RESET}"
}

ok_msg() {
  echo -e "   ${C_BRIGHT_GREEN}✓${C_RESET} $1"
}

info_msg() {
  echo -e "   ${C_CYAN}›${C_RESET} $1"
}

warn_msg() {
  echo -e "   ${C_YELLOW}⚠${C_RESET} $1"
}

err_msg() {
  echo -e "   ${C_BRIGHT_RED}✗${C_RESET} $1"
}

# ─── Interactive language selection ─────────────────────────
select_language() {
  echo ""
  echo -e "  ${C_CYAN}┌─ $(t select_lang) ────────────────────────────┐${C_RESET}"
  echo -e "  ${C_CYAN}│${C_RESET}   1) English"
  echo -e "  ${C_CYAN}│${C_RESET}   2) 中文 (Chinese)"
  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  local choice
  while true; do
    printf "  Enter choice / 请输入选项 [1-2]: "
    read -r choice
    case "$choice" in
      1) LANG="en"; break ;;
      2) LANG="zh"; break ;;
      *) echo -e "  ${C_YELLOW}⚠${C_RESET} Invalid choice. Please enter 1 or 2." ;;
    esac
  done
  echo ""
  ok_msg "$(t lang_selected)"
}

# ─── Detect language from system locale ─────────────────────
detect_lang() {
  if [ "$LANG_SELECTED" != "auto" ]; then
    return
  fi
  local locale="${LANG:-en_US.UTF-8}"
  if echo "$locale" | grep -qiE '^zh|zh_CN|zh_TW'; then
    LANG="zh"
  else
    LANG="en"
  fi
}

# ─── Parse command-line arguments ───────────────────────────
parse_args() {
  LANG_SELECTED="auto"
  local args=()
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --lang)
        if [ -n "${2:-}" ]; then
          LANG_SELECTED="$2"
          shift 2
        else
          shift
        fi
        ;;
      *)
        args+=("$1")
        shift
        ;;
    esac
  done

  if [ "$LANG_SELECTED" != "auto" ]; then
    LANG="$LANG_SELECTED"
  fi
  CMD_ARG="${args[0]:-}"
}

# ─── Detect Docker and docker compose command ───────────────
detect_docker() {
  step_banner "$(t step_detect)"

  # Check Docker daemon
  if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
    DOCKER_INSTALLED=true
    ok_msg "$(t docker_ok)"
  else
    DOCKER_INSTALLED=false
    err_msg "$(t docker_not_found)"
  fi

  # Find docker compose command (v2 plugin first, then v1 standalone)
  if docker compose version &>/dev/null 2>&1; then
    DOCKER_COMPOSE_CMD="docker compose"
    info_msg "$(t docker_compose_v2)"
  elif command -v docker-compose &>/dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
    info_msg "$(t docker_compose_v1)"
  else
    DOCKER_COMPOSE_CMD=""
    info_msg "$(t docker_compose_none)"
  fi
}

# ─── Install Docker if missing ──────────────────────────────
install_docker() {
  echo ""
  read -rp "  $(t ask_install_docker)" answer
  if [ "${answer,,}" != "y" ] && [ "${answer,,}" != "yes" ]; then
    err_msg "$(t error_docker_install)"
    exit 1
  fi

  info_msg "$(t installing_docker)"
  if command -v curl &>/dev/null; then
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    if sudo sh /tmp/get-docker.sh; then
      rm -f /tmp/get-docker.sh
      # Start docker service if systemd is available
      if command -v systemctl &>/dev/null; then
        sudo systemctl enable docker 2>/dev/null || true
        sudo systemctl start docker 2>/dev/null || true
      fi
      # Add current user to docker group
      sudo usermod -aG docker "$(whoami)" 2>/dev/null || true
      ok_msg "$(t install_success)"
      # Reload group membership for this session if possible
      if command -v newgrp &>/dev/null; then
        # We can't newgrp inside a script easily; just warn
        info_msg "You may need to re-login or run: newgrp docker"
      fi
      DOCKER_INSTALLED=true
    else
      err_msg "$(t error_docker_install)"
      exit 1
    fi
  else
    err_msg "curl is required to install Docker. Please install curl first."
    exit 1
  fi

  # Re-detect compose command
  if docker compose version &>/dev/null 2>&1; then
    DOCKER_COMPOSE_CMD="docker compose"
  elif command -v docker-compose &>/dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
  fi
}

# ─── Ensure docker and compose are available ────────────────
ensure_docker() {
  detect_docker

  if [ "$DOCKER_INSTALLED" = false ]; then
    install_docker
  fi

  if [ -z "$DOCKER_COMPOSE_CMD" ]; then
    err_msg "$(t compose_cmd_not_found)"
    exit 1
  fi
}

# ─── Run docker compose with error capture ─────────────────
run_compose() {
  # Executes docker compose command, captures output.
  # If it fails, prints the original output and exits.
  local output
  local rc=0
  output=$($DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" "$@" 2>&1) || rc=$?
  if [ $rc -ne 0 ]; then
    echo ""
    err_msg "$(t error_compose)"
    echo ""
    echo -e "  ${C_RED}$(t error_original_output)${C_RESET}"
    echo -e "  ${C_DIM}────────────────────────────────────────────${C_RESET}"
    echo "$output" | while IFS= read -r line; do
      echo -e "  ${C_RED}│${C_RESET} $line"
    done
    echo -e "  ${C_DIM}────────────────────────────────────────────${C_RESET}"
    echo ""
    exit 1
  fi
  echo "$output"
}

# ─── Configure .env interactively ──────────────────────────
configure_env() {
  step_banner "$(t step_config)"

  info_msg "$(t checking_env)"

  cd "$PROJECT_DIR"

  # Create .env if missing
  if [ ! -f .env ]; then
    warn_msg "$(t env_missing)"
    if [ -f .env.example ]; then
      cp .env.example .env
      ok_msg "$(t env_created)"
    else
      err_msg ".env.example not found. Cannot proceed."
      exit 1
    fi
  fi

  # Load current .env
  set -a
  source .env
  set +a

  # --- Port configuration ---
  echo ""
  echo -e "  ${C_CYAN}┌─ $(t port_config) ──────────────────────────┐${C_RESET}"

  local fe_default="${FRONTEND_PORT:-9001}"
  local be_default="${BACKEND_PORT:-9002}"
  local input

  printf "  ${C_CYAN}│${C_RESET} $(t ask_frontend_port) [${fe_default}]: "
  read -r input
  FRONTEND_PORT="${input:-$fe_default}"

  printf "  ${C_CYAN}│${C_RESET} $(t ask_backend_port) [${be_default}]: "
  read -r input
  BACKEND_PORT="${input:-$be_default}"

  # ENFORCE: VITE_API_PORT must equal BACKEND_PORT
  warn_msg "$(t port_sync_warn)"
  VITE_API_PORT="$BACKEND_PORT"
  ok_msg "$(printf "$(t port_sync_fixed)" "$BACKEND_PORT")"

  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  # --- Version selection ---
  echo ""
  echo -e "  ${C_CYAN}┌─ $(t version_config) ──────────────────────────┐${C_RESET}"
  local ver_default="${PURECORE_VERSION:-latest}"
  printf "  ${C_CYAN}│${C_RESET} $(t ask_version) [${ver_default}]: "
  read -r input
  PURECORE_VERSION="${input:-$ver_default}"
  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  # Write back to .env
  update_env_var "FRONTEND_PORT" "$FRONTEND_PORT"
  update_env_var "BACKEND_PORT" "$BACKEND_PORT"
  update_env_var "VITE_API_PORT" "$VITE_API_PORT"
  update_env_var "PURECORE_VERSION" "$PURECORE_VERSION"

  # Check for default secrets
  local defaults_found=false
  for var in DB_PASSWORD JWT_SECRET; do
    local val
    val=$(grep "^${var}=" .env 2>/dev/null | cut -d= -f2-)
    case "$val" in
      ""|"your_password_here"|"your-jwt-secret-here"|"your-jwt-secret-change-in-production")
        if [ "$defaults_found" = false ]; then
          echo ""
          warn_msg "$(t secret_warn)"
        fi
        echo -e "     ${C_RED}• $var = $val${C_RESET}"
        defaults_found=true
        ;;
    esac
  done
  if [ "$defaults_found" = true ]; then
    echo ""
    warn_msg "$(t env_press_enter)"
    read -r
  fi
}

# ─── Update a key in .env file ─────────────────────────────
update_env_var() {
  local key="$1"
  local value="$2"
  if [ -f .env ]; then
    if grep -q "^${key}=" .env 2>/dev/null; then
      sed -i "s|^${key}=.*|${key}=${value}|" .env
    else
      echo "${key}=${value}" >> .env
    fi
  fi
  export "$key=$value"
}

# ─── ghcr.io login ─────────────────────────────────────────
docker_login() {
  step_banner "$(t step_pull)"

  info_msg "$(t ghcr_login)"

  cd "$PROJECT_DIR"

  # Try a quick pull to check access
  if $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" pull --quiet 2>/dev/null; then
    ok_msg "$(t ghcr_ok)"
    return 0
  fi

  # Try CR_PAT
  if [ -n "${CR_PAT:-}" ]; then
    echo "$CR_PAT" | docker login ghcr.io -u "${GHCR_USER:-}" --password-stdin 2>/dev/null && \
      ok_msg "$(t ghcr_ok)" && return 0
  fi

  # Prompt user
  warn_msg "$(t ghcr_fail)"
  echo -e "  $(t ghcr_cmd_hint)"
  echo ""
  read -rp "  $(t press_enter_continue)" _
}

# ─── Pull images ────────────────────────────────────────────
pull_images() {
  info_msg "$(t pulling)"
  run_compose pull
  ok_msg "$(t pull_done)"
}

# ─── Show container logs ────────────────────────────────────
show_container_logs() {
  local container="$1"
  echo ""
  echo -e "  ${C_YELLOW}┌─ $(t showing_container_logs) ─────────────────────┐${C_RESET}"
  docker logs --tail 40 "$container" 2>&1 | while IFS= read -r line; do
    echo -e "  ${C_YELLOW}│${C_RESET} $line"
  done
  echo -e "  ${C_YELLOW}└────────────────────────────────────────────┘${C_RESET}"
  echo ""
}

# ─── Start services ─────────────────────────────────────────
start_services() {
  step_banner "$(t step_start)"

  info_msg "$(t starting)"

  # Run compose up without strict error handling — it may fail
  # because a container health check times out during startup.
  local up_output
  local up_rc=0
  up_output=$($DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" up -d --remove-orphans 2>&1) || up_rc=$?

  if [ $up_rc -ne 0 ]; then
    echo ""
    warn_msg "$(t compose_up_failed)"

    # Show compose output for context
    echo -e "  ${C_DIM}$(t error_original_output)${C_RESET}"
    echo "$up_output" | while IFS= read -r line; do
      echo -e "  ${C_RED}│${C_RESET} $line"
    done

    # Show backend logs immediately (most likely culprit)
    show_container_logs "purecore-backend" 2>/dev/null || true
  fi

  echo ""
  info_msg "$(t wait_healthy)"

  local max_attempts=30
  local attempt=1
  local all_healthy=false
  local be_unhealthy_reported=false
  local fe_unhealthy_reported=false
  local db_unhealthy_reported=false

  while [ $attempt -le $max_attempts ]; do
    local be_h="${C_RED}⏳${C_RESET}"
    local fe_h="${C_RED}⏳${C_RESET}"
    local db_h="${C_RED}⏳${C_RESET}"

    local be_status
    be_status=$(docker inspect --format='{{.State.Health.Status}}' purecore-backend 2>/dev/null || echo "unknown")
    local fe_status
    fe_status=$(docker inspect --format='{{.State.Health.Status}}' purecore-frontend 2>/dev/null || echo "unknown")
    local db_status
    db_status=$(docker inspect --format='{{.State.Health.Status}}' purecore-db 2>/dev/null || echo "unknown")

    [ "$be_status" = "healthy" ] && be_h="${C_GREEN}✓${C_RESET}"
    [ "$fe_status" = "healthy" ] && fe_h="${C_GREEN}✓${C_RESET}"
    [ "$db_status" = "healthy" ] && db_h="${C_GREEN}✓${C_RESET}"

    # Show logs once if a container becomes unhealthy
    if [ "$be_status" = "unhealthy" ] && [ "$be_unhealthy_reported" = false ]; then
      be_h="${C_RED}✗${C_RESET}"
      be_unhealthy_reported=true
    fi
    if [ "$fe_status" = "unhealthy" ] && [ "$fe_unhealthy_reported" = false ]; then
      fe_h="${C_RED}✗${C_RESET}"
      fe_unhealthy_reported=true
    fi
    if [ "$db_status" = "unhealthy" ] && [ "$db_unhealthy_reported" = false ]; then
      db_h="${C_RED}✗${C_RESET}"
      db_unhealthy_reported=true
    fi

    printf "\r   %-8s %b  %-8s %b  %-8s %b" "backend:" "$be_h" "frontend:" "$fe_h" "database:" "$db_h"

    if [ "$be_status" = "healthy" ] && [ "$fe_status" = "healthy" ] && [ "$db_status" = "healthy" ]; then
      all_healthy=true
      break
    fi

    sleep 2
    attempt=$((attempt + 1))
  done

  echo ""
  echo ""

  if [ "$all_healthy" = true ]; then
    ok_msg "$(t all_healthy)"
  else
    # Show logs for any container that is not healthy
    if [ "$be_status" != "healthy" ]; then
      warn_msg "$(printf "$(t container_unhealthy)" "purecore-backend")"
      show_container_logs "purecore-backend"
    fi
    if [ "$fe_status" != "healthy" ]; then
      warn_msg "$(printf "$(t container_unhealthy)" "purecore-frontend")"
      show_container_logs "purecore-frontend"
    fi
    if [ "$db_status" != "healthy" ]; then
      warn_msg "$(printf "$(t container_unhealthy)" "purecore-db")"
      show_container_logs "purecore-db"
    fi

    warn_msg "$(t health_timeout)"
    echo "  $(t still_starting) $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE logs -f"
  fi
}

# ─── Print success summary ──────────────────────────────────
print_summary() {
  local fe_port="${FRONTEND_PORT:-9001}"
  local be_port="${BACKEND_PORT:-9002}"
  local admin_prefix="${ADMIN_ROUTE_PREFIX:-control-panel}"

  echo ""
  echo -e "  ${C_BRIGHT_GREEN}┌────────────────────────────────────────────┐${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}$(t step_done)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  $(t url_frontend) : ${C_BRIGHT_CYAN}http://localhost:${fe_port}${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  $(t url_backend)  : ${C_BRIGHT_CYAN}http://localhost:${be_port}${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  $(t url_admin)   : ${C_BRIGHT_CYAN}http://localhost:${fe_port}/${admin_prefix}${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_DIM}$(t first_time)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BRIGHT_CYAN}http://localhost:${fe_port}/${admin_prefix}/register${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}└────────────────────────────────────────────┘${C_RESET}"
  echo ""
}

# ─── Stop services ──────────────────────────────────────────
stop_services() {
  cd "$PROJECT_DIR"
  info_msg "$(t stopping)"
  $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" down --remove-orphans 2>/dev/null || true
  ok_msg "$(t stopped)"
}

# ─── Show status ────────────────────────────────────────────
show_status() {
  cd "$PROJECT_DIR"
  echo ""
  echo -e "  ${C_BRIGHT_CYAN}$(t status_title)${C_RESET}"
  echo -e "  ${C_DIM}$(t divider)${C_RESET}"
  $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" ps 2>/dev/null || echo "  No services running."
  echo ""

  # Show additional info if services are up
  local fe_port="${FRONTEND_PORT:-9001}"
  local be_port="${BACKEND_PORT:-9002}"
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "purecore-frontend"; then
    echo -e "  $(t url_frontend): ${C_BRIGHT_CYAN}http://localhost:${fe_port}${C_RESET}"
    echo -e "  $(t url_backend):  ${C_BRIGHT_CYAN}http://localhost:${be_port}${C_RESET}"
  fi
  echo ""
}

# ─── Main ───────────────────────────────────────────────────
main() {
  parse_args "$@"
  detect_lang

  # If no language was forced via --lang, offer interactive selection
  if [ "${LANG_SELECTED:-auto}" = "auto" ]; then
    header
    select_language
  fi

  header

  case "${CMD_ARG:-}" in
    --down)
      ensure_docker
      stop_services
      ;;
    --status)
      ensure_docker
      show_status
      ;;
    *)
      # Full deployment flow
      ensure_docker
      configure_env
      docker_login
      pull_images
      start_services
      print_summary
      ;;
  esac
}

main "$@"
