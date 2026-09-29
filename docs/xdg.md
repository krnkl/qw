# Configuration & XDG Resolution

`qw` follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html).

## Storage Categories & Defaults

| Category | Environment Variable | macOS Fallback | Linux Fallback | Permissions |
| :--- | :--- | :--- | :--- | :--- |
| **Config** | `XDG_CONFIG_HOME` | `$HOME/.config/qw` | `$HOME/.config/qw` | `0755` |
| **Data** | `XDG_DATA_HOME` | `$HOME/.local/share/qw` | `$HOME/.local/share/qw` | `0755` |
| **State** | `XDG_STATE_HOME` | `$HOME/.local/state/qw` | `$HOME/.local/state/qw` | `0755` |
| **Cache** | `XDG_CACHE_HOME` | `$HOME/.cache/qw` | `$HOME/.cache/qw` | `0755` |
| **Runtime** | `XDG_RUNTIME_DIR` | `$TMPDIR/qw` | `/run/user/<UID>/qw` *(or `/tmp/qw`)* | `0700` |

## Workspace Directory

`qw` resolves workspaces inside `<DataDir>/workspaces` (`~/.local/share/qw/workspaces`) by default.

You can customize this location in two ways:

1. **Config File (`config.yaml`)**:
   Create `~/.config/qw/config.yaml`:
   ```yaml
   workspaces: ~/projects
   ```

2. **Environment Variable**:
   Export `QW_WORKSPACES` (takes precedence over `config.yaml`):
   ```bash
   export QW_WORKSPACES="$HOME/projects"
   ```

## Diagnostics

Inspect resolved paths and resolution sources:

```bash
qw config
```

JSON output:
```bash
qw config --json
```
