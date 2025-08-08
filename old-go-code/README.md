# claude-vm

A CLI tool for managing remote Claude Code sessions with git integration.

## Installation

```bash
go install github.com/fidika/claude-vm@latest
```

## Usage

### Start a new session
```bash
# With GitHub repo
claude-vm start --repo github.com/user/repo

# With local directory
claude-vm start
```

### Connect to existing session
```bash
# Connect to specific session
claude-vm connect session-123

# Connect to most recent
claude-vm connect
```

### Other commands
```bash
claude-vm list              # List all sessions
claude-vm status session-123  # Check session status
claude-vm logs session-123    # View session logs
claude-vm delete session-123  # Delete session
```

## Architecture

claude-vm manages remote Docker containers running Claude Code. It provides:
- Git-based synchronization between local and remote
- Real-time streaming of Claude's output
- Interactive terminal sessions
- Session persistence and management

## Development

```bash
# Build
go build

# Run tests
go test ./...

# Install locally
go install
```