# MediaStack CLI

A Go-based command-line tool for managing your Docker media stack.

## Features

- **deploy** - Deploy the full media stack (replaces restart.sh)
- **stop** - Stop all or specific containers
- **restart** - Restart the stack
- **status** - View container status and health
- **logs** - Stream container logs
- **pull** - Update Docker images
- **validate** - Validate configuration
- **apikeys** - Extract API keys from *ARR apps

## Installation

### Prerequisites

- Go 1.22 or later
- Docker with Docker Compose v2

### Build from source

```bash
# Install Go (macOS)
brew install go

# Build the CLI
cd cli
make deps
make build

# Install to /usr/local/bin
sudo make install-local
```

### Build for all platforms

```bash
make build-all
```

This creates binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Usage

### Basic Commands

```bash
# Deploy the stack
mediastack deploy --variant full --pull

# Check status
mediastack status

# View logs
mediastack logs jellyfin -f

# Stop the stack
mediastack stop

# Restart with image updates
mediastack restart --pull

# Extract API keys
mediastack apikeys --json

# Validate configuration
mediastack validate
```

### Global Flags

| Flag | Description |
|------|-------------|
| `-c, --config` | Path to config directory (default: auto-detect) |
| `-v, --variant` | Stack variant: `full`, `mini`, `no-vpn` |
| `--dry-run` | Show what would be done without executing |
| `--verbose` | Enable verbose output |

### Deploy Command

```bash
mediastack deploy [flags]

Flags:
  --pull            Pull images before deploying
  --no-directories  Skip directory creation
  --no-files        Skip config file copying
  --force           Force recreate all containers
  --prune           Prune unused resources (default: true)
```

### Status Command

```bash
mediastack status [flags]

Flags:
  -a, --all      Show all containers (including stopped)
  --health       Show only health status
  --json         Output as JSON
  -w, --watch    Continuously watch status
```

### Logs Command

```bash
mediastack logs [service] [flags]

Flags:
  -f, --follow       Follow log output
  -n, --tail string  Number of lines to show (default "100")
  -t, --timestamps   Show timestamps
```

## Configuration

The CLI looks for configuration in these locations:

1. Path specified by `--config` flag
2. `./base-working-files/` (relative to current directory)
3. `../base-working-files/` (parent directory)
4. `/docker/` (default server location)

### Required Files

- `.env` - Environment variables
- Docker Compose YAML in variant directory

### Stack Variants

| Variant | Description |
|---------|-------------|
| `full-download-vpn` | All traffic through Gluetun VPN |
| `mini-download-vpn` | Only downloads through VPN |
| `no-download-vpn` | Direct internet access |

## Improvements over restart.sh

1. **Fixed container stopping bug** - Properly lists running containers before stopping
2. **Better error handling** - Clear error messages and graceful failures
3. **Dry-run mode** - Preview changes before executing
4. **Individual service control** - Start/stop/restart specific services
5. **Health monitoring** - Built-in health status checking
6. **API key extraction** - Easy access to service API keys
7. **Cross-platform** - Builds for Linux, macOS, and Windows

## Development

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Build and run
make run
```

## Project Structure

```
cli/
├── cmd/mediastack/main.go     # Entry point
├── internal/
│   ├── cli/                   # Cobra commands
│   │   ├── root.go           # Root command and global flags
│   │   ├── deploy.go         # Deploy command
│   │   ├── stop.go           # Stop command
│   │   ├── restart.go        # Restart command
│   │   ├── status.go         # Status command
│   │   ├── logs.go           # Logs command
│   │   ├── pull.go           # Pull command
│   │   ├── validate.go       # Validate command
│   │   └── apikeys.go        # API keys command
│   ├── config/               # Configuration loading
│   │   ├── config.go         # Config struct
│   │   └── env.go            # .env parser
│   ├── docker/               # Docker operations
│   │   ├── client.go         # Docker SDK wrapper
│   │   └── compose.go        # Compose operations
│   └── stack/                # Stack operations
│       ├── directories.go    # Directory creation
│       └── files.go          # Config file copying
├── go.mod
├── Makefile
└── README.md
```
