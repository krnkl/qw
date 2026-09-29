# Release Workflow

This document details the release process and automation for `qw`.

## Overview

1. **`VERSION` File**: Version is tracked in the committed `VERSION` file at the repository root (e.g. `0.1.5`).
2. **PR Gate**: CI runs `scripts/verify-version.sh` to ensure any proposed version is strictly higher than both `main`'s `VERSION` and all existing release tags (no downgrades or conflicts allowed).
3. **Merge to `main`**:
   - **Major / Minor Bumps** (`X.Y.0`): Automatically tagged and released via GoReleaser.
   - **Patch Bumps** (`X.Y.Z`): Tagged for audit; releases are published only if marked with `[release]`/`[hotfix]` or triggered manually via `mise run release`.

## Manual Release Creation

To manually publish release archives for a patch bump or trigger an immediate release:

```bash
# Via mise:
mise run release

# Or directly with GitHub CLI:
gh workflow run release.yml -f force=true
```
