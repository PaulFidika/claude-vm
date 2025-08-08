# Popular Claude Code Containers for Safe Development

## Most Popular Solutions

### 1. **Claude Code YOLO** ⭐ Most Popular
- **GitHub**: thevibeworks/claude-code-yolo
- **Approach**: Docker wrapper that safely enables `--dangerously-skip-permissions`
- **Key Features**:
  - Full dev environment (Python, Node.js, Go, Rust)
  - Non-root execution with UID/GID mapping
  - Authentication passthrough (~/.claude, ~/.aws)
  - Safety checks (warns before running in $HOME)
  - Automatic localhost → host.docker.internal translation
  
```bash
# Usage
$ claude-yolo "build and test the project"
# Claude runs with full permissions inside container
```

### 2. **Claude Code Sandbox**
- **GitHub**: textcortex/claude-code-sandbox
- **Approach**: Web UI + Docker isolation
- **Key Features**:
  - Browser-based terminal with real-time streaming
  - Automatic credential discovery
  - Commit monitoring with notifications
  - Multiple container support
  - Interactive menu for git operations

```bash
# Usage
$ claude-sandbox start
# Access web UI at http://localhost:3000
```

### 3. **Anthropic Official Devcontainer**
- **Docs**: docs.anthropic.com/en/docs/claude-code/devcontainer
- **Approach**: VS Code devcontainer with security firewall
- **Key Features**:
  - Production-ready Node.js 20
  - Custom firewall (whitelist only)
  - Developer tools (git, ZSH, fzf)
  - VS Code integration
  - Session persistence

```json
// .devcontainer/devcontainer.json
{
  "image": "ghcr.io/anthropics/claude-code-devcontainer:latest",
  "features": {
    "ghcr.io/anthropics/devcontainer-features/claude-code:latest": {}
  }
}
```

## Comparison

| Feature | Claude YOLO | Claude Sandbox | Official Devcontainer |
|---------|-------------|----------------|---------------------|
| **Ease of Use** | ⭐⭐⭐ Simple CLI | ⭐⭐ Web UI | ⭐⭐ VS Code only |
| **Dev Tools** | Full stack | Basic | Node.js focused |
| **Security** | Container isolation | Container isolation | Firewall + isolation |
| **Git Integration** | Pass-through | Monitor + notify | Standard |
| **Multi-session** | No | Yes | No |
| **Official Support** | Community | Community | Anthropic |

## Key Insight

All solutions use the same core pattern:
1. Run Claude in Docker container
2. Enable `--dangerously-skip-permissions`
3. Mount only specific directories
4. Isolate from host system

**For claude-vm**: We're essentially building a remote version of these local containers - same safety principles but running on Fly.io instead of local Docker.

## Recommendations

- **Quick local dev**: Use Claude YOLO
- **Web-based workflow**: Use Claude Sandbox  
- **VS Code users**: Use official devcontainer
- **Remote development**: Use our claude-vm (combines best of all)