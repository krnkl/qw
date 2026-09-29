# qw

`qw` stands for **quick workspaces** — an ergonomic 2-key chord (`q` → `w`) designed for rapid muscle memory.

The main goal of `qw` is to help navigate active workspaces and aggregate related external context into a single, concise view, streamlining context switching and tracking across local and remote workspaces.

## Features

- **Strict XDG Base Directory Compliance**: Follows standard XDG paths for configuration, data, state, cache, and runtime directories across macOS and Linux.

## Installation

### Prerequisites

- [mise](https://mise.jdx.dev)
- [fzf](https://github.com/junegunn/fzf)

### Setup

Clone the repository and install via `mise`:

```bash
git clone https://github.com/krnkl/qw.git
cd qw
mise run install
```

## Development

```bash
mise run        # Run all verification, lint, test, and build targets
mise run test   # Run unit test suite
```

## Usage

```bash
qw version
# Output: qw version v0.1.5 (commit: ..., built: ..., go: ..., os/arch: ...)
```

## License

MIT
