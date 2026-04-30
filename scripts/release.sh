#!/usr/bin/env bash
# ============================================================
# PureCore — Release Script / 发布脚本
#
# Automates a new version release:
#   1. Bump version in purecore.json (interactive, auto-increments patch)
#   2. Build Docker images via docker compose
#   3. Tag and push images to GitHub Container Registry (ghcr.io)
#   4. Create and push a Git tag (version suffix like -alpha.1, -beta.1)
#   5. Create a GitHub Release with auto-generated release notes
#
# 自动发布新版本：
#   1. 更新 purecore.json 中的版本号（交互式，自动递增补丁号）
#   2. 通过 docker compose 构建 Docker 镜像
#   3. 为镜像打标签并推送到 GitHub Container Registry (ghcr.io)
#   4. 创建并推送 Git 标签（版本后缀如 -alpha.1, -beta.1）
#   5. 创建 GitHub Release 并自动生成发布说明
#
# Usage / 使用方法:
#   chmod +x scripts/release.sh
#   ./scripts/release.sh              # Interactive / 交互式
#   ./scripts/release.sh 1.0.1        # Non-interactive / 非交互式
#
#   Override release type via env / 通过环境变量覆盖发布类型:
#     RELEASE_TYPE=beta PRE_RELEASE_NUM=2 ./scripts/release.sh 1.0.1
#
#   Override language via env / 通过环境变量覆盖语言:
#     RELEASE_LANG=zh ./scripts/release.sh
#
# Prerequisites / 前置要求:
#   - jq, docker, git, gh (GitHub CLI)
#   - GitHub personal access token with write:packages scope
#     (export CR_PAT=ghp_xxx, or log in via docker login ghcr.io)
#   - gh CLI authenticated: gh auth login
# ============================================================

set -euo pipefail

# ─── Colors / 颜色 ────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${BLUE}[release]${NC} $1"; }
ok()   { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
err()  { echo -e "${RED}✗${NC} $1"; }

# ─── Configuration / 配置 ─────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PURECORE_JSON="$PROJECT_DIR/purecore.json"
LANG_FILE="$SCRIPT_DIR/.release-lang"
# ghcr.io target — derived from purecore.json "repository.url"
GHCR_OWNER=""
GHCR_REPO="purecore"
# Full repository name (e.g., "zhuchunshu/PureCore")
GITHUB_REPO=""
# Owner in lowercase (ghcr.io requires lowercase)
OWNER_LOWER=""
# Current language / 当前语言
LANG=""

# ─── Translation system / 翻译系统 ─────────────────────────
# Each t_* function checks $LANG at runtime and returns the
# appropriate string. English is the default.
# 每个 t_* 函数在运行时检查 $LANG 并返回相应的字符串。
# 默认为英文。

t_select_release_type_title() { case "${LANG:-en}" in zh) echo "选择发布类型:" ;; *) echo "Select release type:" ;; esac; }
t_alpha_desc() { case "${LANG:-en}" in zh) echo "alpha   — 内部测试预发布" ;; *) echo "alpha   — pre-release for internal testing" ;; esac; }
t_beta_desc() { case "${LANG:-en}" in zh) echo "beta    — 公开测试预发布" ;; *) echo "beta    — pre-release for public testing" ;; esac; }
t_stable_desc() { case "${LANG:-en}" in zh) echo "stable  — 正式生产发布" ;; *) echo "stable  — official production release" ;; esac; }
t_enter_choice() { case "${LANG:-en}" in zh) echo "请输入选项 [1-3]: " ;; *) echo "Enter choice [1-3]: " ;; esac; }
t_invalid_choice() { case "${LANG:-en}" in zh) echo "无效选项。请输入 1、2 或 3。" ;; *) echo "Invalid choice. Please enter 1, 2, or 3." ;; esac; }
t_release_type_label() { case "${LANG:-en}" in zh) echo "发布类型" ;; *) echo "Release type" ;; esac; }
t_via_env() { case "${LANG:-en}" in zh) echo "通过环境变量/覆盖" ;; *) echo "via env/override" ;; esac; }
t_current_version() { case "${LANG:-en}" in zh) echo "当前版本" ;; *) echo "Current version" ;; esac; }
t_enter_base_version() { case "${LANG:-en}" in zh) echo "输入新基础版本号" ;; *) echo "Enter new base version" ;; esac; }
t_invalid_version() { case "${LANG:-en}" in zh) echo "无效的版本格式" ;; *) echo "Invalid version format" ;; esac; }
t_expected_semver() { case "${LANG:-en}" in zh) echo "应为 X.Y.Z 格式" ;; *) echo "expected X.Y.Z" ;; esac; }
t_base_version_label() { case "${LANG:-en}" in zh) echo "基础版本" ;; *) echo "Base version" ;; esac; }
t_enter_pre_release_num() { case "${LANG:-en}" in zh) echo "输入预发布编号" ;; *) echo "Enter pre-release number" ;; esac; }
t_invalid_pre_release() { case "${LANG:-en}" in zh) echo "无效的预发布编号" ;; *) echo "Invalid pre-release number" ;; esac; }
t_expected_positive_int() { case "${LANG:-en}" in zh) echo "应为正整数" ;; *) echo "expected positive integer" ;; esac; }
t_version_bumped() { case "${LANG:-en}" in zh) echo "版本已更新" ;; *) echo "Version bumped" ;; esac; }
t_type() { case "${LANG:-en}" in zh) echo "类型" ;; *) echo "type" ;; esac; }
t_github_repo_label() { case "${LANG:-en}" in zh) echo "GitHub 仓库" ;; *) echo "GitHub repository" ;; esac; }
t_ghcr_target_label() { case "${LANG:-en}" in zh) echo "GHCR 目标" ;; *) echo "GHCR target" ;; esac; }
t_prereqs_passed() { case "${LANG:-en}" in zh) echo "前置检查通过" ;; *) echo "Prerequisites check passed" ;; esac; }
t_missing_tools() { case "${LANG:-en}" in zh) echo "缺少必需工具" ;; *) echo "Missing required tools" ;; esac; }
t_install_gh() { case "${LANG:-en}" in zh) echo "安装 gh (GitHub CLI): https://cli.github.com/" ;; *) echo "Install gh (GitHub CLI): https://cli.github.com/" ;; esac; }
t_then_auth() { case "${LANG:-en}" in zh) echo "然后认证: gh auth login" ;; *) echo "Then authenticate: gh auth login" ;; esac; }
t_gh_not_auth() { case "${LANG:-en}" in zh) echo "GitHub CLI (gh) 未认证。" ;; *) echo "GitHub CLI (gh) is not authenticated." ;; esac; }
t_run_gh_auth() { case "${LANG:-en}" in zh) echo "运行: gh auth login" ;; *) echo "Run: gh auth login" ;; esac; }
t_purecore_json_not_found() { case "${LANG:-en}" in zh) echo "未找到 purecore.json 文件，路径" ;; *) echo "purecore.json not found at" ;; esac; }
t_no_repo_url() { case "${LANG:-en}" in zh) echo "在 purecore.json 中未找到 repository.url" ;; *) echo "Could not find repository.url in purecore.json" ;; esac; }
t_no_parse_owner() { case "${LANG:-en}" in zh) echo "无法从仓库 URL 解析 GitHub 所有者" ;; *) echo "Could not parse GitHub owner from repository URL" ;; esac; }
t_no_parse_repo() { case "${LANG:-en}" in zh) echo "无法从 URL 解析 GitHub 仓库" ;; *) echo "Could not parse GitHub repository from URL" ;; esac; }
t_logging_in_ghcr() { case "${LANG:-en}" in zh) echo "正在使用 CR_PAT 登录 ghcr.io..." ;; *) echo "Logging in to ghcr.io with CR_PAT..." ;; esac; }
t_logged_in_ghcr() { case "${LANG:-en}" in zh) echo "已登录 ghcr.io" ;; *) echo "Logged in to ghcr.io" ;; esac; }
t_already_auth_ghcr() { case "${LANG:-en}" in zh) echo "已认证到 ghcr.io" ;; *) echo "Already authenticated to ghcr.io" ;; esac; }
t_not_logged_in_ghcr() { case "${LANG:-en}" in zh) echo "未登录 ghcr.io。请运行:" ;; *) echo "Not logged in to ghcr.io. Please run:" ;; esac; }
t_press_enter_after_login() { case "${LANG:-en}" in zh) echo "登录后按 Enter 继续 (或 Ctrl+C 取消)... " ;; *) echo "Press Enter after logging in (or Ctrl+C to abort)... " ;; esac; }
t_building_docker() { case "${LANG:-en}" in zh) echo "正在构建 Docker 镜像..." ;; *) echo "Building Docker images..." ;; esac; }
t_tagging_backend() { case "${LANG:-en}" in zh) echo "正在为后端打标签" ;; *) echo "Tagging backend" ;; esac; }
t_tagging_frontend() { case "${LANG:-en}" in zh) echo "正在为前端打标签" ;; *) echo "Tagging frontend" ;; esac; }
t_tagging_latest() { case "${LANG:-en}" in zh) echo "正在打 latest 别名标签 (稳定版)..." ;; *) echo "Tagging latest aliases (stable release)..." ;; esac; }
t_pushing_to_ghcr() { case "${LANG:-en}" in zh) echo "正在推送镜像到 ghcr.io..." ;; *) echo "Pushing images to ghcr.io..." ;; esac; }
t_all_images_pushed() { case "${LANG:-en}" in zh) echo "所有镜像已推送到 ghcr.io" ;; *) echo "All images pushed to ghcr.io" ;; esac; }
t_creating_git_tag() { case "${LANG:-en}" in zh) echo "正在创建 Git 标签" ;; *) echo "Creating Git tag" ;; esac; }
t_no_changes() { case "${LANG:-en}" in zh) echo "没有要提交的更改 (版本已是最新)" ;; *) echo "No changes to commit (version already up to date)" ;; esac; }
t_committed_bump() { case "${LANG:-en}" in zh) echo "已提交版本更新" ;; *) echo "Committed version bump" ;; esac; }
t_commit_bump_msg() { case "${LANG:-en}" in zh) echo "chore: 更新版本至" ;; *) echo "chore: bump version to" ;; esac; }
t_release_tag_msg() { case "${LANG:-en}" in zh) echo "发布" ;; *) echo "Release" ;; esac; }
t_tag_exists_locally() { case "${LANG:-en}" in zh) echo "标签在本地已存在" ;; *) echo "Tag already exists locally" ;; esac; }
t_tag_exists_remote() { case "${LANG:-en}" in zh) echo "标签在远程已存在，跳过创建" ;; *) echo "Tag already exists on remote, skipping tag creation" ;; esac; }
t_pushing_existing_tag() { case "${LANG:-en}" in zh) echo "正在推送本地已有标签到远程..." ;; *) echo "Pushing existing local tag to remote..." ;; esac; }
t_pushed_existing_tag() { case "${LANG:-en}" in zh) echo "已推送已有标签" ;; *) echo "Pushed existing tag" ;; esac; }
t_pushing_commits() { case "${LANG:-en}" in zh) echo "正在推送提交和标签到 origin..." ;; *) echo "Pushing commits and tags to origin..." ;; esac; }
t_git_tag_pushed() { case "${LANG:-en}" in zh) echo "Git 标签已推送" ;; *) echo "Git tag pushed" ;; esac; }
t_creating_gh_release() { case "${LANG:-en}" in zh) echo "正在创建 GitHub Release" ;; *) echo "Creating GitHub Release for" ;; esac; }
t_release_title() { echo "PureCore"; }
t_alpha_title_suffix() { case "${LANG:-en}" in zh) echo "(Alpha)" ;; *) echo "(Alpha)" ;; esac; }
t_beta_title_suffix() { case "${LANG:-en}" in zh) echo "(Beta)" ;; *) echo "(Beta)" ;; esac; }
t_docker_images_header() { case "${LANG:-en}" in zh) echo "🐳 Docker 镜像" ;; *) echo "🐳 Docker Images" ;; esac; }
t_deploy_header() { case "${LANG:-en}" in zh) echo "🚀 部署" ;; *) echo "🚀 Deploy" ;; esac; }
t_changes_header() { case "${LANG:-en}" in zh) echo "📝 自上次以来的变更" ;; *) echo "📝 Changes since" ;; esac; }
t_release_created() { case "${LANG:-en}" in zh) echo "GitHub Release 已创建" ;; *) echo "GitHub Release created" ;; esac; }
t_marked_prerelease() { case "${LANG:-en}" in zh) echo "已标记为预发布" ;; *) echo "Marked as pre-release" ;; esac; }
t_failed_gh_release() { case "${LANG:-en}" in zh) echo "创建 GitHub Release 失败" ;; *) echo "Failed to create GitHub Release" ;; esac; }
t_create_manually_at() { case "${LANG:-en}" in zh) echo "你可以手动创建于" ;; *) echo "You can create it manually at" ;; esac; }
t_or_run() { case "${LANG:-en}" in zh) echo "或运行" ;; *) echo "Or run" ;; esac; }
t_using_existing_tag() { case "${LANG:-en}" in zh) echo "使用已有远程标签 — 发布说明将从标签消息中获取" ;; *) echo "Using existing remote tag — release notes will be from tag message" ;; esac; }
t_docker_images_summary() { case "${LANG:-en}" in zh) echo "🐳 Docker 镜像" ;; *) echo "🐳 Docker Images" ;; esac; }
t_git_tag_summary() { case "${LANG:-en}" in zh) echo "🏷️  Git 标签" ;; *) echo "🏷️  Git tag" ;; esac; }
t_github_release_summary() { case "${LANG:-en}" in zh) echo "📦 GitHub Release" ;; *) echo "📦 GitHub Release" ;; esac; }
t_deploy_summary() { case "${LANG:-en}" in zh) echo "🚀 在服务器上部署" ;; *) echo "🚀 Deploy on server" ;; esac; }

# ─── Language selection / 语言选择 ─────────────────────────
detect_or_select_language() {
  # If already set via env, use it
  if [ -n "${RELEASE_LANG:-}" ]; then
    LANG="${RELEASE_LANG}"
    return
  fi

  # Check for saved preference
  if [ -f "$LANG_FILE" ]; then
    LANG=$(cat "$LANG_FILE")
    if [ "$LANG" = "en" ] || [ "$LANG" = "zh" ]; then
      return
    fi
  fi

  # First run — ask user to choose language
  echo ""
  echo "  Please select language / 请选择语言："
  echo "    [1] English"
  echo "    [2] 中文"
  echo ""

  local choice
  while true; do
    printf "  Enter choice / 请输入选项 [1-2]: "
    read -r choice
    case "$choice" in
      1) LANG="en"; break ;;
      2) LANG="zh"; break ;;
      *) echo -e "  ${YELLOW}Invalid / 无效${NC}" ;;
    esac
  done

  # Save preference
  mkdir -p "$(dirname "$LANG_FILE")"
  echo "$LANG" > "$LANG_FILE"

  if [ "$LANG" = "zh" ]; then
    echo ""
    echo -e "  ${GREEN}✓${NC} 语言已设置为中文 (保存在 ${LANG_FILE})"
    echo -e "    下次运行将自动使用中文。可通过设置 RELEASE_LANG=en 覆盖。"
  else
    echo ""
    echo -e "  ${GREEN}✓${NC} Language set to English (saved to ${LANG_FILE})"
    echo -e "    Will auto-detect on next run. Override with RELEASE_LANG=zh."
  fi
  echo ""
}

# ─── Select release type / 选择发布类型 ────────────────────
select_release_type() {
  echo ""
  echo -e "${BLUE}$(t select_release_type_title)${NC}"
  echo "  1) $(t alpha_desc)"
  echo "  2) $(t beta_desc)"
  echo "  3) $(t stable_desc)"
  echo ""

  local choice
  while true; do
    printf "$(t enter_choice)"
    read -r choice
    case "$choice" in
      1) RELEASE_TYPE="alpha"; break ;;
      2) RELEASE_TYPE="beta"; break ;;
      3) RELEASE_TYPE="stable"; break ;;
      *) warn "$(t invalid_choice)" ;;
    esac
  done

  echo ""
  ok "$(t release_type_label): ${YELLOW}${RELEASE_TYPE}${NC}"
}

# ─── Check prerequisites / 检查前置条件 ────────────────────
check_prereqs() {
  local missing=()
  for tool in jq docker git gh; do
    if ! command -v "$tool" &>/dev/null; then
      missing+=("$tool")
    fi
  done
  if [ ${#missing[@]} -gt 0 ]; then
    err "$(t missing_tools): ${missing[*]}"
    echo ""
    echo "  $(t install_gh)"
    echo "  $(t then_auth)"
    exit 1
  fi

  # Verify gh is authenticated
  if ! gh auth status &>/dev/null 2>&1; then
    err "$(t gh_not_auth)"
    echo "  $(t run_gh_auth)"
    exit 1
  fi

  if [ ! -f "$PURECORE_JSON" ]; then
    err "$(t purecore_json_not_found) $PURECORE_JSON"
    exit 1
  fi

  # Derive GHCR_OWNER and GITHUB_REPO from repository URL in purecore.json
  local repo_url
  repo_url=$(jq -r '.repository.url // empty' "$PURECORE_JSON")
  if [ -z "$repo_url" ]; then
    err "$(t no_repo_url)"
    exit 1
  fi
  GHCR_OWNER=$(echo "$repo_url" | sed -n 's|.*github\.com/\([^/]*\)/.*|\1|p')
  if [ -z "$GHCR_OWNER" ]; then
    err "$(t no_parse_owner): $repo_url"
    exit 1
  fi

  GITHUB_REPO=$(echo "$repo_url" | sed -n 's|.*github\.com/\([^/]*/[^/]*\)\.git|\1|p')
  if [ -z "$GITHUB_REPO" ]; then
    err "$(t no_parse_repo): $repo_url"
    exit 1
  fi

  OWNER_LOWER="${GHCR_OWNER,,}"

  log "$(t github_repo_label): ${CYAN}${GITHUB_REPO}${NC}"
  log "$(t ghcr_target_label): ${CYAN}ghcr.io/${OWNER_LOWER}/${GHCR_REPO}${NC}"
  ok "$(t prereqs_passed)"
}

# ─── Bump version / 更新版本号 ─────────────────────────────
bump_version() {
  local old_version new_version input_version
  local base_version pre_release_num

  old_version=$(jq -r '.version // "0.0.0"' "$PURECORE_JSON")

  if [ -n "${1:-}" ]; then
    base_version="$1"
  else
    local suggested
    suggested=$(echo "$old_version" | awk -F. '{printf "%d.%d.%d", $1, $2, $3+1}')

    echo ""
    log "$(t current_version): ${YELLOW}$old_version${NC}"
    printf "$(t enter_base_version) [%s]: " "$suggested"
    read -r input_version
    base_version="${input_version:-$suggested}"
  fi

  if ! echo "$base_version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    err "$(t invalid_version): $base_version ($(t expected_semver))"
    exit 1
  fi

  if [ "$RELEASE_TYPE" = "stable" ]; then
    new_version="$base_version"
  else
    if [ -n "${PRE_RELEASE_NUM:-}" ]; then
      pre_release_num="$PRE_RELEASE_NUM"
    else
      local existing_pre
      existing_pre=$(git tag -l "v${base_version}-${RELEASE_TYPE}.*" 2>/dev/null | sort -V | tail -1 | sed "s/.*${RELEASE_TYPE}\.//")
      local suggested_num
      suggested_num=$(( existing_pre + 1 ))
      if [ -z "$existing_pre" ] || [ "$suggested_num" -le 0 ]; then
        suggested_num=1
      fi

      echo ""
      log "$(t base_version_label): ${YELLOW}$base_version${NC}"
      log "$(t release_type_label): ${YELLOW}$RELEASE_TYPE${NC}"
      printf "$(t enter_pre_release_num) [%d]: " "$suggested_num"
      read -r pre_release_num
      pre_release_num="${pre_release_num:-$suggested_num}"
    fi

    if ! echo "$pre_release_num" | grep -qE '^[1-9][0-9]*$'; then
      err "$(t invalid_pre_release): $pre_release_num ($(t expected_positive_int))"
      exit 1
    fi

    new_version="${base_version}-${RELEASE_TYPE}.${pre_release_num}"
  fi

  jq --arg v "$new_version" --arg rt "$RELEASE_TYPE" \
    '.version = $v | .release_type = $rt' \
    "$PURECORE_JSON" > "${PURECORE_JSON}.tmp" \
    && mv "${PURECORE_JSON}.tmp" "$PURECORE_JSON"

  ok "$(t version_bumped): $old_version → $new_version ($(t type): $RELEASE_TYPE)"

  NEW_VERSION="$new_version"
  TAG_NAME="v${new_version}"
}

# ─── Docker login to ghcr.io / Docker 登录 ghcr.io ────────
docker_login() {
  if [ -n "${CR_PAT:-}" ]; then
    log "$(t logging_in_ghcr)"
    echo "$CR_PAT" | docker login ghcr.io -u "$GHCR_OWNER" --password-stdin 2>/dev/null
    ok "$(t logged_in_ghcr)"
  elif docker pull ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:latest &>/dev/null 2>&1; then
    ok "$(t already_auth_ghcr)"
  else
    warn "$(t not_logged_in_ghcr)"
    echo "  export CR_PAT=ghp_xxxxxxxxxxxx"
    echo "  echo \$CR_PAT | docker login ghcr.io -u $GHCR_OWNER --password-stdin"
    echo ""
    read -rp "$(t press_enter_after_login)"
  fi
}

# ─── Build and push Docker images / 构建推送 Docker 镜像 ───
build_and_push() {
  local version="$1"

  local backend_image="ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:${version}"
  local frontend_image="ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:${version}"
  local backend_latest="ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:latest"
  local frontend_latest="ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:latest"

  log "$(t building_docker)"
  cd "$PROJECT_DIR"

  docker compose -f docker-compose.dev.yml build --pull

  local backend_built="purecore-backend"
  log "$(t tagging_backend): $backend_image"
  docker tag "${backend_built}:latest" "$backend_image"

  local frontend_built="purecore-frontend"
  log "$(t tagging_frontend): $frontend_image"
  docker tag "${frontend_built}:latest" "$frontend_image"

  if [ "$RELEASE_TYPE" = "stable" ]; then
    log "$(t tagging_latest)"
    docker tag "${backend_built}:latest" "$backend_latest"
    docker tag "${frontend_built}:latest" "$frontend_latest"
  fi

  log "$(t pushing_to_ghcr)"
  echo "  → $backend_image"
  docker push "$backend_image"
  echo "  → $frontend_image"
  docker push "$frontend_image"

  if [ "$RELEASE_TYPE" = "stable" ]; then
    echo "  → $backend_latest"
    docker push "$backend_latest"
    echo "  → $frontend_latest"
    docker push "$frontend_latest"
  fi

  ok "$(t all_images_pushed)"
}

# ─── Git tag and push / Git 标签与推送 ─────────────────────
git_tag_and_push() {
  local version="$1"
  local tag_name="v${version}"

  log "$(t creating_git_tag): ${CYAN}${tag_name}${NC}"

  cd "$PROJECT_DIR"

  if ! git diff --quiet purecore.json 2>/dev/null; then
    local commit_msg="$(t commit_bump_msg) ${version}"
    if [ "$RELEASE_TYPE" != "stable" ]; then
      commit_msg="$(t commit_bump_msg) ${version} (${RELEASE_TYPE} release)"
    fi
    git add purecore.json
    git commit -m "$commit_msg"
    ok "$(t committed_bump)"
  else
    log "$(t no_changes)"
  fi

  if git rev-parse "$tag_name" >/dev/null 2>&1; then
    warn "$(t tag_exists_locally): $tag_name"
    if git ls-remote --tags origin "$tag_name" | grep -q "$tag_name"; then
      warn "$(t tag_exists_remote)"
      TAG_EXISTS_REMOTE=true
      return
    fi
    log "$(t pushing_existing_tag)"
    git push origin "$tag_name"
    ok "$(t pushed_existing_tag): $tag_name"
    TAG_EXISTS_REMOTE=false
    return
  fi

  local tag_msg="$(t release_tag_msg) ${tag_name}"
  if [ "$RELEASE_TYPE" != "stable" ]; then
    tag_msg="$(t release_tag_msg) ${tag_name} (${RELEASE_TYPE})"
  fi
  git tag -a "$tag_name" -m "$tag_msg"
  ok "Created tag: $tag_name"

  log "$(t pushing_commits)"
  git push origin HEAD
  git push origin "$tag_name"

  ok "$(t git_tag_pushed): $tag_name"
  TAG_EXISTS_REMOTE=false
}

# ─── Create GitHub Release / 创建 GitHub Release ───────────
create_github_release() {
  local version="$1"
  local tag_name="v${version}"
  local is_pre_release="false"

  [ "$RELEASE_TYPE" != "stable" ] && is_pre_release="true"

  log "$(t creating_gh_release) ${CYAN}${tag_name}${NC}..."

  local release_title="$(t release_title) v${version}"
  [ "$RELEASE_TYPE" = "alpha" ] && release_title="$(t release_title) v${version} $(t alpha_title_suffix)"
  [ "$RELEASE_TYPE" = "beta" ]  && release_title="$(t release_title) v${version} $(t beta_title_suffix)"

  local release_notes="## $(t docker_images_header)

- \`ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:${version}\`
- \`ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:${version}\`
"

  if [ "$RELEASE_TYPE" = "stable" ]; then
    release_notes+="
- \`ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:latest\`
- \`ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:latest\`
"
  fi

  release_notes+="
## $(t deploy_header)

\`\`\`bash
export PURECORE_VERSION=${version}
docker compose pull
docker compose up -d
\`\`\`
"

  local last_tag
  last_tag=$(git describe --tags --abbrev=0 "v${version}" 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null || echo "")
  if [ -n "$last_tag" ] && [ "$last_tag" != "$tag_name" ]; then
    local changelog
    changelog=$(git log "${last_tag}..HEAD" --pretty=format:"- %s (%an)" 2>/dev/null || echo "")
    if [ -n "$changelog" ]; then
      release_notes+="
## $(t changes_header) ${last_tag}

${changelog}
"
    fi
  fi

  local gh_args=(
    release create "$tag_name"
    --repo "$GITHUB_REPO"
    --title "$release_title"
    --notes "$release_notes"
  )

  [ "$is_pre_release" = "true" ] && gh_args+=(--prerelease)

  if [ "${TAG_EXISTS_REMOTE:-false}" = "true" ]; then
    gh_args+=(--notes-from-tag)
    warn "$(t using_existing_tag)"
  fi

  if gh "${gh_args[@]}" 2>&1; then
    ok "$(t release_created): ${CYAN}${release_title}${NC}"
    [ "$is_pre_release" = "true" ] && ok "$(t marked_prerelease)"
  else
    err "$(t failed_gh_release)"
    echo "  $(t create_manually_at):"
    echo "  https://github.com/${GITHUB_REPO}/releases/new?tag=${tag_name}"
    echo ""
    echo "  $(t or_run):"
    echo "  gh release create ${tag_name} --repo ${GITHUB_REPO} --title \"${release_title}\" --notes \"${release_notes}\" --prerelease"
  fi
}

# ─── Summary / 摘要 ────────────────────────────────────────
print_summary() {
  local version="$1"

  echo ""
  echo -e "${BLUE}============================================${NC}"
  echo -e "${BLUE}  PureCore Release ${YELLOW}v${version}${NC}"
  if [ "$RELEASE_TYPE" != "stable" ]; then
    echo -e "${BLUE}  Type: ${YELLOW}${RELEASE_TYPE}${NC}"
  fi
  echo -e "${BLUE}============================================${NC}"
  echo ""
  echo "  $(t docker_images_summary):"
  echo "    ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:${version}"
  echo "    ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:${version}"
  if [ "$RELEASE_TYPE" = "stable" ]; then
    echo "    ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:latest"
    echo "    ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:latest"
  fi
  echo ""
  echo "  $(t git_tag_summary): v${version}"
  echo ""
  echo "  $(t github_release_summary):"
  echo "    https://github.com/${GITHUB_REPO}/releases/tag/v${version}"
  echo ""
  echo "  $(t deploy_summary):"
  echo "    export PURECORE_VERSION=${version}"
  echo "    docker compose pull"
  echo "    docker compose up -d"
  echo ""
}

# ─── Main / 主函数 ─────────────────────────────────────────
main() {
  detect_or_select_language

  check_prereqs

  if [ -z "${1:-}" ] && [ -z "${RELEASE_TYPE_OVERRIDE:-}" ]; then
    select_release_type
  else
    RELEASE_TYPE="${RELEASE_TYPE:-stable}"
    log "$(t release_type_label): ${YELLOW}${RELEASE_TYPE}${NC} ($(t via_env))"
  fi

  bump_version "${1:-}"
  local version="$NEW_VERSION"

  docker_login

  build_and_push "$version"

  git_tag_and_push "$version"

  create_github_release "$version"

  print_summary "$version"
}

main "$@"
