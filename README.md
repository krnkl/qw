# qw (kiwi 🥝): Unified Workspace Navigator

An ultra-fast terminal workspace navigator built in Go.

> **Pronunciation**: *qw* is pronounced **"kiwi"** (🥝) — an ergonomic, 2-key left-hand chord (`q` $\to$ `w`) designed for instant muscle memory.

---

## 🚀 Features

- **⚡ Strict XDG Base Directory Compliance**: Respects `$XDG_CONFIG_HOME`, `$XDG_DATA_HOME`, `$XDG_STATE_HOME`, `$XDG_CACHE_HOME`, and `$XDG_RUNTIME_DIR` with strict fallback handling on macOS and Linux.
- **🛡️ Secure File Permissions**: Initializes runtime directories with `0700` (`rwx------`) and storage directories with `0755` (`rwxr-xr-x`).
- **🔍 Diagnostic CLI**: `qw config` prints exact paths for all affected directories, detailing the precise resolution source (environment variable, config file, or platform fallback).
- **📂 Flexible Workspace Root**: Configurable via `~/.config/qw/config.yaml` (`ws`) and overridable via `$QW_WORKSPACES` (falling back to `$XDG_DATA_HOME/qw/ws`).
- **🧪 Cross-Platform Tested**: Fully verified locally and inside containerized Linux test suites.
- **📦 Multi-Arch Releases**: Automated GoReleaser pipeline for `darwin/arm64`, `darwin/amd64`, `linux/arm64`, and `linux/amd64`.

---

## 🛠️ Installation & Version Management

### 1. Interactive Release Install (`fzf`)

Download and install an official release archive into `mise` and relink `~/.local/bin/qw`:

```bash
# Interactively choose a release via fzf:
mise run install
# or directly:
./install.sh

# Or install a specific version directly without prompting:
./install.sh 0.1.0
```

### 2. Local Development Build (`install:dev`)

Compile the local repository and link it as the active global tool:

```bash
mise run install:dev
# or directly:
./install.sh --dev
```
This builds `./bin/qw` with your current commit SHA and timestamp, configures `mise link -f github:krnkl/qw@dev $(pwd)`, activates it via `mise use -g github:krnkl/qw@dev`, and points `~/.local/bin/qw` to `./bin/qw`.

### 3. Build & Test Targets

```bash
# Build binary to bin/qw
mise run build

# Run unit tests
mise run test

# Run Linux test suite via Docker
mise run test:docker

# Lint & vet
mise run lint
```

---

## ⚙️ Configuration & XDG Resolution

`qw` follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html).

### Storage Categories & Defaults

| Category | Environment Variable | macOS Fallback | Linux Fallback | Permissions |
| :--- | :--- | :--- | :--- | :--- |
| **Config** | `XDG_CONFIG_HOME` | `$HOME/.config/qw` | `$HOME/.config/qw` | `0755` |
| **Data** | `XDG_DATA_HOME` | `$HOME/.local/share/qw` | `$HOME/.local/share/qw` | `0755` |
| **State** | `XDG_STATE_HOME` | `$HOME/.local/state/qw` | `$HOME/.local/state/qw` | `0755` |
| **Cache** | `XDG_CACHE_HOME` | `$HOME/.cache/qw` | `$HOME/.cache/qw` | `0755` |
| **Runtime** | `XDG_RUNTIME_DIR` | `$TMPDIR/qw` | `/run/user/<UID>/qw` *(or `/tmp/qw`)* | `0700` |

### Workspace Directory (`ws`)

`qw` stores new workspaces inside `<DataDir>/ws` (`~/.local/share/qw/ws`) by default.

You can customize this location in two ways:

1. **Config File (`config.yaml`)**:
   Create `~/.config/qw/config.yaml`:
   ```yaml
   ws: ~/projects
   ```

2. **Environment Variable**:
   Export `QW_WORKSPACES` (takes precedence over `config.yaml`):
   ```bash
   export QW_WORKSPACES="$HOME/projects"
   ```

---

## 💻 CLI Commands

### 1. `qw config`

Inspect resolved paths and resolution sources:

```bash
qw config
```

Example output:
```text
Configuration Paths:
  Config File:    /Users/krnkl/.config/qw/config.yaml [missing] (config dir: platform fallback: darwin (~/.config))
  Config Dir:     /Users/krnkl/.config/qw (resolved via: platform fallback: darwin (~/.config))
  Data Dir:       /Users/krnkl/.local/share/qw (resolved via: platform fallback: darwin (~/.local/share))
  State Dir:      /Users/krnkl/.local/state/qw (resolved via: platform fallback: darwin (~/.local/state))
  Cache Dir:      /Users/krnkl/.cache/qw (resolved via: platform fallback: darwin (~/.cache))
  Runtime Dir:    /var/folders/.../T/qw (resolved via: platform fallback: macOS $TMPDIR)

Settings:
  Workspaces Dir: /Users/krnkl/projects (resolved via: env: $QW_WORKSPACES)
```

JSON output:
```bash
qw config --json
```

### 2. `qw version` / `-v`

Print version, commit hash, build date, and toolchain info:

```bash
qw -v
# Output: qw version v0.1.0 (commit: a1b2c3d, built: 2026-09-28T09:30:00Z, go: go1.27.1, os/arch: darwin/arm64)
```

---

## 🔄 Release Workflow

1. **`VERSION` File**: Version is tracked in the committed `VERSION` file (e.g. `0.1.0`).
2. **PR Gate**: CI runs `scripts/verify-version.sh` to ensure any proposed version is strictly higher than previous tags (no downgrades or conflicts allowed).
3. **Merge to `main`**:
   - **Major / Minor Bumps** (`X.Y.0`): Automatically tagged and released via GoReleaser.
   - **Patch Bumps** (`X.Y.Z`): Tagged; releases are published only if marked with `[release]`/`[hotfix]` or triggered manually via `workflow_dispatch`.

---

## 📜 License

MIT
