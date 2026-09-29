#!/usr/bin/env bash
# ==============================================================================
# qw installer script (Option A: mise native tool integration + global link)
# Supports release artifact installation via fzf or local development linking
# ==============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_BIN_DIR="${INSTALL_BIN_DIR:-${HOME}/.local/bin}"

# ANSI Colors
BLUE="\033[38;2;129;162;190m"
GREEN="\033[38;2;181;189;104m"
YELLOW="\033[38;2;240;198;116m"
RED="\033[38;2;204;102;102m"
CYAN="\033[38;2;138;190;183m"
BOLD="\033[1m"
RESET="\033[0m"

info() {
    printf "${BLUE}${BOLD}==>${RESET} %s\n" "$1"
}

success() {
    printf "${GREEN}${BOLD}✓${RESET} %s\n" "$1"
}

warn() {
    printf "${YELLOW}${BOLD}!${RESET} %s\n" "$1"
}

error() {
    printf "${RED}${BOLD}✗${RESET} %s\n" "$1" >&2
    exit 1
}

# Ensure mise is accessible
if ! command -v mise >/dev/null 2>&1; then
    if [ -x "${HOME}/.local/bin/mise" ]; then
        export PATH="${HOME}/.local/bin:$PATH"
    else
        error "'mise' was not found in PATH or ~/.local/bin. Please install mise (https://mise.jdx.dev) to manage qw versions."
    fi
fi

# Detect currently active/installed version
get_current_version() {
    if command -v qw >/dev/null 2>&1; then
        qw -v 2>/dev/null | awk '{print $3}' || echo "unknown"
    elif [ -L "${INSTALL_BIN_DIR}/qw" ] || [ -f "${INSTALL_BIN_DIR}/qw" ]; then
        "${INSTALL_BIN_DIR}/qw" -v 2>/dev/null | awk '{print $3}' || echo "unknown"
    else
        echo "none"
    fi
}

CURRENT_VERSION="$(get_current_version)"

# ------------------------------------------------------------------------------
# Dev Mode: Build local binary and link to mise + ~/.local/bin
# ------------------------------------------------------------------------------
install_dev() {
    info "Installing local development build of qw..."

    cd "${REPO_ROOT}"

    # 1. Build local binary with mise unconditionally
    info "Building ./bin/qw via mise..."
    mise run build
    success "Binary ./bin/qw is ready"

    # 2. Ad-hoc codesign on macOS (prevents arm64 invalidation crashes)
    if [[ "$(uname -s)" == "Darwin" ]] && command -v codesign >/dev/null 2>&1; then
        codesign -s - -f "${REPO_ROOT}/bin/qw" 2>/dev/null || true
    fi

    # 3. Link into mise
    info "Linking local repository into mise as github:krnkl/qw@dev..."
    mise link -f github:krnkl/qw@dev "${REPO_ROOT}"
    mise use -g github:krnkl/qw@dev
    mise reshim 2>/dev/null || true
    success "Configured mise shim pointing to local build"

    # 4. Relink global fallback link (~/.local/bin/qw)
    mkdir -p "${INSTALL_BIN_DIR}"
    ln -sf "${REPO_ROOT}/bin/qw" "${INSTALL_BIN_DIR}/qw"
    success "Relinked ${INSTALL_BIN_DIR}/qw -> ${REPO_ROOT}/bin/qw"

    # 5. Verification
    printf "\n"
    info "Active binary verification:"
    "${INSTALL_BIN_DIR}/qw" -v
    printf "\n${GREEN}${BOLD}Successfully linked qw (dev mode) 🥝${RESET}\n"
    exit 0
}

# ------------------------------------------------------------------------------
# Release Mode: Fetch versions, fzf select, mise use -g, and relink ~/.local/bin
# ------------------------------------------------------------------------------
install_release() {
    local target_version="${1:-}"

    if [ -z "${target_version}" ]; then
        info "Querying available releases for krnkl/qw..."

        # Fetch remote versions available
        local versions_raw=""
        if command -v gh >/dev/null 2>&1; then
            versions_raw="$(gh release list --repo krnkl/qw --limit 50 2>/dev/null | awk '{print $1}' || true)"
        fi

        if [ -z "${versions_raw}" ]; then
            versions_raw="$(mise ls-remote github:krnkl/qw 2>/dev/null || true)"
        fi

        if [ -z "${versions_raw}" ]; then
            error "No releases found for krnkl/qw on GitHub."
        fi

        # Check if fzf is available and interactive
        if [ -t 0 ] && command -v fzf >/dev/null 2>&1; then
            # Format list with currently installed marker
            local fzf_input=""
            while IFS= read -r v; do
                [ -z "$v" ] && continue
                v_clean="${v#v}"
                curr_clean="${CURRENT_VERSION#v}"
                if [ "$v_clean" = "$curr_clean" ]; then
                    fzf_input+="$(printf "%-12s  %b(currently active)%b\n" "$v" "${GREEN}" "${RESET}")"$'\n'
                else
                    fzf_input+="$(printf "%-12s\n" "$v")"$'\n'
                fi
            done <<< "$versions_raw"

            local selected
            selected="$(printf "%s" "$fzf_input" | fzf \
                --header="Active version: ${CURRENT_VERSION}" \
                --prompt="Select qw release to install > " \
                --ansi \
                --height=40% \
                --reverse)" || {
                warn "Selection canceled."
                exit 0
            }

            target_version="$(echo "$selected" | awk '{print $1}')"
        else
            # Non-interactive or fzf unavailable: default to first (latest)
            target_version="$(echo "$versions_raw" | head -n 1)"
            info "Defaulting to latest version: ${target_version}"
        fi
    fi

    # Clean version string (remove leading 'v' for mise)
    local ver_clean="${target_version#v}"

    info "Installing github:krnkl/qw@${ver_clean} via mise..."
    mise use -g "github:krnkl/qw@${ver_clean}"
    mise reshim 2>/dev/null || true

    # Locate installed binary
    local install_dir
    install_dir="$(mise where "github:krnkl/qw@${ver_clean}" 2>/dev/null || true)"
    if [ -z "${install_dir}" ] || [ ! -d "${install_dir}" ]; then
        error "Failed to locate installed directory for github:krnkl/qw@${ver_clean}."
    fi

    local bin_path="${install_dir}/qw"
    if [ ! -f "${bin_path}" ]; then
        bin_path="${install_dir}/bin/qw"
    fi

    if [ ! -x "${bin_path}" ]; then
        error "Executable qw not found in ${install_dir}."
    fi

    # Codesign on macOS
    if [[ "$(uname -s)" == "Darwin" ]] && command -v codesign >/dev/null 2>&1; then
        codesign -s - -f "${bin_path}" 2>/dev/null || true
    fi

    # Relink ~/.local/bin/qw
    mkdir -p "${INSTALL_BIN_DIR}"
    ln -sf "${bin_path}" "${INSTALL_BIN_DIR}/qw"
    success "Relinked ${INSTALL_BIN_DIR}/qw -> ${bin_path}"

    # Verification
    printf "\n"
    info "Active binary verification:"
    "${INSTALL_BIN_DIR}/qw" -v
    printf "\n${GREEN}${BOLD}Successfully installed qw ${target_version}! 🥝${RESET}\n"
}

# ------------------------------------------------------------------------------
# Entrypoint Router
# ------------------------------------------------------------------------------
case "${1:-}" in
    --dev|dev|-d)
        install_dev
        ;;
    --help|-h|help)
        printf "Usage:\n"
        printf "  ./scripts/install.sh          Interactively select and install a release via fzf\n"
        printf "  ./scripts/install.sh <ver>    Directly install a specific release (e.g. 0.1.0 or v0.1.0)\n"
        printf "  ./scripts/install.sh --dev    Build and link the local development checkout into mise and ~/.local/bin\n"
        exit 0
        ;;
    *)
        install_release "${1:-}"
        ;;
esac
