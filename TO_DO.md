

I want to:
- give Claude a to-do-list; it can be a markdown file, with structured waves and individual items.
- give Claude access to a private GitHub repo; it must pull the code in
- Claude runs in the background, working on the codebase. It is not executing on my local machine, so I can lose internet / turn off my machine.
- I must review the changes Claude is making. What is helpful is (1) a diff-view of the changes so far (like filesystem view), and (2) a stream of Claude's output / thoughts.
- I must be able to talk to Claude; ask it questions, and tell it to do things, not included in the original PR. That way I can correct it mid-flight if needed.
- Claude must perform the desired changes, and then open a PR

Two interfaces for this:
- The terminal on my computer
- My phone (presumably on a website)

---

Possible solutions:

VM:
- Spin up a VM (Digital Ocean) ( will it already have docker installed?)
- Pull in a docker container, which has: (1) Ubuntu, (2) Claude Code, (3) git, (4) our custom API-server / client.
- Start the container, providing (1) github credentials, (2) claude code credentials.
- The container receives messages via API; it git-clones the repo, gets the TO_DO.md, and then prompts Claude to start working on it.
- The container streams Claude-output (and maybe filesystem state?) to the API.
- The API can send messages which go into the terminal somehow, so we can live-interact with claude?

Terminal Interaction:
- replace 'claude' with 'claude-vm'. I.e.,'claude-vm' to start a new session, or 'claude-vm --session=abc123' in order to connect to a previous session with the abc123 identifier; this session may currently be running, and it may be possible multiple terminals to connect to the same session at the same time (viewing the history and current stream).
- We would have some sort of web-client, which would connect via claude-vm command, and then provide a web-interface: giving you the stdout and stdin that you would get in a terminal typically.

Security Decision - Git Access:
- Use GitHub Apps instead of deploy keys or credential proxy
- Installation tokens expire in 1 hour (minimal damage if leaked)
- Scoped permissions (contents, PRs, issues only - no admin/secrets)
- Per-repository installation (can't access other repos)
- Shows as "claude-vm[bot]" in GitHub activity
- This matches Anthropic's official approach and is more secure than PATs

---

CLI Commands (Updated Design):

- claude-vm up [repo-url|path] // Create and start new workspace
- claude-vm list // List all workspaces with status
- claude-vm ssh <workspace-id> // SSH into workspace (attaches to running Claude session)
- claude-vm stop <workspace-id> // Stop workspace (can resume later)
- claude-vm delete <workspace-id> // Delete workspace permanently
- claude-vm web <workspace-id> // Opens browser to workspace URL (persistent, OAuth-protected)

Examples:
- claude-vm up github.com/user/repo // Clone from GitHub
- claude-vm up . // Use current directory
- claude-vm up --provider docker // Use local Docker instead of cloud
- claude-vm ssh abc123 // Connects to Claude session, see full history via tmux

---

Github Action Idea:
- Use native Claude Code GitHub Actions (already exists!)
- Create issue with TO_DO list → Claude automatically creates PR
- No VM needed - runs on GitHub's infrastructure
- Monitor progress via GitHub PR comments and commits
- Interact by commenting on the PR (Claude responds)
- Built-in diff view, commit history, and PR review tools
- Limitations: Can't stream real-time thoughts, less interactive than VM

---

Sync Options (Local ↔ Remote VM):

1. Git-based (Non-realtime):
   - Commit locally → push to bridge → VM pulls → VM commits → push → local pulls
   - Pros: Full history, conflict resolution, familiar
   - Cons: Requires commits, not realtime, can pollute history

2. CRDT-based (Realtime):
   - Y.js, Automerge, or custom CRDT server
   - Files sync automatically without conflicts
   - Pros: Realtime, conflict-free, works offline
   - Cons: Complex, needs special server, may lose git history

3. Filesystem Sync Tools:
   - Syncthing: P2P, encrypted, automatic
   - Mutagen: Optimized for code, handles conflicts
   - Unison: Bidirectional, handles conflicts
   - Pros: Mature tools, efficient
   - Cons: Not git-aware, can cause conflicts

4. Mount-based (SSHFS/WebDAV):
   - Mount remote filesystem locally
   - Pros: Transparent, no sync needed
   - Cons: Latency, requires constant connection

5. Event Streaming:
   - File watcher + WebSocket + operational transforms
   - Similar to VS Code Live Share
   - Pros: Realtime, efficient
   - Cons: Complex to implement

6. Hybrid Git+Realtime:
   - Realtime sync for active editing
   - Git commits for checkpoints
   - Best of both worlds

7. Docker Volume Sync:
   - docker-sync, Docker Desktop sync
   - Pros: Docker-native
   - Cons: Docker-specific, performance issues

8. Git-State Sync (Recommended):
   - Instead of syncing files, sync git operations
   - Both local and remote execute same git commands
   - Maintains consistent git state on both sides
   - Implementation:
     * Git wrapper intercepts commands (checkout, commit, stash, etc)
     * Executes locally then remotely via SSH/WebSocket
     * Or: Watch .git/HEAD and .git/index for changes
     * Sync git operations, not file contents
   - Pros: 
     * No file sync conflicts
     * Git handles all file changes
     * Both sides always have same branch/commit state
     * Natural git workflow preserved
   - Cons:
     * Requires git wrapper or hooks
     * Need to handle simultaneous operations
   - Example flow:
     * Local: git checkout main → Remote: git checkout main
     * Remote: git commit → Local: git pull
     * Both stay in perfect sync

9. SSH Terminal + Mutagen Sync (New Recommendation):
   - claude-vm opens SSH session to remote VM
   - All commands (git, claude, etc) run on remote
   - Mutagen syncs files to local directory for editing
   - Remote VM is git master (has .git/)
   - Local has synced working directory only
   - Workflow:
     * `claude-vm` → SSH into VM
     * Type `git commit` normally (runs on VM)
     * Edit files locally, Mutagen syncs to VM
     * Natural git workflow preserved
   - Pros:
     * Most natural - just SSH + file sync
     * No command proxying needed
     * Full terminal access to VM
     * Edit files in local IDE
   - Example:
     ```
     $ claude-vm
     [vm]$ git status     # Runs on VM
     [vm]$ claude "fix auth"
     # Edit files locally, they sync to VM
     [vm]$ git commit -m "done"
     ```

Recommendation: Use SSH Terminal + Mutagen - simplest and most natural
