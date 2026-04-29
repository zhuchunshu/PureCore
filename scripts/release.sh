#!/usr/bin/env bash
# ============================================================
# PureCore — Release Script
#
# Automates a new version release:
#   1. Bump version in purecore.json (interactive, auto-increments patch)
#   2. Build Docker images via docker compose
#   3. Tag and push images to GitHub Container Registry (ghcr.io)
#   4. Create and push a Git tag
#
# Usage:
#   chmod +x scripts/release.sh
#   ./scripts/release.sh              # Interactive
#   ./scripts/release.sh 1.0.1        # Non-interactive: use given version
#
# Prerequisites:
#   - jq, docker, git
#   - GitHub personal access token with write:packages scope
#     (export CR_PAT=ghp_xxx, or log in via docker login ghcr.io)
# ============================================================

set -euo pipefail

# ─── Colors ────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log()  { echo -e "${BLUE}[release]${NC} $1"; }
ok()   { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
err()  { echo -e "${RED}✗${NC} $1"; }

# ─── Configuration ─────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PURECORE_JSON="$PROJECT_DIR/purecore.json"
# ghcr.io target — derived from purecore.json "repository.url"
# Example: https://github.com/zhuchunshu/PureCore.git → ghcr.io/zhuchunshu/purecore
GHCR_OWNER=""
GHCR_REPO="purecore"

# ─── Select release type ───────────────────────────────────
select_release_type() {
  echo ""
  echo -e "${BLUE}Select release type:${NC}"
  echo "  1) alpha   — pre-release for internal testing"
  echo "  2) beta    — pre-release for public testing"
  echo "  3) stable  — official production release"
  echo ""

  local choice
  while true; do
    printf "Enter choice [1-3]: "
    read -r choice
    case "$choice" in
      1) RELEASE_TYPE="alpha"; break ;;
      2) RELEASE_TYPE="beta"; break ;;
      3) RELEASE_TYPE="stable"; break ;;
      *) warn "Invalid choice. Please enter 1, 2, or 3." ;;
    esac
  done

  echo ""
  ok "Release type: ${YELLOW}${RELEASE_TYPE}${NC}"
}

# ─── Check prerequisites ───────────────────────────────────
check_prereqs() {
  local missing=()
  for tool in jq docker git; do
    if ! command -v "$tool" &>/dev/null; then
      missing+=("$tool")
    fi
  done
  if [ ${#missing[@]} -gt 0 ]; then
    err "Missing required tools: ${missing[*]}"
    exit 1
  fi

  if [ ! -f "$PURECORE_JSON" ]; then
    err "purecore.json not found at $PURECORE_JSON"
    exit 1
  fi

  # Derive GHCR_OWNER from repository URL in purecore.json
  local repo_url
  repo_url=$(jq -r '.repository.url // empty' "$PURECORE_JSON")
  if [ -z "$repo_url" ]; then
    err "Could not find repository.url in purecore.json"
    exit 1
  fi
  # Extract owner from GitHub URL: https://github.com/OWNER/REPO.git → OWNER
  GHCR_OWNER=$(echo "$repo_url" | sed -n 's|.*github\.com/\([^/]*\)/.*|\1|p')
  if [ -z "$GHCR_OWNER" ]; then
    err "Could not parse GitHub owner from repository URL: $repo_url"
    exit 1
  fi

  log "GHCR target: ghcr.io/${GHCR_OWNER}/${GHCR_REPO}"
  ok "Prerequisites check passed"
}

# ─── Bump version ──────────────────────────────────────────
bump_version() {
  local old_version new_version input_version
  local base_version pre_release_num

  old_version=$(jq -r '.version // "0.0.0"' "$PURECORE_JSON")

  if [ -n "${1:-}" ]; then
    # Non-interactive mode — use provided version
    base_version="$1"
  else
    # Interactive mode
    # Auto-increment patch version as default suggestion
    local suggested
    suggested=$(echo "$old_version" | awk -F. '{printf "%d.%d.%d", $1, $2, $3+1}')

    echo ""
    log "Current version: ${YELLOW}$old_version${NC}"
    printf "Enter new base version [%s]: " "$suggested"
    read -r input_version
    base_version="${input_version:-$suggested}"
  fi

  # Basic semver validation
  if ! echo "$base_version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    err "Invalid version format: $base_version (expected X.Y.Z)"
    exit 1
  fi

  # Build final version based on release type
  if [ "$RELEASE_TYPE" = "stable" ]; then
    new_version="$base_version"
  else
    # For alpha/beta, ask for pre-release number
    if [ -n "${PRE_RELEASE_NUM:-}" ]; then
      pre_release_num="$PRE_RELEASE_NUM"
    else
      # Try to auto-detect the next pre-release number from existing tags
      local existing_pre
      existing_pre=$(git tag -l "v${base_version}-${RELEASE_TYPE}.*" 2>/dev/null | sort -V | tail -1 | sed "s/.*${RELEASE_TYPE}\.//")
      local suggested_num
      suggested_num=$(( existing_pre + 1 ))
      if [ -z "$existing_pre" ] || [ "$suggested_num" -le 0 ]; then
        suggested_num=1
      fi

      echo ""
      log "Base version: ${YELLOW}$base_version${NC}"
      log "Release type: ${YELLOW}$RELEASE_TYPE${NC}"
      printf "Enter pre-release number [%d]: " "$suggested_num"
      read -r pre_release_num
      pre_release_num="${pre_release_num:-$suggested_num}"
    fi

    # Validate pre-release number is a positive integer
    if ! echo "$pre_release_num" | grep -qE '^[1-9][0-9]*$'; then
      err "Invalid pre-release number: $pre_release_num (expected positive integer)"
      exit 1
    fi

    new_version="${base_version}-${RELEASE_TYPE}.${pre_release_num}"
  fi

  # Update purecore.json with the full version
  jq --arg v "$new_version" '.version = $v' "$PURECORE_JSON" > "${PURECORE_JSON}.tmp" \
    && mv "${PURECORE_JSON}.tmp" "$PURECORE_JSON"
  ok "Version bumped: $old_version → $new_version"

  NEW_VERSION="$new_version"
}

# ─── Docker login to ghcr.io ───────────────────────────────
docker_login() {
  # Check if already authenticated
  if echo "$GHCR_OWNER" | grep -q '[A-Z]'; then
    warn "GHCR_OWNER contains uppercase letters. Docker requires lowercase for ghcr.io."
    warn "Owner will be lowercased for image tags."
  fi

  # Try CR_PAT (GitHub Personal Access Token) first
  if [ -n "${CR_PAT:-}" ]; then
    log "Logging in to ghcr.io with CR_PAT..."
    echo "$CR_PAT" | docker login ghcr.io -u "$GHCR_OWNER" --password-stdin 2>/dev/null
    ok "Logged in to ghcr.io"
  elif docker pull ghcr.io/${GHCR_OWNER,,}/${GHCR_REPO}-backend:latest &>/dev/null 2>&1; then
    # Pull succeeded — we are likely already logged in
    ok "Already authenticated to ghcr.io"
  else
    warn "Not logged in to ghcr.io. Please run:"
    echo "  export CR_PAT=ghp_xxxxxxxxxxxx"
    echo "  echo \$CR_PAT | docker login ghcr.io -u $GHCR_OWNER --password-stdin"
    echo ""
    read -rp "Press Enter after logging in (or Ctrl+C to abort)... "
  fi
}

# ─── Build and push Docker images ──────────────────────────
build_and_push() {
  local version="$1"
  local owner_lower="${GHCR_OWNER,,}"  # ghcr.io requires lowercase

  local backend_image="ghcr.io/${owner_lower}/${GHCR_REPO}-backend:${version}"
  local frontend_image="ghcr.io/${owner_lower}/${GHCR_REPO}-frontend:${version}"
  local backend_latest="ghcr.io/${owner_lower}/${GHCR_REPO}-backend:latest"
  local frontend_latest="ghcr.io/${owner_lower}/${GHCR_REPO}-frontend:latest"

  log "Building Docker images..."
  cd "$PROJECT_DIR"

  # Build both images via docker-compose.dev.yml (local build), then tag for ghcr.io
  docker compose -f docker-compose.dev.yml build --pull

  # Tag backend
  local backend_built="purecore-backend"
  log "Tagging backend: $backend_image"
  docker tag "${backend_built}:latest" "$backend_image"

  # Tag frontend
  local frontend_built="purecore-frontend"
  log "Tagging frontend: $frontend_image"
  docker tag "${frontend_built}:latest" "$frontend_image"

  # Only tag :latest for stable releases
  if [ "$RELEASE_TYPE" = "stable" ]; then
    log "Tagging latest aliases (stable release)..."
    docker tag "${backend_built}:latest" "$backend_latest"
    docker tag "${frontend_built}:latest" "$frontend_latest"
  fi

  # Push tagged images
  log "Pushing images to ghcr.io..."
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

  ok "All images pushed to ghcr.io"
}

# ─── Git tag and push ──────────────────────────────────────
git_tag_and_push() {
  local version="$1"
  local tag_name="v${version}"

  log "Creating Git tag: $tag_name"

  cd "$PROJECT_DIR"

  # Commit purecore.json version bump
  if ! git diff --quiet purecore.json 2>/dev/null; then
    local commit_msg="chore: bump version to ${version}"
    if [ "$RELEASE_TYPE" != "stable" ]; then
      commit_msg="chore: bump version to ${version} (${RELEASE_TYPE} release)"
    fi
    git add purecore.json
    git commit -m "$commit_msg"
    ok "Committed version bump"
  else
    log "No changes to commit (version already up to date)"
  fi

  # Create and push tag
  if git rev-parse "$tag_name" >/dev/null 2>&1; then
    warn "Tag $tag_name already exists locally, skipping tag creation"
  else
    local tag_msg="Release ${tag_name}"
    if [ "$RELEASE_TYPE" != "stable" ]; then
      tag_msg="Release ${tag_name} (${RELEASE_TYPE})"
    fi
    git tag -a "$tag_name" -m "$tag_msg"
    ok "Created tag: $tag_name"
  fi

  log "Pushing commits and tags to origin..."
  git push origin HEAD
  git push origin "$tag_name"

  ok "Git tag pushed: $tag_name"
}

# ─── Summary ───────────────────────────────────────────────
print_summary() {
  local version="$1"
  local owner_lower="${GHCR_OWNER,,}"

  echo ""
  echo -e "${BLUE}============================================${NC}"
  echo -e "${BLUE}  PureCore Release ${YELLOW}v${version}${NC}"
  if [ "$RELEASE_TYPE" != "stable" ]; then
    echo -e "${BLUE}  Type: ${YELLOW}${RELEASE_TYPE}${NC}"
  fi
  echo -e "${BLUE}============================================${NC}"
  echo ""
  echo "  Images:"
  echo "    ghcr.io/${owner_lower}/${GHCR_REPO}-backend:${version}"
  echo "    ghcr.io/${owner_lower}/${GHCR_REPO}-frontend:${version}"
  if [ "$RELEASE_TYPE" = "stable" ]; then
    echo "    ghcr.io/${owner_lower}/${GHCR_REPO}-backend:latest"
    echo "    ghcr.io/${owner_lower}/${GHCR_REPO}-frontend:latest"
  fi
  echo ""
  echo "  Git tag: v${version}"
  echo ""
  echo "  Pull on server:"
  echo "    docker pull ghcr.io/${owner_lower}/${GHCR_REPO}-backend:${version}"
  echo "    docker pull ghcr.io/${owner_lower}/${GHCR_REPO}-frontend:${version}"
  echo ""
}

# ─── Main ──────────────────────────────────────────────────
main() {
  check_prereqs

  # Step 1: Select release type (skip in non-interactive mode if RELEASE_TYPE is set)
  if [ -z "${1:-}" ] && [ -z "${RELEASE_TYPE_OVERRIDE:-}" ]; then
    select_release_type
  else
    RELEASE_TYPE="${RELEASE_TYPE:-stable}"
    log "Release type: ${YELLOW}${RELEASE_TYPE}${NC} (via env/override)"
  fi

  # Step 2: Bump version
  bump_version "${1:-}"
  local version="$NEW_VERSION"

  # Step 3: Login to ghcr.io
  docker_login

  # Step 4: Build and push Docker images
  build_and_push "$version"

  # Step 5: Git tag and push
  git_tag_and_push "$version"

  # Step 6: Summary
  print_summary "$version"
}

main "$@"
