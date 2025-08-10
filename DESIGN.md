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

├── provider                                           # Provider management (CRUD)
│   ├── add <name>                                     # Add new provider
│   │   └── --option <key=value>                       #   Set provider options
│   │
│   ├── list                                           # List all configured providers
│   │   └── -v, --verbose                              #   Show all options (secrets hidden)
│   │
│   ├── set-options <name>                             # Update provider configuration
│   │   └── --option <key=value>                       #   Set/update options
│   │
│   └── delete <name>                                  # Remove provider
│       └── -y, --yes                                  #   Skip confirmation

├── workspace                                           # Workspace management
│   ├── up [workspace-id|repo-url|path/to/local-repo]   # Create and start new workspace
│   │   ├── --provider <name>                           #   Cloud provider (fly|aws|docker)
│   │   ├── --image <name>                              #   Override container image
│   │   └── --branch <name>                             #   Git branch to checkout
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
├── web                                       # Open web interface (you can pick between all workspaces)
│
├── chat                                      # Talk to coding agents
│   ├── -w, --workspace <workspace-id>        #   Interactive: list/select conversations in workspace
│   └── -c, --conversation <conversation-id>  #   Direct: connect to specific conversation
│
└── new-chat <workspace-id>                   # Create new conversation with coding agent
    └── --agent <name>                        #   Coding agent: claude, qwen, codex, goose (default: claude)

EXAMPLES:
  # Provider management
  claude-vm provider add digitalocean         # Interactive setup
  claude-vm provider add fly \
    --option token=fo1_xxx \
    --option organization=my-org
  claude-vm provider list                     # Show all providers
  claude-vm provider list -v                  # Show with options (secrets hidden)
  claude-vm provider set-options digitalocean \
    --option region=sfo3 \
    --option droplet_size=s-4vcpu-8gb
  claude-vm provider delete aws               # Remove AWS provider
  
  # Workspace management
  claude-vm workspace up bold-fire-1234       # restarts an existing workspace on fly.io
  claude-vm workspace up . --image python:3.11
  claude-vm workspace list --status running --long
  claude-vm workspace down 3f2504e0bb11        # Stop workspace, preserve state
  claude-vm workspace up 3f2504e0bb11          # Resume stopped workspace
  claude-vm workspace delete 45678901 --yes
  claude-vm workspace delete quiet-lake-5678 --purge    # Complete deletion including S3 backup
  
  # SSH access
  claude-vm ssh quiet-lake-5678 --user developer
  claude-vm ssh 8a9b2c3d4e5f --port 2222
  
  # Web interface
  claude-vm web                                # Open web interface (pick between workspaces)
  
  # Chat with coding agents
  claude-vm chat                               # Interactive: pick workspace → pick conversation
  claude-vm chat -w bold-fire-1234             # Interactive: pick conversation in workspace bold-fire-1234
  claude-vm chat -c conv-456                   # Direct: connect to conversation conv-456
  
  # Create new conversations
  claude-vm new-chat 3f2504e0bb11              # Create new conversation with default claude agent
  claude-vm new-chat 87654321 --agent qwen     # Create new conversation with qwen agent

GLOBAL OPTIONS:
  --provider fly|aws|docker     # Default cloud provider
  --config <file>               # Config file path
  --help, -h                    # Show help
```

---



### Volume Persistence

All Providers allow us to mount volumes; we mount them in /workspace. We can stop a container, and then start it again with the same volume preserving all data (such as edited code).

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

### Building Devcontainer

*Parsing User's Devcontainer.json:*
- Users should supply a devcontainer.json optionally, but we should also be able to generate a good starter image by parsing the repo.

*Inserting Processes into Devcontainer:*

**Agent Installation via Devcontainer Features:**
```json
// Generated devcontainer.json based on config.yaml
{
  "name": "claude-vm-workspace",
  "image": "mcr.microsoft.com/devcontainers/universal:2-linux",
  "features": {
    // Install all enabled agents from config
    "ghcr.io/anthropics/devcontainer-features/claude-code:latest": {},
    "ghcr.io/qwen/devcontainer-features/qwen-coder:latest": {},
    "ghcr.io/openai/devcontainer-features/codex:latest": {},
    "ghcr.io/goose-ai/devcontainer-features/goose:latest": {},
    "ghcr.io/google/devcontainer-features/gemini:latest": {},
    
    // Our custom feature for workspace management
    "ghcr.io/claude-vm/devcontainer-features/workspace-manager:latest": {
      "defaultAgent": "claude",
      "enabledAgents": ["claude", "qwen", "codex", "goose", "gemini"],
      "workspaceId": "${localEnv:WORKSPACE_ID}"
    }
  }
}
```

**How Multi-Agent Installation Works:**

1. **Config-driven**: User's `config.yaml` lists all desired agents:
   ```yaml
   agents:
     enabled: [claude, qwen, codex, goose, gemini]
     defaults:
       agent: claude  # Default for new-chat
   ```

2. **Parallel Installation**: All agents installed via devcontainer features:
   - Each agent gets its own binary/runtime
   - Credentials injected as environment variables
   - No conflicts between agents

3. **Runtime Selection**: User can switch agents anytime:
   ```bash
   claude-vm new-chat workspace-123           # Uses default (claude)
   claude-vm new-chat workspace-123 --agent qwen     # Use Qwen
   claude-vm new-chat workspace-123 --agent gemini   # Use Gemini
   ```

4. **Credential Injection**: Each agent gets its own auth:
   ```bash
   CLAUDE_CREDENTIALS=/workspace/.claude/.credentials.json
   OPENAI_API_KEY=sk-xxx        # For codex
   GOOGLE_API_KEY=xxx            # For gemini
   QWEN_API_KEY=xxx              # For qwen
   GOOSE_API_KEY=xxx             # For goose
   ```

5. **Benefits**:
   - **Compare models**: Test same task across different agents
   - **Fallback options**: If one agent fails, try another
   - **Specialized tasks**: Use best agent for specific jobs
   - **Cost optimization**: Use cheaper models when appropriate

### Provider Orchestration

(Devpod already has this code all written; we do not need to write it again.)

### Task Orchestration

We need both SSH and Web Access into every workspace (container).

Container:
- Claude code runs in a named tmux session
- Claude output is written to local log file (so SSH users can see)
- Log Trailer process sends changes (claude log and file changes) to S3 for persistence
- API Server: Container exposes a REST API + SSE endpoint for web access. The web can send commands via the REST API, and observe responses via SSE.

API Server:
- (needs to be specified)

Centralized Task Server:
- (needs to be specified)

S3 Storage:
- (needs to be specified)

Security:
- (how do we allow ourselves to login while protecting against others?)

### Workspace Lifecycle

- How long to keep containers alive after claude?

(???)

### Web UI

(???)

---

### Task Specification

(open issue)

### Code Review

(open issue)

---

### Future Work

We may want to add a `claude-vm backup (export?)` and `claude-vm restore (import?)` command which uses tarball files, so that users can manually save and restore devcontainer state.
