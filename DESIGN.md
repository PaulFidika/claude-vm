# Claude VM Design Document

## Goal
A system that can run Claude Code (or any coding agent) in a remote VM, so that we the developer can turn off their laptop / go offline and Claude can still continue on its own. The developer should be able to (1) view and modify remote-VM files on their local machine, and (2) talk to Claude directly. This allows the developer to scale Claude out to many machines at once, and let the developer supervise its work.

We aim for it to be a CLI tool for developers, with a mobile-compatible web-interface for supervising AI.

We need to deal with two separate problems:

1. managing dev-environment lifecycles
2. allocating work within the dev-environments, along with collecting results


### Design and Monetization

Should the user be operating at the workspace level, or the task level? Probably the task level (which is more abstract), with the option to manage workspaces as needed. For simplicity, we will have a rule of `one coding agent + task per workspace`. We want to be a _task orchestrator_ (new), not a _workspace orchestrator_ (like DevPod).

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
- Provider Orchestration: launching containers + volumes on local or remote providers
- Task Lifecycle: giving tasks to workspaces, keeping track of progress (Claude Code Logs), recording results in S3.
- Workspace Lifecycle: stopping containers when the task is complete.
- UI: mobile UI + remote VSCode for supervising / talking to claude. Terminal (ssh) access to workspace.

Provider scope:
- Infrastructure lifecycle (VM, network)
- Volume management

---

### OpenAI Codex notes:
- Turns off the container immediately after the task is completed, meaning the continaer needs to be started up again for every user-interaction.
- Takes about 60 seconds to startup the container again; runs a bunch of installs, meaning the dev environment (like node_modules) is not being cached; only file changes are.
- Does not support queued-messages yet. You literally cannot talk to it while it's working; you have to wait for it to finish.
- For each session it really only stores whatever file chanes are tracked in git. Non-git changes are not tracked or displayed.
- Honestly this tool sucks and is barely functional.

---

### Claude Code UIs

The main point is to allow you to use Claude Code outside of the CLI.

- getAsterisk/Claudia: 11k stars. Rust and Typescript. A Tauri-based desktop app

- siteboon/claudecodeui: react-based, supports both desktop and mobile web views. GPL license.

- sugyan/claude-code-webui: react based, not quite as sexy as Claude-Code-UI (in my opnion). MIT license

- wbopan/cui: react-based; UI is a complete clone of OpenAI's codex. Apache license. My favorite so far. This is not just a UI; it also orchestrates Claude instances that run locally. There is no isolation between Claude Code instances (they all work in the same git branch). All claude code instances run locally, although the server itself can be viewed remotely.

---

### Claude Code Containers:

The maint point of these is to make it easier to run claude-code with the --dangerously-skip-permissions flag, so you do not need to manually approve stuff. Running claude code in a container, with its own copy of the code, means that claude cannot destroy the codebase (or your computer) easily. Another concern is exfiltration / prompt injection hacks caused by rogue websites.

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

### Marketing:

- Submit commits to the 2 main claude-awesome repos to have our project listed as well
https://github.com/hesreallyhim/awesome-claude-code
https://github.com/jqueryscript/awesome-claude-code

---

### DevPod notes:
- stores a list of your running workspaces locally, and queries providers to get their status.
- DevPod Pro stores a list of your providers in your DevPod account.

---

## CLI Design

```bash
claude-vm - Run Claude Code on remote VMs

USAGE:
  claude-vm [command] [options]

COMMANDS:
  up [repo-url|path]        Create and start a new workspace (with or without a devcontainer.json)
  list                      List all workspaces with status
  ssh <workspace-id>        SSH into workspace container
  stop <workspace-id>       Stop workspace (can resume later)
  delete <workspace-id>     Delete workspace permanently
  web <workspace-id>        Open browser to workspace URL

OPTIONS:
  --provider fly|aws|docker  Cloud provider (default: docker)
  --image <image-name>       Override container image
  -y, --yes                  Skip confirmation prompts
  --help, -h                 Show help

COMMAND-SPECIFIC FLAGS:
  up:
    --provider               Choose cloud provider (default: docker)
    --image                  Override devcontainer.json image
    --public-web             Make web interface publicly accessible
  
  list:
    -l, --long               Show detailed workspace info
  
  delete:
    -y, --yes                Skip deletion confirmation

EXAMPLES:
  claude-vm up github.com/user/repo       # Create a remote workspace from a GitHub repo
  claude-vm up .                          # Create a renite workspace from the current directory
  claude-vm up --provider local           # Use local Docker container instead of a remote workspace
  claude-vm ssh abc123                    # SSH into workspace abc123
  claude-vm web abc123                    # View the workspace abc123 in a web browser (https://abc123.fly.io)
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

*Parsing Devcontainer.json:*
- Devpod can parse these container specs, but can also work without them by looking through the codebase.
- A user-supplied devcontainer.json spec should be optional.

*Adding Coding Agent into Devcontainer:*
- (For now, we can just use the official devcontainer feature, but we may want to build something more secure / custom.)
- (For now, we can just support Claude Code, but in the future we will want to suppor other agents.)

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

Workspace Server:
- 




### Future Work

We may want to add a `claude-vm backup (export?)` and `claude-vm restore (import?)` command which uses tarball files, so that users can manually save and restore devcontainer state.
