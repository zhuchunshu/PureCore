#!/usr/bin/env bash
# ============================================================
#  ██████╗ ██╗   ██╗██████╗ ███████╗ ██████╗ ██████╗ ██████╗ ███████╗
#  ██╔══██╗██║   ██║██╔══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝
#  ██████╔╝██║   ██║██████╔╝█████╗  ██║     ██║   ██║██████╔╝█████╗
#  ██╔═══╝ ██║   ██║██╔══██╗██╔══╝  ██║     ██║   ██║██╔══██╗██╔══╝
#  ██║     ╚██████╔╝██║  ██║███████╗╚██████╗╚██████╔╝██║  ██║███████╗
#  ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝
#
#  PureCore — Management Script / 管理脚本
#
#  Manage an already-installed PureCore instance:
#  管理已安装的 PureCore 实例：
#    - Switch admin route prefix / 切换后台路由后缀
#    - Switch frontend theme / 切换前端默认主题
#    - Update to a specific version / 更新到指定版本
#    - View status, logs, restart services / 查看状态、日志、重启服务
#
#  Usage / 用法:
#    ./scripts/manage.sh                    # Interactive / 交互式
#    curl -fsSL <url> | bash               # Pipe mode / 管道模式
#    ./scripts/manage.sh --lang zh          # Force Chinese
#    ./scripts/manage.sh --lang en          # Force English
# ============================================================

# ─── Detect pipe mode (curl | bash) ────────────────────────
PIPE_MODE=false
if [[ ! -t 0 ]] || [[ "${0##*/}" == "bash" ]] || [[ "$0" == "/dev/stdin" ]] || [[ "$0" == /dev/fd/* ]]; then
  PIPE_MODE=true
fi

set -euo pipefail

# ─── Global state ──────────────────────────────────────────
if [ "$PIPE_MODE" = true ]; then
  SCRIPT_DIR="$PWD"
  PROJECT_DIR="$PWD"
else
  SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
fi
ENV_FILE="$PROJECT_DIR/.env"
COMPOSE_FILE="$PROJECT_DIR/docker-compose.yml"
LANG="auto"
DOCKER_COMPOSE_CMD=""
GITHUB_API_RELEASES="https://api.github.com/repos/zhuchunshu/PureCore/releases"
INSTALLED=false

# ─── Color palette ─────────────────────────────────────────
C_RESET='\033[0m'
C_BOLD='\033[1m'
C_DIM='\033[2m'
C_CYAN='\033[0;36m'
C_GREEN='\033[0;32m'
C_YELLOW='\033[1;33m'
C_RED='\033[0;31m'
C_MAGENTA='\033[0;35m'
C_BLUE='\033[0;34m'
C_BRIGHT_CYAN='\033[1;36m'
C_BRIGHT_GREEN='\033[1;32m'
C_BRIGHT_RED='\033[1;31m'
C_BRIGHT_YELLOW='\033[1;93m'
C_BG_CYAN='\033[46m'
C_BG_BLACK='\033[40m'

# ─── Multi-language message catalog ────────────────────────
declare -A MSG_EN MSG_ZH

# --- English ---
MSG_EN[title]="PureCore Management"
MSG_EN[title_sub]="Manage your PureCore instance"
MSG_EN[divider]="━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
MSG_EN[lang_selected]="Language: English"
MSG_EN[select_lang]="Select Language / 选择语言"
MSG_EN[lang_option_en]="English"
MSG_EN[lang_option_zh]="中文 (Chinese)"
MSG_EN[hint_stuck]="If the script appears stuck, try pressing Enter."
MSG_EN[check_installation]="Checking PureCore installation..."
MSG_EN[installed_ok]="PureCore installation detected ✓"
MSG_EN[installed_compose_ok]="docker-compose.yml found"
MSG_EN[installed_env_ok]=".env file found"
MSG_EN[installed_containers]="Running containers found"
MSG_EN[not_installed_title]="PureCore Not Detected"
MSG_EN[not_installed_desc]="This directory does not appear to contain a valid PureCore installation."
MSG_EN[not_installed_hint]="Please run this script from the PureCore project directory, or deploy PureCore first using:"
MSG_EN[not_installed_deploy_cmd]="  curl -fsSL https://raw.githubusercontent.com/zhuchunshu/PureCore/main/scripts/deploy.sh | bash"
MSG_EN[menu_title]="Management Menu"
MSG_EN[menu_switch_prefix]="Switch Admin Route Prefix"
MSG_EN[menu_switch_theme]="Switch Frontend Theme"
MSG_EN[menu_update]="Update PureCore Version"
MSG_EN[menu_status]="View Service Status"
MSG_EN[menu_restart]="Restart Services"
MSG_EN[menu_logs]="View Service Logs"
MSG_EN[menu_shell]="Open Shell in Container"
MSG_EN[menu_lang]="Switch Language"
MSG_EN[menu_exit]="Exit"
MSG_EN[enter_choice]="Enter choice [1-8]: "
MSG_EN[invalid_choice]="Invalid choice. Please try again."
MSG_EN[press_enter]="Press Enter to continue..."
MSG_EN[current_prefix]="Current admin route prefix"
MSG_EN[enter_new_prefix]="Enter new admin route prefix"
MSG_EN[prefix_updated]="Admin route prefix updated to: %s"
MSG_EN[prefix_restart_note]="Restart services to apply the change? [Y/n]: "
MSG_EN[prefix_restarting]="Restarting services to apply prefix change..."
MSG_EN[prefix_no_restart]="Prefix updated in .env. Restart services manually to apply."
MSG_EN[current_theme]="Current theme"
MSG_EN[available_themes]="Available DaisyUI themes"
MSG_EN[enter_theme]="Enter theme name"
MSG_EN[theme_updated]="Theme updated to: %s"
MSG_EN[theme_restart_note]="Restart services to apply the theme? [Y/n]: "
MSG_EN[theme_restarting]="Restarting services to apply theme change..."
MSG_EN[theme_no_restart]="Theme updated in .env. Restart services manually to apply."
MSG_EN[update_title]="Update PureCore"
MSG_EN[current_version]="Current version"
MSG_EN[fetching_versions]="Fetching available versions from GitHub..."
MSG_EN[fetch_failed]="Failed to fetch versions from GitHub API."
MSG_EN[fetch_fallback]="You can still enter a version manually."
MSG_EN[select_version]="Select a version to update to"
MSG_EN[manual_input_option]="[Manual input]"
MSG_EN[enter_version_manual]="Enter version (e.g., 1.0.7, 1.0.0-alpha.1, latest): "
MSG_EN[navigate_hint]="Use ↑/↓ arrows or j/k to navigate, Enter to select"
MSG_EN[selecting_latest]="Selecting latest version..."
MSG_EN[version_selected]="Selected version: %s"
MSG_EN[update_pulling]="Pulling PureCore images (version: %s)..."
MSG_EN[update_pull_done]="Images pulled successfully."
MSG_EN[update_starting]="Starting services with new version..."
MSG_EN[update_done]="Update complete! PureCore is now running version %s."
MSG_EN[update_failed]="Update failed. Check logs with: docker compose logs"
MSG_EN[status_title]="PureCore Service Status"
MSG_EN[status_no_containers]="No PureCore containers are running."
MSG_EN[restarting]="Restarting all services..."
MSG_EN[restart_done]="Services restarted successfully."
MSG_EN[restart_failed]="Failed to restart services."
MSG_EN[logs_title]="Service Logs"
MSG_EN[logs_choose]="Which service logs to view?"
MSG_EN[logs_all]="All services"
MSG_EN[logs_backend]="Backend"
MSG_EN[logs_frontend]="Frontend"
MSG_EN[logs_database]="Database"
MSG_EN[logs_follow]="Follow logs (tail -f)? [y/N]: "
MSG_EN[shell_title]="Open Shell in Container"
MSG_EN[shell_choose]="Which container?"
MSG_EN[shell_opening]="Opening shell in %s container..."
MSG_EN[shell_failed]="Failed to open shell. Is the container running?"
MSG_EN[goodbye]="Goodbye!"
MSG_EN[docker_not_running]="Docker does not appear to be running."
MSG_EN[no_compose]="Could not find docker compose command."

# --- 中文 ---
MSG_ZH[title]="PureCore 管理工具"
MSG_ZH[title_sub]="管理你的 PureCore 实例"
MSG_ZH[divider]="━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
MSG_ZH[lang_selected]="语言：中文"
MSG_ZH[select_lang]="Select Language / 选择语言"
MSG_ZH[lang_option_en]="English"
MSG_ZH[lang_option_zh]="中文 (Chinese)"
MSG_ZH[hint_stuck]="如果脚本看起来卡住了，请尝试按回车键。"
MSG_ZH[check_installation]="正在检查 PureCore 安装状态..."
MSG_ZH[installed_ok]="已检测到 PureCore 安装 ✓"
MSG_ZH[installed_compose_ok]="找到 docker-compose.yml"
MSG_ZH[installed_env_ok]="找到 .env 文件"
MSG_ZH[installed_containers]="发现运行中的容器"
MSG_ZH[not_installed_title]="未检测到 PureCore"
MSG_ZH[not_installed_desc]="当前目录似乎不包含有效的 PureCore 安装。"
MSG_ZH[not_installed_hint]="请在 PureCore 项目目录中运行此脚本，或先使用以下命令部署 PureCore："
MSG_ZH[not_installed_deploy_cmd]="  curl -fsSL https://raw.githubusercontent.com/zhuchunshu/PureCore/main/scripts/deploy.sh | bash"
MSG_ZH[menu_title]="管理菜单"
MSG_ZH[menu_switch_prefix]="切换后台路由后缀"
MSG_ZH[menu_switch_theme]="切换前端默认主题"
MSG_ZH[menu_update]="更新 PureCore 版本"
MSG_ZH[menu_status]="查看服务状态"
MSG_ZH[menu_restart]="重启服务"
MSG_ZH[menu_logs]="查看服务日志"
MSG_ZH[menu_shell]="进入容器 Shell"
MSG_ZH[menu_lang]="切换语言"
MSG_ZH[menu_exit]="退出"
MSG_ZH[enter_choice]="请输入选项 [1-8]: "
MSG_ZH[invalid_choice]="无效选项，请重试。"
MSG_ZH[press_enter]="按 Enter 继续..."
MSG_ZH[current_prefix]="当前后台路由前缀"
MSG_ZH[enter_new_prefix]="输入新的后台路由前缀"
MSG_ZH[prefix_updated]="后台路由前缀已更新为：%s"
MSG_ZH[prefix_restart_note]="是否重启服务以应用更改？[Y/n]: "
MSG_ZH[prefix_restarting]="正在重启服务以应用前缀更改..."
MSG_ZH[prefix_no_restart]="前缀已在 .env 中更新。请手动重启服务以应用更改。"
MSG_ZH[current_theme]="当前主题"
MSG_ZH[available_themes]="可用的 DaisyUI 主题"
MSG_ZH[enter_theme]="输入主题名称"
MSG_ZH[theme_updated]="主题已更新为：%s"
MSG_ZH[theme_restart_note]="是否重启服务以应用主题？[Y/n]: "
MSG_ZH[theme_restarting]="正在重启服务以应用主题更改..."
MSG_ZH[theme_no_restart]="主题已在 .env 中更新。请手动重启服务以应用更改。"
MSG_ZH[update_title]="更新 PureCore"
MSG_ZH[current_version]="当前版本"
MSG_ZH[fetching_versions]="正在从 GitHub 获取可用版本..."
MSG_ZH[fetch_failed]="无法从 GitHub API 获取版本列表。"
MSG_ZH[fetch_fallback]="你仍可以手动输入版本号。"
MSG_ZH[select_version]="选择要更新到的版本"
MSG_ZH[manual_input_option]="[手动输入]"
MSG_ZH[enter_version_manual]="输入版本号（例如：1.0.7、1.0.0-alpha.1、latest）："
MSG_ZH[navigate_hint]="使用 ↑/↓ 方向键或 j/k 键导航，Enter 确认"
MSG_ZH[selecting_latest]="正在选择最新版本..."
MSG_ZH[version_selected]="已选择版本：%s"
MSG_ZH[update_pulling]="正在拉取 PureCore 镜像（版本：%s）..."
MSG_ZH[update_pull_done]="镜像拉取成功。"
MSG_ZH[update_starting]="正在使用新版本启动服务..."
MSG_ZH[update_done]="更新完成！PureCore 现在运行的是 %s 版本。"
MSG_ZH[update_failed]="更新失败。使用以下命令查看日志：docker compose logs"
MSG_ZH[status_title]="PureCore 服务状态"
MSG_ZH[status_no_containers]="没有运行中的 PureCore 容器。"
MSG_ZH[restarting]="正在重启所有服务..."
MSG_ZH[restart_done]="服务已成功重启。"
MSG_ZH[restart_failed]="服务重启失败。"
MSG_ZH[logs_title]="服务日志"
MSG_ZH[logs_choose]="查看哪个服务的日志？"
MSG_ZH[logs_all]="所有服务"
MSG_ZH[logs_backend]="后端"
MSG_ZH[logs_frontend]="前端"
MSG_ZH[logs_database]="数据库"
MSG_ZH[logs_follow]="是否持续跟踪日志（tail -f）？[y/N]: "
MSG_ZH[shell_title]="进入容器 Shell"
MSG_ZH[shell_choose]="选择容器？"
MSG_ZH[shell_opening]="正在进入 %s 容器..."
MSG_ZH[shell_failed]="无法进入容器 Shell。容器是否在运行？"
MSG_ZH[goodbye]="再见！"
MSG_ZH[docker_not_running]="Docker 似乎未在运行。"
MSG_ZH[no_compose]="无法找到 docker compose 命令。"

# DaisyUI themes list
DAISYUI_THEMES=(
  "light" "dark" "cupcake" "bumblebee" "emerald" "corporate"
  "synthwave" "retro" "cyberpunk" "valentine" "halloween"
  "garden" "forest" "aqua" "lofi" "pastel" "fantasy"
  "wireframe" "black" "luxury" "dracula" "cmyk" "autumn"
  "business" "acid" "lemonade" "night" "coffee" "winter"
  "dim" "nord" "sunset"
)

# ─── Helper: translate ─────────────────────────────────────
t() {
  local key="$1"
  local msg=""
  if [ "$LANG" = "zh" ]; then
    msg="${MSG_ZH[$key]:-${MSG_EN[$key]:-$key}}"
  else
    msg="${MSG_EN[$key]:-$key}"
  fi
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
  echo -e "  ${C_DIM}💡 $(t hint_stuck)${C_RESET}"
  echo ""
}

ok_msg()   { echo -e "   ${C_BRIGHT_GREEN}✓${C_RESET} $1"; }
info_msg() { echo -e "   ${C_CYAN}›${C_RESET} $1"; }
warn_msg() { echo -e "   ${C_YELLOW}⚠${C_RESET} $1"; }
err_msg()  { echo -e "   ${C_BRIGHT_RED}✗${C_RESET} $1"; }

# ─── Update a key in .env file ─────────────────────────────
update_env_var() {
  local key="$1"
  local value="$2"
  if [ -f "$ENV_FILE" ]; then
    if grep -q "^${key}=" "$ENV_FILE" 2>/dev/null; then
      sed -i "s|^${key}=.*|${key}=${value}|" "$ENV_FILE"
    else
      echo "${key}=${value}" >> "$ENV_FILE"
    fi
  fi
  export "$key=$value"
}

# ─── Get current value from .env ───────────────────────────
get_env_var() {
  local key="$1"
  local default="${2:-}"
  if [ -f "$ENV_FILE" ]; then
    local val
    val=$(grep "^${key}=" "$ENV_FILE" 2>/dev/null | cut -d= -f2-)
    echo "${val:-$default}"
  else
    echo "$default"
  fi
}

# ─── Detect docker compose command ─────────────────────────
detect_docker_compose() {
  if docker compose version &>/dev/null 2>&1; then
    DOCKER_COMPOSE_CMD="docker compose"
  elif command -v docker-compose &>/dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
  else
    err_msg "$(t no_compose)"
    return 1
  fi
  return 0
}

# ─── Check if docker is available ──────────────────────────
check_docker() {
  if ! command -v docker &>/dev/null; then
    err_msg "$(t docker_not_running)"
    return 1
  fi
  if ! docker info &>/dev/null 2>&1; then
    err_msg "$(t docker_not_running)"
    return 1
  fi
  return 0
}

# ─── Check PureCore installation ───────────────────────────
check_installation() {
  echo ""
  info_msg "$(t check_installation)"
  echo ""

  local checks_ok=0
  local checks_total=0

  # Check docker-compose.yml
  checks_total=$((checks_total + 1))
  if [ -f "$COMPOSE_FILE" ]; then
    ok_msg "$(t installed_compose_ok)"
    checks_ok=$((checks_ok + 1))
  else
    err_msg "docker-compose.yml not found"
  fi

  # Check .env
  checks_total=$((checks_total + 1))
  if [ -f "$ENV_FILE" ]; then
    ok_msg "$(t installed_env_ok)"
    checks_ok=$((checks_ok + 1))
  else
    err_msg ".env not found"
  fi

  # Check for running containers
  checks_total=$((checks_total + 1))
  if check_docker 2>/dev/null && detect_docker_compose 2>/dev/null; then
    local running
    running=$($DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" ps -q 2>/dev/null | wc -l)
    if [ "$running" -gt 0 ]; then
      ok_msg "$(t installed_containers) (${running})"
      checks_ok=$((checks_ok + 1))
    else
      warn_msg "No running containers found (services may be stopped)"
    fi
  fi

  echo ""

  if [ "$checks_ok" -ge 2 ]; then
    ok_msg "$(t installed_ok)"
    INSTALLED=true
    return 0
  else
    echo ""
    echo -e "  ${C_BRIGHT_RED}┌─ $(t not_installed_title) ────────────────────────────┐${C_RESET}"
    echo -e "  ${C_BRIGHT_RED}│${C_RESET}  $(t not_installed_desc)"
    echo -e "  ${C_BRIGHT_RED}│${C_RESET}"
    echo -e "  ${C_BRIGHT_RED}│${C_RESET}  $(t not_installed_hint)"
    echo -e "  ${C_BRIGHT_RED}│${C_RESET}  ${C_DIM}$(t not_installed_deploy_cmd)${C_RESET}"
    echo -e "  ${C_BRIGHT_RED}└────────────────────────────────────────────┘${C_RESET}"
    echo ""
    INSTALLED=false
    return 1
  fi
}

# ─── Prompt to restart services ────────────────────────────
prompt_restart() {
  local reason="$1"
  echo ""
  read -rp "  $(t "${reason}")" answer
  if [ "${answer,,}" = "y" ] || [ "${answer,,}" = "yes" ] || [ -z "$answer" ]; then
    return 0
  fi
  return 1
}

# ─── Restart services ──────────────────────────────────────
do_restart() {
  info_msg "$(t restarting)"
  cd "$PROJECT_DIR"
  if $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" up -d --remove-orphans 2>&1; then
    ok_msg "$(t restart_done)"
    return 0
  else
    err_msg "$(t restart_failed)"
    return 1
  fi
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

# ─── Switch admin route prefix ─────────────────────────────
switch_prefix() {
  echo ""
  echo -e "  ${C_CYAN}┌─ $(t menu_switch_prefix) ──────────────────────────┐${C_RESET}"

  local current
  current=$(get_env_var "ADMIN_ROUTE_PREFIX" "control-panel")

  echo -e "  ${C_CYAN}│${C_RESET}  $(t current_prefix): ${C_BRIGHT_CYAN}${current}${C_RESET}"
  echo -e "  ${C_CYAN}│${C_RESET}"
  printf "  ${C_CYAN}│${C_RESET}  $(t enter_new_prefix) [${C_BRIGHT_GREEN}${current}${C_RESET}]: "
  read -r new_prefix
  new_prefix="${new_prefix:-$current}"

  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"
  echo ""

  # Update both vars
  update_env_var "ADMIN_ROUTE_PREFIX" "$new_prefix"
  update_env_var "VITE_ADMIN_ROUTE_PREFIX" "$new_prefix"

  ok_msg "$(printf "$(t prefix_updated)" "$new_prefix")"

  if prompt_restart "prefix_restart_note"; then
    info_msg "$(t prefix_restarting)"
    do_restart
  else
    warn_msg "$(t prefix_no_restart)"
  fi
}

# ─── Interactive theme selector with arrow keys ────────────
# Stores result in global SELECTED_THEME variable (avoids subshell stdin issue)
select_theme_interactive() {
  local current="$1"
  local count=${#DAISYUI_THEMES[@]}
  local selected=0
  local page_size=0

  # Find current theme index
  for i in "${!DAISYUI_THEMES[@]}"; do
    if [ "${DAISYUI_THEMES[$i]}" = "$current" ]; then
      selected=$i
      break
    fi
  done

  # Calculate how many themes fit on screen (leave room for UI elements)
  local max_visible=12
  if [ $count -lt $max_visible ]; then
    max_visible=$count
  fi

  echo ""
  echo -e "  ${C_CYAN}┌─ $(t available_themes) ─────────────────────────┐${C_RESET}"
  echo -e "  ${C_DIM}│  $(t navigate_hint)${C_RESET}"
  echo -e "  ${C_CYAN}│${C_RESET}"

  draw_theme_list() {
    local start_idx=$1
    local sel=$2
    local visible=$3
    local total=$4

    # Clear previous list if any (move cursor up)
    if [ "${5:-}" != "first" ]; then
      tput cuu $((visible + 2)) 2>/dev/null || true
    fi

    for ((i = 0; i < visible; i++)); do
      local idx=$((start_idx + i))
      if [ $idx -ge $total ]; then
        break
      fi
      local theme_name="${DAISYUI_THEMES[$idx]}"
      # Pad to consistent width
      if [ $idx -eq $sel ]; then
        echo -e "  ${C_CYAN}│${C_RESET} ${C_BG_CYAN}${C_BOLD} ▶ ${theme_name} ${C_RESET}"
      else
        echo -e "  ${C_CYAN}│${C_RESET}    ${theme_name}"
      fi
    done

    # Show scroll indicator
    if [ $total -gt $visible ]; then
      local pct=$(( (start_idx + visible) * 100 / total ))
      echo -e "  ${C_CYAN}│${C_RESET} ${C_DIM}── showing $(($start_idx + 1))-$(($start_idx + visible)) of $total ($pct%) ──${C_RESET}"
    fi
  }

  local scroll_offset=0
  draw_theme_list $scroll_offset $selected $max_visible $count "first"

  # Save terminal settings and set raw mode for key reading
  local old_stty
  old_stty=$(stty -g 2>/dev/null || true)
  stty raw -echo 2>/dev/null || true

  local key
  local done=false
  while [ "$done" = false ]; do
    key=$(dd bs=1 count=3 2>/dev/null || true)

    case "$key" in
      $'\x1b[A'|'k')
        # Up
        if [ $selected -gt 0 ]; then
          selected=$((selected - 1))
          if [ $selected -lt $scroll_offset ]; then
            scroll_offset=$selected
          fi
          draw_theme_list $scroll_offset $selected $max_visible $count
        fi
        ;;
      $'\x1b[B'|'j')
        # Down
        if [ $selected -lt $((count - 1)) ]; then
          selected=$((selected + 1))
          if [ $selected -ge $((scroll_offset + max_visible)) ]; then
            scroll_offset=$((selected - max_visible + 1))
          fi
          draw_theme_list $scroll_offset $selected $max_visible $count
        fi
        ;;
      '')
        # Enter
        done=true
        ;;
    esac
  done

  # Restore terminal settings
  stty "$old_stty" 2>/dev/null || true

  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  SELECTED_THEME="${DAISYUI_THEMES[$selected]}"
}

# ─── Switch frontend theme ─────────────────────────────────
switch_theme() {
  local current
  current=$(get_env_var "THEME" "sunset")

  echo ""
  echo -e "  ${C_CYAN}┌─ $(t menu_switch_theme) ──────────────────────────┐${C_RESET}"
  echo -e "  ${C_CYAN}│${C_RESET}  $(t current_theme): ${C_MAGENTA}${current}${C_RESET}"
  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  # Use interactive theme selector (direct call, result via global var)
  select_theme_interactive "$current"
  local new_theme="$SELECTED_THEME"

  echo ""
  ok_msg "$(printf "$(t theme_updated)" "$new_theme")"

  # Update both vars
  update_env_var "THEME" "$new_theme"
  update_env_var "VITE_THEME" "$new_theme"

  if prompt_restart "theme_restart_note"; then
    info_msg "$(t theme_restarting)"
    do_restart
  else
    warn_msg "$(t theme_no_restart)"
  fi
}

# ─── Fetch versions from GitHub API ────────────────────────
fetch_versions() {
  info_msg "$(t fetching_versions)"

  local versions_json
  versions_json=$(curl -fsSL "${GITHUB_API_RELEASES}?per_page=30" 2>/dev/null || echo "")

  if [ -z "$versions_json" ]; then
    err_msg "$(t fetch_failed)"
    info_msg "$(t fetch_fallback)"
    return 1
  fi

  # Extract tag names and strip 'v' prefix, also get the "latest" release
  local tags
  tags=$(echo "$versions_json" | jq -r '.[].tag_name' 2>/dev/null | sed 's/^v//' | grep -v '^$' || echo "")

  if [ -z "$tags" ]; then
    err_msg "$(t fetch_failed)"
    return 1
  fi

  # Add "latest" at the top
  echo "latest"
  echo "$tags"
}

# ─── Interactive version selector with arrow keys ──────────
# Stores result in global SELECTED_VERSION variable (avoids subshell stdin issue)
select_version_interactive() {
  local versions=()
  local line
  while IFS= read -r line; do
    [ -n "$line" ] && versions+=("$line")
  done < <(fetch_versions)

  # Add manual input option
  versions+=("__MANUAL_INPUT__")

  local count=${#versions[@]}
  if [ $count -eq 0 ]; then
    printf "  $(t enter_version_manual)"
    read -r ver
    SELECTED_VERSION="${ver:-latest}"
    return
  fi

  local selected=0
  local max_visible=10
  if [ $count -lt $max_visible ]; then
    max_visible=$count
  fi

  echo ""
  echo -e "  ${C_CYAN}┌─ $(t select_version) ───────────────────────────┐${C_RESET}"
  echo -e "  ${C_DIM}│  $(t navigate_hint)${C_RESET}"
  echo -e "  ${C_CYAN}│${C_RESET}"

  draw_version_list() {
    local start_idx=$1
    local sel=$2
    local visible=$3
    local total=$4

    if [ "${5:-}" != "first" ]; then
      tput cuu $((visible + 2)) 2>/dev/null || true
    fi

    for ((i = 0; i < visible; i++)); do
      local idx=$((start_idx + i))
      if [ $idx -ge $total ]; then
        break
      fi
      local vname="${versions[$idx]}"
      if [ "$vname" = "__MANUAL_INPUT__" ]; then
        vname="$(t manual_input_option)"
      fi
      if [ $idx -eq $sel ]; then
        if [ "$vname" = "latest" ]; then
          echo -e "  ${C_CYAN}│${C_RESET} ${C_BRIGHT_GREEN}${C_BOLD} ▶ ${vname} ${C_DIM}(most recent)${C_RESET}"
        elif [ "$vname" = "$(t manual_input_option)" ]; then
          echo -e "  ${C_CYAN}│${C_RESET} ${C_BRIGHT_YELLOW}${C_BOLD} ▶ ${vname}${C_RESET}"
        else
          echo -e "  ${C_CYAN}│${C_RESET} ${C_BG_CYAN}${C_BOLD} ▶ ${vname} ${C_RESET}"
        fi
      else
        if [ "$vname" = "latest" ]; then
          echo -e "  ${C_CYAN}│${C_RESET}    ${C_GREEN}${vname}${C_RESET} ${C_DIM}(most recent)${C_RESET}"
        elif [ "$vname" = "$(t manual_input_option)" ]; then
          echo -e "  ${C_CYAN}│${C_RESET}    ${C_YELLOW}${vname}${C_RESET}"
        else
          echo -e "  ${C_CYAN}│${C_RESET}    ${vname}"
        fi
      fi
    done

    if [ $total -gt $visible ]; then
      local pct=$(( (start_idx + visible) * 100 / total ))
      echo -e "  ${C_CYAN}│${C_RESET} ${C_DIM}── showing $(($start_idx + 1))-$(($start_idx + visible)) of $total ($pct%) ──${C_RESET}"
    fi
  }

  local scroll_offset=0
  draw_version_list $scroll_offset $selected $max_visible $count "first"

  local old_stty
  old_stty=$(stty -g 2>/dev/null || true)
  stty raw -echo 2>/dev/null || true

  local key
  local done=false
  while [ "$done" = false ]; do
    key=$(dd bs=1 count=3 2>/dev/null || true)

    case "$key" in
      $'\x1b[A'|'k')
        if [ $selected -gt 0 ]; then
          selected=$((selected - 1))
          if [ $selected -lt $scroll_offset ]; then
            scroll_offset=$selected
          fi
          draw_version_list $scroll_offset $selected $max_visible $count
        fi
        ;;
      $'\x1b[B'|'j')
        if [ $selected -lt $((count - 1)) ]; then
          selected=$((selected + 1))
          if [ $selected -ge $((scroll_offset + max_visible)) ]; then
            scroll_offset=$((selected - max_visible + 1))
          fi
          draw_version_list $scroll_offset $selected $max_visible $count
        fi
        ;;
      '')
        done=true
        ;;
    esac
  done

  stty "$old_stty" 2>/dev/null || true

  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  local chosen="${versions[$selected]}"
  if [ "$chosen" = "__MANUAL_INPUT__" ]; then
    echo ""
    printf "  $(t enter_version_manual)"
    read -r chosen
    if [ -z "$chosen" ]; then
      chosen="latest"
    fi
  fi

  echo ""
  if [ "$chosen" = "latest" ]; then
    info_msg "$(t selecting_latest)"
  fi
  ok_msg "$(printf "$(t version_selected)" "$chosen")"

  SELECTED_VERSION="$chosen"
}

# ─── Update PureCore version ───────────────────────────────
update_version() {
  echo ""
  echo -e "  ${C_CYAN}┌─ $(t update_title) ──────────────────────────────┐${C_RESET}"

  local current
  current=$(get_env_var "PURECORE_VERSION" "latest")
  echo -e "  ${C_CYAN}│${C_RESET}  $(t current_version): ${C_BRIGHT_CYAN}${current}${C_RESET}"
  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  # Select version interactively (direct call, result via global var)
  select_version_interactive
  local new_version="$SELECTED_VERSION"
  if [ -z "$new_version" ]; then
    new_version="latest"
  fi

  # Update .env
  update_env_var "PURECORE_VERSION" "$new_version"

  # Pull new images
  echo ""
  info_msg "$(printf "$(t update_pulling)" "$new_version")"
  cd "$PROJECT_DIR"

  if ! $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" pull 2>&1; then
    err_msg "$(t update_failed)"
    echo ""
    return 1
  fi
  ok_msg "$(t update_pull_done)"

  # Restart services
  info_msg "$(t update_starting)"
  if $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" up -d --remove-orphans 2>&1; then
    ok_msg "$(printf "$(t update_done)" "$new_version")"
  else
    err_msg "$(t update_failed)"
    return 1
  fi
}

# ─── View service status ───────────────────────────────────
view_status() {
  echo ""
  echo -e "  ${C_BRIGHT_CYAN}$(t status_title)${C_RESET}"
  echo -e "  ${C_DIM}$(t divider)${C_RESET}"

  cd "$PROJECT_DIR"
  if ! $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" ps 2>/dev/null; then
    warn_msg "$(t status_no_containers)"
  fi

  echo ""

  # Show version info
  local ver
  ver=$(get_env_var "PURECORE_VERSION" "latest")
  local prefix
  prefix=$(get_env_var "ADMIN_ROUTE_PREFIX" "control-panel")
  local theme
  theme=$(get_env_var "THEME" "sunset")
  local fe_port
  fe_port=$(get_env_var "FRONTEND_PORT" "9001")

  echo -e "  ${C_DIM}Version:${C_RESET}       ${C_CYAN}${ver}${C_RESET}"
  echo -e "  ${C_DIM}Admin prefix:${C_RESET}  ${C_CYAN}${prefix}${C_RESET}"
  echo -e "  ${C_DIM}Theme:${C_RESET}         ${C_MAGENTA}${theme}${C_RESET}"
  echo -e "  ${C_DIM}Frontend URL:${C_RESET}  ${C_BRIGHT_CYAN}http://localhost:${fe_port}${C_RESET}"
  echo -e "  ${C_DIM}Admin URL:${C_RESET}     ${C_BRIGHT_CYAN}http://localhost:${fe_port}/${prefix}${C_RESET}"
  echo ""
}

# ─── Restart services ──────────────────────────────────────
restart_services() {
  echo ""
  do_restart
}

# ─── View logs ─────────────────────────────────────────────
view_logs() {
  echo ""
  echo -e "  ${C_CYAN}┌─ $(t logs_title) ──────────────────────────────┐${C_RESET}"
  echo -e "  ${C_CYAN}│${C_RESET}  $(t logs_choose)"
  echo -e "  ${C_CYAN}│${C_RESET}   1) $(t logs_all)"
  echo -e "  ${C_CYAN}│${C_RESET}   2) $(t logs_backend)"
  echo -e "  ${C_CYAN}│${C_RESET}   3) $(t logs_frontend)"
  echo -e "  ${C_CYAN}│${C_RESET}   4) $(t logs_database)"
  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  local choice
  printf "  Enter choice [1-4]: "
  read -r choice

  local service=""
  case "$choice" in
    1) service="" ;;
    2) service="backend" ;;
    3) service="frontend" ;;
    4) service="postgres" ;;
    *) warn_msg "$(t invalid_choice)"; return ;;
  esac

  echo ""
  read -rp "  $(t logs_follow)" follow
  echo ""

  cd "$PROJECT_DIR"
  if [ "${follow,,}" = "y" ] || [ "${follow,,}" = "yes" ]; then
    $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" logs -f --tail=50 $service 2>/dev/null
  else
    $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" logs --tail=100 $service 2>/dev/null | less -R
  fi
}

# ─── Open shell in container ───────────────────────────────
open_shell() {
  echo ""
  echo -e "  ${C_CYAN}┌─ $(t shell_title) ──────────────────────────────┐${C_RESET}"
  echo -e "  ${C_CYAN}│${C_RESET}  $(t shell_choose)"
  echo -e "  ${C_CYAN}│${C_RESET}   1) $(t logs_backend)   (purecore-backend)"
  echo -e "  ${C_CYAN}│${C_RESET}   2) $(t logs_frontend)  (purecore-frontend)"
  echo -e "  ${C_CYAN}│${C_RESET}   3) $(t logs_database)  (purecore-db)"
  echo -e "  ${C_CYAN}└────────────────────────────────────────────┘${C_RESET}"

  local choice
  printf "  Enter choice [1-3]: "
  read -r choice

  local container=""
  case "$choice" in
    1) container="purecore-backend" ;;
    2) container="purecore-frontend" ;;
    3) container="purecore-db" ;;
    *) warn_msg "$(t invalid_choice)"; return ;;
  esac

  echo ""
  info_msg "$(printf "$(t shell_opening)" "$container")"

  if docker exec -it "$container" /bin/sh 2>/dev/null || docker exec -it "$container" /bin/bash 2>/dev/null; then
    :
  else
    err_msg "$(t shell_failed)"
  fi
}

# ─── Interactive menu ──────────────────────────────────────
show_menu() {
  echo ""
  echo -e "  ${C_BRIGHT_GREEN}┌─ $(t menu_title) ────────────────────────────────┐${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}1)${C_RESET} ${C_CYAN}$(t menu_switch_prefix)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}2)${C_RESET} ${C_MAGENTA}$(t menu_switch_theme)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}3)${C_RESET} ${C_BRIGHT_CYAN}$(t menu_update)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}4)${C_RESET} ${C_GREEN}$(t menu_status)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}5)${C_RESET} ${C_YELLOW}$(t menu_restart)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}6)${C_RESET} ${C_BLUE}$(t menu_logs)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}7)${C_RESET} ${C_DIM}$(t menu_shell)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}8)${C_RESET} ${C_DIM}$(t menu_lang)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}  ${C_BOLD}0)${C_RESET} ${C_RED}$(t menu_exit)${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}│${C_RESET}"
  echo -e "  ${C_BRIGHT_GREEN}└────────────────────────────────────────────┘${C_RESET}"
}

# ─── Main loop ─────────────────────────────────────────────
main_loop() {
  local running=true
  while [ "$running" = true ]; do
    show_menu
    echo ""
    printf "  $(t enter_choice)"
    read -r choice
    echo ""

    case "$choice" in
      1) switch_prefix ;;
      2) switch_theme ;;
      3) update_version ;;
      4) view_status ;;
      5) restart_services ;;
      6) view_logs ;;
      7) open_shell ;;
      8)
        # Toggle language
        if [ "$LANG" = "zh" ]; then
          LANG="en"
        else
          LANG="zh"
        fi
        ok_msg "$(t lang_selected)"
        ;;
      0)
        echo -e "  ${C_GREEN}$(t goodbye)${C_RESET}"
        echo ""
        running=false
        ;;
      *)
        warn_msg "$(t invalid_choice)"
        ;;
    esac

    if [ "$running" = true ]; then
      echo ""
      read -rp "  $(t press_enter)"
    fi
    header
  done
}

# ─── Detect language from system locale ─────────────────────
detect_lang() {
  local locale="${LANG:-en_US.UTF-8}"
  if echo "$locale" | grep -qiE '^zh|zh_CN|zh_TW'; then
    LANG="zh"
  else
    LANG="en"
  fi
}

# ─── Parse command-line arguments ───────────────────────────
parse_args() {
  local lang_selected="auto"
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --lang)
        if [ -n "${2:-}" ]; then
          lang_selected="$2"
          shift 2
        else
          shift
        fi
        ;;
      *)
        shift
        ;;
    esac
  done

  if [ "$lang_selected" != "auto" ]; then
    LANG="$lang_selected"
  fi
}

# ─── Main ───────────────────────────────────────────────────
main() {
  # Reconnect stdin to terminal if running in pipe mode
  if [ "$PIPE_MODE" = true ]; then
    exec </dev/tty 2>/dev/null || true
  fi

  parse_args "$@"

  # Detect language if not forced
  if [ "${LANG:-auto}" = "auto" ]; then
    detect_lang
  fi

  # Show header
  header

  # Interactive language selection on first run (only in pipe mode or when explicitly interactive)
  if [ "$PIPE_MODE" = true ]; then
    select_language
    header
  fi

  # Check installation
  if ! check_installation; then
    exit 1
  fi

  # Ensure Docker is available
  if ! check_docker; then
    exit 1
  fi

  if ! detect_docker_compose; then
    exit 1
  fi

  # Source .env for runtime values
  if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE" 2>/dev/null || true
    set +a
  fi

  # Enter main interactive loop
  main_loop
}

main "$@"
