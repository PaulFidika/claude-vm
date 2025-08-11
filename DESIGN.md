# Claude VM Design Document

## Goal
A system that can run Claude Code (or any coding agent) in a remote VM, so that we the developer can turn off their laptop / go offline and Claude can still continue on its own. The developer should be able to (1) view and modify remote-VM files on their local machine, and (2) talk to Claude directly. This allows the developer to scale Claude out to many machines at once, and let the developer supervise its work.

We aim for it to be a CLI tool for developers, with a mobile-compatible web-interface for supervising AI.

We need to deal with two separate problems:

1. managing dev-environment lifecycles
2. allocating work within the dev-environments, along with collecting results


### Design

What level of abstraction should we be working at? Levels:
- Project-level artifacts (code, documentation, binaries)
- Workspace
- Task
- Conversation history

A 'task' is not well-defined here. It could be (1) just some random text I type into a webUI chatbox, (2) an open GitHub issue, (3) a collection of checklist items in a markdown file. We do not want to be opinionated and say 'this is what a task MUST BE!'. Ideally, our API should be able to hook into many different task-types; we'll probably create our own task-layer later, but it will hook into our workspace-management layer.

In the 'ideal dev setup' every task would get its own branch / workspace, and then we'd use git to merge them all together. I don't really like this though; (1) some users might not be using git (i.e., write a research paper, produce a bunch of PDFs), (2) if we have a dozen linearly dependent tasks, we need a dozen branches, PRs, and environments.

The only assumptions we will make are:
- Every workspace has one project; a project may have many workspaces
- Every workspace has 0 or more converational histories; every conversational history has only one workspace

This assumption is not limiting; if we want a chat to span multiple workspaces, we can always output project-level artificats (claude memory) that can be used in other workspaces / conversations.

```text
                     ┌───────────────────────────────────────────────────────────────┐
                     │         🔵 PROJECT: ecommerce-app (single repo)               │
                     │                                                               │
   ┌─────────────────│─────┐            ┌─────────────────┐            ┌─────────────│─────┐
   │  🟣 WORKSPACE:        │            │  🟣 WORKSPACE:  │            │  🟣 WORKSPACE:    │
   │    branch: auth       │            │    branch: dev  │            │    branch: dev    │
   │                       │            │                 │            │                   │
   │  ⚪ Claude: JWT setup │            │  ⚪ Claude:     │            │  ⚪ Claude:       │
   │  ⚪ Claude: login UI  │            │    cart UI      │            │    stripe setup   │
   │                       │            │  ⚪ Claude:     │            │  ⚪ Claude:       │
   │                       │            │    validation   │            │    webhook handler│
   │                       │            │                 │            │                   │
   └───────────────────────┘            └─────────────────┘            └───────────────────┘
                     │                                                               │
                     └───────────────────────────────────────────────────────────────┘
```



### Monetization

Probably the best bet is (1) an open-source self-hosted tool anyone can use, and then (2) a hosted "premium" solution we can sell to enterprises. This is how DevPod works.

Monetization:
- Hosting: customers may pay us to host their data (Pro / Enterprise Plan).
- Model Routing: first party providers (OpenAI, Anthropic), and third-party providers (Fireworks, Groq) may pay us a referrel fee to send our user traffic to their endpoints.

Acqusition:
- Cloud Providers (Digital Ocean, AWS) would value us.
- AI Providers (OpenRouter) would value us.


### Scope of Project

User scope:
- Task Authoring: what work should Claude do next?
- Devcontainer Configuration: what all do we need in the enviornment to run the codebase?
- Source Code Review: which changes should be committed?

Our scope:
- Adding Agent to Devcontainer: adding a secure version of Claude Code, or other coding agents, to the user's provided devcontainer spec. The agent should have maximum permissions, but the user's code, secrets, and machine should be kept safe.
- Agent Lifecycle: providing the user with input / output to the remote agent, keeping track of file changes, persisting changes.
- Provider Orchestration: launching containers + volumes on local or remote providers.
- Workspace Lifecycle: stopping containers when the task is complete so they don't consume resources.
- UI: mobile UI + remote VSCode + terminal (ssh) for supervising claude and viewing the workspace.

Provider scope:
- Infrastructure lifecycle (VM, network)
- Volume management

---

### OpenAI Codex notes:
- Turns off the container immediately after the task is completed, meaning the continaer needs to be started up again for every user-interaction.
- Takes about 60 seconds to startup the container again; runs a bunch of installs, meaning the dev environment (like node_modules) is not being cached; only file changes are.
- Does not support queued-messages yet. You literally cannot talk to it while it's working; you have to wait for it to finish.
- For each session it really only stores whatever file chanes are tracked in git. Non-git changes are not tracked or displayed.
- Honestly this tool sucks and is barely functional. I can't believe a billion-dollar company made this.

---

### Claude Code UIs

The main point is to allow you to use Claude Code outside of the CLI.

- getAsterisk/Claudia: 11k stars. Rust and Typescript. A Tauri-based desktop app

- siteboon/claudecodeui: react-based, supports both desktop and mobile web views. GPL license.

- sugyan/claude-code-webui: react based, not quite as sexy as Claude-Code-UI (in my opnion). MIT license

- wbopan/cui: react-based; UI is a complete clone of OpenAI's codex. Apache license. My favorite so far. This is not just a UI; it also orchestrates Claude instances that run locally. There is no isolation between Claude Code instances (they all work in the same git branch). All claude code instances run locally, although the server itself can be viewed remotely.

---

### Claude Code Containers:

The maint point of these is to make it easier to run claude-code with the --dangerously-skip-permissions flag, so you do not need to manually approve stuff. Running claude code in a container, with its own copy of the code, means that claude cannot destroy the codebase (or your computer!) easily. A still-valid concern is prompt injection hacks caused by rogue websites that can exfiltrate your code or env-secrets.

Containers:
- RchGrav/claudebox: sounds the most promising
- eastlondoner/claude-yolo
- thevibeworks/claude-code-yolo
- textcortex/claude-code-sandbox
- VishalJ99/claude-docker

Devcontainers: 
- anthropics/claude-code/tree/main/.devcontainer: official devcontainer from anthropic
- anthropics/devcontainer-features: official devcontainer feature for claude code

Misc:
- claude-did-this/claude-hub: webhooks for Github. No longer maintained.

### Competitors:

People building similar projects to us.

- dtormoen/tsk: 50 stars: built in Rust. The closest to what we're building. It clones the entire repo (not git worktrees) locally, mounts it in a container, and then runs claude code inside the container.

- dagger/container-use: 3k stars, written in Go. MCP Server that gives any agent (running in your terminal) the ability to create a git worktree, mount it in a container, then modify the code inside of it. Claude DOES NOT run inside of the container. This works with local containers only.

- smtg-ai/claude-squad: 4k stars, written in Go: Uses git worktrees, not containers. It runs multiple agents on your machine in separate terminals. Provides a terminal UI.

- superagent-ai/vibekit: 1k stars: Written in Typescript. Intercepts Claude's commands and runs them inside of a local or remote container. Claude runs outside of the container (on the host machine) not inside of it. It's like a transparent virtualization layer; dangerous stuff happens inside the containers. Unlike dagger's container-use, this is not an MCP-server.

- parruda/claude-swarm: 1k stars; written in Ruby????. Over-engineered garbage; basically just an MCP server so multiple claude instances running on the same machine can talk to each other.

### Tools That Are The Most Useful:

wbopan/cui: could be used as our frontend
dtormoen/task: copy some of its design?

---

### User Interactions

- VSCode connects into the workspace; users can review file changes and perform changes (write git commit, ask Claude to make changes, make changes manually, etc). Unfortunately VSCode is not good at switching between environments (no multi-tab view).
- Users can SSH into the workspace, and talk to Claude and issue commands. Unfortunately it's hard to view file changes from the CLI. (Lazygit might work? Some terminal UI?)
- Our WebUI can fill the gap; users can review files, talk to Claude, and issue shell commands.

---

### Marketing:

- Submit commits to the 2 main claude-awesome repos to have our project listed as well
https://github.com/hesreallyhim/awesome-claude-code
https://github.com/jqueryscript/awesome-claude-code

---

### DevPod notes:
- stores a list of your running workspaces locally, and queries providers to get their status.
- DevPod Pro stores a list of your providers in your DevPod account.

---

## CLI Design (Alternative)

```bash
claude-vm - Run Claude Code on remote VMs

USAGE:
  claude-vm <command> [options]

COMMAND TREE:
├── login                                               # OAuth login to claude-vm.com account
│
├── workspace                                           # Workspace management
│   ├── up [workspace-id|repo-url|path/to/local-repo]   # Create and start new workspace
│   │   ├── --cloud <name>                              #   Cloud platform (docker|digitalocean|fly|aws)
│   │   ├── --agent <name>                              #   Agents to include (claude|codex|qwen|goose|gemini)
│   │   ├── --devcontainer <path>                       #   DevContainer spec (takes priority over --image)
│   │   ├── --image <name>                              #   Docker image override (fallback)
│   │   ├── --region <region>                           #   Cloud region (optional)
│   │   ├── --size <size>                               #   Machine size (optional)
│   │   ├── --branch <name>                             #   Git branch to checkout
│   │   └── --non-interactive                           #   Disable prompts, fail if config missing
│   │
│   ├── list                                  # List all workspaces 
│   │   ├── -l, --long                        #   Show detailed workspace info
│   │   └── --status <filter>                 #   Filter by status (running|stopped|error)
│   │
│   ├── down <workspace-id>                   # Stop workspace (can resume later)  
│   │   └── -f, --force                       #   Force stop without graceful shutdown
│   │
│   └── delete <workspace-id>                 # Delete provider resource, keep S3 backup
│       ├── -y, --yes                         #   Skip confirmation prompt
│       └── --purge                           #   Delete provider resource AND S3 backup
│
├── ssh <workspace-id>                        # SSH into workspace container
│   ├── --user <name>                         #   SSH username (default: current user)
│   └── -p, --port <port>                     #   SSH port (default: 22)
│
├── chat [conversation-name]                   # Talk to coding agents (drop into named conversation)
│   ├── -w, --workspace <workspace-id>        #   Interactive: list/select conversations in workspace  
│   ├── --new <conversation-name>             #   Create new conversation (name required, like git branch -b)
│   ├── --agent <agent-name>                  #   Agent for new conversation (must be available in workspace)
│   ├── --list                                #   List all conversations with their names  
│   ├── --rename <old-name> <new-name>        #   Rename conversation (like git branch -m)
│   └── --non-interactive                     #   Send stdin message, output response to stdout, exit (for piping)
│
├── agent                                     # Agent configuration management
│   ├── set-config <agent-name>              # Configure agent settings
│   │   └── --option <key=value>             #   Agent-specific configuration options
│   ├── list                                 # Show all agents with their configuration
│   └── clear-config <agent-name>            # Clear agent configuration
│
├── cloud                                     # Cloud platform management
│   ├── set-config <cloud-name>              # Add or update cloud configuration
│   │   └── --option <key=value>             #   Set cloud options (api-key, region, etc.)
│   ├── list                                 # List all clouds with configuration
│   └── clear-config <cloud-name>            # Clear all options for cloud
│
├── provider                                  # LLM API provider management  
│   ├── set-config <provider-name>           # Add or update provider configuration
│   │   └── --option <key=value>             #   Set provider options (api-key, oauth-token, etc.)
│   ├── list                                 # List all providers with configuration
│   └── clear-config <provider-name>         # Clear all options for provider
│
├── web                                       # Open web interface (you can pick between all workspaces)

EXAMPLES:
  # Workspace management  
  claude-vm workspace up .                   # Auto-detect .devcontainer/ or generate; use default cloud and default agent
  claude-vm workspace up . --cloud digitalocean --agent claude  # DO with Claude agent only
  claude-vm workspace up . --cloud fly --agent codex,goose     # Fly with Codex + Goose agents  
  claude-vm workspace up . --agent null                       # Use default cloud, no agents in devcontainer
  claude-vm workspace up . --devcontainer .devcontainer/ai-agents.json  # Custom devcontainer
  claude-vm workspace up . --devcontainer https://github.com/user/devcontainers/claude.json
  claude-vm workspace up . --image node:18-alpine         # Simple Docker image override
  
  # Agent specification examples
  claude-vm workspace up . --agent claude                 # Bakes claude into container (explicitly)
  claude-vm workspace up . --agent claude,goose,gemini    # Bakes multiple agents into container
  claude-vm workspace up github.com/user/repo --agent goose --cloud fly  # Goose on Fly.io
  claude-vm workspace up bold-fire-1234                   # Restart existing workspace
  claude-vm workspace list --status running --long        # List running workspaces
  claude-vm workspace down 3f2504e0bb11                   # Stop workspace, preserve state
  claude-vm workspace up 3f2504e0bb11                     # Resume stopped workspace
  claude-vm workspace delete 45678901 --yes               # Delete, keep S3 backup
  claude-vm workspace delete quiet-lake-5678 --purge      # Complete deletion including S3 backup
  
  # Cloud platform configuration
  claude-vm cloud set-config digitalocean --option api-key=dop_xxx --option region=nyc1
  claude-vm cloud set-config fly --option api-key=fly_xxx --option region=iad
  claude-vm cloud set-config aws --option access-key=AKIA... --option secret-key=xxx
  claude-vm cloud set-config docker  # Local Docker (no credentials needed)
  claude-vm cloud list                                     # Show all clouds with config
  claude-vm cloud clear-config digitalocean               # Clear DO configuration
  
  # LLM API provider configuration
  claude-vm provider set-config anthropic --option api-key=sk-ant-xxx
  claude-vm provider set-config anthropic --option oauth-token=/path/to/token.json
  claude-vm provider set-config openai --option api-key=sk-xxx
  claude-vm provider set-config openai --option oauth-token=/path/to/token.json
  claude-vm provider set-config google --option api-key=AIza-xxx
  claude-vm provider list                                  # Show all providers with config
  claude-vm provider clear-config anthropic               # Clear Anthropic configuration
  
  # Agent configuration (agents use their provider's credentials)
  claude-vm agent set-config claude --option model=opus --option auth_preference=oauth
  claude-vm agent set-config claude --option model=sonnet --option auth_preference=api-key
  claude-vm agent set-config codex --option model=gpt-4o --option auth_preference=oauth
  claude-vm agent set-config qwen --option model=qwen3-coder-plus --option temperature=0.8
  claude-vm agent set-config goose --option provider=openrouter --option model=anthropic/claude-3.5-sonnet
  claude-vm agent list                                           # Show all agents with config
  claude-vm agent clear-config claude                            # Clear Claude configuration
  
  # SSH access
  claude-vm ssh quiet-lake-5678 --user developer
  claude-vm ssh 8a9b2c3d4e5f --port 2222
  
  # Web interface
  claude-vm web                                # Open web interface (pick between workspaces)
  
  # Chat with coding agents (git-like syntax) 
  claude-vm chat --new "fix-auth-bug"                     # Create new conversation (agent selected from available)
  claude-vm chat --new "refactor-database" --agent goose  # Create new conversation with specific agent
  claude-vm chat "fix-auth-bug"                           # Continue conversation (interactive terminal)
  
  # Interactive mode (drops into terminal, stdin becomes initial prompt)
  claude-vm chat "fix-auth-bug"                           # Start interactive session
  echo "debug this error" | claude-vm chat "fix-auth-bug" # Start with stdin as initial prompt
  
  # Non-interactive mode (stdin → stdout, for piping)
  echo "what's the status?" | claude-vm chat "refactor-database" --non-interactive
  cat error.log | claude-vm chat "debug-session" --non-interactive | grep -i "solution"
  
  claude-vm chat --list                                    # List all conversations with names
  # Output: 
  # fix-auth-bug (claude) - 45 messages - Last: 2 hours ago
  # refactor-database (goose) - 12 messages - Last: 1 day ago
  
  claude-vm chat --rename "fix-auth-bug" "auth-system-overhaul"  # Rename conversation (like git branch -m)
  
  # Git-like branch workflow
  claude-vm chat --list                                    # List conversations (like git branch)
  claude-vm chat --new "feature-user-auth"                # Create conversation with default agent (like git branch -b)
  claude-vm chat --new "feature-user-auth" --agent goose  # Create conversation with specific agent
  claude-vm chat "feature-user-auth"                       # Switch to conversation (like git checkout)
  claude-vm chat --rename "feature-user-auth" "auth-v2"   # Rename conversation (like git branch -m)

GLOBAL OPTIONS:
  --non-interactive             # Disable interactive prompts (for CI/automation)
  --help, -h                    # Show help
```

---

## Interactive Configuration System

claude-vm uses interactive prompts to guide users through missing configuration, making it beginner-friendly while remaining automation-safe.

**Note:** The `agent set-config`, `cloud set-config`, and `provider set-config` commands are **always explicit** and do not have interactive modes (similar to `git config` or `docker config`).

**Config Command Behavior:**
```bash
# Config commands require explicit parameters
claude-vm agent set-config claude --option provider=anthropic   # ✅ Works
claude-vm agent set-config claude                               # ❌ Fails with usage error

# ERROR: Missing required --option flag  
#   Usage: claude-vm agent set-config <agent-name> --option <key=value>
#   Example: claude-vm agent set-config claude --option provider=anthropic
```

### Interactive Mode (Default Behavior)

When required configuration is missing, claude-vm automatically drops into interactive mode:

```bash
# Example: User runs command without configuration
claude-vm workspace up .
# ✓ Cloud: docker (default)
# ✓ Agent: claude (built-in default)
# ✗ Provider: anthropic not configured for claude agent

┌─ Claude Code Setup Required ───────────────────────────────┐
│ Claude agent requires Anthropic API configuration.             │
│                                                              │
│ Choose authentication method:                                │
│ 1) OAuth login (recommended) - uses ~/.claude-code/         │
│ 2) API key - enter Anthropic API key                        │
│                                                              │
│ Select [1-2]: 1                                              │
└──────────────────────────────────────────────────────────────┘

❯ Opening browser for Anthropic OAuth...
❯ Waiting for authentication...
✓ Successfully configured Anthropic provider for claude agent
❯ Continuing with workspace creation...
```

**Conversation Naming Example:**
```bash
# User forgets to provide name (like git checkout without branch name)
claude-vm chat --new
# ERROR: Conversation name required
#   Usage: claude-vm chat --new <conversation-name> [--agent agent-name]
#   Example: claude-vm chat --new "fix-auth-bug" --agent claude

# Correct usage (like git branch -b branch-name)  
claude-vm chat --new "fix-auth-bug"
# ✓ Workspace agents: claude, goose
# ✓ Conversation name: fix-auth-bug
# ✓ Agent: claude (built-in default)

❯ Created conversation "fix-auth-bug" with Claude agent
❯ Connecting to interactive terminal...

# Interactive conversation selection when no name provided
claude-vm chat
┌─ Select Conversation ──────────────────────────────────────┐
│ Choose existing conversation or create new:                │
│                                                            │
│ 1) fix-auth-bug (claude) - 45 messages - 2 hours ago     │
│ 2) refactor-database (goose) - 12 messages - 1 day ago   │
│ 3) Create new conversation...                              │
│                                                            │
│ Select [1-3]: 3                                           │
└────────────────────────────────────────────────────────────┘

# Agent selection when default agent not available in workspace
┌─ Select Agent ─────────────────────────────────────────────┐
│ Default agent 'claude' not available in this workspace.    │
│ Available agents:                                          │
│                                                            │
│ 1) goose - Block's Goose AI                               │
│ 2) gemini - Google's Gemini CLI                           │
│ 3) codex - OpenAI Codex                                   │
│                                                            │
│ Select agent [1-3]: 1                                     │
└────────────────────────────────────────────────────────────┘
```

**Interactive Triggers:**
- Missing LLM provider configuration for selected agent
- Missing cloud platform credentials (when not using docker)  
- Missing GitHub authentication for private repositories
- No conversation name provided for bare `claude-vm chat` command
- Default agent not available in workspace (conversation creation)
- Ambiguous configuration choices (multiple valid options)

### Non-Interactive Mode (CI/Automation)

Use `--non-interactive` flag to disable prompts and fail fast with clear errors:

```bash
# Fails immediately with actionable error message
claude-vm workspace up . --non-interactive
# ERROR: Anthropic provider not configured for agent 'claude'
#   Run: claude-vm provider set-config anthropic --option api-key=sk-ant-xxx
#   Or:  claude-vm agent set-config claude --option auth_preference=oauth

# Conversation naming in non-interactive mode
claude-vm chat --new --non-interactive
# ERROR: Conversation name required in non-interactive mode
#   Run: claude-vm chat --new "my-conversation-name" --agent claude

# Agent selection in non-interactive mode (when default not available)
claude-vm chat --new "test-conversation" --non-interactive  
# ERROR: Default agent 'claude' not available in workspace
#   Available agents: goose, gemini, codex
#   Run: claude-vm chat --new "test-conversation" --agent goose --non-interactive
```

**Commands Supporting `--non-interactive`:**
- `workspace up` - Workspace creation and configuration
- `chat --new` - Starting conversations (requires name + agent if default not available)
- `chat [conversation-name]` - Conversation continuation by name

**Global Flag:**
```bash
# Apply to subcommands that support interactive mode
claude-vm --non-interactive workspace up .                            # Uses default agent
claude-vm --non-interactive workspace up . --agent "claude,goose"     # Multiple agents
claude-vm --non-interactive chat --new "automated-test" --agent claude
echo "run the tests" | claude-vm --non-interactive chat "fix-auth-bug"

# Config commands are always explicit (no interactive mode)
claude-vm agent set-config claude --option provider=anthropic         # No --non-interactive needed
claude-vm cloud set-config fly --option api-key=xxx                   # Always explicit
claude-vm provider set-config anthropic --option api-key=sk-ant-xxx   # Always explicit
```

**Benefits:**
- **Interactive**: Beginner-friendly, discoverable, reduces friction for workspace and chat operations
- **Non-Interactive**: CI/CD safe, predictable, scriptable for automation
- **Explicit Config**: Config commands are always explicit (like git config), no interactive mode needed
- **Clear Errors**: Actionable error messages with exact commands to fix issues

## Default Agent Configuration

claude-vm uses **sensible defaults** with explicit override options:

**Default Behavior:**
- **Default cloud**: `docker` (local Docker, no credentials needed)
- **Default agent**: `claude` (most popular agent)

**Workspace Creation Behavior:**
```bash
# Uses built-in defaults
claude-vm workspace up .
# → Cloud: docker (built-in default) 
# → Agent: claude (built-in default)

# Explicit overrides
claude-vm workspace up . --agent "claude,goose"     # Multiple agents
claude-vm workspace up . --agent "codex"            # Single different agent
claude-vm workspace up . --agent "null"             # No agents in devcontainer
claude-vm workspace up . --cloud fly                # Different cloud
```

**Conversation Creation Behavior:**
```bash
# Uses claude agent when available in workspace
claude-vm chat --new "hello-world"
# → Agent: claude (if available in workspace)

# Explicit agent specification
claude-vm chat --new "hello-world" --agent goose
# → Agent: goose (if available in workspace)
```

**Agent-Provider Relationships:**
- **Agents** are configured to use specific **providers**
- **Providers** have no concept of "default" - they're just configured endpoints
- Example: Claude agent configured to use Anthropic provider with API key

```bash
# Providers are just endpoints
claude-vm provider set-config anthropic --option api-key=sk-ant-xxx

# Agents are configured to use providers
claude-vm agent set-config claude --option provider=anthropic
```

**Benefits:**
- **Simple defaults**: One sensible default agent (claude), explicit when you want different
- **Explicit multi-agent**: Use `--agent "claude,goose"` when you want multiple
- **Explicit no-agent**: Use `--agent "null"` when you want none
- **Override flexibility**: Easy to specify exactly what you want when defaults don't fit

---

### Volume Persistence

Every cloud platform allows us to mount volumes; we mount them in /workspace. We can stop a container, and then start it again with the same volume preserving all data (such as edited code).

```dockerfile
# Dockerfile
WORKDIR /workspace  # This is your persistent volume

# Store these in volume:
# - /workspace/projects (git repos)
# - /workspace/.cache (package caches)
# - /workspace/.config (user configs)

# Optimize with Symlinks

# In your container startup script
ln -s /workspace/.cache/npm ~/.npm
ln -s /workspace/.cache/pnpm ~/.pnpm
ln -s /workspace/.cargo ~/.cargo
ln -s /workspace/go ~/go
```

### Persistence with S3 Object Store

The only things we really need to save from each devcontainer session is:

- 1. Logs from Claude
- 2. Git diffs that were not committed and pushed (all files tracked by git)

We probably also want to store files that are not tracked by git; in the future non-devs will use things outside of git (such as excel spreadsheets or PDFs). We should not assume that everything not tracked by git is worthless.

This means we do not really need persistent volumes; if the user wants to start up workspace-abc123 again, we can just start a fresh container, pull in the files it had (using git? tarball?) and the pull in the above from our S3 object store to recreate state. This is a lot cheaper.

*Issue:* Also, when you 'delete' a workspace, what does that mean? Do we delete the volume + object-store + record of it from history?

### Listing Workspaces

How does clauce-vm run it's `list` command? We can store local files on the dev's machine describing the workspaces that exist. For pro users, we can store an authoratative state in our database for each user / enterprise. This means devs are not tied to a specific machine, and can work together. We will need to store:

- 1. list of all workspaces, and their statuses
- 2. S3 object store with logs + files

---

### Container Specification Priority

Workspace containers are configured using this priority order:

1. **--devcontainer <path>** (highest priority) - Explicit devcontainer file
2. **--image <name>** - Simple Docker image override
3. **Auto-detected** `.devcontainer/devcontainer.json` - Project's devcontainer
4. **Generated devcontainer** (lowest priority) - Created from repo analysis

**Examples:**
```bash
# Priority 1: Explicit devcontainer (wins over everything)
claude-vm workspace up . --devcontainer .devcontainer/ai-agents.json

# Priority 2: Auto-detected (if .devcontainer/devcontainer.json exists)
claude-vm workspace up .  # Uses .devcontainer/devcontainer.json automatically

# Priority 3: Docker image override (fallback when no devcontainer)
claude-vm workspace up . --image node:18-alpine

# Priority 4: Generated (when nothing else available)
claude-vm workspace up .  # Analyzes repo and generates devcontainer
```

**Key Architecture Decision: Fixed Agent Set**

Agents are **permanently baked into the workspace container** at creation time:

```bash
# This bakes claude, goose, and gemini into the container
claude-vm workspace up github.com/user/repo --agent claude,goose,gemini

# Later usage - only these agents are available:
claude-vm new-chat workspace-123 claude    # ✅ Available
claude-vm new-chat workspace-123 goose     # ✅ Available  
claude-vm new-chat workspace-123 gemini    # ✅ Available
claude-vm new-chat workspace-123 codex     # ❌ ERROR: codex not baked into this workspace

# Cannot add agents later - would require new workspace
claude-vm workspace up github.com/user/repo --agent claude,goose,gemini,codex  # New workspace
```

**Benefits:**
- **Immediate availability**: Agents ready when container starts
- **Predictable environment**: Fixed set of tools per workspace  
- **No runtime failures**: Cannot request unavailable agents
- **Build-time optimization**: Docker layer caching for agent installation

### Building Devcontainer

*Parsing User's Devcontainer.json:*
- Users should supply a devcontainer.json optionally, but we should also be able to generate a good starter image by parsing the repo.

*Inserting Processes into Devcontainer:*

**Agent Installation via Devcontainer Features:**
```json
// Generated devcontainer.json based on --agent flag
// Example: claude-vm workspace up . --agent claude,goose,gemini
{
  "name": "claude-vm-workspace",
  "image": "mcr.microsoft.com/devcontainers/universal:2-linux",
  "features": {
    // Install only requested agents
    "ghcr.io/anthropics/devcontainer-features/claude-code:latest": {},
    "ghcr.io/block/devcontainer-features/goose:latest": {},
    "ghcr.io/google/devcontainer-features/gemini:latest": {},
    
    // Our custom feature for workspace management
    "ghcr.io/claude-vm/devcontainer-features/workspace-manager:latest": {
      "enabledAgents": ["claude"],  // From --agent flag or built-in default
      "workspaceId": "${localEnv:WORKSPACE_ID}"
    }
  }
}
```

**How Multi-Agent Installation Works:**

1. **Workspace Creation**: User specifies agents when creating workspace:
   ```bash
   claude-vm workspace up . --agent claude,goose,gemini
   ```

2. **DevContainer Generation**: claude-vm generates devcontainer.json with requested agents:
   - Only specified agents get devcontainer features
   - Each agent gets its own binary/runtime baked into container
   - Credentials injected as environment variables during build
   - No conflicts between agents

3. **Fixed Agent Set**: Agents are baked into the container and cannot be changed:
   ```bash
   # These agents are available in the workspace
   claude-vm new-chat workspace-123 claude    # Uses baked-in claude
   claude-vm new-chat workspace-123 goose     # Uses baked-in goose  
   claude-vm new-chat workspace-123 gemini    # Uses baked-in gemini
   
   # This would fail - agent not baked into container
   claude-vm new-chat workspace-123 qwen      # ERROR: qwen not available
   ```

4. **Credential Injection**: Each baked-in agent gets its configured credentials:
   ```bash
   # Only for agents specified during workspace creation
   CLAUDE_CREDENTIALS=/workspace/.claude/.credentials.json
   GOOGLE_API_KEY=xxx            # For gemini (if requested)
   GOOSE_PROVIDER=anthropic      # For goose (if requested)
   ```

### Provider Orchestration

(Devpod already has this code all written; we do not need to write it again.)

### API Server Architecture & Authentication

claude-vm supports **three distinct access patterns** with unified authentication via GitHub:

1. **Local CLI → Devcontainers** (CLI generates bearer tokens)
2. **Hosted Service → Devcontainers** (OAuth proxy with bearer tokens)  
3. **Direct Browser → Container** (GitHub organization membership)

## Multi-Pattern Authentication Strategy

### Pattern 1: Local CLI → Devcontainers
**Authentication Flow:**
```bash
# User runs: claude-vm workspace up
# 1. CLI generates unique bearer token for container
# 2. Token injected into container at startup via environment
# 3. Token stored locally: ~/.claude-vm/workspaces.json
# 4. Local web server uses stored tokens to call container APIs

# Example ~/.claude-vm/workspaces.json:
{
  "bold-fire-1234": {
    "url": "https://bold-fire-1234.fly.dev",
    "bearer_token": "cvt_abc123...",  # CLI-generated token
    "status": "running"
  }
}

# API Request:
# GET https://bold-fire-1234.fly.dev/api/conversations
# Authorization: Bearer cvt_abc123...
```

### Pattern 2: Hosted Service → Devcontainers  
**Authentication Flow:**
```bash
# User creates workspace via claude-vm.com (hosted service)
# 1. User authenticates: OAuth with GitHub (organization membership verified)
# 2. Hosted service generates bearer token for container
# 3. Token injected into container at startup
# 4. Token stored in hosted service database
# 5. Hosted service proxies user requests to container APIs

# User → Hosted Service: GitHub OAuth + organization membership check
# Hosted Service → Container: Bearer token (same as CLI pattern)
```

### Pattern 3: Direct Browser → Container (GitHub Organization Auth)
**Authentication Flow:**
```yaml
Container GitHub OAuth Setup:
  # Each container has GitHub OAuth app configuration
  GITHUB_CLIENT_ID: "container-oauth-app-id"
  GITHUB_CLIENT_SECRET: "oauth-secret"
  GITHUB_ORG_NAME: "your-company"           # Required organization
  GITHUB_ORG_REQUIRED: "true"               # Enforce membership
  
Direct Browser Access:
  1. User visits: https://bold-fire-1234.fly.dev
  2. Container redirects: GitHub OAuth login
  3. GitHub OAuth callback with user info + organization membership
  4. Container validates: User is member of required organization
  5. Container sets session cookie with GitHub user info
  6. User accesses workspace with authenticated session

Organization Membership Validation:
  # Container calls GitHub API to verify membership
  GET https://api.github.com/orgs/{org}/members/{username}
  Authorization: token {oauth_token}
  # Response: 200 (member) or 404 (not member or private membership)
```

## Workspace API Server (runs inside each container)

**Container Configuration:**
```yaml
Port: 8080
Base Authentication Methods:
  1. Bearer Token: CLI/hosted service access (Authorization: Bearer cvt_...)
  2. GitHub OAuth Session: Direct browser access (session cookies)
  3. SSH Key: Terminal access (ssh user@container.host)

GitHub OAuth Environment Variables:
  GITHUB_CLIENT_ID: "oauth-app-client-id"
  GITHUB_CLIENT_SECRET: "oauth-app-secret"  
  GITHUB_ORG_NAME: "required-organization"
  GITHUB_ORG_REQUIRED: "true|false"
  
Authentication Priority:
  1. Valid bearer token → immediate API access
  2. Valid GitHub session cookie → API access (organization verified)
  3. No auth → redirect to GitHub OAuth (direct browser access)
  4. Invalid/expired auth → 401 Unauthorized
```

**API Endpoints:**
```yaml
# Authentication Endpoints (GitHub OAuth for direct browser access)
GET    /auth/github                             # Initiate GitHub OAuth flow
GET    /auth/github/callback                    # OAuth callback handler  
GET    /auth/logout                             # Clear session, redirect to login
GET    /auth/status                             # Current auth status + user info

# Core Web Interface Support  
WS     /api/terminal/{conversation-name}        # Stream tmux session to browser
GET    /api/conversations/{name}/data           # Get parsed conversation data (JSONL → JSON)
POST   /api/conversations/{name}/messages       # Send message via tmux + parse response
SSE    /api/events                              # Real-time conversation updates

# Conversation Management
GET    /api/conversations                       # List conversations from conversations.json
POST   /api/conversations                       # Create conversation + tmux session
  - name: "fix-auth-bug" (required)
  - agent: "claude" (required) 
PUT    /api/conversations/{name}/rename         # Update conversation name
DELETE /api/conversations/{name}                # Delete conversation + kill tmux session

# Workspace Operations
GET    /api/status                              # Workspace health + running tmux sessions
POST   /api/backup                              # Trigger S3 backup of workspace files
GET    /api/backup/status                       # Check backup progress/status
POST   /api/shutdown                            # Graceful workspace shutdown

# Event Stream (SSE /api/events):
Events:
  - conversation.message    # New agent responses (parsed from JSONL/terminal)
  - conversation.created    # New conversation started  
  - conversation.renamed    # Conversation name changed
  - backup.completed        # S3 backup finished
  - workspace.idle          # No activity detected (approaching auto-shutdown)
  - auth.login              # User logged in via GitHub OAuth
  - auth.logout             # User logged out
```

## Centralized Task Server (claude-vm.com)

**Hosted Service Configuration:**
```yaml
Purpose: Coordinate workspaces + GitHub OAuth proxy for hosted users
Database: PostgreSQL
GitHub OAuth App: "claude-vm-hosted-service"

Services:
  # User Authentication (GitHub OAuth)
  - GitHub organization membership verification
  - Session management with organization context
  - Multi-organization support for enterprise users
  
  # Workspace Registry  
  - Track all user workspaces across providers
  - Store workspace metadata (provider, status, URL, org context)
  - Handle workspace discovery for web UI
  
  # Bearer Token Management
  - Generate unique bearer tokens for user workspaces
  - Token rotation and expiration management
  - Workspace access control by organization membership
  
  # GitHub Integration
  - Store GitHub App credentials for managed repositories  
  - Generate fresh installation tokens for workspace git access
  - Organization repository access management
  
  # S3 Coordination
  - Generate pre-signed URLs for workspace backups (org-scoped)
  - Manage backup lifecycle policies per organization
  - Handle restore operations with access control
```

**Hosted Service API Endpoints:**
```yaml
# User Authentication (GitHub OAuth)
POST   /api/auth/github                    # Initiate GitHub OAuth flow
GET    /api/auth/github/callback           # OAuth callback handler
POST   /api/auth/logout                    # Clear session
GET    /api/auth/user                      # Current user + organization info

# Organization Management  
GET    /api/orgs                           # List user's organizations
POST   /api/orgs/{org}/workspaces          # Create workspace in organization context
GET    /api/orgs/{org}/workspaces          # List organization workspaces
GET    /api/orgs/{org}/members             # List organization members (admin only)

# Workspace Management (organization-scoped)
GET    /api/workspaces                     # List user's accessible workspaces
POST   /api/workspaces                     # Register new workspace  
DELETE /api/workspaces/{id}                # Unregister workspace
GET    /api/workspaces/{id}/access         # Get workspace bearer token
POST   /api/workspaces/{id}/invite         # Invite organization member to workspace

# GitHub Integration
POST   /api/github/refresh-token           # Get fresh GitHub token (managed mode)
GET    /api/github/repos                   # List accessible repositories
POST   /api/github/repos/{repo}/workspace  # Create workspace from repository

# S3 Backup (organization-scoped)
GET    /api/backup/upload-url              # Get S3 pre-signed upload URL
GET    /api/backup/download-url            # Get S3 pre-signed download URL
POST   /api/backup/restore                 # Restore workspace from backup
```

## Authentication Implementation Examples

### GitHub Organization Membership Validation
```javascript
// Container-side organization membership check
async function validateGitHubOrgMembership(accessToken, username) {
  const requiredOrg = process.env.GITHUB_ORG_NAME;
  if (!requiredOrg) return true; // No organization requirement
  
  try {
    const response = await fetch(
      `https://api.github.com/orgs/${requiredOrg}/members/${username}`,
      { headers: { 'Authorization': `token ${accessToken}` } }
    );
    return response.status === 200; // 200 = member, 404 = not member/private
  } catch (error) {
    console.error('GitHub org membership check failed:', error);
    return false;
  }
}

// OAuth callback handler in container
app.get('/auth/github/callback', async (req, res) => {
  const { code } = req.query;
  
  // Exchange code for access token
  const tokenResponse = await exchangeCodeForToken(code);
  const { access_token } = tokenResponse;
  
  // Get user info
  const userResponse = await fetch('https://api.github.com/user', {
    headers: { 'Authorization': `token ${access_token}` }
  });
  const user = await userResponse.json();
  
  // Validate organization membership
  const isMember = await validateGitHubOrgMembership(access_token, user.login);
  if (!isMember) {
    return res.status(403).send(`Access denied. Must be member of ${process.env.GITHUB_ORG_NAME} organization.`);
  }
  
  // Set session cookie
  req.session.user = user;
  req.session.githubToken = access_token;
  res.redirect('/'); // Redirect to workspace
});
```

### Bearer Token Generation (CLI Pattern)
```javascript
// CLI generates workspace bearer token
function generateWorkspaceToken(workspaceId, userId) {
  const payload = {
    workspace: workspaceId,
    user: userId,
    type: 'workspace',
    created: Date.now(),
    expires: Date.now() + (7 * 24 * 60 * 60 * 1000) // 7 days
  };
  return jwt.sign(payload, process.env.WORKSPACE_SECRET);
}

// Container validates bearer token
function validateBearerToken(token) {
  try {
    const payload = jwt.verify(token, process.env.WORKSPACE_SECRET);
    if (payload.expires < Date.now()) throw new Error('Token expired');
    return payload;
  } catch (error) {
    return null;
  }
}
```

## Authentication Security Strategy

**Multi-Layer Security:**
```yaml
Layer 1 - User Authentication:
  GitHub OAuth: Primary authentication method
  Organization Membership: Access control mechanism  
  Session Management: Secure cookies with CSRF protection
  
Layer 2 - Workspace Access:
  Bearer Tokens: Unique per workspace, limited lifetime
  Token Rotation: Regular refresh for long-running workspaces
  Scope Limitation: Tokens only valid for specific workspace
  
Layer 3 - Container Security:
  Environment Isolation: Separate containers per workspace
  Network Segmentation: Containers cannot access each other
  Resource Limits: CPU/memory limits per container
  
Layer 4 - Data Security:  
  S3 Encryption: All backups encrypted at rest
  Pre-signed URLs: Time-limited, scope-limited S3 access
  GitHub Token Security: Minimal scope, secure storage
```

**Benefits of GitHub Organization Authentication:**
- ✅ **Unified Identity**: Same GitHub identity across all access patterns
- ✅ **Organization Control**: Companies can manage workspace access via GitHub teams
- ✅ **Granular Permissions**: Different organization roles can have different workspace permissions  
- ✅ **Audit Trail**: GitHub provides comprehensive audit logs for authentication events
- ✅ **Developer Familiar**: All developers already have GitHub accounts and understand org membership
- ✅ **Zero Additional Setup**: No separate user management system needed
- ✅ **Enterprise Ready**: Integrates with GitHub Enterprise and SAML/SSO

## S3 Backup Strategy

Based on research from GitHub Codespaces, Gitpod, DevPod and backup best practices, we need a comprehensive strategy for:
1. **Conversation History**: JSONL files from agents for web UI display
2. **File Changes**: Git-tracked files + selective untracked files for commit workflow
3. **Intelligent File Selection**: Exclude build artifacts while preserving important generated content

### File Selection Strategy

**Always Backup:**
```yaml
Critical Files:
  - All git-tracked files (git ls-files)
  - All staged changes (git diff --staged --name-only)  
  - All unstaged changes to tracked files (git diff --name-only)
  - Conversation history: ~/.claude/projects/*/*.jsonl (Claude Code conversation chains)
  - Agent session data: ~/.goose/sessions/*, ~/.codex/*, etc.
  - Workspace config: /workspace/.claude-vm/* 
  - Project config: package.json, requirements.txt, Dockerfile, .devcontainer/*, etc.

Git Metadata:
  - .git/config, .git/HEAD, .git/refs/* (essential git state)
  - .git/hooks/* (custom git hooks)
  - Note: Skip .git/objects/* (handled by remote git repo)
```

**Simplified Backup Strategy (deliverables folder + git changes):**
```yaml
Core Principle: "Simple folder convention + git change tracking"

Always Backup:
  - Git-tracked files with changes (git ls-files + git diff --name-only)
  - Agent deliverables folder: /workspace/deliverables/** (everything)
  - Conversation history: ~/.claude/projects/*/*.jsonl, ~/.goose/sessions/*, etc.
  - Conversation mappings: /workspace/.claude-vm/conversations.json (only this file)

Never Backup:
  - Build artifacts: node_modules/, dist/, build/, target/, __pycache__/
  - Package managers: .npm/, .pnpm/, .cargo/, .gradle/, venv/, env/
  - IDE/Editor: .vscode/, .idea/, *.swp, *.swo, .DS_Store, Thumbs.db
  - Logs/Cache: *.log, .cache/, temp/, tmp/, *.tmp
  - Large binaries: *.iso, *.dmg, *.exe >50MB (unless in /workspace/deliverables/)
  - System files: .Trash/, lost+found/, .fseventsd/

Agent Folder Convention:
  - Agents put final deliverables in /workspace/deliverables/
  - Everything else gets wiped on container restart
  - No complex categorization or metadata needed

Note: We do NOT back up claude-vm configuration files. The main claude-vm config 
lives on the user's local machine (~/.claude-vm/). Inside containers, we only 
preserve the conversation mapping file that tracks conversation names to tmux sessions.
```

**Example Agent Workflow:**
```yaml
User Request: "Analyze our sales data and create a comprehensive report"

Agent Workflow:
  1. Downloads data: wget sales.csv → /workspace/sales.csv (temporary)
  2. Processes data: python analyze.py → /workspace/processed.json (temporary)  
  3. Creates charts: matplotlib → /workspace/chart1.png, /workspace/chart2.png (temporary)
  4. Generates report: pandoc → /workspace/sales-report.pdf (temporary)
  5. Agent moves final output: mv sales-report.pdf /workspace/deliverables/
  6. Agent: "Your sales report is ready in /workspace/deliverables/ and will be preserved."

Backup Result:
  ✅ BACKED UP: /workspace/deliverables/sales-report.pdf (agent deliverable)
  ❌ NOT BACKED UP: All other files in /workspace/ (temporary processing artifacts)
  
Clean separation: deliverables vs temporary work files.
```

**Configuration File Support:**
```yaml
# /workspace/.claude-vm-backup (optional project-level config)
include:
  - "*.custom"           # Include custom file types
  - "data/*.processed"   # Include specific processed data
  - "outputs/*.pdf"      # Include generated reports

exclude:
  - "temp-*"             # Exclude temp files
  - "*.local"            # Exclude local config
  - "debug/"             # Exclude debug directories

# Global defaults (built into claude-vm)
max_file_size: "10MB"    # Skip individual files larger than this
max_backup_size: "1GB"   # Warn if total backup exceeds this
```

### Backup Storage Structure

```yaml
Provider: Any S3-compatible (AWS S3, Cloudflare R2, MinIO)

Bucket Structure:
  /{user-id}/
    /workspaces/
      /{workspace-id}/
        /conversations/
          /claude/
            /{conversation-name}/
              /{timestamp}-{uuid}.jsonl         # Original conversation
              /{timestamp}-{uuid}.jsonl         # Resumed conversation (linked by parentUuid)
              /conversation-chain.json          # Metadata linking UUIDs
          /goose/
            /{conversation-name}/
              /session-{timestamp}.log          # Goose session logs
              /session-metadata.json            # Session info
          /other-agents/
            /{conversation-name}/
              /agent-specific-files
              
        /file-snapshots/
          /{timestamp}/
            /workspace-snapshot.tar.gz          # Selected files per strategy above  
            /git-diff.patch                     # Uncommitted changes as patch
            /file-manifest.json                 # List of included files with sizes/hashes
            /backup-metadata.json               # Backup config used, timing, statistics
            
        /workspace-metadata.json                # Workspace config, agents, creation info
        
Backup Timing Strategy:
  Conversation Backups:
    - Immediate: After each agent response completes
    - Chain linking: Update conversation-chain.json when Claude resumes sessions
    
  File Backups:
    - Every 10 minutes: If files changed (git status porcelain check)
    - On demand: Via API endpoint POST /api/backup
    - Before shutdown: Automatic final backup before container stops
    - Size-based: If workspace grows >100MB since last backup
    
  Retention Policy:
    - Keep last 50 file snapshots per workspace (rolling window)
    - Keep conversation history indefinitely (small size)
    - Delete snapshots older than 30 days (configurable)
    - Compress old snapshots (>7 days) with higher compression
```

### Implementation Details

**Conversation History Backup:**
```bash
# Claude Code conversation handling
# Problem: Claude creates new .jsonl files when resuming (linked by parentUuid)
# Solution: Track conversation chains and backup all related files

backup_claude_conversation() {
  local conversation_name="$1"
  local workspace_id="$2"
  
  # Find current conversation UUID from conversations.json
  current_uuid=$(jq -r ".conversations[\"$conversation_name\"].conversationId" /workspace/.claude-vm/conversations.json)
  
  # Find all related JSONL files by following parentUuid chain
  local project_dir=$(python3 -c "import urllib.parse; print(urllib.parse.quote('/workspace', safe=''))")
  local claude_dir="$HOME/.claude/projects/$project_dir"
  
  # Upload all files in conversation chain
  for jsonl_file in "$claude_dir"/*.jsonl; do
    if grep -q "\"id\":\"$current_uuid\"" "$jsonl_file" 2>/dev/null || \
       grep -q "\"parentUuid\":\"$current_uuid\"" "$jsonl_file" 2>/dev/null; then
      aws s3 cp "$jsonl_file" "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/conversations/claude/$conversation_name/"
    fi
  done
  
  # Create conversation chain metadata
  create_conversation_chain_metadata "$conversation_name" "$workspace_id"
}
```

**File Selection Implementation:**
```bash
# Simplified file selection for workspace backup
create_file_snapshot() {
  local workspace_id="$1"
  local timestamp="$2"
  
  cd /workspace
  
  # Always include: git-tracked files with changes
  git ls-files > /tmp/backup-files.txt
  git diff --name-only >> /tmp/backup-files.txt        # Unstaged changes
  git diff --staged --name-only >> /tmp/backup-files.txt  # Staged changes
  
  # Add conversation history (all agents)
  find ~/.claude/projects -name "*.jsonl" -type f >> /tmp/backup-files.txt
  find ~/.goose/sessions -type f >> /tmp/backup-files.txt
  find ~/.codex -type f >> /tmp/backup-files.txt 2>/dev/null || true
  
  # Add conversation mappings (only this file from .claude-vm)
  echo ".claude-vm/conversations.json" >> /tmp/backup-files.txt 2>/dev/null || true
  
  # Add ALL agent deliverable files (simple folder convention)
  find deliverables -type f >> /tmp/backup-files.txt 2>/dev/null || true
  
  # Remove excluded patterns  
  grep -v -E "(node_modules|__pycache__|\.cache|build/|dist/|\.log$)" /tmp/backup-files.txt > /tmp/backup-files-filtered.txt
  
  # Create tar with file manifest
  tar -czf "/tmp/workspace-snapshot-$timestamp.tar.gz" --files-from=/tmp/backup-files-filtered.txt
  
  # Generate manifest with file sizes and hashes
  generate_file_manifest /tmp/backup-files-filtered.txt > "/tmp/file-manifest-$timestamp.json"
  
  # Upload to S3
  aws s3 cp "/tmp/workspace-snapshot-$timestamp.tar.gz" "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/file-snapshots/$timestamp/"
  aws s3 cp "/tmp/file-manifest-$timestamp.json" "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/file-snapshots/$timestamp/"
}
```

## LLM System Prompt Injection

To ensure agents understand their environment and organize outputs correctly, we inject contextual information into their system prompt:

**Context Injection Strategy:**
```yaml
Agent Identity Context:
  "You are {agent_name} running in a claude-vm workspace.
   Agent: {agent_name} (Claude Code, Goose, Codex, Gemini, etc.)
   Environment: DevContainer on {cloud_provider} ({region})
   Workspace: {workspace_id} 
   Project: {repo_url} (branch: {branch_name})"

Persistence Rules:
  "IMPORTANT - File Persistence:
  - Git-tracked files with changes are automatically preserved
  - Files in /workspace/deliverables/ are preserved across container restarts
  - ALL OTHER FILES are temporary and will be lost when container stops
  - Put deliverables/results the user should keep in /workspace/deliverables/
  - Work files, downloads, temp processing can go anywhere else in /workspace/"

Workspace Behavior Context:
  "This workspace runs in a remote container that can stop/restart.
   Your conversation history is preserved, but the filesystem resets except for:
   - Git repository state (tracked files + changes)
   - Files you place in /workspace/deliverables/
   Plan your file organization accordingly."
```

**Complete System Prompt Examples:**

```yaml
# Claude Code System Prompt Addition
CLAUDE_VM_CONTEXT = """
You are Claude Code running in a claude-vm workspace.
Environment: DevContainer on DigitalOcean (nyc1)
Workspace: bold-fire-1234
Project: github.com/user/myproject (branch: main)

IMPORTANT - File Persistence Rules:
- Git-tracked files with changes are automatically backed up
- Files in /workspace/deliverables/ are preserved when container stops/restarts  
- ALL OTHER FILES are temporary and lost on container restart
- Put final deliverables (reports, generated code, artifacts) in /workspace/deliverables/
- Temporary downloads, build files, processing data can go anywhere else

Your conversation history persists across container restarts, but organize your file outputs carefully.
"""

# Goose System Prompt Addition  
GOOSE_VM_CONTEXT = """
You are Goose AI running in a claude-vm remote workspace.
Environment: DevContainer on Fly.io (iad region) 
Workspace: quiet-lake-5678
Project: github.com/company/backend (branch: feature-auth)

File Organization:
- Modified git files are backed up automatically
- /workspace/deliverables/ contents survive container restarts
- Everything else is ephemeral - plan accordingly
- Save important results to /workspace/deliverables/ before finishing tasks

This is a persistent coding environment - your session continues across container lifecycle.
"""

# Implementation in claude-vm
inject_system_context() {
  local agent_name="$1"
  local workspace_id="$2" 
  local cloud_provider="$3"
  local region="$4"
  local repo_url="$5"
  local branch_name="$6"
  
  local context_prompt="You are $agent_name running in a claude-vm workspace.
Environment: DevContainer on $cloud_provider ($region)
Workspace: $workspace_id
Project: $repo_url (branch: $branch_name)

IMPORTANT - File Persistence Rules:
- Git-tracked files with changes are automatically backed up
- Files in /workspace/deliverables/ are preserved when container stops/restarts
- ALL OTHER FILES are temporary and lost on container restart
- Put final deliverables in /workspace/deliverables/
- Temporary files can go anywhere else in /workspace/

Your conversation history persists across container restarts."

  # Inject into agent's environment/config
  case "$agent_name" in
    "claude")
      export CLAUDE_ADDITIONAL_SYSTEM_PROMPT="$context_prompt"
      ;;
    "goose")
      echo "$context_prompt" > ~/.goose/system_context.txt
      ;;
    "codex")
      export CODEX_SYSTEM_CONTEXT="$context_prompt" 
      ;;
  esac
}
```

**Benefits of Simplified Backup Strategy:**

✅ **Ultra-simple convention** - Single `/workspace/deliverables/` folder, no complex categorization  
✅ **Perfect precision** - Only git changes + explicit agent deliverables backed up  
✅ **Zero configuration** - No manifest files, metadata, or preserve-file tools needed  
✅ **Universal approach** - Same simple pattern works for all agents (Claude, Goose, Codex, etc.)  
✅ **Storage efficiency** - No accidental backup of downloads, temp files, or build artifacts  
✅ **Clear user experience** - `/workspace/deliverables/` clearly indicates preserved outputs  
✅ **Agent context awareness** - System prompt injection ensures agents understand their environment  
✅ **Git workflow integration** - All code changes preserved, ready for commits via web UI  
✅ **Minimal backup scope** - Only conversation mappings backed up from container state, main config stays local  
✅ **Implementation simplicity** - Clean backup code, no complex heuristics or file detection

### Restore Process

**File Restoration:**
```bash
# Restore workspace from S3 backup
restore_workspace() {
  local workspace_id="$1"
  local snapshot_timestamp="$2"  # Optional: latest if not specified
  
  cd /workspace
  
  # 1. Fresh git clone (clean state)
  git clone "$repo_url" .
  git checkout "$branch_name"
  
  # 2. Download and extract file snapshot
  aws s3 cp "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/file-snapshots/$snapshot_timestamp/workspace-snapshot.tar.gz" /tmp/
  tar -xzf /tmp/workspace-snapshot.tar.gz -C /workspace
  
  # 3. Apply uncommitted changes
  if aws s3api head-object --bucket claude-vm-backups --key "$user_id/workspaces/$workspace_id/file-snapshots/$snapshot_timestamp/git-diff.patch" 2>/dev/null; then
    aws s3 cp "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/file-snapshots/$snapshot_timestamp/git-diff.patch" /tmp/
    git apply /tmp/git-diff.patch
  fi
  
  # 4. Restore conversation history
  restore_conversation_history "$workspace_id"
  
  # 5. Update conversation mappings
  update_conversation_mappings
}

restore_conversation_history() {
  local workspace_id="$1"
  
  # Restore Claude conversations
  aws s3 sync "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/conversations/claude/" "$HOME/.claude/projects/"
  
  # Restore other agent conversations
  aws s3 sync "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/conversations/goose/" "$HOME/.goose/sessions/"
  
  # Restore conversation mappings
  aws s3 cp "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/conversations.json" /workspace/.claude-vm/conversations.json
}
```

### Benefits of Agent-Directed S3 Backup Strategy

✅ **Complete conversation history** - All JSONL files and conversation chains preserved for web UI  
✅ **Agent-controlled outputs** - Agents decide what files are important, no guesswork needed  
✅ **Perfect precision** - Only meaningful deliverables backed up, temp files excluded automatically  
✅ **Storage efficiency** - No accidental backup of downloads, build artifacts, or intermediate files  
✅ **Universal approach** - Same pattern works for Claude, Goose, Codex, Gemini, and future agents  
✅ **Semantic organization** - Files organized by category (reports, data, artifacts) with metadata  
✅ **Web UI integration** - Backed up files displayed without container restart + git workflow support  
✅ **Simple implementation** - Clean backup logic replaces complex heuristics and file detection  
✅ **User clarity** - `/workspace/outputs/` clearly shows what will be preserved across restarts  
✅ **Fast restoration** - Git clone + agent outputs + conversation history restoration

Security:
```yaml
Authentication Layers:
  1. User → claude-vm.com:
     - OAuth (GitHub, Google, Email magic link)
     - Session cookies with CSRF protection
     - Rate limiting per user
  
  2. claude-vm.com → Workspace API:
     - Unique bearer token per workspace
     - Token rotation on each deployment
     - IP allowlist (only claude-vm.com)
  
  3. User → Workspace (direct SSH):
     - SSH key authentication
     - Keys injected at container creation
     - Optional: Teleport/Boundary for enterprise
  
  4. Workspace → S3:
     - Pre-signed URLs with expiration
     - Workspace-scoped IAM roles
     - Encryption at rest

Container Security:
  - Agent runs with restricted permissions
  - No access to provider metadata service
  - Secure credential storage (gh CLI pattern)
  - Network isolation between workspaces
  - Read-only mount for sensitive configs

Token Security:
  - GitHub tokens: Secure storage via gh CLI
  - API tokens: Short-lived (1 hour)
  - Workspace tokens: Unique per session
  - No tokens in environment variables (except PAT mode)
```

### Workspace Lifecycle

**Idle Timeout Strategy by Provider:**

```yaml
Fly.io: 
  - Native auto-suspend after no HTTP requests (2-3 min)
  - Auto-resume on first request
  - No container credentials needed

Google Cloud Run:
  - Native scale-to-zero when idle
  - Auto-resume on HTTP request
  - No container credentials needed

AWS ECS/Fargate:
  - Container exits after idle timeout
  - ECS handles stopped containers
  - Optional: CloudWatch + Lambda for cleanup

DigitalOcean:
  - Container exits after idle timeout  
  - External monitor cleans up stopped droplets
  - Or: manual cleanup via claude-vm CLI

Docker (Local):
  - No auto-shutdown (free resources)
  - User controlled via claude-vm commands
```

**Container Self-Management (No Provider Credentials Required):**
- Container monitors its own activity (SSH, API calls, tmux sessions)
- On idle timeout: backup to S3, then `exit 0` 
- Cloud platform handles stopped containers per its configuration
- No need for containers to have cloud provider API access

### Agent Integration Strategy

claude-vm supports **two distinct modes** for each agent:
1. **Interactive mode**: Drop user into agent's terminal UI via tmux
2. **Non-interactive mode**: Send stdin, get stdout response, exit

## Implementation Approaches by Agent

### Claude Code (JSONL File Watching)
**Key Insight**: Claude Code stores conversation history in structured JSONL files, allowing us to parse rich data instead of terminal output.

```bash
# Interactive mode: Drop into Claude Code terminal  
claude-vm chat "fix-auth-bug" 
# → tmux attach-session -t workspace-123:claude

# Non-interactive mode: Parse JSONL files for structured output
echo "debug error" | claude-vm chat "fix-auth-bug" --non-interactive
# → tmux send-keys: claude --print "debug error"
# → Watch: ~/.claude/projects/[encoded-path]/*.jsonl for new entries
# → Parse: JSON responses with tool calls, file changes, metadata
# → Return: Clean structured data to stdout

# Conversation Chaining: Claude creates new .jsonl files when resuming
# fix-auth-bug.jsonl → abc123.jsonl → def456.jsonl (linked by parentUuid)
```

### Other Agents (Terminal Output Parsing)
**Approach**: Parse terminal output since these agents lack structured conversation storage.

```bash
# Interactive mode: Drop into agent terminal
claude-vm chat "code-session" --agent goose
# → tmux attach-session -t workspace-123:goose

# Non-interactive mode: Send message via existing tmux session
echo "analyze code" | claude-vm chat "code-session" --non-interactive
# → tmux send-keys -t workspace-123:goose "analyze code" C-m
# → tmux capture-pane: Monitor terminal output for completion
# → Parse: Detect agent response completion
# → Return: Cleaned terminal output to stdout
```

## Key Implementation Details

**Tmux Wrapper Implementation Strategy:**

The key challenge is wrapping interactive CLI tools like Claude Code with tmux to enable remote access while preserving full functionality.

## Session Architecture

**Tmux Session Architecture (Corrected):**

**One tmux session per conversation** - this enables multiple concurrent conversations with the same agent:

```bash
# Inside workspace container, tmux sessions are named by conversation:
tmux list-sessions
# fix-auth-bug: 1 windows (claude running) 
# refactor-db: 1 windows (claude running)
# analyze-performance: 1 windows (goose running)
# add-tests: 1 windows (claude running)

# Multiple Claude conversations can run simultaneously
# Each gets its own dedicated tmux session
```

**Why One Session Per Conversation:**
```bash
# Benefits:
# ✅ Multiple concurrent conversations per agent (2+ Claude sessions)
# ✅ Direct mapping: conversation name = tmux session name  
# ✅ Simple attachment: tmux attach-session -t "fix-auth-bug"
# ✅ Clean programmatic control: tmux send-keys -t "fix-auth-bug" 
# ✅ No workspace prefix needed (tmux is container-scoped)
# ✅ Each conversation completely isolated

# Example: 3 concurrent Claude conversations
Session: "fix-auth-bug"     (claude --resume abc123)
Session: "refactor-db"      (claude --resume def456) 
Session: "add-tests"        (claude --resume ghi789)
```

**Session Name Simplification:**
```bash
# WRONG (previous approach): workspace-123:claude  
# - Can't handle multiple Claude conversations
# - Unnecessary workspace prefix inside container

# CORRECT (new approach): "fix-auth-bug"
# - Conversation name directly becomes tmux session name
# - Clean, intuitive, supports concurrent conversations
# - No prefixes needed (tmux sessions are container-scoped)
```

**For Advanced Users Who Want Multi-Window Sessions:**
SSH users can still create their own multi-window setups:

```bash
# Create personal multi-agent session
ssh user@workspace-123.provider.com
tmux new-session -s my-work
tmux new-window -n claude
tmux new-window -n goose  
tmux new-window -n shell

# But these won't integrate with claude-vm chat commands
# (trade-off: flexibility vs. integration)
```

## Workspace Initialization Process

**Container Startup Sequence:**
```bash
# 1. Container starts, workspace-manager service begins
# 2. Set up agent environment variables and credentials
# 3. tmux sessions created on-demand when conversations start
# 4. No pre-created sessions needed - cleaner startup

# Environment setup (done once at container start):
export CLAUDE_DANGEROUS_SKIP_PERMISSIONS=true
export GOOSE_API_KEY=$ANTHROPIC_API_KEY
cd /workspace

# Sessions created dynamically when users start conversations
```

## Conversation Management

**Starting New Conversations:**
```bash
# User runs: claude-vm chat --new "fix-auth-bug" --agent claude
# 1. Create new tmux session with conversation name
# 2. Start agent in that session

tmux new-session -d -s "fix-auth-bug" -c /workspace
tmux send-keys -t "fix-auth-bug" 'claude --project /workspace' C-m

# Session "fix-auth-bug" now running Claude Code
# Ready for interactive or non-interactive use
```

**Continuing Existing Conversations:**
```bash
# User runs: claude-vm chat "fix-auth-bug"
# 1. Check if tmux session "fix-auth-bug" exists
# 2. If exists: attach or send message
# 3. If not exists: resume from conversation mapping

# Session exists - direct attachment
tmux attach-session -t "fix-auth-bug"

# Session doesn't exist - recreate and resume
tmux new-session -d -s "fix-auth-bug" -c /workspace
tmux send-keys -t "fix-auth-bug" 'claude --resume abc123-def456-ghi789' C-m
tmux attach-session -t "fix-auth-bug"
```

## Access Method Implementations

### claude-vm CLI Access

**Interactive Mode:**
```bash
# claude-vm chat "fix-auth-bug"
# Implementation:
1. Check if tmux session "fix-auth-bug" exists
2. If exists: attach directly
3. If not: recreate from conversation mapping and attach
   tmux attach-session -t "fix-auth-bug"
```

**Non-Interactive Mode:**
```bash
# echo "what's the status" | claude-vm chat "fix-auth-bug" --non-interactive
# Implementation:
1. Check if tmux session "fix-auth-bug" exists
2. If not: recreate session and resume conversation
3. Send message via tmux:
   tmux send-keys -t "fix-auth-bug" "what's the status" C-m
4. Monitor tmux output until completion signal detected
5. Capture and clean output, return to stdout
6. Exit (leave session running in background)
```

### SSH User Access

**Recommended Approach: Always Use Tmux Sessions**

When SSH users access workspaces, they should **always use tmux sessions** for agent interactions:
```bash
# SSH into workspace container
ssh user@workspace-123.provider.com

# List claude-vm managed conversations
tmux list-sessions
# Output:
# fix-auth-bug: 1 windows (created Thu Aug 11 04:13:07 2025) [80x24] (claude)
# refactor-db: 1 windows (created Thu Aug 11 04:15:22 2025) [80x24] (claude)
# analyze-performance: 1 windows (created Thu Aug 11 04:16:45 2025) [80x24] (goose)

# Attach to any conversation by name
tmux attach-session -t "fix-auth-bug"         # Join Claude conversation
tmux attach-session -t "analyze-performance"  # Join Goose conversation

# Why this is the only recommended approach:
# ✅ Full integration with claude-vm chat commands
# ✅ Web interface can see the same conversations  
# ✅ Multiple users can collaborate on same conversation
# ✅ Conversation names work consistently ("fix-auth-bug")
# ✅ State persists across SSH disconnects
# ✅ Same experience as claude-vm CLI users
# ✅ No broken workflows or orphaned sessions
```

**Direct Agent Usage: Not Recommended**

Running agents directly (`claude --continue`, `goose run`) creates **orphaned sessions** that break claude-vm integration:

```bash
# ❌ DON'T DO THIS - creates orphaned sessions
ssh user@workspace-123.provider.com
claude --continue  # This session won't be accessible via claude-vm chat
goose run          # This session won't appear in web interface

# Problems this creates:
# ❌ claude-vm chat commands cannot access these sessions
# ❌ Web interface cannot see these conversations  
# ❌ No conversation name mapping ("fix-auth-bug" won't work)
# ❌ No collaborative access with other users
# ❌ Confusing for users who expect consistency
# ❌ State fragmentation across different access methods
```

**Guiding Users to the Correct Approach:**
We should provide clear guidance and helper commands to make tmux attachment easy:

```bash
# Helper script: /workspace/bin/claude-vm-attach
#!/bin/bash
echo "Available claude-vm managed conversations:"
cat ~/.claude-vm/conversations.json | jq -r 'keys[]' 
echo ""
echo "Tmux sessions:"
tmux list-sessions
echo ""
echo "To join a conversation: tmux attach-session -t \"<conversation-name>\""
echo "To start claude-vm CLI: claude-vm chat --list"

# SSH login message in ~/.bashrc
cat >> ~/.bashrc << 'EOF'
echo "🚀 Claude VM Workspace"
echo "Active conversations: $(cat ~/.claude-vm/conversations.json 2>/dev/null | jq -r 'keys | length // 0')"
echo ""
echo "📋 To join a conversation:"
echo "  tmux list-sessions                    # List all conversations"
echo "  tmux attach-session -t \"<name>\"      # Join specific conversation"
echo "  claude-vm chat --list                # Manage via CLI"
echo ""
echo "⚠️  Always use tmux sessions - don't run agents directly (claude, goose, etc.)"
echo ""
EOF
```

**SSH Access Guide:**

```markdown
## SSH Workspace Access - Always Use Tmux

When you SSH into a claude-vm workspace:

### ✅ DO: Use tmux sessions for all agent interactions
- `tmux list-sessions` - See all active conversations
- `tmux attach-session -t "conversation-name"` - Join any conversation
- Full integration with claude-vm CLI and web interface
- Multiple people can collaborate on the same conversation
- Consistent experience across all access methods

### ❌ DON'T: Run agents directly 
- Never run `claude --continue`, `goose run`, etc. directly
- Creates orphaned sessions that break claude-vm integration
- Web interface won't see these conversations
- Other users can't collaborate on these sessions
- Causes state fragmentation and user confusion
```

**Orphaned Session Detection and Prevention:**
We should actively detect and warn about orphaned agent sessions:

```bash
# In workspace-manager service, monitor for orphaned agent processes
ps aux | grep -E "(claude|goose|codex)" | grep -v tmux
# If found, show prominent warning and guidance

# SSH login check for orphaned sessions
cat >> ~/.bashrc << 'EOF'
if pgrep -f "claude|goose|codex" >/dev/null && ! pgrep -f tmux >/dev/null; then
  echo "⚠️  WARNING: Detected orphaned agent sessions!"
  echo "    These sessions won't work with claude-vm commands."
  echo "    Kill them with: pkill -f \"claude|goose|codex\""
  echo "    Then use: claude-vm chat --list"
  echo ""
fi
EOF

# Show warnings in claude-vm CLI
claude-vm chat --list
# ⚠️  Warning: Found 1 orphaned Claude session (PID 1234)
#     Kill it with: kill 1234
#     Then use proper tmux sessions: claude-vm chat "conversation-name"
```

## Agent-Specific Implementation Details

### Claude Code Wrapper
```bash
# New conversation session
tmux new-session -d -s "fix-auth-bug" -c /workspace
tmux send-keys -t "fix-auth-bug" 'export CLAUDE_DANGEROUS_SKIP_PERMISSIONS=true' C-m
tmux send-keys -t "fix-auth-bug" 'claude --project /workspace' C-m

# Resume existing conversation session
tmux new-session -d -s "fix-auth-bug" -c /workspace  
tmux send-keys -t "fix-auth-bug" 'claude --resume abc123-def456-ghi789' C-m

# Non-interactive message
tmux send-keys -t "fix-auth-bug" 'hello claude' C-m
# Wait for response completion (detect via JSONL file updates + shell prompt return)
```

### Goose Wrapper
```bash
# New conversation session
tmux new-session -d -s "analyze-performance" -c /workspace
tmux send-keys -t "analyze-performance" 'export GOOSE_API_KEY=$ANTHROPIC_API_KEY' C-m
tmux send-keys -t "analyze-performance" 'goose run' C-m

# Send message within conversation
tmux send-keys -t "analyze-performance" 'analyze this code' C-m
# Wait for response completion (detect via terminal output patterns)
```

## State Management Architecture

**Two-Level Tracking System:**

### 1. Workspace Registry (Local to User's Machine)
```json
// ~/.claude-vm/workspaces.json (on user's local machine)
{
  "workspaces": {
    "bold-fire-1234": {
      "name": "bold-fire-1234",
      "status": "running|stopped|error",
      "provider": "fly.io",
      "region": "iad",
      "url": "https://bold-fire-1234.fly.dev",
      "sshHost": "bold-fire-1234.fly.dev",
      "agents": ["claude", "goose"],
      "created": "2025-08-11T04:13:07Z",
      "lastActivity": "2025-08-11T08:22:11Z",
      "project": {
        "repo": "github.com/user/project",
        "branch": "main",
        "path": "/workspace"
      }
    },
    "quiet-lake-5678": {
      "name": "quiet-lake-5678", 
      "status": "stopped",
      "provider": "digitalocean",
      "region": "nyc1",
      "agents": ["claude"],
      "created": "2025-08-10T14:22:11Z",
      "lastActivity": "2025-08-10T16:33:55Z"
    }
  }
}
```

### 2. Conversation Tracking (Per Workspace Container)
```json
// /workspace/.claude-vm/conversations.json (inside each workspace container)
// This is the ONLY container state file we back up
{
  "metadata": {
    "version": "1.0",
    "lastUpdated": "2025-08-11T08:22:11Z"
  },
  "conversations": {
    "fix-auth-bug": {
      "agent": "claude",
      "tmuxSession": "fix-auth-bug",
      "status": "active",
      "created": "2025-08-11T04:13:07Z",
      "lastActivity": "2025-08-11T06:45:22Z",
      "messageCount": 45,
      "agentSpecific": {
        "conversationChain": ["abc123-original", "def456-resumed", "ghi789-current"],
        "activeUuid": "ghi789-current"
      }
    },
    "refactor-db": {
      "agent": "goose",
      "tmuxSession": "refactor-db", 
      "status": "idle",
      "created": "2025-08-10T14:22:11Z",
      "lastActivity": "2025-08-11T08:22:11Z",
      "messageCount": 12,
      "agentSpecific": {
        "sessionId": "goose-session-xyz789"
      }
    }
  }
}
```

**Container State Philosophy:**
- Keep minimal state in containers (only conversation mappings)
- Workspace metadata belongs in local CLI or hosted service
- Agent credentials managed by agents themselves
- This single JSON file is easy to backup, manipulate with `jq`, and extend

**How the Two-Level System Works:**

```bash
# User runs: claude-vm chat "fix-auth-bug"
# 1. Local CLI reads ~/.claude-vm/workspaces.json to find current workspace
# 2. Connects to workspace (SSH/API) and reads /workspace/.claude-vm/conversations.json  
# 3. Finds conversation "fix-auth-bug" → agent "claude", conversationId "abc123..."
# 4. Attaches to tmux session "fix-auth-bug" or recreates it if needed

# User runs: claude-vm workspace list
# 1. Local CLI reads ~/.claude-vm/workspaces.json
# 2. Shows all workspaces with their status, provider, agents

# User runs: claude-vm chat --list  
# 1. Local CLI connects to current/selected workspace
# 2. Reads /workspace/.claude-vm/conversations.json from that workspace
# 3. Shows conversations specific to that workspace
```

**Session Recovery After Container Restart:**
```bash
# Container restart procedure (happens inside workspace):
# 1. Read /workspace/.claude-vm/conversations.json 
# 2. For conversations marked as "active", attempt to resume:
#    - Claude: recreate tmux session + claude --resume <conversation-id>
#    - Goose: recreate tmux session + restore from goose session files
# 3. Update conversation status based on recovery success

# Local workspace status updated via:
# 1. Periodic health checks from claude-vm CLI
# 2. Provider status APIs (fly.io, digitalocean, etc.)
# 3. Manual refresh: claude-vm workspace list --refresh
```

## Benefits of This Architecture

✅ **Native Agent Experience**: Each agent works exactly as designed, no limitations
✅ **Multiple Access Methods**: CLI, SSH, web interface all work with same sessions  
✅ **Persistent Conversations**: Sessions survive disconnects, container restarts (with recovery)
✅ **Collaborative Access**: Multiple users can attach to same tmux session
✅ **Multiple Concurrent Conversations**: 2+ Claude conversations can run simultaneously
✅ **Clean Session Mapping**: Conversation name = tmux session name (intuitive)
✅ **Agent Agnostic**: Same pattern works for Claude, Goose, Codex, Gemini, etc.

**Conversation Continuation:**
- **Claude Code**: Use `claude --continue` or `claude --resume [id]` + JSONL tracking
  - **Critical**: Each resume creates new `.jsonl` file linked by `parentUuid` 
  - **Implementation**: Watch entire project directory, parse conversation chains
  - **Benefit**: Rich structured data (tool calls, file changes, metadata)
- **Other agents**: Resume via tmux session persistence (conversation state in terminal history)

### Human-Friendly Conversation Names

Similar to git branches, users can assign memorable names to conversations instead of working with UUIDs:

**Conversation Naming System:**
```bash
# Name mapping (stored in workspace)
~/.claude-vm/conversations.json:
{
  "fix-auth-bug": {
    "agent": "claude",
    "uuidChain": ["abc123-original", "def456-resumed", "ghi789-current"],
    "activeUuid": "ghi789-current",
    "created": "2025-08-11T04:13:07Z",
    "lastMessage": "2025-08-11T06:45:22Z",
    "messageCount": 45
  },
  "refactor-database": {
    "agent": "goose", 
    "uuidChain": ["xyz789-original"],
    "activeUuid": "xyz789-original",
    "created": "2025-08-10T14:22:11Z",
    "lastMessage": "2025-08-10T16:33:55Z", 
    "messageCount": 12
  }
}
```

**Name Resolution Process:**
1. User runs `claude-vm chat "fix-auth-bug"`
2. System looks up name in `conversations.json` 
3. Finds `activeUuid: "ghi789-current"`
4. Attaches to tmux session for that workspace
5. Claude Code resumes using `claude --resume ghi789-current`
6. Claude Code creates new UUID `jkl012-new-session`
7. System updates mapping: `"activeUuid": "jkl012-new-session"`
8. Appends to `uuidChain` for conversation history

**Features:**
- **Name validation**: Enforce git branch-like naming rules (no spaces, special chars)
- **Conversation chaining**: Track UUID chains automatically as Claude resumes sessions
- **History preservation**: Access full conversation across all UUID files in chain
- **Renaming**: `claude-vm chat --rename "old-name" "new-name"`
- **Listing**: `claude-vm chat --list` shows names, agents, message counts, timestamps

**Benefits:**
- **Human-friendly**: "fix-auth-bug" vs "d7f6c714-b5e7-4d40-822b-778b4aa7c646"
- **Context switching**: Easy to switch between different development tasks
- **Organization**: Group related conversations by feature/bug/topic
- **Team collaboration**: Share conversation names instead of UUIDs

### Stdin/Stdout Behavior

claude-vm chat supports **two distinct modes** with different stdin/stdout behavior:

## Interactive Mode (Default)
**Command:** `claude-vm chat "conversation-name"`

**Behavior:**
- Reads stdin (if provided) as **initial prompt**
- Drops user into **interactive terminal** (tmux attach)
- User can continue conversation interactively
- Terminal UI persists until user exits

**Examples:**
```bash
# Start interactive terminal with no initial prompt
claude-vm chat "fix-auth-bug"

# Start interactive terminal with stdin as initial prompt
echo "debug this error" | claude-vm chat "fix-auth-bug"
cat error.log | claude-vm chat "debug-session"
```

## Non-Interactive Mode (Piping)
**Command:** `claude-vm chat "conversation-name" --non-interactive`

**Behavior:**  
- Reads stdin (required) as **message to send**
- Sends message to agent via tmux
- Waits for agent response
- Outputs **final response to stdout**
- **Exits immediately** (for piping)

**Examples:**
```bash
# Send message, get response, exit
echo "what's the status?" | claude-vm chat "project" --non-interactive

# Pipe through multiple tools
cat error.log | claude-vm chat "debug" --non-interactive | grep "solution"

# Chain with other commands
echo "analyze code" | claude-vm chat "review" --non-interactive | tee analysis.txt
```

## Key Differences

| Mode | Stdin Usage | Terminal | Output | Exit |
|------|-------------|----------|--------|------|
| **Interactive** | Initial prompt (optional) | Drops into tmux terminal | Interactive UI | User exits manually |
| **Non-Interactive** | Message to send (required) | Background only | Response to stdout | Exits immediately |

**Both modes maintain conversation history** - the difference is whether you get an interactive terminal or a pipeable stdout response.

We also need to capture this input/output and provide it via the web interface through the Container API Server's SSE endpoints.

### Web UI Architecture - Two Phase Approach

We'll implement the web interface in **two phases**, each with distinct advantages:

## Phase 1: Terminal Streaming (xterm.js + pty/tmux)

**Implementation Strategy:**
```javascript
// Direct tmux session streaming to browser
const terminal = new Terminal();
const socket = new WebSocket('/api/terminal/fix-auth-bug');
terminal.open(document.getElementById('terminal'));

// Alternative approaches:
// Option A: tmux → pty → WebSocket → xterm.js
// Option B: tmux → WebSocket → xterm.js (direct streaming, no pty needed)
```

**Phase 1 Benefits:**
- ✅ **Rapid MVP**: Get web UI working in days, not weeks
- ✅ **100% Feature Parity**: Every Claude Code feature works identically  
- ✅ **Zero Translation Layer**: No parsing, no data loss, no bugs from interpretation
- ✅ **Native Experience**: Colors, cursor positioning, keyboard shortcuts, Esc+Esc branching
- ✅ **Multi-Agent Support**: Works with all agents (Claude, Goose, Codex, etc.)
- ✅ **Debugging**: Can see exactly what CLI tools are doing

**Phase 1 Limitations:**
- ❌ **Desktop Only**: Not mobile-friendly (tiny terminal text, no touch)
- ❌ **No Conversation Management**: Can't easily switch between named conversations
- ❌ **Limited History**: Can't search/filter past conversations
- ❌ **Raw Terminal**: No syntax highlighting enhancements, file diff views

## Phase 2: Custom Parsed Interface (JSONL + Mobile UI)

**Implementation Strategy:**
```javascript
// Parse JSONL files and present custom interface
const conversationWatcher = watchDirectory('~/.claude/projects/-home-user-project/');
conversationWatcher.on('new-message', (message) => {
  if (message.type === 'assistant' && message.message.content) {
    renderMessage(message, conversationName);
  }
});

// Mobile-optimized conversation UI
fetch('/api/conversations/fix-auth-bug')
  .then(response => response.json())
  .then(conversation => renderMobileConversation(conversation));
```

**Phase 2 Benefits:**
- ✅ **Mobile First**: Touch-friendly interface, readable on phones
- ✅ **Rich Conversation Management**: Named conversations, search, filtering, history
- ✅ **Enhanced Presentation**: Syntax highlighting, file diffs, tool call summaries
- ✅ **Cross-Device Sync**: Start on mobile, continue on desktop terminal
- ✅ **Offline Capable**: PWA with cached conversation history
- ✅ **Team Features**: Share conversation links, comment on specific messages

**Phase 2 Limitations:**
- ❌ **Development Complexity**: Weeks/months to implement full feature parity
- ❌ **Parsing Bugs**: Risk of misinterpreting terminal output during format changes
- ❌ **Feature Lag**: New agent features may not work until we update parsers

## Hybrid Implementation Plan

**Container API Server:** 
The unified API server definition above (line 685) supports both Phase 1 and Phase 2 approaches with minimal endpoints focused on our actual needs.

**Multi-Agent Support in Phase 2:**

Each agent requires custom parsing, but this enables a **unified experience**:

```yaml
Agent-Specific Parsers:
  Claude Code: 
    - Parse ~/.claude/projects/[encoded-path]/*.jsonl files
    - Extract tool calls, file changes, timestamps from rich JSON data
    
  Goose AI:
    - Monitor goose log files and terminal output  
    - Parse task completion status, file modifications
    
  OpenAI Codex:
    - Track Codex approval workflows and file edits
    - Parse interactive approval responses
    
  Gemini CLI:
    - Monitor terminal output and conversation state
    - Extract responses and command results
    
  Qwen-Coder:
    - Parse conversation flows and code generation
    - Track file modifications and suggestions

Unified Conversation Schema:
{
  "conversation": "fix-auth-bug",
  "agent": "claude",
  "messages": [
    {
      "type": "user|assistant", 
      "content": "...",
      "timestamp": "...",
      "toolCalls": [...],
      "fileChanges": [...]
    }
  ]
}
```

**Benefits of Custom Parsing per Agent:**
- ✅ **Consistent UX**: Same mobile interface regardless of which agent you're using
- ✅ **Rich Data**: Extract structured information each agent provides differently  
- ✅ **Cross-Agent Features**: Search conversations across all agents, unified history
- ✅ **Enhanced Presentation**: Syntax highlighting, file diffs work for all agents
- ✅ **Agent Switching**: Start with Claude, continue with Goose, consistent experience

**Technology Stack:**
- **Phase 1**: xterm.js + WebSocket + tmux (simple proxy)
- **Phase 2**: React/Vue + agent-specific parsers + unified schema + SSE + PWA
- **Backend**: Node.js/Go API server supporting both approaches simultaneously
- **Storage**: Named conversation mapping + agent-specific conversation tracking

---

### Task Specification

(open issue)

### Code Review

(open issue)

---

### Future Work

We may want to add a `claude-vm backup (export?)` and `claude-vm restore (import?)` command which uses tarball files, so that users can manually save and restore devcontainer state.
