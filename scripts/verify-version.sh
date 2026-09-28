#!/usr/bin/env bash
# ==============================================================================
# verify-version.sh
# Validates semver format, prevents downgrades, enforces PR version bumping,
# and provides semver utility subcommands (compare, bump-type, highest-tag).
# ==============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

# Strip leading 'v' and whitespace
clean_ver() {
    local v="${1#v}"
    echo "${v}" | tr -d '[:space:]'
}

# Validate semver format: X.Y.Z or X.Y.Z-prerelease (with optional build metadata)
is_valid_semver() {
    local v
    v="$(clean_ver "$1")"
    if [[ "$v" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
        return 0
    else
        return 1
    fi
}

# Compare two non-negative integers: 1 if a > b, -1 if a < b, 0 if equal
cmp_num() {
    local a="$1"
    local b="$2"
    a="${a#"${a%%[!0]*}"}"
    [ -z "$a" ] && a=0
    b="${b#"${b%%[!0]*}"}"
    [ -z "$b" ] && b=0

    if [ "$a" -gt "$b" ]; then
        echo 1
    elif [ "$a" -lt "$b" ]; then
        echo -1
    else
        echo 0
    fi
}

# Compare two prerelease dot-separated parts according to SemVer 2.0.0
cmp_prerelease_part() {
    local p1="$1"
    local p2="$2"
    local is_num1=0
    local is_num2=0

    [[ "$p1" =~ ^[0-9]+$ ]] && is_num1=1
    [[ "$p2" =~ ^[0-9]+$ ]] && is_num2=1

    if [ "$is_num1" -eq 1 ] && [ "$is_num2" -eq 1 ]; then
        cmp_num "$p1" "$p2"
    elif [ "$is_num1" -eq 1 ] && [ "$is_num2" -eq 0 ]; then
        echo -1 # numeric has lower precedence than non-numeric
    elif [ "$is_num1" -eq 0 ] && [ "$is_num2" -eq 1 ]; then
        echo 1  # non-numeric has higher precedence than numeric
    else
        if [ "$p1" \> "$p2" ]; then
            echo 1
        elif [ "$p1" \< "$p2" ]; then
            echo -1
        else
            echo 0
        fi
    fi
}

# Compare two prerelease strings
cmp_prerelease() {
    local pre1="$1"
    local pre2="$2"

    if [ -z "$pre1" ] && [ -z "$pre2" ]; then
        echo 0
        return
    fi
    # Normal version without prerelease has higher precedence than with prerelease
    if [ -z "$pre1" ] && [ -n "$pre2" ]; then
        echo 1
        return
    fi
    if [ -n "$pre1" ] && [ -z "$pre2" ]; then
        echo -1
        return
    fi

    IFS='.' read -r -a parts1 <<< "$pre1"
    IFS='.' read -r -a parts2 <<< "$pre2"

    local len1="${#parts1[@]}"
    local len2="${#parts2[@]}"
    local min_len="$len1"
    [ "$len2" -lt "$min_len" ] && min_len="$len2"

    local i
    for (( i=0; i<min_len; i++ )); do
        local c
        c="$(cmp_prerelease_part "${parts1[i]}" "${parts2[i]}")"
        if [ "$c" -ne 0 ]; then
            echo "$c"
            return
        fi
    done

    if [ "$len1" -gt "$len2" ]; then
        echo 1
    elif [ "$len1" -lt "$len2" ]; then
        echo -1
    else
        echo 0
    fi
}

# Semver comparison: outputs 1 if v1 > v2, -1 if v1 < v2, 0 if v1 == v2
semver_compare() {
    local v1 v2
    v1="$(clean_ver "$1")"
    v2="$(clean_ver "$2")"

    v1="${v1%%+*}"
    v2="${v2%%+*}"

    local core1="${v1%%-*}"
    local pre1=""
    [[ "$v1" == *-* ]] && pre1="${v1#*-}"

    local core2="${v2%%-*}"
    local pre2=""
    [[ "$v2" == *-* ]] && pre2="${v2#*-}"

    local maj1 min1 pat1
    IFS='.' read -r maj1 min1 pat1 <<< "$core1"

    local maj2 min2 pat2
    IFS='.' read -r maj2 min2 pat2 <<< "$core2"

    local c
    c="$(cmp_num "$maj1" "$maj2")"
    if [ "$c" -ne 0 ]; then echo "$c"; return; fi

    c="$(cmp_num "$min1" "$min2")"
    if [ "$c" -ne 0 ]; then echo "$c"; return; fi

    c="$(cmp_num "$pat1" "$pat2")"
    if [ "$c" -ne 0 ]; then echo "$c"; return; fi

    cmp_prerelease "$pre1" "$pre2"
}

# Determine bump type (major, minor, patch, none)
semver_bump_type() {
    local old_v new_v
    old_v="$(clean_ver "$1")"
    new_v="$(clean_ver "$2")"

    local core_old="${old_v%%-*}"
    local core_new="${new_v%%-*}"

    local maj_old min_old pat_old
    IFS='.' read -r maj_old min_old pat_old <<< "$core_old"

    local maj_new min_new pat_new
    IFS='.' read -r maj_new min_new pat_new <<< "$core_new"

    if [ "$(cmp_num "$maj_new" "$maj_old")" -eq 1 ]; then
        echo "major"
    elif [ "$(cmp_num "$min_new" "$min_old")" -eq 1 ]; then
        echo "minor"
    elif [ "$(cmp_num "$pat_new" "$pat_old")" -eq 1 ]; then
        echo "patch"
    else
        echo "none"
    fi
}

# Find highest existing git tag (v*)
get_highest_tag() {
    local highest=""
    for t in $(git tag -l "v*"); do
        is_valid_semver "$t" || continue
        if [ -z "$highest" ]; then
            highest="$t"
        else
            if [ "$(semver_compare "$t" "$highest")" -eq 1 ]; then
                highest="$t"
            fi
        fi
    done
    echo "$highest"
}

run_verify() {
    local explicit_base="${1:-}"

    if [ ! -f "VERSION" ]; then
        echo "Error: VERSION file not found in $(pwd)" >&2
        exit 1
    fi

    local proposed_ver
    proposed_ver="$(clean_ver "$(cat VERSION)")"

    if [ -z "$proposed_ver" ]; then
        echo "Error: VERSION file is empty" >&2
        exit 1
    fi

    if ! is_valid_semver "$proposed_ver"; then
        echo "Error: VERSION '${proposed_ver}' does not match semantic versioning format (e.g. 0.1.0)" >&2
        exit 1
    fi

    echo "✓ Proposed VERSION format is valid: v${proposed_ver}"

    # Determine reference branch / context
    local is_pr=0
    local base_ref=""

    if [ -n "$explicit_base" ]; then
        is_pr=1
        base_ref="$explicit_base"
    elif [ "${GITHUB_EVENT_NAME:-}" = "pull_request" ] && [ -n "${GITHUB_BASE_REF:-}" ]; then
        is_pr=1
        base_ref="${GITHUB_BASE_REF}"
    elif [ -n "${GITHUB_BASE_REF:-}" ]; then
        is_pr=1
        base_ref="${GITHUB_BASE_REF}"
    else
        local current_branch
        current_branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")"
        if [ -n "$current_branch" ] && [ "$current_branch" != "main" ] && [ "$current_branch" != "master" ] && [ "$current_branch" != "HEAD" ]; then
            is_pr=1
            base_ref="main"
        fi
    fi

    # Check against base branch VERSION (e.g. origin/main or main)
    local main_ver=""
    if [ -n "$base_ref" ]; then
        if git rev-parse --verify "origin/${base_ref}" >/dev/null 2>&1; then
            main_ver="$(git show "origin/${base_ref}:VERSION" 2>/dev/null | tr -d '[:space:]' || true)"
        elif git rev-parse --verify "${base_ref}" >/dev/null 2>&1; then
            main_ver="$(git show "${base_ref}:VERSION" 2>/dev/null | tr -d '[:space:]' || true)"
        fi
    fi

    if [ -n "$main_ver" ]; then
        local main_clean
        main_clean="$(clean_ver "$main_ver")"
        echo "ℹ Base branch (${base_ref}) VERSION: ${main_clean}"

        local cmp_main
        cmp_main="$(semver_compare "$proposed_ver" "$main_clean")"

        if [ "$is_pr" -eq 1 ]; then
            if [ "$cmp_main" -le 0 ]; then
                echo "✗ Error: Proposed VERSION (${proposed_ver}) is <= ${base_ref} branch VERSION (${main_clean})." >&2
                echo "  Every PR must bump the VERSION file for audit trail." >&2
                exit 1
            fi
            echo "✓ Proposed VERSION (${proposed_ver}) is strictly greater than ${base_ref} VERSION (${main_clean})"
        fi
    fi

    # Ensure git tags are up-to-date
    if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        git fetch --tags origin 2>/dev/null || true
    fi

    # Check against highest existing git tag
    local highest_tag
    highest_tag="$(get_highest_tag)"

    if [ -n "$highest_tag" ]; then
        local tag_clean
        tag_clean="$(clean_ver "$highest_tag")"
        echo "ℹ Highest existing release tag: ${highest_tag}"

        local cmp_tag
        cmp_tag="$(semver_compare "$proposed_ver" "$tag_clean")"

        if [ "$is_pr" -eq 1 ]; then
            if [ "$cmp_tag" -le 0 ]; then
                echo "✗ Error: Proposed VERSION (${proposed_ver}) is <= highest existing tag (${highest_tag})." >&2
                echo "  Downgrades or reusing an existing release version is not allowed." >&2
                exit 1
            fi
            echo "✓ Proposed VERSION (${proposed_ver}) is strictly greater than latest tag (${highest_tag})"
        else
            if [ "$cmp_tag" -lt 0 ]; then
                echo "✗ Error: VERSION (${proposed_ver}) on main is lower than existing tag (${highest_tag})." >&2
                exit 1
            elif [ "$cmp_tag" -eq 0 ]; then
                echo "ℹ VERSION (${proposed_ver}) matches existing tag (${highest_tag})."
            else
                echo "✓ VERSION (${proposed_ver}) is higher than latest tag (${highest_tag}) (new release pending)."
            fi
        fi
    else
        echo "ℹ No previous release tags found. Initial release will be v${proposed_ver}"
    fi
}

# Entrypoint router
case "${1:-verify}" in
    verify)
        run_verify "${2:-}"
        ;;
    compare)
        if [ "$#" -lt 3 ]; then
            echo "Usage: $0 compare <v1> <v2>" >&2
            exit 1
        fi
        semver_compare "$2" "$3"
        ;;
    bump-type)
        if [ "$#" -lt 3 ]; then
            echo "Usage: $0 bump-type <old_v> <new_v>" >&2
            exit 1
        fi
        semver_bump_type "$2" "$3"
        ;;
    highest-tag)
        get_highest_tag
        ;;
    *)
        echo "Unknown subcommand: $1" >&2
        echo "Usage: $0 [verify [base-branch]|compare <v1> <v2>|bump-type <old> <new>|highest-tag]" >&2
        exit 1
        ;;
esac
