# snip — Clipboard History Manager

A lightweight CLI clipboard history manager. Use `snip start` to run a long-running clipboard watcher that stores every copied item in a local database. Search, recall, and paste any entry from the terminal.

## Features

- Clipboard watcher with automatic deduplication (`snip start`)
- Fuzzy search with highlighted matches
- JSON output for scripting
- Configurable via `~/.snip/config.yaml`
- Color output (disable with `--no-color` or `NO_COLOR`)
- Single binary, no external services

## Installation

### 1. Quick Install (macOS / Linux)

The easiest way to install `snip` without needing Go is via the install script. It automatically downloads the correct pre-compiled binary for your OS and Architecture.

```bash
curl -fsSL https://raw.githubusercontent.com/srmoralesomar/snip/main/install.sh | bash
```
*(Requires: `curl`, `tar`, and `sudo` access to install to `/usr/local/bin`)*

### 2. Using Go (All Platforms)

If you already have Go installed, you can install the latest version directly:

```bash
go install github.com/srmoralesomar/snip@latest
```
*(Requires: Go 1.20+)*

### 3. Build from Source

```bash
git clone https://github.com/srmoralesomar/snip
cd snip
go build -o snip .
# Move to a directory on your PATH, e.g.:
mv snip /usr/local/bin/
```
*(Requires: Go 1.20+, Git)*

## Quick Start

```bash
# Start the watcher in the shell background (suppresses output). 
snip start > /dev/null 2>&1 &
# Closing the terminal may stop it; 
# use nohup, disown, tmux, or a user LaunchAgent/systemd unit if you need it to survive the terminal closing.

# Copy something to your clipboard, then list history
snip list

# Search history
snip search "hello"

# Copy a clip back to your clipboard
snip copy 3

# Stop the watcher
snip stop
```

## Commands

### `snip start`

Start the long-running clipboard watcher. It polls the clipboard every 500 ms (configurable) and saves new entries to `~/.snip/history.db`. Press Ctrl+C or run `snip stop` to exit.

```bash
snip start
snip start --max-history 1000     # Keep up to 1000 entries
snip start --poll-interval-ms 250 # Poll every 250ms
snip start --storage-path /tmp/snip.db
```

### `snip list`

Show recent clipboard history with index, relative timestamp, and content preview.

```bash
snip list                # Show last 20 entries
snip list --count 50     # Show last 50 entries
snip list --json         # Machine-readable JSON output
```

Example output:

```
INDEX  TIME          PREVIEW
-----  ------------  ----------------------------------------
1      just now      func main() { fmt.Println("hello") }
2      3m ago        https://example.com/some/long/url
3      1h ago        TODO: fix the authentication bug
```

### `snip search <query>`

Fuzzy-search all stored entries ranked by relevance. Matched characters are highlighted.

```bash
snip search "main"
snip search "http" --limit 5   # Cap results
snip search "todo" --json      # JSON output
```

### `snip copy <index>`

Copy a clip back to the system clipboard.

```bash
snip copy 3       # Copy clip at index 3
snip copy --last  # Copy the most recent clip
```

### `snip delete <index>`

Remove a single entry from history.

```bash
snip delete 5
```

### `snip clear`

Delete all clipboard history.

```bash
snip clear           # Prompts for confirmation
snip clear --force   # Skip confirmation
```

### `snip status`

Show whether the clipboard watcher is running, along with its PID, uptime, and total clip count.

```bash
snip status
```

Example output:

```
watcher: running
  PID:    12345
  uptime: 2h 15m 3s
  clips:  142
```

### `snip stop`

Send SIGTERM to the running snip process (started with `snip start`).

```bash
snip stop
```

### `snip config`

Print the current effective configuration (defaults + config file + CLI flags).

```bash
snip config
```

### Global Flags

| Flag        | Description                  |
|-------------|------------------------------|
| `--no-color`| Disable color output         |
| `--version` | Print version and exit       |
| `--help`    | Show help for any command    |

## Configuration

Create `~/.snip/config.yaml` to set persistent defaults:

```yaml
max_history: 500        # Maximum number of clips to keep (default: 500)
poll_interval_ms: 500   # Clipboard poll interval in milliseconds (default: 500)
storage_path: ""        # Custom DB path; default: ~/.snip/history.db
```

CLI flags always override config file values. Missing config file silently uses built-in defaults.

## File Layout

```
~/.snip/
├── history.db    # BoltDB database of clipboard entries
├── config.yaml   # Optional configuration file
└── snip.pid      # PID file written while snip start is running
```

## Architecture

```
snip/
├── main.go                        # Entry point — calls cmd.Execute()
├── cmd/                           # Cobra CLI commands
│   ├── root.go                    # Root command, --no-color flag, loadConfig()
│   ├── start.go                   # snip start — watcher + storage loop
│   ├── list.go                    # snip list
│   ├── search.go                  # snip search (fuzzy)
│   ├── copy.go                    # snip copy
│   ├── delete.go                  # snip delete
│   ├── clear.go                   # snip clear
│   ├── status.go                  # snip status
│   ├── stop.go                    # snip stop
│   └── config_cmd.go              # snip config
└── internal/
    ├── clipboard/
    │   ├── watcher.go             # Polling loop, Clip type, Reader interface
    │   └── system.go              # SystemReader wrapping atotto/clipboard
    ├── config/
    │   └── config.go              # Config struct, Load(), defaults
    ├── pidfile/
    │   └── pidfile.go             # PID file read/write/remove/IsRunning
    └── store/
        └── store.go               # BoltDB wrapper: Save, List, Get, Delete, Prune, Clear
```

**Key dependencies:**

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `go.etcd.io/bbolt` | Embedded key-value database |
| `github.com/sahilm/fuzzy` | Fuzzy string matching |
| `github.com/atotto/clipboard` | Cross-platform clipboard access |
| `github.com/spf13/viper` | Configuration file support |
| `github.com/fatih/color` | Terminal color output |

## License

MIT
