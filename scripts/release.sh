#!/usr/bin/env bash
# ============================================================
# PureCore — Release Script
#
# Automates a new version release:
#   1. Bump version in purecore.json (interactive, auto-increments patch)
#   2. Build Docker images via docker compose
#   3. Tag and push images to GitHub Container Registry (ghcr.io)
#   4. Create and push a Git tag (version suffix like -alpha.1, -beta.1)
#   5. Create a GitHub Release with auto-generated release notes
#
# Usage:
#   chmod +x scripts/release.sh
#   ./scripts/release.sh              # Interactive
#   ./scripts/release.sh 1.0.1        # Non-interactive: use given version
#
#   Override release type via env:
#     RELEASE_TYPE=beta PRE_RELEASE_NUM=2 ./scripts/release.sh 1.0.1
#
# Prerequisites:
#   - jq, docker, git, gh (GitHub CLI)
#   - GitHub personal access token with write:packages scope
#     (export CR_PAT=ghp_xxx, or log in via docker login ghcr.io)
#   - gh CLI authenticated: gh auth login
# ============================================================

set -euo pipefail

# ─── Colors ────────────────────────────────────────────────
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

# ─── Configuration ─────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PURECORE_JSON="$PROJECT_DIR/purecore.json"
# ghcr.io target — derived from purecore.json "repository.url"
# Example: https://github.com/zhuchunshu/PureCore.git → ghcr.io/zhuchunshu/purecore
GHCR_OWNER=""
GHCR_REPO="purecore"
# Full repository name (e.g., "zhuchunshu/PureCore")
GITHUB_REPO=""
# Owner in lowercase (ghcr.io requires lowercase)
OWNER_LOWER=""

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
  for tool in jq docker git gh; do
    if ! command -v "$tool" &>/dev/null; then
      missing+=("$tool")
    fi
  done
  if [ ${#missing[@]} -gt 0 ]; then
    err "Missing required tools: ${missing[*]}"
    echo ""
    echo "  Install gh (GitHub CLI): https://cli.github.com/"
    echo "  Then authenticate: gh auth login"
    exit 1
  fi

  # Verify gh is authenticated
  if ! gh auth status &>/dev/null 2>&1; then
    err "GitHub CLI (gh) is not authenticated."
    echo "  Run: gh auth login"
    exit 1
  fi

  if [ ! -f "$PURECORE_JSON" ]; then
    err "purecore.json not found at $PURECORE_JSON"
    exit 1
  fi

  # Derive GHCR_OWNER and GITHUB_REPO from repository URL in purecore.json
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

  # Extract full repo name: https://github.com/OWNER/REPO.git → OWNER/REPO
  GITHUB_REPO=$(echo "$repo_url" | sed -n 's|.*github\.com/\([^/]*/[^/]*\)\.git|\1|p')
  if [ -z "$GITHUB_REPO" ]; then
    err "Could not parse GitHub repository from URL: $repo_url"
    exit 1
  fi

  OWNER_LOWER="${GHCR_OWNER,,}"

  log "GitHub repository: ${CYAN}${GITHUB_REPO}${NC}"
  log "GHCR target: ${CYAN}ghcr.io/${OWNER_LOWER}/${GHCR_REPO}${NC}"
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

  # Also update release_type in purecore.json
  jq --arg v "$new_version" --arg rt "$RELEASE_TYPE" \
    '.version = $v | .release_type = $rt' \
    "$PURECORE_JSON" > "${PURECORE_JSON}.tmp" \
    && mv "${PURECORE_JSON}.tmp" "$PURECORE_JSON"

  ok "Version bumped: $old_version → $new_version (type: $RELEASE_TYPE)"

  NEW_VERSION="$new_version"
  TAG_NAME="v${new_version}"
}

# ─── Docker login to ghcr.io ───────────────────────────────
docker_login() {
  # Try CR_PAT (GitHub Personal Access Token) first
  if [ -n "${CR_PAT:-}" ]; then
    log "Logging in to ghcr.io with CR_PAT..."
    echo "$CR_PAT" | docker login ghcr.io -u "$GHCR_OWNER" --password-stdin 2>/dev/null
    ok "Logged in to ghcr.io"
  elif docker pull ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:latest &>/dev/null 2>&1; then
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

  local backend_image="ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:${version}"
  local frontend_image="ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:${version}"
  local backend_latest="ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:latest"
  local frontend_latest="ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:latest"

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

  log "Creating Git tag: ${CYAN}${tag_name}${NC}"

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
    warn "Tag $tag_name already exists locally"
    # Check if the tag is already on remote
    if git ls-remote --tags origin "$tag_name" | grep -q "$tag_name"; then
      warn "Tag $tag_name already exists on remote, skipping tag creation"
      TAG_EXISTS_REMOTE=true
      return
    fi
    # Tag exists locally but not remotely — push it
    log "Pushing existing local tag $tag_name to remote..."
    git push origin "$tag_name"
    ok "Pushed existing tag: $tag_name"
    TAG_EXISTS_REMOTE=false
    return
  fi

  local tag_msg="Release ${tag_name}"
  if [ "$RELEASE_TYPE" != "stable" ]; then
    tag_msg="Release ${tag_name} (${RELEASE_TYPE})"
  fi
  git tag -a "$tag_name" -m "$tag_msg"
  ok "Created tag: $tag_name"

  log "Pushing commits and tags to origin..."
  git push origin HEAD
  git push origin "$tag_name"

  ok "Git tag pushed: $tag_name"
  TAG_EXISTS_REMOTE=false
}

# ─── Create GitHub Release ─────────────────────────────────
create_github_release() {
  local version="$1"
  local tag_name="v${version}"
  local is_pre_release="false"

  if [ "$RELEASE_TYPE" != "stable" ]; then
    is_pre_release="true"
  fi

  log "Creating GitHub Release for ${CYAN}${tag_name}${NC}..."

  # Build release title
  local release_title="PureCore v${version}"
  if [ "$RELEASE_TYPE" = "alpha" ]; then
    release_title="PureCore v${version} (Alpha)"
  elif [ "$RELEASE_TYPE" = "beta" ]; then
    release_title="PureCore v${version} (Beta)"
  fi

  # Build release notes
  local release_notes="## 🐳 Docker Images

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
## 🚀 Deploy

\`\`\`bash
export PURECORE_VERSION=${version}
docker compose pull
docker compose up -d
\`\`\`
"

  # Try to include changes since last tag
  local last_tag
  last_tag=$(git describe --tags --abbrev=0 "v${version}" 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null || echo "")
  if [ -n "$last_tag" ] && [ "$last_tag" != "$tag_name" ]; then
    local changelog
    changelog=$(git log "${last_tag}..HEAD" --pretty=format:"- %s (%an)" 2>/dev/null || echo "")
    if [ -n "$changelog" ]; then
      release_notes+="
## 📝 Changes since ${last_tag}

${changelog}
"
    fi
  fi

  # Create the GitHub Release using gh CLI
  local gh_args=(
    release create "$tag_name"
    --repo "$GITHUB_REPO"
    --title "$release_title"
    --notes "$release_notes"
  )

  if [ "$is_pre_release" = "true" ]; then
    gh_args+=(--prerelease)
  fi

  # If tag was already pushed remotely, use it directly
  if [ "${TAG_EXISTS_REMOTE:-false}" = "true" ]; then
    gh_args+=(--notes-from-tag)
    warn "Using existing remote tag — release notes will be from tag message"
  fi

  if gh "${gh_args[@]}" 2>&1; then
    ok "GitHub Release created: ${CYAN}${release_title}${NC}"
    if [ "$is_pre_release" = "true" ]; then
      ok "Marked as pre-release"
    fi
  else
    err "Failed to create GitHub Release"
    echo "  You can create it manually at:"
    echo "  https://github.com/${GITHUB_REPO}/releases/new?tag=${tag_name}"
    echo ""
    echo "  Or run:"
    echo "  gh release create ${tag_name} --repo ${GITHUB_REPO} --title \"${release_title}\" --notes \"${release_notes}\" --prerelease"
  fi
}

# ─── Summary ───────────────────────────────────────────────
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
  echo "  🐳 Docker Images:"
  echo "    ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:${version}"
  echo "    ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:${version}"
  if [ "$RELEASE_TYPE" = "stable" ]; then
    echo "    ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-backend:latest"
    echo "    ghcr.io/${OWNER_LOWER}/${GHCR_REPO}-frontend:latest"
  fi
  echo ""
  echo "  🏷️  Git tag: v${version}"
  echo ""
  echo "  📦 GitHub Release:"
  echo "    https://github.com/${GITHUB_REPO}/releases/tag/v${version}"
  echo ""
  echo "  🚀 Deploy on server:"
  echo "    export PURECORE_VERSION=${version}"
  echo "    docker compose pull"
  echo "    docker compose up -d"
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
  # (this triggers the GitHub Actions workflow in .github/workflows/docker-publish.yml,
  #  but we've already pushed images manually, so the CI is a fallback safety net)
  git_tag_and_push "$version"

  # Step 6: Create GitHub Release
  create_github_release "$version"

  # Step 7: Summary
  print_summary "$version"
}

main "$@"
