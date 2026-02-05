# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**reap** — Interactive TUI for viewing and killing processes on ports. Like `htop` meets `lsof`. Part of the cull ecosystem.

## Development Commands

```bash
go run ./cmd/reap                  # run TUI
go run ./cmd/reap list             # run CLI list
go run ./cmd/reap kill 3000        # run CLI kill
go test ./...                      # run all tests
go test ./internal/ports/...       # run a single package's tests
go test -run TestParseLsof ./internal/ports/  # run a single test
goreleaser release --snapshot      # test release build
```

## Stack

- **Go** with **Bubble Tea** (charmbracelet/bubbletea) for TUI
- **Lip Gloss** (charmbracelet/lipgloss) for styling — never use raw ANSI codes
- **charmbracelet/bubbles/table** for table component
- **Cobra** for CLI commands
- Config: TOML at `~/.config/reap/config.toml`

## Architecture

The project follows Bubble Tea's Elm architecture (Model → Update → View).

**Entry points** (`cmd/reap/`): Cobra commands — `main.go` (root, launches TUI), `list.go` (non-interactive list), `kill.go` (non-interactive kill).

**TUI layer** (`internal/tui/`): Bubble Tea model in `model.go` with sub-components for table, detail panel, filter input, and kill confirmation dialog. Keep the `Update` function clean — extract logic to helpers.

**Port scanning** (`internal/ports/`): Platform-abstracted port scanning behind a common interface. Each platform has its own build-tagged file (`scanner_darwin.go`, `scanner_linux.go`, `scanner_windows.go`). Types in `types.go` define `PortInfo` and `ProcessInfo` structs.

**Config** (`internal/config/`): Loads user config (refresh interval, custom port colors/labels, show_system flag).

### Platform-Specific Scanning

| Platform | Primary Command | Enrichment |
|----------|----------------|------------|
| macOS | `lsof -iTCP -sTCP:LISTEN -P -n` | — |
| Linux | `ss -tlnp` | `/proc/[pid]/cmdline`, `/proc/[pid]/status` |
| Windows | `netstat -ano` + `tasklist` | `wmic`/PowerShell `Get-Process` |

## Code Style

- Follow Go conventions: `gofmt`, `golint`
- Error messages: lowercase, no punctuation, include context
- Max line length: 100 chars (soft limit)
- Use `lipgloss.NewStyle()` — never raw ANSI codes
- Tests mock all system calls — no external dependencies in tests

## Key Design Decisions

- **Port color coding**: Ports are color-coded by category (green=frontend, yellow=backend, cyan=Flask/Vite, magenta=Postgres, red=Redis, blue=MySQL/MongoDB, white=HTTP/S, dim=other). Colors are customizable via config.
- **Kill signals**: `k` sends SIGTERM (TerminateProcess on Windows), `K` sends SIGKILL (taskkill /F on Windows). Always show confirmation before kill.
- **Two modes**: Interactive TUI (default `reap`) and non-interactive CLI (`reap list`, `reap kill`). CLI supports `--json` output.
- **Cross-platform**: macOS, Linux, Windows. Use build tags for platform-specific code.

## Non-Goals

Do not add: network traffic monitoring, remote port scanning, firewall management, Docker-specific features, or service management.
