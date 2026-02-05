# reap

Interactive TUI for viewing and killing processes on ports. Like `htop` meets `lsof`.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Screenshot

<!-- TODO: Add terminal screenshot/GIF demo -->
```
┌─────────────────────────────────────────────────────────────────────────────┐
│ PORT   PID     PROCESS        USER    MEMORY   UPTIME   CONTAINER   DIR     │
│ 3000   12345   node           dev     128MB    2h 15m   -           ~/app   │
│ 5432   1234    postgres       post    256MB    5d 3h    postgres    /var    │
│ 8080   5678    java           dev     512MB    1h 30m   -           ~/api   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Features

- **Interactive TUI** with real-time process monitoring
- **Color-coded ports** by service type (frontend, backend, databases)
- **Docker container detection** - see which processes run in containers
- **Process tree grouping** - parent-child relationships, same PID with multiple ports, shared PPID
- **Flexible filtering** - filter by port, process name, user, or container
- **Kill processes** - send SIGTERM or SIGKILL with confirmation
- **Kill parent process** - terminate the parent when needed
- **Cross-platform** - works on macOS, Linux, and Windows

## Installation

### Using Go

```bash
go install github.com/legostin/reap/cmd/reap@latest
```

### Homebrew (macOS/Linux)

```bash
# Coming soon
# brew install legostin/tap/reap
```

### Binary Download

<!-- TODO: Add release download links -->
Download pre-built binaries from the [Releases](https://github.com/legostin/reap/releases) page.

## Usage

### Interactive TUI

Launch the interactive interface:

```bash
reap
```

### Non-Interactive List

List all listening ports:

```bash
reap list
```

Filter by port number:

```bash
reap list --port 3000
reap list -p 3000
```

Filter by process name:

```bash
reap list --name node
reap list -n node
```

Output as JSON:

```bash
reap list --json
```

Combine filters:

```bash
reap list --port 8080 --name java --json
```

### Non-Interactive Kill

Kill process on a specific port:

```bash
reap kill 3000
```

Kill processes on multiple ports:

```bash
reap kill 3000 5000 8080
```

Force kill (SIGKILL instead of SIGTERM):

```bash
reap kill --force 3000
reap kill -f 3000
```

Skip confirmation prompt:

```bash
reap kill --yes 3000
reap kill -y 3000
```

Combine flags:

```bash
reap kill -f -y 3000 5000
```

## Keybindings

| Key | Action |
|-----|--------|
| `↑` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Enter` | Show process details |
| `k` | Kill process (SIGTERM) |
| `K` | Force kill process (SIGKILL) |
| `p` | Kill parent process |
| `/` | Filter processes |
| `s` | Cycle sort column |
| `S` | Reverse sort order |
| `a` | Toggle system processes |
| `t` | Toggle tree view |
| `r` | Refresh process list |
| `?` | Show help |
| `Esc` | Go back / close dialog |
| `q` / `Ctrl+C` | Quit |

## Port Colors

Ports are color-coded by their typical service type:

| Color | Service Type | Default Ports |
|-------|--------------|---------------|
| Green | Frontend | 3000, 3001, 4200, 5173, 8080 |
| Yellow | Backend | 4000, 5000, 8000, 8081, 9000 |
| Cyan | Flask / Vite | 5001, 5174, 5175 |
| Magenta | PostgreSQL | 5432 |
| Red | Redis | 6379 |
| Blue | MySQL / MongoDB | 3306, 27017 |
| White | HTTP / HTTPS | 80, 443 |
| Dim | Other | All other ports |

Colors are customizable via configuration (see below).

## Configuration

Configuration file location: `~/.config/reap/config.toml`

### Example Configuration

```toml
# Refresh interval in seconds (default: 2)
refresh_interval = 2

# Show system processes by default (default: false)
show_system = false

# Custom port colors
# Available colors: green, yellow, cyan, magenta, red, blue, white, dim
[port_colors]
"3333" = "green"      # Custom dev server
"4444" = "yellow"     # Custom API
"9200" = "magenta"    # Elasticsearch

# Custom port labels (shown in details view)
[port_labels]
"3333" = "My Dev Server"
"4444" = "Custom API"
"9200" = "Elasticsearch"
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `refresh_interval` | int | 2 | Auto-refresh interval in seconds |
| `show_system` | bool | false | Show system processes by default |
| `port_colors` | map | {} | Override default port colors |
| `port_labels` | map | {} | Custom labels for ports |

## Building from Source

```bash
# Clone the repository
git clone https://github.com/legostin/reap.git
cd reap

# Build
go build -o reap ./cmd/reap

# Run tests
go test ./...

# Install to $GOPATH/bin
go install ./cmd/reap
```

### Development Commands

```bash
go run ./cmd/reap                  # Run TUI
go run ./cmd/reap list             # Run CLI list
go run ./cmd/reap kill 3000        # Run CLI kill
go test ./...                      # Run all tests
go test ./internal/ports/...       # Run package tests
goreleaser release --snapshot      # Test release build
```

## Platform Support

| Platform | Port Detection Method |
|----------|----------------------|
| macOS | `lsof -iTCP -sTCP:LISTEN -P -n` |
| Linux | `ss -tlnp` with `/proc` enrichment |
| Windows | `netstat -ano` + `tasklist` |

## License

MIT License - see [LICENSE](LICENSE) for details.
