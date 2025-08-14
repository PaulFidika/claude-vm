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
- Agent Installation in Devcontainer: installing coding agents (Claude Code, Goose, etc.) inside devcontainers with full development capabilities. Multiple agents can run within the same devcontainer environment.
- Agent Lifecycle: providing the user with input / output to the remote agent, keeping track of file changes, persisting changes.
- Provider Orchestration: launching VMs (droplets/machines/instances) + volumes on local or remote providers, then creating devcontainers with agents inside.
- Workspace Lifecycle: stopping VMs when the task is complete so they don't consume resources.
- UI: mobile UI + remote VSCode + terminal (devcontainer shell) for supervising claude and viewing the workspace.

Provider scope:
- Infrastructure lifecycle (VM, network)
- Volume management

---

### OpenAI Codex notes:
- Turns off the execution environment immediately after the task is completed, meaning the environment needs to be started up again for every user-interaction.
- Takes about 60 seconds to startup the environment again; runs a bunch of installs, meaning the dev environment (like node_modules) is not being cached; only file changes are.
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

The main point of these is to make it easier to run claude-code with the --dangerously-skip-permissions flag, so you do not need to manually approve stuff. Running claude code in a devcontainer provides strong security isolation: Claude cannot destroy your host system, and critically, Claude cannot access sensitive credentials (GitHub keys, S3 credentials, cloud provider tokens) which are stored only at the VM level. Even if Claude is compromised via prompt injection, the blast radius is limited to the development sandbox.

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

- daytonaio/daytona: 21k stars; written in Typescript + Go (for some reason). An over-engineered way of running arbitrary code in a container safely. Basically you have CLI -> API-server (Typescript) -> running (Go) -> container that executes code. They raised $7 million. They mostly want you to use their hosted-service since it's complex to self-host. Charges $0.46 for a 4 CPU, 16 GB instance per hour, which is a 2x markup compared to Digital Ocean / fly.io ($0.22 and $0.29 respectively); bare-metal you can rent this for $0.10 per hour.

### OpenAI Agent Mode

- Launches containers using a setting like `{"type":"code_interpreter","container":{"type":"auto","file_ids":[]}}`. This indicates that it can start up a container, and mount a selected set of files. It specifies this using a manifest (list of files).
- Agent mode takes about ~5 seconds to boot up its container
- Containers are useful for tasks that require state. An LLM's text-browser, or calls to python code-interpreter are stateless (i.e., here is a query, give me text of results, here is a URL, give me text on page, here is python code, run it and give me result), but stateful workflows require a container. Exmaples (other than coding):
  - Login state
  - Browser-navigation state on SPAs (visual browser)
  - File aggregation (i.e., download 3 PDFs and put them together into a single PDF)
  - Multi-file steps (i.e., create a chart, then place that chart in a PowerPoint presentation)

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
├── logout                                              # Logout from SaaS
│
├── workspace                                           # Workspace management
│   ├── up [workspace-id|repo-url|path/to/local-repo]   # Create and start new workspace
│   │   ├── --cloud <name>                              #   Cloud platform (docker|digitalocean|fly|aws) - uses VM services: droplets/machines/ec2
│   │   ├── --agent <name>                              #   Agents to include (claude|codex|qwen|goose|gemini)
│   │   ├── --credentials <list>                        #   Credentials to include (comma-separated)
│   │   ├── --devcontainer <path>                       #   DevContainer spec (takes priority over --image)
│   │   ├── --image <name>                              #   Docker image override (fallback)
│   │   ├── --region <region>                           #   Cloud region (optional)
│   │   ├── --size <size>                               #   Machine size (optional)
│   │   ├── --branch <name>                             #   Git branch to checkout
│   │   └── --non-interactive                           #   Disable prompts, fail if config missing
│   │
│   ├── list                                  # List all workspaces 
│   │   ├── -l, --long                        #   Show detailed workspace info
│   │   ├── --status <filter>                 #   Filter by status (running|stopped|error)
│   │   └── --json                            #   Output in JSON format
│   │
│   ├── token <workspace-id>                  # Generate JWT for workspace access
│   │
│   ├── down <workspace-id>                   # Stop workspace (can resume later)  
│   │   └── -f, --force                       #   Force stop without graceful shutdown
│   │
│   └── delete <workspace-id>                 # Delete provider resource, keep S3 backup
│       ├── -y, --yes                         #   Skip confirmation prompt
│       └── --purge                           #   Delete provider resource AND S3 backup
│
├── shell <workspace-id>                      # Open shell in workspace devcontainer
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
├── cloud                                     # SYSTEM: Cloud platform management (used by claude-vm for VM provisioning)
│   ├── set-config <cloud-name>              # Add or update cloud configuration
│   │   └── --option <key=value>             #   Set cloud options (api-key, region, etc.)
│   ├── list                                 # List all clouds with configuration
│   └── clear-config <cloud-name>            # Clear all options for cloud
│
├── provider                                  # SYSTEM: LLM API provider management (used by claude-vm for agent APIs)
│   ├── set-config <provider-name>           # Add or update provider configuration
│   │   └── --option <key=value>             #   Set provider options (api-key, oauth-token, etc.)
│   ├── list                                 # List all providers with configuration
│   └── clear-config <provider-name>         # Clear all options for provider
│
├── storage                                   # SYSTEM: S3-compatible backup storage (used by claude-vm for backups)
│   ├── set-config                           # Configure S3-compatible storage for backups
│   │   └── --option <key=value>             #   Set storage options (endpoint, bucket, access-key-id, etc.)
│   ├── list                                 # Show current storage configuration
│   ├── test                                 # Test storage connection and permissions
│   └── clear-config                         # Clear storage configuration
│
├── credential                               # DEVELOPMENT: Agent credential management (used by agents for development work)
│   ├── add <name>                           # Add development credential for agents to use  
│   │   ├── --ssh-key <path>                 #   SSH private key for git operations
│   │   ├── --profile <name>                 #   AWS/cloud profile name
│   │   ├── --username <user>                #   Username (for registries)
│   │   ├── --token <token>                  #   API token or PAT
│   │   ├── --api-key <key>                  #   API key (for services like Stripe)
│   │   └── --connection-string <url>        #   Database connection string
│   ├── list                                 # Show registered development credentials (keys masked)
│   ├── remove <name>                        # Remove credential from keychain
│   └── test <name>                          # Test credential connection
│
├── web                                     # Open web interface dashboard
│   │   --port, -p <PORT>                   # Port to run on (default: 3000)
│   │   --host <HOST>                       # Host to bind to (default: 127.0.0.1)
│   │   --ssh-key <PATH>                    # SSH private key path (default: auto-detect)
│   │   --config <PATH>                     # Config file path (default: ~/.claude-vm/config.yaml)
│   │   --no-browser                        # Don't auto-open browser
│   │   --timeout <DURATION>                # JWT expiration timeout (default: 1h)
│   │   --auth-password <PASSWORD>          # Simple password protection
│   │   --auth-github-org <ORG>             # Required GitHub organization
│   │   --auth-github-teams <TEAMS>         # GitHub teams (comma-separated)
│   │   --auth-config <PATH>                # Auth configuration file (YAML)

EXAMPLES:
  # Workspace management  
  claude-vm workspace up .                   # Auto-detect .devcontainer/ or generate; use default cloud and default agent
  claude-vm workspace up . --cloud digitalocean --agent claude  # DO with Claude agent only
  claude-vm workspace up . --cloud fly --agent codex,goose     # Fly with Codex + Goose agents  
  claude-vm workspace up . --agent null                       # Use default cloud, no agents in devcontainer
  claude-vm workspace up . --credentials github,aws,stripe,pnpm  # Specify workspace credentials
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
  
  # SYSTEM cloud platform configuration (used by claude-vm to provision VMs)
  claude-vm cloud set-config digitalocean --option api-key=dop_xxx --option region=nyc1
  claude-vm cloud set-config fly --option api-key=fly_xxx --option region=iad
  claude-vm cloud set-config aws --option access-key=AKIA... --option secret-key=xxx
  claude-vm cloud set-config docker  # Local Docker (no credentials needed)
  claude-vm cloud list                                     # Show all clouds with config
  claude-vm cloud clear-config digitalocean               # Clear DO configuration
  
  # SYSTEM LLM API provider configuration (used by claude-vm to power agents)
  claude-vm provider set-config anthropic --option api-key=sk-ant-xxx
  claude-vm provider set-config anthropic --option oauth-token=/path/to/token.json
  claude-vm provider set-config openai --option api-key=sk-xxx
  claude-vm provider set-config openai --option oauth-token=/path/to/token.json
  claude-vm provider set-config google --option api-key=AIza-xxx
  claude-vm provider list                                  # Show all providers with config
  claude-vm provider clear-config anthropic               # Clear Anthropic configuration
  
  # DEVELOPMENT credentials (used by agents for development work)
  claude-vm credential add github ~/.ssh/id_rsa          # SSH key (auto-detected) - for agent git operations
  claude-vm credential add aws my-dev-profile             # AWS profile (auto-detected) - for agent AWS CLI usage
  claude-vm credential add stripe sk_test_abc123          # API key → STRIPE_API_KEY - for agent API testing
  claude-vm credential add openai sk-abc123def456         # API key → OPENAI_API_KEY - for agent OpenAI usage
  claude-vm credential add npm ~/.npmrc                   # Config file (auto-detected) - for agent npm operations
  claude-vm credential add pnpm ~/.npmrc                  # Config file (auto-detected) - for agent pnpm operations
  claude-vm credential add docker myuser:dckr_pat_abc123  # Registry auth (auto-detected) - for agent docker pulls
  
  # Custom credential patterns for unlisted services
  claude-vm credential add my-api --env-var MY_API_KEY=abc123
  claude-vm credential add internal-db --proxy postgresql://host:5432/db --local-port 5432
  
  claude-vm credential list                               # Show registered development credentials (keys masked)  
  claude-vm credential test github                        # Test credential connection
  claude-vm credential remove old-credential             # Remove credential from keychain
  
  # S3-compatible backup storage configuration (for self-hosted users)
  claude-vm storage set-config --option endpoint=https://s3.amazonaws.com --option bucket=my-claude-vm-backups
  claude-vm storage set-config --option access-key-id=AKIAIOSFODNN7EXAMPLE --option secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYzEXAMPLEKEY
  claude-vm storage set-config --option endpoint=https://r2.cloudflarestorage.com --option bucket=my-backups  # Cloudflare R2
  claude-vm storage test                                   # Test storage connection and permissions
  claude-vm storage list                                   # Show current storage configuration (credentials masked)
  claude-vm storage clear-config                           # Clear storage configuration
  
  # Agent configuration (agents use their provider's credentials)
  claude-vm agent set-config claude --option model=opus --option auth_preference=oauth
  claude-vm agent set-config claude --option model=sonnet --option auth_preference=api-key
  claude-vm agent set-config codex --option model=gpt-4o --option auth_preference=oauth
  claude-vm agent set-config qwen --option model=qwen3-coder-plus --option temperature=0.8
  claude-vm agent set-config goose --option provider=openrouter --option model=anthropic/claude-3.5-sonnet
  claude-vm agent list                                           # Show all agents with config
  claude-vm agent clear-config claude                            # Clear Claude configuration
  
  # Workspace management with JSON output
  claude-vm workspace list                                       # List all workspaces
  claude-vm workspace list --json                               # Export workspace registry as JSON
  claude-vm workspace token bold-fire-1234                      # Generate JWT for workspace access
  
  # SaaS authentication
  claude-vm login                                                # Login to SaaS (claude-vm.com)
  claude-vm logout                                               # Logout from SaaS
  
  # SSH access
  claude-vm ssh quiet-lake-5678 --user developer
  claude-vm ssh 8a9b2c3d4e5f --port 2222
  
  # Web interface
  claude-vm web                                                  # Local web UI (localhost:3000)
  claude-vm web --host 0.0.0.0 --auth-password team-secret      # Self-hosted with password
  claude-vm web --host 0.0.0.0 --auth-github-org acme-corp     # Self-hosted with GitHub org
  claude-vm web --host 0.0.0.0 --auth-github-org acme-corp --auth-github-teams dev,ops  # With teams
  
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

## Configuration Management

claude-vm uses a simple split-file approach for clean separation of concerns:

**Storage Locations:**
- **User preferences**: `~/.claude-vm/config.yaml` (only configured settings, no defaults)
- **Workspace registry**: `~/.claude-vm/workspaces.yaml` (managed by CLI)
- **Secrets**: OS keychain (API keys, access tokens, OAuth credentials)

**Keychain Storage Pattern:**
```
Service: claude-vm
Account: <type>.<name>.<field>
Password: <secret_value>

Examples:
- claude-vm.cloud.digitalocean.api_key
- claude-vm.cloud.fly.api_key
- claude-vm.cloud.aws.access_key_id
- claude-vm.cloud.aws.secret_access_key
- claude-vm.cloud.google.gcp_credentials
- claude-vm.provider.anthropic.api_key
- claude-vm.provider.anthropic.oauth_token
- claude-vm.provider.openai.api_key
- claude-vm.provider.openai.oauth_token
- claude-vm.provider.google.api_key
- claude-vm.provider.alibaba.api_key
- claude-vm.provider.groq.api_key
- claude-vm.storage.access_key_id
- claude-vm.storage.secret_access_key
```

**Why OS Keychain:**
- ✅ **Security**: Encrypted at rest, OS-managed access control
- ✅ **Cross-platform**: Works on macOS, Linux (Secret Service), Windows (Credential Manager)
- ✅ **Integration**: Same pattern used by Docker, AWS CLI, Google Cloud SDK
- ✅ **No plaintext**: Secrets never stored in config files or environment variables

**Configuration Commands:**
All `set-config` commands automatically detect secrets vs non-secrets and store them appropriately:
- API keys, access tokens → OS keychain  
- Regions, sizes, model names → config.yaml

**Complete Configuration Details:**
See DESIGN_CONFIG.md for the complete file format documentation:
- config.yaml structure (user preferences only) 
- workspaces.yaml structure (workspace registry)
- keychain storage patterns for all secrets

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

**Default Behavior:**
- **Default cloud**: `docker` (local Docker, no credentials needed)
- **Default agent**: `claude` (most popular agent)

**Agent-Provider Relationships:**
- **Agents** are configured to use specific **providers**
- **Providers** have no concept of "default" - they're just configured endpoints
- Example: Claude agent configured to use Anthropic provider with API key

```bash
# Providers are just endpoints
claude-vm provider set-config anthropic --option api-key=sk-ant-xxx

# Some agents support multip providers, and can be configured to use them
claude-vm agent set-config goose --option provider=anthropic
```

---

## Cloud Infrastructure Architecture

**VM + Devcontainer Design (Codespaces Model):**
Claude VM launches **virtual machines** on cloud providers, then runs coding agents **inside devcontainers** for development work:

- **Digital Ocean**: Uses Droplets (not Digital Ocean Apps or container services)
- **Fly.io**: Uses Machines (VM instances, not container deployment)  
- **AWS**: Uses EC2 instances (not ECS/Fargate container services)
- **Docker**: Uses local Docker daemon

**Why agents in devcontainers?**
1. **Full development capabilities**: Agents can install packages, run services via Docker-in-Docker
2. **VS Code integration**: Seamless Remote-Containers experience for users
3. **Environment consistency**: Agents and users work in identical, reproducible environments
4. **Industry standard**: Follows GitHub Codespaces, Gitpod patterns
5. **Appropriate scope**: Development-focused, not system administration

**Architecture Flow:**
```
claude-vm CLI → Launch VM (droplet/machine/ec2) → Create devcontainer → Agents run inside devcontainer
```

**System Architecture:**
```
VM (managed infrastructure):
├── Docker daemon
├── SSH daemon (for VM management)
└── Devcontainer (development environment):
    ├── Claude Code + other agents (multiple agents per container)
    ├── SSH server (for direct container access)
    ├── VS Code Server
    ├── Project code (/workspaces/repo/)
    ├── Development tools
    ├── Docker-in-Docker support
    └── /workspaces/ (persistent workspace data)
```

**Key Points:**
- **Multiple agents per container**: Claude Code, Goose, Gemini can all run in same devcontainer
- **VM is background infrastructure**: Users never directly access the VM
- **Direct devcontainer access**: SSH server installed in devcontainer for shell access


### Volume Persistence

Every cloud platform allows us to mount volumes; we mount them in /workspace. We can stop a VM, and then start it again with the same volume preserving all data (such as edited code).

**Container Setup:**
- WORKDIR /workspace (persistent volume mount point)  
- Store git repos, package caches, and user configs in volume
- Create symlinks from home directory to workspace cache directories for optimization:

```bash
ln -s /workspace/.cache/npm ~/.npm
ln -s /workspace/.cache/pnpm ~/.pnpm
ln -s /workspace/.cargo ~/.cargo
ln -s /workspace/go ~/go
```

### Persistence with S3 Object Store

**Git Credential Security:**
GitHub credentials (SSH keys, personal access tokens) are stored securely on the VM and never exposed to containers:

```bash
# VM Level (Trusted):
~/.ssh/id_rsa                    # GitHub SSH key (never copied to container)
ssh-agent running with keys     # SSH agent with loaded keys
git credential helper           # Provides tokens from keychain

# Container Level (Sandboxed):
SSH_AUTH_SOCK forwarded         # Points to VM's SSH agent
git configured to use agent     # All git operations work normally
```

**Result**: Claude gets full git access (`git commit`, `git push`, `git pull`, branch operations) without ever seeing the actual SSH keys or tokens. Credentials remain on VM, authentication happens via agent forwarding.

## Credential Management System

**Two Distinct Credential Scopes:**

### Development Credentials vs System Configuration

**Development Credentials** (`claude-vm credential`):
- **Used BY**: AI agents (Claude, Goose, etc.) running inside devcontainers
- **Purpose**: Enable agents to access development services during coding work
- **Examples**: GitHub repos, Stripe test APIs, development databases, NPM registries
- **Security**: Stored on VM, securely forwarded to containers via helpers/proxies
- **Lifecycle**: Per-project, selected per workspace

**System Configuration** (`claude-vm cloud`, `claude-vm provider`, `claude-vm storage`):
- **Used BY**: claude-vm CLI itself for infrastructure operations
- **Purpose**: Enable claude-vm to provision VMs, run agents, manage backups
- **Examples**: DigitalOcean API for VMs, Anthropic API for Claude agent, S3 for backups
- **Security**: Stored locally, never exposed to containers
- **Lifecycle**: Set once globally, used by claude-vm infrastructure

---

## Development Credential Management

**Three-Step Development Credential Approach:**

### 1. Credential Registration (One-Time Setup)

**Smart Defaults for Common Services:**
```bash
# Reserved keywords with automatic credential type detection:
claude-vm credential add github ~/.ssh/id_rsa              # SSH key (covers Git, Go modules, Zig packages)
claude-vm credential add aws my-dev-profile                # AWS profile
claude-vm credential add stripe sk_test_abc123             # API key → STRIPE_API_KEY
claude-vm credential add openai sk-abc123def456            # API key → OPENAI_API_KEY  
claude-vm credential add npm ~/.npmrc                      # Config file (npm)
claude-vm credential add pnpm ~/.npmrc                     # Config file (pnpm) 
claude-vm credential add yarn ~/.yarnrc.yml                # Config file (yarn)
claude-vm credential add pip ~/.pip/pip.conf               # Config file (Python traditional)
claude-vm credential add uv ~/.config/uv/uv.toml           # Config file (Python modern)
claude-vm credential add docker myuser:dckr_pat_abc123     # Registry auth
claude-vm credential add anthropic claude_api_key_123     # API key → ANTHROPIC_API_KEY
claude-vm credential add supabase https://abc.supabase.co:service_key_xyz  # API key → SUPABASE_URL + SUPABASE_SERVICE_ROLE_KEY
```

**Manual Patterns for Custom/Unlisted Services:**
```bash
# Full control for edge cases or custom services:
claude-vm credential add my-api --env-var MY_API_KEY=abc123
claude-vm credential add internal-db --proxy postgresql://internal:5432/app --local-port 5432
claude-vm credential add github-token --env-var GITHUB_TOKEN=ghp_123  # Override default SSH
claude-vm credential add custom-service --http-header "Authorization: Bearer jwt123"
```

**Supported Credential Patterns:**
- `--ssh-key PATH`: SSH agent forwarding (Git, SSH, rsync)
- `--env-var KEY=VALUE`: Environment variable injection (REST APIs)
- `--credential-helper TOOL --profile NAME`: System credential helpers (AWS, GCP, Azure) 
- `--proxy CONNECTION_STRING --local-port PORT`: Database/service proxies
- `--config-file PATH`: Mount configuration files (.npmrc, .gitconfig)
- `--registry-auth USERNAME:TOKEN`: Container registry authentication
- `--http-header "HEADER: VALUE"`: HTTP proxy with auth header injection

**Reserved Service Keywords:**
```bash
github      → SSH key (covers Git, Go modules, Zig packages)
gitlab      → SSH key (covers GitLab repos, Go modules, Zig packages)
aws         → AWS credential helper with profile
gcp         → GCP credential helper  
azure       → Azure credential helper
stripe      → STRIPE_API_KEY environment variable
openai      → OPENAI_API_KEY environment variable
anthropic   → ANTHROPIC_API_KEY environment variable
npm         → .npmrc config file mounting (npm)
pnpm        → .npmrc config file mounting (pnpm)
yarn        → .yarnrc.yml config file mounting (yarn)
pip         → pip config file mounting (Python traditional)
uv          → uv config file mounting (Python modern)
cargo       → .cargo/config.toml config file mounting (Rust)
docker      → Docker registry authentication
supabase    → SUPABASE_URL + SUPABASE_SERVICE_ROLE_KEY
neon        → Database proxy (PostgreSQL)
planetscale → Database proxy (MySQL)
railway     → Database proxy
vercel      → VERCEL_TOKEN environment variable
netlify     → NETLIFY_AUTH_TOKEN environment variable
```

# List registered credentials:
claude-vm credential list
# github          SSH Key        Ready (covers Git + Go + Zig)
# aws             AWS Profile    Ready (my-dev-profile)
# stripe          API Key        Ready
# openai          API Key        Ready  
# npm             Config File    Ready
# pnpm            Config File    Ready
# yarn            Config File    Ready
# pip             Config File    Ready (Python traditional)
# uv              Config File    Ready (Python modern)
# docker          Registry       Ready
```

### 2. Workspace Credential Selection

**Preferred: devcontainer.json specification:**
```json
{
  "name": "My Fullstack App",
  "image": "node:18",
  "customizations": {
    "claudeVM": {
      "credentials": [
        "github",
        "aws", 
        "stripe",
        "openai",
        "pnpm",
        "docker"
      ]
    }
  }
}
```

**Alternative: CLI flags:**
```bash
claude-vm workspace up . --credentials github,aws,stripe,pnpm
```

### 3. Secure Credential Forwarding

**VM Level (Trusted):**
- Stores actual credentials in keychain/filesystem
- Runs SSH agents, credential helpers, database proxies
- Provides secure credential forwarding to containers

**Container Level (Sandboxed):**
- Gets access via forwarded agents/helpers only
- SSH_AUTH_SOCK, AWS credential helpers, environment variable helpers
- Never sees raw credentials in filesystem or environment

**Runtime Examples:**
```bash
# Claude gets full access via secure forwarding:
git push origin main              # SSH agent forwarding
aws s3 ls my-bucket              # AWS credential helper
docker push myregistry/image     # Docker registry auth helper
psql $DATABASE_URL               # Database connection proxy
npm publish                      # NPM token helper
curl -H "Auth: Bearer $STRIPE_API_KEY" api.stripe.com  # Environment variable helper
```

This ensures Claude can use all necessary services for development while maintaining complete credential isolation.

**Automated Backup Process:**

**Self-Hosted (No API Server):**
- VM has direct access to S3-compatible storage credentials (DigitalOcean Spaces, Cloudflare R2, AWS S3, etc.)
- VM monitors `/workspace/deliverables/` and uploads directly to configured storage
- Git operations work through credential forwarding (described above)

**API Server Model (Self-Hosted or SaaS):**
- Containers upload files to claude-vm API server via HTTP
- API server handles all S3-compatible storage operations using server-managed credentials  
- Prevents chaotic direct S3 access from every container
- Git operations still use credential forwarding for direct container access

**What Gets Backed Up:**
1. **Git Repository**: Automatic commits and pushes (VM-level or API server)
2. **Deliverables**: Files in `/workspace/deliverables/` → S3-compatible storage
3. **Agent State**: Conversation logs and workspace state

**Security Model:**
- **Self-hosted**: VM-level credential isolation acceptable (user controls infrastructure)
- **API Server**: Centralized credential management, containers never see storage credentials
- **Agents**: Never see external credentials in either model

**S3-Compatible Storage:**
We support any S3-compatible API including DigitalOcean Spaces, Cloudflare R2, AWS S3, Minio, etc. - not just AWS S3.

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

claude-vm uses **simplified token-based authentication** inspired by HashiCorp Vault, Docker CLI, and GitHub CLI patterns. This approach eliminates OAuth complexity from containers while supporting three distinct web UI hosting patterns:

1. **Localhost Web UI** (like `vault ui`)
2. **User's Own Web Server** (like self-hosted services)
3. **Hosted Service** (like claude-vm.com)

## Unified Token-Based Strategy

**Core Principles:**
- Containers are pure API servers with bearer token authentication only
- No OAuth complexity per container (learned from Vault approach)
- Token storage follows proven CLI patterns (Docker, GitHub CLI, Vault)
- Web UIs act as authenticated proxies to container APIs

## Container API Server (Simplified)

**Architecture Overview:**
```
WebUI (Browser) ←→ Container API Server ←→ tmux sessions ←→ Agents
  WebSocket/SSE        RESTful API           Terminal I/O      Claude/Goose/etc.
```

**Communication Flow:**
1. **WebUI** sends message via WebSocket to Container API
2. **Container API** forwards message to named tmux session via `tmux send-keys`
3. **tmux session** delivers input to running agent (Claude, Goose, etc.)
4. **Agent** processes input, writes response to terminal
5. **Container API** captures terminal output, streams back to WebUI via WebSocket
6. **WebUI** displays agent response in real-time

**Pure API Server Configuration:**
```yaml
Port: 8080
Authentication: Bearer tokens ONLY (no OAuth, no sessions)

Environment Variables:
  WORKSPACE_SECRET: "jwt-signing-key"           # For token validation
  WORKSPACE_ID: "bold-fire-1234"               # Workspace identifier

Authentication Logic:
  1. Check Authorization: Bearer <token> header
  2. Validate JWT token signature and expiration
  3. Extract workspace/user claims from token
  4. Allow/deny API access based on token validity
  
# No GitHub OAuth, no sessions, no cookies - pure stateless API
```

**API Endpoints:**
```yaml
# Core Web Interface Support
WS     /api/terminal/{conversation-name}        # Stream tmux session to browser
GET    /api/conversations/{name}/data           # Get parsed conversation data (JSONL → JSON)
POST   /api/conversations/{name}/messages       # Send message via tmux + parse response
SSE    /api/events                              # Real-time conversation updates

# Conversation Management
GET    /api/conversations                       # List conversations from conversations.yaml
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
```

## Four Web UI Access Patterns

Users can access their containers through four distinct patterns, each serving different use cases:

### Pattern 1: Localhost Dashboard (`claude-vm web`)
**Use Case:** Developer's primary workflow on their local machine

```bash
# Local usage (no auth)
claude-vm web

# ERROR: Public host without auth
claude-vm web --host 0.0.0.0
# → ERROR: Public host requires authentication. Use --auth-password or --auth-github-org

# Self-hosted with simple password
claude-vm web --host 0.0.0.0 --auth-password team-secret

# Self-hosted with GitHub org membership
claude-vm web --host 0.0.0.0 --auth-github-org mycompany

# Self-hosted with GitHub org + teams
claude-vm web --host 0.0.0.0 --auth-github-org mycompany --auth-github-teams dev-team,ops

# Self-hosted with explicit GitHub OAuth credentials
claude-vm web --host 0.0.0.0 --auth-github-org mycompany \
              --auth-github-client-id abc123 \
              --auth-github-client-secret def456

# Using auth config file  
claude-vm web --host 0.0.0.0 --auth-config ./auth.yaml
```

**Command Options:**
```
AUTHENTICATION (required for non-localhost hosts):
  --auth-password <PASSWORD>         Simple password protection
  --auth-github-org <ORG>            Required GitHub organization
  --auth-github-teams <TEAMS>        GitHub teams (comma-separated, optional)
  --auth-github-client-id <ID>       GitHub OAuth client ID (optional, uses defaults)
  --auth-github-client-secret <SECRET>  GitHub OAuth client secret (optional)
  --auth-config <PATH>               Auth configuration file (YAML)

CORE OPTIONS:
  --port, -p <PORT>                  Port to run on (default: 3000)
  --host <HOST>                      Host to bind to (default: 127.0.0.1)  
  --ssh-key <PATH>                   SSH private key path (default: auto-detect)
  --config <PATH>                    Config file path (default: ~/.claude-vm/config.yaml)
  --no-browser                       Don't auto-open browser
  --timeout <DURATION>               JWT expiration timeout (default: 1h)

EXAMPLES:
  # Local development (no auth required)
  claude-vm web

  # Simple password protection for team access
  claude-vm web --host 0.0.0.0 --auth-password team-secret-123
  
  # GitHub organization restriction
  claude-vm web --host 0.0.0.0 --auth-github-org acme-corp
  
  # GitHub org + specific teams
  claude-vm web --host 0.0.0.0 --auth-github-org acme-corp --auth-github-teams dev-team,ops
  
  # Custom port with auth
  claude-vm web --host 0.0.0.0 --port 8080 --auth-github-org mycompany
```

**Auth Configuration File (Standard Pattern):**

**Option 1: Environment Variables (Recommended)**
```yaml
# auth.yaml
auth:
  type: github
  github:
    org: "mycompany"
    teams: ["dev-team", "ops-team"]     # Optional
    client_id: "${GITHUB_CLIENT_ID}"    # From environment
    client_secret: "${GITHUB_CLIENT_SECRET}"  # From environment
    callback_url: "https://dev-tools.company.com/auth/github/callback"

# Alternative: simple password
# auth:
#   type: password
#   password: "${AUTH_PASSWORD}"
```

**Option 2: Direct Values (Less Secure)**
```yaml
# auth.yaml
auth:
  type: github
  github:
    org: "mycompany"
    teams: ["dev-team", "ops-team"]
    client_id: "Iv1.abc123def456"       # Direct value
    client_secret: "secret789xyz"       # Direct value (not recommended)
    callback_url: "https://dev-tools.company.com/auth/github/callback"
```

**How It Works:**
- **Security Check**: Errors if non-localhost host specified without authentication
- **GitHub OAuth Flow**: User logs in → Check org/team membership → Create session → Access dashboard  
- **Container Access**: Uses SSH private key to generate JWTs on-demand for container API calls
- **Session Management**: Web session for dashboard access, JWTs for container communication

### Pattern 2: SaaS Dashboard (claude-vm.com)
**Use Case:** Team collaboration, mobile access, unified management

```bash
# Access via web browser after login
open https://claude-vm.com
# → OAuth login (GitHub/Google)
# → See all SaaS-registered containers
```

**How It Works:**
- OAuth login establishes user identity
- Shows only SaaS-registered containers (created while logged in)
- Uses SaaS service SSH keys for container authentication
- Provides team features, billing, advanced management

### Pattern 3: Individual Container Web UI
**Use Case:** Direct access to specific workspace, sharing with others

```bash
# Generate JWT for specific container
claude-vm auth token bold-fire-1234

# Visit container URL directly
open https://bold-fire-1234.fly.dev
# → Shows JWT authentication form
# → Paste token → Full web UI access
```

**How It Works:**
- Each container has web UI built-in (runs on port 8080)
- Simple JWT authentication form
- Scoped access to single workspace only
- Perfect for sharing access to specific project

### Pattern 4: Self-Hosted Dashboard
**Use Case:** Enterprise compliance, custom domains, full control

```bash
# Deploy self-hosted dashboard with GitHub OAuth (using config file)
docker run -d --name claude-vm-dashboard \
  -p 443:3000 \
  -v ~/.ssh/id_rsa:/app/.ssh/id_rsa:ro \
  -v ~/.claude-vm/config.yaml:/app/config.yaml:ro \
  -v ./auth.yaml:/app/auth.yaml:ro \
  -e CLAUDE_VM_AUTH_CONFIG="/app/auth.yaml" \
  -e CLAUDE_VM_DOMAIN="dev-containers.acme.com" \
  claude-vm/web-ui

# Deploy using standard pattern: config file + environment variables
docker run -d --name claude-vm-dashboard \
  -p 443:3000 \
  -v ~/.ssh/id_rsa:/app/.ssh/id_rsa:ro \
  -v ~/.claude-vm/config.yaml:/app/config.yaml:ro \
  -v ./auth.yaml:/app/auth.yaml:ro \
  -e GITHUB_CLIENT_ID="Iv1.abc123def456" \
  -e GITHUB_CLIENT_SECRET="secret789xyz" \
  claude-vm/web-ui --host 0.0.0.0 --auth-config /app/auth.yaml

# Kubernetes deployment with ConfigMap + Secrets
kubectl create secret generic github-oauth \
  --from-literal=GITHUB_CLIENT_ID="Iv1.abc123def456" \
  --from-literal=GITHUB_CLIENT_SECRET="secret789xyz"
kubectl apply -f claude-vm-dashboard.yaml
```

**How It Works:**
- **Same Authentication**: Uses identical auth system as `claude-vm web`
- **Container Access**: Uses mounted SSH private key for JWT generation
- **Enterprise Features**: Custom domains, branding, and team management
- **Flexible Deployment**: Docker, Kubernetes, or traditional hosting

## GitHub OAuth Setup for Self-Hosting

### Why OAuth Over PATs?

**Personal Access Token Approach (❌ Not Recommended):**
```bash
# User would need to manually create PAT with read:org scope
# Then paste it into self-hosted dashboard
# Problems: cumbersome UX, broader access than needed, manual management
```

**OAuth App Approach (✅ Recommended):**
- ✅ **Better UX**: Users just click "Login with GitHub"
- ✅ **Secure**: Limited scope, proper OAuth flow
- ✅ **Organization Control**: Admins approve which apps can access org
- ✅ **Industry Standard**: Same as Grafana, GitLab, ArgoCD

### Setup Process

**1. Create GitHub OAuth App:**
```bash
# Organization admin goes to:
# GitHub → Settings → Developer settings → OAuth Apps → New OAuth App

Application name: "Acme Corp Claude-VM Dashboard"  
Homepage URL: "https://dev-tools.acme.com"
Authorization callback URL: "https://dev-tools.acme.com/auth/github/callback"

# GitHub provides:
Client ID: "Iv1.abc123def456"
Client Secret: "secret789xyz"
```

**2. Organization Approval:**
```bash
# Organization admin goes to:
# GitHub Org → Settings → Third-party access → OAuth Apps
# → Review pending requests → Approve "Acme Corp Claude-VM Dashboard"
# → Set organization member access policy (if needed)
```

**3. Deploy Self-Hosted Dashboard:**
```yaml
# auth.yaml
auth:
  type: github
  github:
    org: "acme-corp"                    # Your GitHub org
    teams: ["dev-team", "ops"]          # Optional: restrict to teams
    client_id: "${GITHUB_CLIENT_ID}"    # From environment variables
    client_secret: "${GITHUB_CLIENT_SECRET}"  # From environment variables
    callback_url: "https://dev-tools.acme.com/auth/github/callback"
```

### User Login Flow

**"Login with GitHub" Experience:**
```bash
# 1. User visits https://dev-tools.acme.com
# 2. Sees "Login with GitHub" button
# 3. Clicks button → redirected to github.com/login/oauth/authorize
# 4. GitHub shows: "Acme Corp Claude-VM Dashboard wants access to verify your organization membership"
# 5. User clicks "Authorize"
# 6. GitHub redirects back to dev-tools.acme.com with OAuth code
# 7. Dashboard exchanges code for access token
# 8. Dashboard calls GitHub API to verify org membership
# 9. If user in required org/teams → create session → access granted
```

### GitHub API Integration

**Organization Membership Check:**
```bash
# Dashboard makes these API calls with user's OAuth token:

# 1. Get user info
GET https://api.github.com/user
# Returns: {"login": "john-doe", "id": 12345, ...}

# 2. Check organization membership  
GET https://api.github.com/user/memberships/orgs/acme-corp
# Returns: {"state": "active", "role": "member", ...} or 404 if not member

# 3. Check team membership (if teams specified)
GET https://api.github.com/user/teams
# Returns: [{"name": "dev-team", "organization": {"login": "acme-corp"}}, ...]
```

**Required OAuth Scopes:**
```bash
# OAuth app requests these permissions:
read:org    # Check organization membership
read:user   # Get user profile info
# Note: Very limited scope, much safer than broad PAT
```

**Security Benefits:**
- ✅ **No Password Management**: Uses existing GitHub accounts
- ✅ **Organization Control**: Admins control access via GitHub org membership  
- ✅ **Team Granularity**: Restrict to specific teams within organization
- ✅ **Audit Trail**: GitHub provides login/access logs
- ✅ **2FA Support**: Inherits GitHub's security policies (2FA requirements, etc.)

## Pattern Comparison

| Pattern | Access | Web Auth | Container Auth | Best For |
|---------|--------|----------|---------------|----------|
| **Localhost** | `localhost:3000` | None (localhost) / Required auth (external) | SSH-signed JWTs | Daily development |
| **SaaS Dashboard** | `claude-vm.com` | OAuth (GitHub/Google) | Service SSH JWTs | Team collaboration |
| **Individual Container** | `container-url.fly.dev` | Pasted JWT token | Same JWT | Sharing projects |
| **Self-Hosted** | `company.com` | GitHub OAuth + Org membership | SSH-signed JWTs | Enterprise compliance |

**Authentication Methods Available:**
- **No auth** - Default for localhost (127.0.0.1), errors for external hosts
- `--auth-password <secret>` - Simple password protection
- `--auth-github-org <org>` - GitHub OAuth with organization membership requirement  
- `--auth-config <file>` - YAML configuration file with detailed auth settings

**Two-Tier Authentication Architecture:**
1. **User → SaaS**: OAuth (GitHub, Google) for web interface access
2. **SaaS → Container**: Service SSH keys for backend operations

## Simplified SaaS Integration (Login-Based)

**Core Principle**: User's login state determines whether new containers get SaaS integration.

### Login-Based Container Creation
```bash
# Login to SaaS (optional - enables SaaS integration for new containers)
claude-vm login
# → OAuth flow with claude-vm.com 
# → Stores account ID and access token locally

# Check login status
claude-vm auth status
# Output: Logged in as user-123 (claude-vm.com)
#         SSH Key: ~/.ssh/claude-vm_ed25519

# Logout (disables SaaS integration for future containers)
claude-vm logout
# → Clears login state
# → Existing SaaS-managed containers remain accessible
```

### Container Creation Based on Login State
```bash
claude-vm workspace up .

# IF user is logged in:
# → Injects user SSH key + SaaS service key + account ID
# → Container automatically registers with SaaS
# → Visible in both local CLI and SaaS web dashboard

# IF user is NOT logged in:
# → Injects only user SSH key  
# → Container is local-only (no SaaS integration)
# → Only accessible via local CLI and local web UI
```

### Container Ownership via Account ID Injection
```yaml
# Local-only container (user not logged in)
USER_SSH_PUBLIC_KEY: "ssh-rsa AAAAB3... user@machine"
WORKSPACE_ID: "bold-fire-1234"

# SaaS-integrated container (user logged in)
USER_SSH_PUBLIC_KEY: "ssh-rsa AAAAB3... user@machine"
SAAS_SSH_PUBLIC_KEY: "ssh-rsa AAAAB3... service@claude-vm.com"
WORKSPACE_ID: "bold-fire-1234"
USER_ACCOUNT_ID: "user-123"  # ← SaaS knows this container belongs to user-123
```

### Container Self-Registration

**Container Startup Logic:**
- Always setup user SSH access with user's public key
- If SaaS-integrated (USER_ACCOUNT_ID present): add SaaS service public key  
- If SaaS-integrated: register with SaaS API endpoint with container ID and account ID
- Start SSH daemon and API server

## Container Discovery & Registration Architecture

### Local Container Discovery (JSON File Tracking)
```bash
# ~/.claude-vm/workspaces.yaml tracks all known containers
claude-vm web
# → Reads config file to discover containers
# → Makes API calls to each container URL to get status
# → Shows unified dashboard of local + SaaS containers
```

**Local Config File:**
```yaml
# ~/.claude-vm/config.yaml
auth:
  logged_in: true
  account_id: "user-123"
  access_token: "oauth2_token_abc123..."
  token_expires: "2025-09-11T04:13:07Z"
  endpoint: "https://claude-vm.com"

workspaces:
  bold-fire-1234:
    url: "https://bold-fire-1234.fly.dev"
    saas_managed: true
    created: "2025-08-11T04:13:07Z"
  quiet-lake-5678:
    url: "https://quiet-lake-5678.do.dev"
    saas_managed: false
    created: "2025-08-10T14:22:11Z"
```

### SaaS Container Discovery (Self-Registration)

**Self-Registration Process:**
- Containers register themselves with SaaS (not CLI registration)  
- More reliable - container startup always happens even if CLI crashes
- If USER_ACCOUNT_ID present: container calls SaaS registration API
- Sends container ID, container URL, and account ID to establish ownership

### Complete Container Creation Flow (Logged In User)

**High-Level Process:**
1. User runs `claude-vm workspace up .`
2. CLI checks login state from local config file
3. CLI fetches SaaS service public key using stored OAuth token  
4. CLI creates container with user SSH key + SaaS service key + account ID
5. Container starts and self-registers with SaaS API
6. CLI updates local config file with new container info
7. Result: Container visible in both local web UI and SaaS dashboard

## Simplified Access Control (Single User)

**SaaS Database Schema:**
- Users table: basic user info (ID, email, GitHub ID, name, timestamps)
- User_containers table: container ownership mapping (user ID → container ID, URL, status, timestamps)
- Simple 1:1 relationship - each container belongs to exactly one user
- No teams, organizations, or complex IAM structures

**Local Configuration (Login State):**
```yaml
# ~/.claude-vm/config.yaml
auth:
  logged_in: true
  account_id: "user-123"
  access_token: "jwt-access-token..."
  endpoint: "https://claude-vm.com"

ssh:
  key: "~/.ssh/claude-vm_ed25519"

workspaces:
  bold-fire-1234:
    url: "https://bold-fire-1234.fly.dev"
    saas_managed: true
    created: "2025-08-11T04:13:07Z"
```

**SaaS API Endpoints:**
- OAuth login/callback endpoints for web interface authentication
- GET /api/containers - returns user's registered containers
- POST /api/containers - container self-registration endpoint (called by containers on startup)
- Simple authentication - no team/organization complexity

**Container Environment Variables:**
```yaml
USER_SSH_PUBLIC_KEY: "ssh-rsa AAAAB3... user@machine"
SAAS_SSH_PUBLIC_KEY: "ssh-rsa AAAAB3... service@claude-vm.com"  # Only if logged in
WORKSPACE_ID: "bold-fire-1234"
USER_ACCOUNT_ID: "user-123"  # Only if logged in
```

**Container Startup Process:**
1. Setup SSH access with user's public key (always)
2. Add SaaS service public key if USER_ACCOUNT_ID present
3. Set proper SSH permissions and start SSH daemon
4. Configure API server for dual-key JWT validation (user + SaaS)
5. Start workspace API server

**SaaS Background Operations:**
- Health monitoring: Check all registered containers periodically using service SSH keys
- Backup scheduling: Trigger container backups on schedule via API calls
- Resource monitoring: Track container usage for billing and scaling
- All operations use service SSH keys and JWT tokens for authentication

**JWT Token Types:**
- **User tokens**: Signed with user's SSH private key for direct access (1 hour expiration)
- **Service tokens**: Signed with SaaS SSH private key for backend operations (24 hour expiration)

**Container Dual-Key Validation:**
Containers validate JWTs using both user and SaaS SSH public keys, trying user key first, then SaaS service key for backend operations.

## SaaS Service Architecture (claude-vm.com)

Our centralized SaaS service manages container registration, user authentication, and background operations. Here's how containers register and establish user ownership:

### Container Registration Flow

**1. Container Creation with User Context:**
When a logged-in user creates a container, CLI injects the account ID:
```bash
# CLI sets these environment variables during container creation
export USER_ACCOUNT_ID="user-123"           # Links container to user account
export SAAS_SSH_PUBLIC_KEY="ssh-rsa AAAAB..." # Enables SaaS backend operations
export WORKSPACE_ID="bold-fire-1234"        # Unique container identifier
```

**2. Container Self-Registration on Startup:**
```bash
# Container startup script checks for SaaS integration
if [ ! -z "$USER_ACCOUNT_ID" ]; then
  # Generate service JWT for authentication
  service_jwt=$(generate_service_jwt "$WORKSPACE_ID" "$USER_ACCOUNT_ID")
  
  # Register with SaaS service
  curl -X POST "https://claude-vm.com/api/containers/register" \
    -H "Authorization: Bearer $service_jwt" \
    -H "Content-Type: application/json" \
    -d "{
      \"container_id\": \"$WORKSPACE_ID\",
      \"account_id\": \"$USER_ACCOUNT_ID\",
      \"container_url\": \"$CONTAINER_PUBLIC_URL\",
      \"ssh_public_keys\": {
        \"user\": \"$USER_SSH_PUBLIC_KEY\",
        \"saas\": \"$SAAS_SSH_PUBLIC_KEY\"
      }
    }"
fi
```

**3. SaaS Service Registration API:**
```
POST /api/containers/register
Authentication: Service JWT (signed with SaaS SSH key)
Rate Limiting: 10 requests/minute per container

Request Body:
{
  "container_id": "bold-fire-1234",
  "account_id": "user-123", 
  "container_url": "https://bold-fire-1234.fly.dev",
  "ssh_public_keys": {...}
}

Response (Success):
{
  "status": "registered",
  "container_id": "bold-fire-1234",
  "owner": "user-123",
  "registered_at": "2025-08-11T04:13:07Z"
}
```

### Database Schema

**Users Table:**
```sql
users (
  id VARCHAR PRIMARY KEY,           -- "user-123"
  email VARCHAR UNIQUE NOT NULL,    -- "user@example.com" 
  github_id INT UNIQUE,             -- OAuth provider ID
  name VARCHAR,                     -- "John Doe"
  created_at TIMESTAMP,
  last_login TIMESTAMP
)
```

**Containers Table:**
```sql
containers (
  id VARCHAR PRIMARY KEY,           -- "bold-fire-1234" 
  account_id VARCHAR NOT NULL,      -- References users.id
  container_url VARCHAR NOT NULL,   -- "https://bold-fire-1234.fly.dev"
  status VARCHAR DEFAULT 'active',  -- active, stopped, error
  ssh_user_key TEXT,               -- User's SSH public key
  ssh_saas_key TEXT,               -- SaaS service SSH public key  
  created_at TIMESTAMP,
  last_heartbeat TIMESTAMP,        -- Health monitoring
  FOREIGN KEY (account_id) REFERENCES users(id)
)
```

### User Ownership Determination

**How we know which user owns a container:**
1. **Account ID Injection**: CLI injects `USER_ACCOUNT_ID` during container creation
2. **Authenticated Registration**: Container uses service JWT to prove legitimacy  
3. **Database Link**: Registration API creates `containers` record linking container ID to user account
4. **Verification**: SaaS validates account ID exists and container URL is reachable

**Security Measures:**
- Service JWT prevents spoofed registrations (only real containers can generate valid JWTs)
- Account ID must exist in users table (prevents registration to non-existent accounts)
- Container URL must be publicly accessible (prevents registration of fake containers)
- Rate limiting prevents abuse (10 registrations/minute per container)

### Background SaaS Operations

**Health Monitoring:**
```bash
# Cron job every 5 minutes
for container in $(get_all_containers); do
  # Use SaaS service JWT to check container health
  curl -H "Authorization: Bearer $SAAS_SERVICE_JWT" \
       "$container_url/api/health" || mark_container_unhealthy "$container"
done
```

**Automated Backups:**
- Daily S3 backups for all registered containers using service JWTs
- Conversation history, file snapshots, and workspace metadata
- User can access backups via SaaS web dashboard

**Resource Monitoring:**
- Track container CPU/memory usage for billing
- Monitor storage usage across workspace volumes
- Generate usage reports for user dashboard

### API Endpoints

**Core SaaS APIs:**
```
Authentication: OAuth 2.0 (GitHub/Google) for web UI
                Service JWTs for container operations

POST   /api/auth/login              # OAuth login flow
GET    /api/auth/user               # Get current user info  
POST   /api/auth/logout             # Clear session

POST   /api/containers/register     # Container self-registration
GET    /api/containers              # List user's containers
GET    /api/containers/{id}         # Get container details
DELETE /api/containers/{id}         # Unregister container

GET    /api/backups                 # List user's backups
POST   /api/backups/{container_id}  # Trigger manual backup
```

**Why Container Self-Registration Works:**
- ✅ **Authoritative**: Container startup is the definitive event
- ✅ **Reliable**: Works even if CLI crashes after container creation  
- ✅ **Secure**: Service JWT prevents unauthorized registrations
- ✅ **Simple**: No complex orchestration between CLI and SaaS needed
- ✅ **Standard**: Same pattern as AWS Auto Scaling, Kubernetes service discovery

## Unified Web UI Experience (Like GitHub Codespaces Dashboard)

**Local Web UI (`claude-vm web`) - Shows All Containers:**
Unified dashboard showing both local-only and SaaS-managed containers with categorized workspaces (local_only, saas_managed, saas_only, unified view).

**SaaS Web UI (claude-vm.com) - Team Collaboration:**
Team-focused dashboard with role-based access showing owned, collaborating, team, and recent workspaces.

## SSH Keypair Authentication (Unified Approach)

Uses standard SSH keys for both SSH access and JWT authentication - no separate key management needed.

### SSH Key Generation & Management (Following GitHub CLI Pattern)

**Key Detection Priority:**
```bash
# CLI automatically detects SSH keys in this order:
claude-vm workspace up .

# 1. User-configured key (explicit choice)
claude-vm config get ssh-key  # → ~/.ssh/my-custom-key (if set)

# 2. SSH agent keys (passwordless convenience)
ssh-add -l  # → Use first available agent key if found

# 3. Standard keys in priority order  
test -f ~/.ssh/id_ed25519     # Modern default (preferred)
test -f ~/.ssh/id_rsa         # Traditional default
test -f ~/.ssh/id_ecdsa       # Alternative

# 4. Generate new key with user consent (if none found)
```

**First-Time SSH Key Setup:**
```bash
claude-vm workspace up .

# If no SSH key found, prompts user:
┌─ SSH Key Setup Required ────────────────────────────────────────┐
│ No SSH key found for claude-vm authentication.                  │
│                                                                  │
│ Options:                                                         │
│ 1) Generate new SSH key (recommended)                           │
│ 2) Use existing SSH key                                         │
│ 3) Specify custom SSH key path                                  │
│                                                                  │
│ Select [1-3]: 1                                                 │
└──────────────────────────────────────────────────────────────────┘

❯ Generating new Ed25519 SSH key...
❯ Saved to ~/.ssh/claude-vm_ed25519
❯ Added to SSH agent  
✓ SSH key ready for claude-vm workspaces
```

**SSH Key Generation Process:**
```bash
# Generates Ed25519 key (modern, secure, fast)
ssh-keygen -t ed25519 \
  -f ~/.ssh/claude-vm_ed25519 \
  -C "claude-vm-$(whoami)@$(hostname)" \
  -N ""  # No passphrase for automation

# Sets proper permissions
chmod 600 ~/.ssh/claude-vm_ed25519
chmod 644 ~/.ssh/claude-vm_ed25519.pub

# Adds to SSH agent if available
ssh-add ~/.ssh/claude-vm_ed25519

# Updates claude-vm config
claude-vm config set ssh-key ~/.ssh/claude-vm_ed25519
```

### SSH Key Management Commands:
```bash
# Show current SSH key status (like gh auth status)
claude-vm auth status
# Output: SSH Key: ~/.ssh/claude-vm_ed25519 (Ed25519, 256-bit)
#         SSH Agent: Key loaded, 1 identity

# List available SSH keys (like gh auth keys) 
claude-vm auth keys
# Output: 
# ~/.ssh/claude-vm_ed25519 (Ed25519) [current] [auto-generated]
# ~/.ssh/id_rsa (RSA 4096-bit)
# ~/.ssh/work_key (RSA 2048-bit)

# Switch to different SSH key
claude-vm auth use-key ~/.ssh/id_rsa
# Output: Updated SSH key to ~/.ssh/id_rsa

# Generate new SSH key (explicit command)
claude-vm auth generate-key  
# Output: Generated new Ed25519 key: ~/.ssh/claude-vm_ed25519_2

# Use SSH agent for signing (passwordless)
ssh-add ~/.ssh/claude-vm_ed25519
claude-vm chat "fix-bug"  # Uses SSH agent automatically
```

### User Experience Examples:

**Scenario 1: New User (No Existing SSH Keys)**
```bash
claude-vm workspace up .
# → Auto-detects no SSH keys exist
# → Prompts user to generate new key  
# → Generates ~/.ssh/claude-vm_ed25519
# → Adds to SSH agent
# → Stores in config, proceeds with workspace creation
```

**Scenario 2: Developer with Existing SSH Keys**
```bash
claude-vm workspace up .
# → Detects existing ~/.ssh/id_ed25519

┌─ SSH Key Detected ──────────────────────────────────────────────┐
│ Found SSH key: ~/.ssh/id_ed25519 (Ed25519)                     │
│                                                                  │  
│ Use this key for claude-vm? [Y/n]: y                           │
└──────────────────────────────────────────────────────────────────┘

✓ Using existing SSH key: ~/.ssh/id_ed25519
❯ Proceeding with workspace creation...
```

**Scenario 3: SSH Agent Integration**
```bash
# User has SSH agent with loaded keys
ssh-add ~/.ssh/work_key

claude-vm workspace up .
# → Auto-detects SSH agent
# → Uses first available agent key 
# → No file I/O needed for signing (uses agent)
```

**Why Ed25519 Keys (Industry Standard):**
- **Modern**: Recommended by security experts and GitHub
- **Fast**: Much faster signing than RSA (important for frequent JWT generation)
- **Small**: 68-character public keys vs 800+ for RSA
- **Secure**: Resistant to timing attacks and quantum-resistant precursor

**Configuration Storage:**
```yaml
# ~/.claude-vm/config.yaml (user preferences only)
defaults:
  cloud: "docker"
  agent: "claude"
  auto_stop_minutes: 60

# Only configured clouds with non-default settings
clouds:
  fly:
    region: "lax"

# Only configured agents with non-default settings  
agents:
  claude:
    model: "opus"
```

**~/.claude-vm/workspaces.yaml (workspace registry):**
```yaml
version: "1.0"

workspaces:
  bold-fire-1234:
    name: "bold-fire-1234"
    status: "running"
    provider: "fly"
    url: "https://bold-fire-1234.fly.dev"
    ssh_key: "~/.ssh/claude-vm_ed25519"
    saas_managed: false
    created: "2025-08-11T04:13:07Z"
```

### SSH Key Management Commands:
```bash
# Show current SSH key status
claude-vm auth status
# Output: SSH Key: ~/.ssh/id_rsa (RSA 4096-bit), SSH Agent: 1 key loaded

# List available SSH keys
claude-vm auth keys  
# Output: ~/.ssh/id_rsa (RSA 4096-bit) [current], ~/.ssh/id_ed25519 (ED25519)

# Use specific SSH key for new workspaces
claude-vm auth use-key ~/.ssh/id_ed25519

# Generate SSH key if none exists  
claude-vm auth generate-key
# Runs: ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519

# Update existing workspace SSH key
claude-vm workspace update bold-fire-1234 --ssh-key ~/.ssh/new_key.pub
```

### SSH Agent Integration:
```bash
# Load key into SSH agent for automatic signing
ssh-add ~/.ssh/id_rsa

# All claude-vm operations use SSH agent automatically
claude-vm chat "fix-bug"                   # Gets private key from SSH agent
ssh user@bold-fire-1234.fly.dev          # Uses same key for SSH access
```

## SSH-Based Authentication Flow

### Container SSH Key Configuration:
```yaml
# Container receives single SSH public key for both purposes
SSH_PUBLIC_KEY: "ssh-rsa AAAAB3NzaC1yc2EAAAA... user@machine"
WORKSPACE_ID: "bold-fire-1234"
USER_ID: "user123"
```

### Container Setup (Unified):
```bash
#!/bin/bash
# Container startup script - single key serves both purposes

# 1. SSH access setup  
mkdir -p ~/.ssh
echo "$SSH_PUBLIC_KEY" > ~/.ssh/authorized_keys
chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys
service ssh start

# 2. API server JWT validation (same key)
export JWT_PUBLIC_KEY="$SSH_PUBLIC_KEY" 
node /workspace/api-server.js
```

### JWT Generation (CLI Side):
CLI signs JWTs with SSH private key using RS256 algorithm. JWT payload structure:

```yaml
# JWT Payload (JSON format when encoded)
workspace: "workspaceId"
user: "userId"
exp: 1691756787
iat: 1691753187
```

**CLI Commands for JWT Generation:**
```bash
# Generate JWT for specific workspace
claude-vm workspace token bold-fire-1234
# Output: eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ3b3Jrc3BhY2UiOiJib2xkLWZpcmUtMTIzNCIsInVzZXIiOiJ1c2VyLTEyMyIsImV4cCI6MTY5MTc1Njc4NywiaWF0IjoxNjkxNzUzMTg3fQ...

# Copy token to clipboard (like gh auth token | pbcopy) 
claude-vm workspace token bold-fire-1234 | pbcopy

# Show current workspace registry (for self-hosted web UI setup)
claude-vm workspace list --json
# Output: {"bold-fire-1234": {"url": "https://bold-fire-1234.fly.dev", "saas_managed": true}}
```

**Direct Container Web Access Flow:**
```bash
# User gets JWT for specific container
claude-vm workspace token bold-fire-1234

# User visits container URL: https://bold-fire-1234.fly.dev  
# Container shows simple login form: "Paste your JWT token"
# User pastes token and gets full web UI access
```

SSH Agent integration provides passwordless signing when keys are loaded in agent.

### JWT Validation (Container Side):
Container validates JWTs with SSH public key from environment variables, verifying workspace ID matches and using RS256 algorithm with 60-second clock tolerance. Express middleware handles Bearer token extraction and validation.


## Complete Authentication Matrix

| Access Method | Authentication | Key Used | User Experience |
|---------------|----------------|----------|------------------|
| **Local CLI/Web Patterns** | | | |
| `claude-vm chat "fix-bug"` | JWT signed with user SSH key | `~/.ssh/id_rsa` | Automatic (no login) |
| `claude-vm web` | JWT signed with user SSH key | `~/.ssh/id_rsa` | Automatic (reads config) |
| `ssh user@container.host` | SSH public key authentication | `~/.ssh/id_rsa` | Standard SSH |
| Direct container web visit | JWT (paste token from CLI) | `~/.ssh/id_rsa` | Simple token form |
| **Hybrid SaaS Patterns (Industry Standard)** | | | |
| SaaS web dashboard (claude-vm.com) | GitHub/Google OAuth → Service JWT | SaaS SSH key | OAuth login, unified dashboard |
| SaaS background operations | Service JWT signed with SaaS SSH key | SaaS SSH key | Automatic (health, backups) |
| Local web UI (`claude-vm web`) | SSH-signed JWT | User's SSH key | Shows local + SaaS containers |
| **Container Access (Local or SaaS-created)** | | | |
| CLI to any container | JWT signed with user SSH key | User's `~/.ssh/id_rsa` | Always works (user key always injected) |
| SSH to any container | SSH public key authentication | User's `~/.ssh/id_rsa` | Standard SSH access |
| **SaaS Login Commands** | | | |
| `claude-vm login` | OAuth flow, stores account ID locally | N/A | Enables SaaS for new containers |
| `claude-vm logout` | Clears login state | N/A | Disables SaaS for new containers |

**Simplified Login-Based Flow Examples:**

### Scenario 1: Privacy-First Developer (Local-Only)
```bash
# No login required - works locally by default
claude-vm workspace up .              # Creates local-only container
claude-vm chat "fix-bug"              # CLI access  
claude-vm web                         # Local web UI access
ssh user@workspace.host               # Direct SSH access

# Container has only user's SSH key
# No SaaS integration or account tracking
```

### Scenario 2: SaaS-Integrated Developer  
```bash
# Login once to enable SaaS integration
claude-vm login
# → OAuth with claude-vm.com, stores account ID

claude-vm workspace up .              # Creates SaaS-integrated container
# → Automatically injects user + SaaS service keys + account ID
# → Container self-registers with SaaS service
# → Visible in both local CLI and claude-vm.com dashboard

claude-vm chat "fix-bug"              # CLI access (same as before)
ssh user@workspace.host               # SSH access (same as before)
# + Now also visible in SaaS web interface for management
```

### Scenario 3: Mixed Local + SaaS Usage
```bash
claude-vm login                       # Enable SaaS integration
claude-vm workspace up project1      # → SaaS-integrated container

claude-vm logout                      # Disable SaaS for future containers  
claude-vm workspace up project2      # → Local-only container

# Result:
# - project1: Visible in local CLI + SaaS dashboard 
# - project2: Only visible in local CLI
# - User can access both normally via CLI/SSH
```

### Scenario 4: SaaS Dashboard Management
```bash
# User visits claude-vm.com (after logging in via CLI)
# → See all SaaS-integrated containers
# → Monitor container status, resource usage
# → Access containers via web terminal
# → View conversation history, backups

# CLI still works normally for direct access
claude-vm workspace list              # Shows local + SaaS containers
claude-vm chat "review-code"          # Direct CLI access
```

**SaaS Background Operations (Independent):**
```bash
# SaaS manages all registered containers 24/7 using service SSH keys:
# - Health monitoring and alerts
# - Scheduled backups to S3
# - Resource scaling based on usage  
# - Billing calculations and reporting
# - Team access control enforcement
# - All using service JWTs signed with SaaS SSH private key
```

## Simplified Architecture Validation

This simplified architecture addresses the three core requirements:

1. **User Opt-in/Opt-out Control**: Default privacy-first local-only containers with explicit login-based opt-in to SaaS integration
2. **Centralized Access Control**: Single user accounts with container ownership via account ID injection 
3. **Dynamic SSH Key Injection**: Login state determines container configuration (user key only vs user + SaaS service keys)

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
  - Conversation mappings: /workspace/.claude-vm/conversations.yaml (only this file)

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
# Track conversation chains and backup all related JSONL files
current_uuid=$(yq -r ".conversations[\"$conversation_name\"].conversationId" /workspace/.claude-vm/conversations.yaml)
for jsonl_file in "$claude_dir"/*.jsonl; do
  if grep -q "\"id\":\"$current_uuid\"" "$jsonl_file"; then
    aws s3 cp "$jsonl_file" "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/conversations/"
  fi
done
```

**File Selection Implementation:**
```bash
# Create comprehensive workspace backup
git ls-files > /tmp/backup-files.txt
git diff --name-only >> /tmp/backup-files.txt
find ~/.claude/projects -name "*.jsonl" >> /tmp/backup-files.txt
find deliverables -type f >> /tmp/backup-files.txt 2>/dev/null || true
grep -v -E "(node_modules|__pycache__|build/)" /tmp/backup-files.txt > /tmp/filtered.txt
tar -czf "workspace-$timestamp.tar.gz" --files-from=/tmp/filtered.txt
aws s3 cp "workspace-$timestamp.tar.gz" "s3://claude-vm-backups/$user_id/workspaces/"
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
  "You are running inside a devcontainer on a remote VM with full development capabilities.
   Your conversation history and all /workspace/ files persist across container restarts:
   - Git repository state (tracked files + changes)
   - Files in /workspace/deliverables/
   - Build caches in /workspace/.cache/
   - Your state in /workspace/.claude-vm/
   You can install packages, run docker-compose via Docker-in-Docker, and have access to all development tools."
```

**Complete System Prompt Examples:**

```yaml
# Claude Code System Prompt Addition
CLAUDE_VM_CONTEXT = """
You are Claude Code running inside a devcontainer on DigitalOcean infrastructure.
Workspace: bold-fire-1234
Project: github.com/user/myproject (branch: main)

IMPORTANT - You have full development capabilities:
- You run inside a devcontainer with development tools pre-installed
- Git repository is at /workspace/repo/ (always persistent)
- Put final deliverables in /workspace/deliverables/ (always persistent)
- Build caches persist in /workspace/.cache/ (npm, pip, cargo, etc.)
- You can install packages: apt install <package>, pip install <package>, npm install <package>
- You can run docker-compose stacks via Docker-in-Docker for local development
- Multiple agents may be running in this same container environment

IMPORTANT - Git and external operations:
- You have full git access: git add, git commit, git push, git pull, branching, etc.
- Git credentials (SSH keys/tokens) are securely managed via agent forwarding - you'll never see the actual keys
- AWS CLI, Docker registry, database connections work normally via credential forwarding/proxying
- All sensitive credentials (API keys, passwords, certificates) are VM-managed - you never see raw credentials
- Deliverables backup handled automatically (VM → S3-compatible storage or Container → API server → storage)
- Use external services normally for development - authentication is handled transparently

All your work persists across container restarts. When user runs 'claude-vm shell', they'll connect to this same environment.
"""

# Goose System Prompt Addition  
GOOSE_VM_CONTEXT = """
You are Goose AI running inside a devcontainer on Fly.io infrastructure.
Workspace: quiet-lake-5678  
Project: github.com/company/backend (branch: feature-auth)

Development Environment:
- You run inside a devcontainer with development tools available
- Git repository at /workspace/repo/ persists across restarts
- Save deliverables to /workspace/deliverables/ (always persistent)
- Build caches in /workspace/.cache/ persist automatically
- Install packages: apt install <package>, pip install <package>, etc.
- Run docker-compose for local services via Docker-in-Docker
- Multiple agents may share this container environment

External Operations (Credential Forwarding):
- You have full git access via secure credential forwarding - use git commands normally
- AWS CLI, Docker registry, external APIs work via credential helpers - you never see raw credentials
- All sensitive credentials are VM-managed for security
- Deliverables backup handled automatically (VM → S3-compatible storage or Container → API server → storage)
- Focus on development - credential management is handled transparently

This is a persistent development environment - everything in /workspace/ survives restarts.
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
  aws s3 cp "s3://claude-vm-backups/$user_id/workspaces/$workspace_id/conversations.yaml" /workspace/.claude-vm/conversations.yaml
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
cat ~/.claude-vm/conversations.yaml | yq -r 'keys[]' 
echo ""
echo "Tmux sessions:"
tmux list-sessions
echo ""
echo "To join a conversation: tmux attach-session -t \"<conversation-name>\""
echo "To start claude-vm CLI: claude-vm chat --list"

# SSH login message in ~/.bashrc
cat >> ~/.bashrc << 'EOF'
echo "🚀 Claude VM Workspace"
echo "Active conversations: $(cat ~/.claude-vm/conversations.yaml 2>/dev/null | yq -r 'keys | length // 0')"
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
```yaml
# ~/.claude-vm/workspaces.yaml (managed by CLI)
version: "1.0"
last_updated: "2025-08-11T08:22:11Z"

workspaces:
  bold-fire-1234:
    name: "bold-fire-1234"
    status: "running"
    provider: "fly"
    region: "iad"
    url: "https://bold-fire-1234.fly.dev"
    ssh_host: "bold-fire-1234.fly.dev"
    agents: ["claude", "goose"]
    created: "2025-08-11T04:13:07Z"
    last_activity: "2025-08-11T08:22:11Z"
    project:
      repo: "github.com/user/project"
      branch: "main"
      path: "/workspace"
      
  quiet-lake-5678:
    name: "quiet-lake-5678"
    status: "stopped"
    provider: "digitalocean"
    region: "nyc1"
    agents: ["claude"]
    created: "2025-08-10T14:22:11Z"
    last_activity: "2025-08-10T16:33:55Z"
```

### 2. Conversation Tracking (Per Workspace Container)
```yaml
# /workspace/.claude-vm/conversations.yaml (inside each workspace container)
# This is the ONLY container state file we back up
metadata:
  version: "1.0"
  last_updated: "2025-08-11T08:22:11Z"

conversations:
  fix-auth-bug:
    agent: "claude"
    tmux_session: "fix-auth-bug"
    status: "active"
    created: "2025-08-11T04:13:07Z"
    last_activity: "2025-08-11T06:45:22Z"
    message_count: 45
    agent_specific:
      conversation_chain: ["abc123-original", "def456-resumed", "ghi789-current"]
      active_uuid: "ghi789-current"
  refactor-db:
    agent: "goose"
    tmux_session: "refactor-db"
    status: "idle"
    created: "2025-08-10T14:22:11Z"
    last_activity: "2025-08-11T08:22:11Z"
    message_count: 12
    agent_specific:
      session_id: "goose-session-xyz789"
```

**Container State Philosophy:**
- Keep minimal state in containers (only conversation mappings)
- Workspace metadata belongs in local CLI or hosted service
- Agent credentials managed by agents themselves
- This single JSON file is easy to backup, manipulate with `jq`, and extend

**How the Two-Level System Works:**

```bash
# User runs: claude-vm chat "fix-auth-bug"
# 1. Local CLI reads ~/.claude-vm/workspaces.yaml to find current workspace
# 2. Connects to workspace (SSH/API) and reads /workspace/.claude-vm/conversations.yaml  
# 3. Finds conversation "fix-auth-bug" → agent "claude", conversationId "abc123..."
# 4. Attaches to tmux session "fix-auth-bug" or recreates it if needed

# User runs: claude-vm workspace list
# 1. Local CLI reads ~/.claude-vm/workspaces.yaml
# 2. Shows all workspaces with their status, provider, agents

# User runs: claude-vm chat --list  
# 1. Local CLI connects to current/selected workspace
# 2. Reads /workspace/.claude-vm/conversations.yaml from that workspace
# 3. Shows conversations specific to that workspace
```

**Session Recovery After Container Restart:**
```bash
# Container restart procedure (happens inside workspace):
# 1. Read /workspace/.claude-vm/conversations.yaml 
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
~/.claude-vm/conversations.yaml:
fix-auth-bug:
  agent: "claude"
  uuid_chain: ["abc123-original", "def456-resumed", "ghi789-current"]
  active_uuid: "ghi789-current"
  created: "2025-08-11T04:13:07Z"
  last_message: "2025-08-11T06:45:22Z"
  message_count: 45

refactor-database:
  agent: "goose"
  uuid_chain: ["xyz789-original"]
  active_uuid: "xyz789-original"
  created: "2025-08-10T14:22:11Z"
  last_message: "2025-08-10T16:33:55Z"
  message_count: 12
```

**Name Resolution Process:**
1. User runs `claude-vm chat "fix-auth-bug"`
2. System looks up name in `conversations.yaml` 
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
Direct tmux session streaming to browser using xterm.js with WebSocket connections (tmux → pty → WebSocket → xterm.js).

## Phase 2: Custom Parsed Interface (JSONL + Mobile UI)

**Implementation Strategy:**
Parse JSONL files and present custom mobile-optimized interface with directory watching for new messages and conversation management APIs.

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
