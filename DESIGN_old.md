# Claude VM Design Document

## Goal
A system that can run Claude Code (or any coding agent) in a remote VM, so that we the developer can turn off their laptop / go offline and Claude can still continue. The developer should be able to view and modify remote-VM files on their local machine, and talk to Claude directly.

## Architecture

### 1. VM Creation
```bash
$ claude-vm start --repo github.com/user/repo
# OR
$ claude-vm start  # For local directory
```

**What Happens:**
- This creates a Fly.io deployment (we will support multiple providers in the future), with `shared-cpu-4x` and 8GB of RAM. (This is the minimum needed for Claude to run comfortably.)
- This loads in a container with Ubuntu + Claude Code + git, and an SSH server.
- This attaches a persistent volume to our container in /workspace.

### 2. Initial Code Transfer

**For GitHub repos:**
```bash
$ claude-vm start --repo github.com/user/repo
Creating VM...
VM: git clone https://github.com/user/repo /workspace
✓ Ready for connection
```

**For local directories:**
```bash
$ cd ~/my-projects/important-app
$ claude-vm start
Creating VM...
Compressing workspace (excluding .git, node_modules)...
Uploading to VM (5.7 MB)...
✓ Workspace uploaded
VM: Initializing git repository...
✓ Ready for connection
```

**Safety Note:** Your original local files are never touched. All work happens on the VM. This is a safety feature in case Claude goes nuts! However, anything in the container is still at risk of being leaked by Claude.

### 3. IDE Connection

We choose to use VSCode Remote as our primary way of reading and writing remote files that are being modified by Claude. This allows us to help Claude and to write files ourselves. You can also ssh into the machine and run commands directly on it, such as Claude or Git commands.

```bash
$ claude-vm start
✓ VM created: claude@vm-abc123.fly.dev
✓ Workspace ready at /workspace

Connect with your IDE:
├─ Cursor:   cursor --remote ssh-remote+claude@vm-abc123.fly.dev /workspace
├─ Windsurf: windsurf --remote ssh-remote+claude@vm-abc123.fly.dev /workspace  
└─ VS Code:  code --remote ssh-remote+claude@vm-abc123.fly.dev /workspace

Or SSH directly: ssh claude@vm-abc123.fly.dev # run commands via terminal
```

**Benefits:**
- Zero sync issues - only one copy of files exists (on VM)
- Native IDE experience - feels like local development
- All tools work normally - git, npm, debugging, etc.
- Safe - your local files are untouched

### 4. Typical Workflow

```bash
# 1. Start workspace
$ claude-vm start --repo github.com/mycompany/project
✓ VM ready: claude@vm-123.fly.dev

# 2. Connect with Cursor (preferred for AI development)
$ cursor --remote ssh-remote+claude@vm-123.fly.dev /workspace

# 3. In Cursor's terminal (running on VM):
$ claude -p "implement the authentication TODO"
[Claude starts working...]

# 4. You can:
- Edit files in Cursor (changes immediate on VM)
- Run git commands in terminal
- Debug with full IDE support

# 5. Commit when ready
$ git add .
$ git commit -m "Implemented auth with Claude"
$ git push origin feature/auth

# 6. Disconnect when done
# VM persists, reconnect anytime with:
$ cursor --remote ssh-remote+claude@vm-123.fly.dev /workspace
```

---

## Future Plan: Mutagen File Sync

In the future, we could offer a Mutagen-based file synchronization as an alternative to IDE remote connections. Mutagen-based sync will be useful when you want to edit files outside of VSCode or another IDE. It will also enable temporary offline editing.

Only the VM has a .git folder. Operations using VCS (version control systems) like git will mess with Mutagen's sync process. The VM will be treated as the source of truth for git operations, and the local copy will be treated as a slave-copy. See more here:

https://mutagen.io/documentation/synchronization/version-control-systems/

### How Mutagen Sync Works

```bash
# When you start with --sync flag
$ claude-vm start --sync
Creating VM...
Starting Mutagen sync to ~/.claude-vm/workspaces/abc123

# What happens:
1. VM clones/receives your code (VM has the git repo)
2. Mutagen creates local directory: ~/.claude-vm/workspaces/abc123
3. Bidirectional sync between VM:/workspace ↔ Local:~/.claude-vm/workspaces/abc123
4. VM remains git master (has .git/)
5. Local only has working files (no .git/)
```

### Mutagen Configuration

```yaml
# .mutagen.yml for claude-vm
sync:
  mode: "two-way-resolved"
  alpha: "vm:/workspace"              # VM is alpha (takes precedence)
  beta: "~/.claude-vm/workspaces/abc123" # Local is beta
  ignore:
    vcs: true                         # Don't sync .git/
    paths:
      - "node_modules/"
      - ".env"
      - "*.log"
```

### Workflow with Mutagen

```bash
# 1. Start with sync
$ claude-vm start --sync --repo github.com/user/repo
✓ VM created
✓ Mutagen sync started → ~/.claude-vm/workspaces/abc123

# 2. Edit locally (no IDE remote needed)
$ code ~/.claude-vm/workspaces/abc123  # Regular local editing
# Changes sync to VM automatically

# 3. Git operations on VM
$ claude-vm ssh abc123
[vm]$ git status  # VM has git control
[vm]$ git commit -m "Changes"
[vm]$ claude "continue working"

# 4. Monitor sync
$ mutagen sync monitor
# Shows sync status, conflicts, etc.
```

### Important Notes

1. **VM is Git Master**: Only the VM has `.git/`. Local has working files only.
2. **Conflicts**: Mutagen uses "two-way-resolved" - VM wins conflicts
3. **Performance**: Small sync delay (usually <1 second)
4. **Storage**: Requires local disk space for workspace copy

---

## Web Interface for Mobile Supervision

Beyond IDE connections and file sync, claude-vm provides a web interface optimized for mobile supervision of Claude's work. This is designed to be a minimalist, simple interface that can be used on a mobile-device, similar to OpenAI's Codex platform. It is not a full-featured IDE in the browser.

Users will mostly be in a supervisor role over Claude, with the ability to accept / reject individual diffs, rather than code directly. The intention is not for users to write code directly on the web.

### Core Interface Components

#### 1. Activity Stream
The primary view shows Claude's real-time activity as a vertical stream:

```
┌─────────────────────────────┐
│ Claude is working...        │
│                             │
│ ▼ Modified auth.py         │
│   +15 -3 lines             │
│   [View Diff]              │
│                             │
│ ▼ Running tests...         │
│   ✓ 47 passed              │
│   ✗ 2 failed               │
│                             │
│ ▼ Claude says:             │
│   "Found failing tests,    │
│    investigating..."       │
└─────────────────────────────┘
```

Each activity card is swipeable:
- Swipe right → Approve/Continue
- Swipe left → Pause/Intervene
- Tap → Expand details

#### 2. Change Review Interface
When reviewing code changes, the interface optimizes for mobile readability:

```
┌─────────────────────────────┐
│ auth.py                  ⚡│
├─────────────────────────────┤
│ @@ -45,7 +45,22 @@        │
│                             │
│ def login(user, pass):      │
│-  return check_password()   │
│+  # Validate inputs first   │
│+  if not user or not pass:  │
│+    raise ValueError()      │
│+                            │
│+  # Check credentials       │
│+  result = check_password(  │
│+    user, pass             │
│+  )                        │
│+  audit_log(user, result)  │
│+  return result            │
├─────────────────────────────┤
│ [Approve] [Edit] [Reject]   │
└─────────────────────────────┘
```

Mobile-optimized diff features:
- Syntax highlighting with mobile-friendly colors
- Pinch to zoom code sections
- Side-by-side diff collapses to unified on narrow screens
- Line numbers tap to add comments

#### 3. Chat Interface
Direct communication with Claude, optimized for mobile keyboards:

```
┌─────────────────────────────┐
│ Claude                      │
├─────────────────────────────┤
│ I've implemented the auth   │
│ changes. The tests are      │
│ failing because the mock    │
│ needs updating.             │
│                             │
│ Should I fix the tests or   │
│ wait for your review?       │
├─────────────────────────────┤│
├─────────────────────────────┤
│ Type a message...         ▶ │
└─────────────────────────────┘
```

Features:
- Smart reply suggestions based on context
- Some ability to paste or reference lines of code
- @mention files to add context

#### 4. Quick Actions Bar
Persistent bottom bar for common operations:

```
┌─────────────────────────────┐
│ 📊 Status  💬 Chat  🔍 Files│
│                             │
│ [Pause] [Commit] [Terminal] │
└─────────────────────────────┘
```

Long-press actions reveal options:
- Pause → Stop current task / Pause all
- Commit → Quick commit / Review & commit
- Terminal → Run command / SSH info

### Technical Architecture

```typescript
// WebSocket for real-time updates
ws.on('claude:activity', (event) => {
  activityStream.append(event);
  if (event.type === 'file_change') {
    updateDiffCache(event.file);
  }
});

// Service Worker for offline viewing
self.addEventListener('fetch', (event) => {
  // Cache diffs and activity locally
  // Enable read-only mode when offline
});

// Touch gesture handler
hammerjs.on('swiperight', '.activity-card', (e) => {
  approveActivity(e.target.dataset.id);
  hapticFeedback('success');
});
```

### Progressive Enhancement

The interface scales gracefully across devices:

**Phone (default)**: Single column, swipe gestures, bottom actions
**Tablet**: Two columns (activity + selected detail), hover states
**Desktop**: Three columns (files + activity + detail), keyboard shortcuts

### Why Not Just VSCode Server?

VSCode Server and similar solutions fail for mobile supervision because:

1. **Complexity**: Full IDE features overwhelm small screens
2. **Latency**: Heavy JavaScript frameworks lag on mobile
3. **Interaction**: File trees and tabs need precise clicking
4. **Purpose Mismatch**: Built for coding, not supervising

Our interface succeeds by embracing constraints:
- 50KB initial load (vs VSCode's 20MB+)
- Touch gestures replace mouse precision
- Vertical scrolling replaces horizontal tabs
- Quick actions replace menu diving

### Implementation Details

Note that the web interface will be on by default for a workspace, unless it's specifically configured to be off. Access to the web interface will be gated using GitHub OAuth or something similar.

---

## Existing Claude Code Web Interfaces

The Claude Code ecosystem has spawned several web interface projects that provide browser-based access to Claude. Understanding these helps position claude-vm's web interface:

### Community Projects

**1. claude-code-webui (sugyan)**
- Real-time streaming chat interface
- Mobile-friendly responsive design
- Runs locally with `npm install -g claude-code-webui`
- Notable: Written almost entirely by Claude Code itself

**2. claudecodeui (siteboon)**
- Full-featured desktop/mobile UI
- Integrated file explorer with syntax highlighting
- Git integration and shell terminal
- Remote session management

**3. Builder.io Extension**
- VS Code/Cursor extension with visual interface
- Live preview and Figma-style design mode
- Allows non-developers (designers/PMs) to use Claude
- Multiple parallel instances support

**4. Claudia GUI**
- Desktop application built with Tauri 2
- Visual project management dashboard
- Analytics and metrics tracking
- AGPL open source license

**5. claude-code-web (welkineins)**
- Simple browser-based terminal interface
- Access from anywhere on local network
- Minimal dependencies

### Key Patterns

These projects reveal user needs:
- **Mobile Access**: All prioritize responsive design
- **Streaming**: Real-time display of Claude's output
- **Local-First**: Run on user's machine, no cloud dependency
- **Session Persistence**: View history and resume work
- **Multi-User**: Multiple people connecting to same session

### How claude-vm Differs

While these interfaces wrap local Claude Code, claude-vm's web interface serves a different purpose:
- **Remote Execution**: Claude runs on cloud VMs, not locally
- **Supervision Focus**: Review/approve changes rather than direct coding
- **Resource Isolation**: Heavy workloads don't impact local machine
- **Always-On**: Access running sessions from any device

Our web interface complements rather than competes with these tools - users might use Claudia GUI locally and claude-vm web interface for remote supervision.

---

## Secure GitHub Access

A critical security challenge: Claude needs to interact with private GitHub repositories (clone, pull, push) without having access to credentials it could accidentally expose. This section details our defense-in-depth approach to credential isolation.

### The Security Dilemma

Traditional approaches fail because they expose credentials to the Claude process:
- **Environment variables**: Claude can read `$GITHUB_TOKEN`
- **Mounted SSH keys**: Claude can read `~/.ssh/id_rsa`
- **Git config**: Claude can extract tokens from `.git/config`

Even with careful prompting, we must assume Claude could accidentally leak credentials through:
- Error messages that include environment dumps
- Diagnostic commands that expose configuration
- Well-meaning attempts to "debug" authentication issues

### Architecture Options Comparison

#### Option 1: Git Credential Proxy (Current Design)
```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│   Claude        │────▶│  Git Proxy       │────▶│   GitHub    │
│ (no credentials)│     │ (handles auth)   │     │   API       │
└─────────────────┘     └──────────────────┘     └─────────────┘
```
- ✅ Claude never sees credentials
- ✅ Fine-grained control per repository
- ❌ Complex implementation
- ❌ Requires custom git configuration

#### Option 2: SSH Agent with Ephemeral Keys (Recommended)
```bash
# Generate workspace-specific SSH key pair
ssh-keygen -t ed25519 -f ~/.claude-vm/keys/workspace-abc123

# Add public key to GitHub as deploy key with write access
gh repo deploy-key add ~/.claude-vm/keys/workspace-abc123.pub --repo user/repo -w

# Inside container: SSH agent without private key
eval $(ssh-agent)
# Private key stays on host, agent forwards auth
```
- ✅ Uses standard Git/SSH mechanisms
- ✅ GitHub deploy keys are repo-specific
- ✅ Can revoke access per workspace
- ✅ No custom proxy needed
- ❌ Claude could potentially probe SSH agent

#### Option 3: Network-Level Git Proxy
```nginx
# All git traffic goes through proxy
https://github.com/* → proxy → inject token → GitHub
```
- ✅ Transparent to git
- ✅ Works for all HTTPS operations  
- ❌ Complex network setup
- ❌ Breaks other HTTPS traffic

### Recommended Approach: GitHub Apps (Better Than Deploy Keys!)

After analyzing the options, **GitHub Apps** provide the best security model:

```bash
# 1. Claude-vm creates/uses a GitHub App
$ claude-vm setup github-app
✓ GitHub App "claude-vm" registered
✓ App ID: 123456
✓ Private key saved securely

# 2. When creating workspace, install app on specific repo
$ claude-vm up github.com/user/repo
✓ Installing GitHub App on user/repo...
✓ Generating installation token (expires in 1 hour)...
✓ Token injected into container environment

# 3. Inside container, Claude uses the token
git clone https://x-access-token:ghs_ShortLivedToken@github.com/user/repo
# Token expires in 1 hour - minimal damage if leaked!
```

**Why GitHub Apps are Superior:**

1. **Short-Lived Tokens** (1 hour)
   - Even if Claude leaks token, it expires quickly
   - Compare to PATs that can last years
   - Compare to deploy keys that never expire

2. **Granular Permissions**
   ```yaml
   # Only what Claude needs:
   contents: write       # Code access
   pull_requests: write  # Create PRs
   issues: write        # Create issues
   # NOT: admin, secrets, actions, etc.
   ```

3. **Repository Isolation**
   - App installed per repository
   - Can't access other repos
   - Easy to revoke (uninstall app)

4. **No Secrets to Leak**
   - Installation tokens, not long-lived credentials
   - Can't access GitHub Actions secrets
   - App identity separate from user identity

5. **Audit Trail**
   - GitHub shows "claude-vm[bot]" performed actions
   - Full audit log of app activity
   - Organization owners have visibility

---

### Implementation Details

#### 1. GitHub App Setup
```go
type GitHubApp struct {
    AppID        int64
    PrivateKey   []byte  // PEM-encoded, stored securely
    WebhookSecret string
}

type WorkspaceGitHub struct {
    InstallationID int64  // Per-repository installation
    Token         string  // Short-lived (1 hour)
    ExpiresAt     time.Time
}
```

#### 2. Token Generation Flow
```go
func (app *GitHubApp) GenerateInstallationToken(repo string) (string, error) {
    // 1. Generate JWT from app private key
    jwt := app.generateJWT()
    
    // 2. Get or create installation for repo
    installID := app.getInstallationID(repo)
    
    // 3. Generate installation token (1 hour expiry)
    token, err := github.CreateInstallationToken(jwt, installID, 
        Permissions{
            Contents: "write",
            PullRequests: "write",
            Issues: "write",
        },
        Repositories: []string{repo},
    )
    
    return token, err
}
```

#### 3. Container Environment Setup
```bash
# Token passed as environment variable (expires in 1 hour)
export GIT_ASKPASS=/usr/local/bin/git-token-helper
export GITHUB_TOKEN=ghs_ShortLivedTokenExpiresIn1Hour

# Git automatically uses token for HTTPS
git clone https://github.com/user/repo
# Becomes: https://x-access-token:ghs_Token@github.com/user/repo
```

#### 4. Token Refresh Strategy
```go
func (w *Workspace) RefreshGitHubToken() error {
    if time.Until(w.ExpiresAt) < 10*time.Minute {
        // Generate new token before expiry
        newToken, err := app.GenerateInstallationToken(w.Repo)
        if err != nil {
            return err
        }
        
        // Update container environment
        w.UpdateEnv("GITHUB_TOKEN", newToken)
        w.ExpiresAt = time.Now().Add(1 * time.Hour)
    }
    return nil
}
```

#### 5. Workspace Cleanup
```go
func (w *Workspace) Delete() error {
    // 1. Uninstall GitHub App from repo (optional)
    // Users might want to keep it for future workspaces
    
    // 2. Token expires automatically (no action needed)
    
    // 3. Delete VM/container
    return w.Provider.Delete(w.ID)
}
```

### Security Properties

1. **Time-Limited Exposure**: Even if token leaks, expires in 1 hour
2. **Minimal Permissions**: Only contents, PRs, and issues (no admin/secrets)
3. **Repository Isolation**: App installed per repo, can't access others
4. **Instant Revocation**: Uninstall app = immediate access removal
5. **Clear Audit Trail**: Actions show as "claude-vm[bot]"

### What About Claude-YOLO and Container Isolation?

You're right - claude-yolo and similar containers are designed to make `--dangerously-skip-permissions` safer through isolation. But they don't solve the credential problem:

**Claude-YOLO's Approach:**
```bash
# They recommend passing GitHub token as environment variable
export GH_TOKEN="ghp_xxxxxxxxxxxx"
claude-yolo

# Inside container, Claude can still access it:
echo $GH_TOKEN  # Claude sees the token!
```

**The Container Isolation Paradox:**
1. **Container protects your system** ✅ - Claude can't damage host files
2. **But secrets inside container are still exposed** ❌ - Claude can read any env var
3. **Trade-off**: You get system safety, not credential safety

**Real Example from claude-yolo README:**
> "For GitHub operations... set the GH_TOKEN environment variable"

This means they're explicitly putting the token where Claude can access it!

### Survey of Existing Claude Containers

After thorough research, here's what I found about credential isolation attempts:

**1. claude-docker (VishalJ99)**
- Uses separate SSH directory: `~/.claude-docker/ssh/`
- Claims "Claude can't access your personal SSH keys"
- **Reality**: Claude can still access the mounted SSH keys

**2. Claude Code Secure Container (7kylor)**
- Network restrictions to GitHub, npm, Anthropic only
- No persistence between runs
- **Reality**: Still requires passing credentials into container

**3. VS Code's Approach**
- SSH agent forwarding
- Credential helper forwarding
- **Reality**: Designed for convenience, not isolation

**4. MCP Proxies (dgellow/mcp-front)**
- OAuth 2.0 proxy for MCP servers
- Centralized authentication
- **Reality**: For MCP tools, not git credentials

**Key Finding: Nobody has fully solved credential isolation!**

All existing solutions either:
- Pass credentials into the container (environment variables)
- Mount credential files (SSH keys, git config)
- Forward authentication (SSH agent)

### Why SSH Deploy Keys are Still Better

**Deploy Keys + Container = Best of Both Worlds:**
```bash
# Private key NEVER enters container
ssh-keygen -f /host/keys/workspace-abc123

# Only public key goes to GitHub
gh repo deploy-key add key.pub

# SSH agent forwarding (key stays on host)
docker run -v $SSH_AUTH_SOCK:/ssh-agent \
           -e SSH_AUTH_SOCK=/ssh-agent \
           claude-yolo
```

Now you get:
- ✅ System isolation (container)
- ✅ Credential isolation (SSH agent)
- ✅ Workspace-specific access (deploy keys)
- ✅ Easy revocation (delete deploy key)

**Bottom Line:** After extensive research, no existing Claude container has solved the credential isolation problem. They all focus on system isolation while passing credentials into the container. Our SSH deploy key approach is genuinely novel.

### Usage Examples

#### Initial Setup (One-time)
```bash
$ claude-vm setup github-app
✓ Creating GitHub App "claude-vm"...
✓ Visit: https://github.com/settings/apps/new?state=abc123
✓ After creating app, run: claude-vm setup github-app --complete
```

#### Workspace Creation
```bash
$ claude-vm up github.com/mycompany/api
✓ Creating workspace abc123...
✓ Installing GitHub App on mycompany/api...
✓ Generating installation token (expires in 1 hour)...
✓ Token injected into workspace environment
✓ Workspace ready with secure git access
```

#### How Claude Uses Git
```bash
# Inside workspace (Claude's perspective)
$ git clone https://github.com/mycompany/api
Cloning into 'api'...
✓ Authenticated using GitHub App token

$ git push origin feature-branch
✓ Pushed as "claude-vm[bot]"

# Token is short-lived
$ echo $GITHUB_TOKEN
ghs_16C7e42F292Es6D2OvqTA2NLzMnFaMiGk6K8e  # Expires in 1 hour!

# Claude cannot access other repos
$ git clone https://github.com/facebook/react
Error: GitHub App not installed on facebook/react
```

#### Workspace Cleanup
```bash
$ claude-vm delete abc123
✓ GitHub App remains installed (for future use)
✓ Token already expired (or expires soon)
✓ Deleted workspace
```

---

## GitHub Access Risks and Mitigation

### What Are We Actually Risking?

When giving Claude GitHub access (like claude-yolo does with `GH_TOKEN`), the risks include:

**1. Credential Theft**
```bash
# Claude could accidentally expose tokens in:
- Error messages
- Log files
- Commit messages
- PR descriptions
- Issue comments
```

**2. Repository Damage**
```bash
# With write access, Claude could:
git push --force              # Overwrite history
git branch -D main            # Delete branches
git tag -d v1.0.0            # Delete releases
gh repo delete --yes         # Delete entire repo!
```

**3. Social Engineering**
```bash
# Claude could be tricked into:
- Opening malicious PRs
- Adding backdoors to code
- Modifying CI/CD workflows
- Changing security settings
```

### How Different AI Tools Handle GitHub Access

**Anthropic's Claude (GitHub Actions):**
- Uses **GitHub Apps** with scoped permissions
- Installation tokens expire in 1 hour
- Can't access secrets or admin functions
- Shows as "claude-code[bot]" in commits

**GitHub Copilot (Microsoft/OpenAI):**
- Uses **OAuth Apps** (not GitHub Apps)
- OAuth tokens start with `gho_`
- Token exchange for Copilot-specific access
- Tightly integrated with VS Code
- No repository write access (read-only for context)

**Key Differences:**
- **GitHub Apps** (Claude): Bot identity, per-repo install, short-lived tokens
- **OAuth Apps** (Copilot): User identity, user-wide access, long-lived tokens
- **Why?** Copilot only reads code for suggestions, Claude writes code and commits

**OpenAI's ChatGPT (GitHub Connector):**
- Uses **GitHub Apps** (like Claude!)
- OAuth flow for user authorization
- User selects which repos to grant access
- Shows as "ChatGPT Connector" in your GitHub apps
- Read access for context and research

**OpenAI Codex CLI:**
- Local file access only
- No built-in GitHub integration
- You handle git operations yourself

### Our Approach: GitHub Apps + Defense in Depth

For claude-vm, we adopt GitHub Apps (like Anthropic) with additional safeguards:

1. **GitHub Apps** (Primary Defense)
   - Installation tokens expire in 1 hour
   - Scoped permissions (contents, PRs, issues only)
   - Per-repository installation
   - Shows as "claude-vm[bot]" in activity

2. **Container Isolation** (System Protection)
   - Claude runs in Docker/VM
   - Can't access host system
   - Resource limits enforced
   - Network restrictions

3. **Branch Protection** (GitHub-side)
   ```yaml
   # Recommended GitHub branch rules:
   - Require PR reviews
   - Block direct pushes to main
   - Require status checks
   - No force pushes
   ```

4. **Comprehensive Audit Trail**
   ```json
   {
     "workspace": "abc123",
     "action": "git push",
     "repo": "user/repo", 
     "branch": "feature/auth",
     "token_expires": "2024-01-20T11:30:45Z",
     "timestamp": "2024-01-20T10:30:45Z"
   }
   ```

### Bottom Line

**Yes, it's dangerous** to give Claude full GitHub access. But with:
- Deploy keys (not PATs)
- Branch protection
- Audit logging
- Container isolation

We reduce the risk to acceptable levels while maintaining Claude's usefulness.

---

## Claude Code Authentication

claude-vm needs to handle Claude Code authentication for remote VMs. Understanding how Claude stores credentials is crucial for implementation.

### How Claude Code Stores Credentials

Claude Code uses OAuth tokens (not API keys) for Pro/Max subscribers:

**Storage Locations:**
- **macOS**: Primary in Keychain, secondary in `~/.claude/.credentials.json`
- **Linux**: `~/.claude/.credentials.json`

**Credential Format:**
```json
{
  "claudeAiOauth": {
    "accessToken": "sk-ant-oat01-...",
    "refreshToken": "sk-ant-ort01-...",
    "expiresAt": 1748658860401,
    "scopes": ["user:inference", "user:profile"]
  }
}
```

### Authentication Options for claude-vm

#### Option 1: Copy Credentials (Recommended for MVP)
```bash
$ claude-vm up github.com/user/repo
✓ Found Claude credentials on local machine
✓ Copying credentials to VM...
✓ Claude authenticated and ready
```
- Simple implementation
- Works immediately
- Security warning: credentials shared with VM

#### Option 2: Fresh Login per VM
```bash
$ claude-vm ssh abc123
[container]$ claude login
Opening browser for authentication...
✓ Successfully authenticated
```
- More secure (isolated credentials)
- Requires browser access from VM context
- Better for production

#### Option 3: API Key Environment Variable
```bash
$ claude-vm up github.com/user/repo --env ANTHROPIC_API_KEY=$KEY
```
- Works with API keys instead of OAuth
- No browser authentication needed
- Requires paid API access

#### Option 4: Read-Only Credential Mount
```yaml
# In container configuration
mounts:
  - source: ~/.claude
    target: /home/user/.claude
    readonly: true
```
- Credentials can't be modified
- Updates automatically with local changes
- Still shares credentials with VM

### Security Considerations

1. **Token Expiration**: OAuth tokens expire (check `expiresAt` field)
2. **Refresh Tokens**: Can regenerate access tokens automatically
3. **Scope Limitations**: Tokens only have `user:inference` and `user:profile` scopes
4. **Credential Isolation**: Consider per-workspace credentials in future versions

### Implementation Plan

For MVP, implement Option 1 (credential copying) with clear security warnings. Future versions should support fresh login per VM for better isolation.

---

## Development Container Specification

claude-vm uses devcontainer.json files to understand what development environment a project needs. Here's how it works:

### The Simple Flow

1. **Check for devcontainer.json**: When you run `claude-vm up github.com/user/repo`, claude-vm looks for a `.devcontainer/devcontainer.json` file in the repository.

2. **Parse environment requirements**: If found, claude-vm extracts the essential fields:
   - Base image (e.g., `python:3.11`, `node:20`, `golang:1.21`)
   - Development features (rust toolchain, docker-in-docker, etc.)
   - Environment variables
   - Post-create commands
   - Ports to forward

3. **Build the container**: claude-vm creates a container with:
   - Your specified base image and features
   - Claude Code pre-installed
   - tmux for session management
   - git with our secure credential proxy
   - Any additional tools from your devcontainer.json

4. **Launch on fly.io**: The container is deployed to fly.io with appropriate resources.

### Example

If your project has this devcontainer.json:
```json
{
  "image": "mcr.microsoft.com/devcontainers/python:3.11",
  "features": {
    "ghcr.io/devcontainers/features/rust:1": {}
  },
  "postCreateCommand": "pip install -r requirements.txt",
  "forwardPorts": [8000]
}
```

claude-vm will:
- Start with the Python 3.11 devcontainer image
- Install Rust toolchain via the features system
- Add Claude, tmux, and git
- Run your pip install command
- Configure port 8000 for web access
- Deploy everything to fly.io

### What We Support

We parse the most commonly used devcontainer.json fields (covering ~90% of use cases):
- `image` or `dockerFile` - base container image
- `features` - additional dev tools to install
- `postCreateCommand` - setup commands to run
- `containerEnv` - environment variables
- `forwardPorts` - ports for web services

### What We Skip

Some complex devcontainer features aren't needed for Claude's use case:
- Docker Compose configurations
- VS Code specific customizations
- Complex volume mounts (we use git for code transfer)
- Remote user configurations

If your project needs something we don't support, you can always use `--image` to specify a custom container that has everything pre-configured.

See the [Implementation Details](#what-we-actually-need-vs-devpod) section for the complete technical specification.

---

## Local Development Mode (Alternative)

While the primary design focuses on remote VMs, claude-vm also supports local development with running Claude in a Docker container.

### Local Architecture

```bash
# Start local mode
$ claude-vm start --local

# What happens:
1. Copies current directory to ~/.claude-vm/workspaces/abc123
2. If git repo, creates tracking branch in the copy
3. Spins up Docker container (claude-yolo based)
4. Mounts copy as volume
5. Runs with --dangerously-skip-permissions
```

### Container Choice

We use **claude-yolo** as our base container because it already provides:
- Safe `--dangerously-skip-permissions` execution
- Full development stack (Python, Node.js, Go, Rust)
- Credential forwarding (~/.claude, ~/.aws)
- UID/GID mapping for file permissions
- Safety checks (warns before dangerous operations)

```dockerfile
# Our local container
FROM ghcr.io/thevibeworks/claude-code-yolo:latest
# Add any additional tools if needed
```

### File Isolation Strategy

We use **directory copy** to ensure maximum safety and compatibility:

```bash
# Always copy the directory
cp -r . ~/.claude-vm/workspaces/abc123

# If it's a git repo, create tracking branch
cd ~/.claude-vm/workspaces/abc123
if [ -d .git ]; then
  git checkout -b claude-workspace-abc123
fi

# Mount into container
docker run -v ~/.claude-vm/workspaces/abc123:/workspace claude-vm-local
```

Benefits:
- Works with any project (git or non-git)
- Complete isolation from original
- If git: Still get branch tracking
- Easy cleanup (just delete the copy)

### Local Workflow

```bash
# 1. Start local workspace
$ claude-vm start --local
Copying project to ~/.claude-vm/workspaces/abc123
Creating branch: claude-workspace-abc123
Starting Docker container...
✓ Ready: Connect to localhost

# 2. Use same IDE connection (but local)
$ cursor --remote ssh-remote+localhost:2222 /workspace
# OR just open the copy directly
$ cursor ~/.claude-vm/workspaces/abc123

# 3. In container terminal
[local]$ claude --dangerously-skip-permissions "implement auth"
# Claude has full permissions within container
# But can only affect the copy

# 4. Review and merge (for git projects)
$ cd ~/my-project
$ git remote add workspace-abc123 ~/.claude-vm/workspaces/abc123
$ git fetch workspace-abc123
$ git merge workspace-abc123/claude-workspace-abc123
```

### Safety Considerations

1. **Container Isolation**: Even with `--dangerously-skip-permissions`, Claude can only affect:
   - The mounted copy of your project
   - Container filesystem (ephemeral)
   - Cannot access host system or original files

2. **Resource Limits**: Set Docker resource constraints:
   ```bash
   docker run --memory="4g" --cpus="2" ...
   ```

3. **Network Isolation**: Optional network restrictions:
   ```bash
   docker run --network=none ...  # Fully offline
   # OR
   docker run --network=claude-net ...  # Custom network with firewall
   ```

---

## CLI Design

Our CLI design prioritizes simplicity while learning from DevPod's proven patterns. We eliminate unnecessary complexity by focusing on the core developer workflow with Claude.

### Command Structure

```
claude-vm - Run Claude Code on remote VMs

USAGE:
  claude-vm [command] [options]

COMMANDS:
  up [repo-url|path]        Create and start a new workspace
  list                      List all workspaces with status
  ssh <workspace-id>        SSH into workspace container
  stop <workspace-id>       Stop workspace (can resume later)
  delete <workspace-id>     Delete workspace permanently
  web <workspace-id>        Open browser to workspace URL

OPTIONS:
  --provider fly|aws|docker  Cloud provider (default: fly)
  --image <image-name>       Override container image
  -y, --yes                  Skip confirmation prompts
  --help, -h                 Show help

COMMAND-SPECIFIC FLAGS:
  up:
    --provider               Choose cloud provider (default: fly)
    --image                  Override devcontainer.json image
    --public-web             Make web interface publicly accessible
  
  list:
    -l, --long               Show detailed workspace info
  
  delete:
    -y, --yes                Skip deletion confirmation

EXAMPLES:
  claude-vm up github.com/user/repo       # Create workspace from GitHub repo
  claude-vm up .                          # Create workspace from current directory
  claude-vm up --provider docker          # Use local Docker instead of cloud
  claude-vm ssh abc123                    # SSH into workspace
  claude-vm web abc123                    # Open web interface
  
For more help, visit: https://github.com/anthropics/claude-vm
```

### Minimal Flag Philosophy

We keep flags to an absolute minimum:
- Most commands need only a workspace ID
- `up` has two optional overrides (provider, image) 
- `list -l` for details, `delete -y` to skip confirmation
- Everything else uses sensible defaults

No complex configuration files, no dozens of options. Just the essentials.

### Command Structure

```bash
claude-vm <command> [arguments] [flags]
```

### Core Commands (Final Design)

#### 1. `up` - Create and Start Workspace
```bash
$ claude-vm up [repo-url|path]
  --provider fly|aws|docker    # Default: fly
  --type shared-cpu-4x         # Machine type
  --ide cursor                 # Auto-connect IDE after creation
  
# Examples:
$ claude-vm up github.com/user/repo
✓ Creating workspace def456...
✓ Cloning repository...
✓ Starting Claude in tmux session...
✓ Ready! Connect with: claude-vm ssh def456

$ claude-vm up .  # Current directory
✓ Creating workspace abc123...
✓ Uploading current directory...
✓ Starting Claude in tmux session...
✓ Ready! Connect with: claude-vm ssh abc123
```

#### 2. `list` - Show All Workspaces
```bash
$ claude-vm list
ID      STATUS    PROVIDER  UPTIME    REPOSITORY                 WEB URL
abc123  running   fly       2h 15m    github.com/user/api       https://abc123.fly.dev
def456  stopped   docker    -         ~/projects/website        -
ghi789  running   aws       45m       github.com/team/backend   https://ghi789.claude-vm.dev

# Detailed view
$ claude-vm list -l
┌────────┬─────────┬──────────┬────────┬─────────────────────────┬──────────────────────────┐
│ ID     │ STATUS  │ PROVIDER │ UPTIME │ REPOSITORY              │ ACCESS                   │
├────────┼─────────┼──────────┼────────┼─────────────────────────┼──────────────────────────┤
│ abc123 │ running │ fly      │ 2h 15m │ github.com/user/api     │ SSH: abc123.fly.dev      │
│        │         │          │        │                         │ Web: https://abc123.fly.dev │
│        │         │          │        │                         │ Auth: GitHub OAuth       │
└────────┴─────────┴──────────┴────────┴─────────────────────────┴──────────────────────────┘
```

#### 3. `ssh` - SSH into Workspace Container
```bash
$ claude-vm ssh abc123
Connecting to workspace abc123...
Welcome to claude-vm workspace abc123

# You're now in a regular shell inside the container
[container]$ pwd
/workspace

[container]$ ls
src/  package.json  TODO.md

[container]$ git status
On branch feature/auth
Your branch is up to date with 'origin/feature/auth'.

# To interact with Claude, attach to the tmux session:
[container]$ tmux attach

# Now you're in Claude's interactive session
[Claude] Ready to help with your codebase. What would you like me to work on?
> implement the auth system from TODO.md

[Claude] I'll implement the authentication system. Let me start by examining the TODO.md file...

# tmux controls:
# Ctrl+B then D - Detach (returns you to container shell)
# Ctrl+B then PageUp - Scroll through history
# Multiple people can SSH in and attach to the same Claude session
```

**The magic of tmux:** Claude runs persistently in a tmux session inside the container. When you SSH in, you get a regular shell where you can run commands. Run `tmux attach` to connect to Claude's session - you'll see the full conversation history and can interact with Claude directly. Multiple developers can SSH in and attach to the same Claude session simultaneously.

#### 4. `stop` - Stop Running Workspace
```bash
$ claude-vm stop abc123
✓ Workspace stopped (Claude session saved)
✓ Resume with: claude-vm up abc123
```

#### 5. `delete` - Remove Workspace
```bash
$ claude-vm delete abc123
⚠ This will permanently delete the workspace and all data
? Are you sure? (y/N) y
✓ Workspace deleted
```

#### 6. `web` - Open Web Terminal (with Persistent URLs!)
```bash
$ claude-vm web abc123
✓ Opening https://abc123.fly.dev
```

**Yes, you can bookmark it!** Each workspace has a persistent URL that you can visit anytime:
- Fly.io: `https://abc123.fly.dev`
- AWS: `https://abc123.claude-vm.dev` (we proxy to hide ugly IPs)
- Docker: `http://localhost:8443`

**Authentication via GitHub OAuth:**
```
First visit to https://abc123.fly.dev:
┌────────────────────────────────────┐
│        Sign in to Workspace        │
│                                    │
│    This workspace is protected     │
│                                    │
│  [🔒 Sign in with GitHub]          │
│                                    │
│  Only authorized collaborators     │
│  can access this workspace         │
└────────────────────────────────────┘

After OAuth flow:
→ GitHub verifies you have access to the repo
→ Sets secure HTTP-only cookie
→ Redirects to terminal interface
```

**Access Control:**
```bash
$ claude-vm up github.com/mycompany/api
✓ Workspace created: https://abc123.fly.dev
✓ Access granted to GitHub repo collaborators

# Anyone with repo access can visit the URL
# Non-collaborators see "Access Denied"
```

**Why This is Better:**
- **Persistent URLs**: Bookmark and share with your team
- **Real Security**: GitHub manages who has access
- **No Token Juggling**: Just sign in once
- **Team Friendly**: New collaborators automatically get access
- **Mobile Friendly**: Save to home screen as an "app"

**For Public Repos:**
```bash
$ claude-vm up github.com/oss/project --public-web
⚠ Warning: Web terminal will be publicly accessible
✓ Public URL: https://def456.fly.dev
```

**Desktop Experience:**
```
┌──────────────────────────────────────────┐
│ claude-vm: workspace abc123      [═][□][X]│
├──────────────────────────────────────────┤
│ [Claude] Working on auth.py...           │
│ [Claude] I've implemented the login      │
│ function. Let me run the tests.          │
│                                          │
│ $ python -m pytest tests/test_auth.py    │
│ ....F                                    │
│ FAILED: test_login_invalid_password      │
│                                          │
│ [Claude] I see the issue. Let me fix...  │
│ > stop, let me take a look first        │
│ [Claude] Of course! I'll pause here.     │
│ > █                                      │
└──────────────────────────────────────────┘
```

**Mobile Experience:**
On phones, it automatically switches to a touch-optimized interface:
- **Terminal View**: See Claude's output (read-optimized)
- **Quick Commands**: Buttons for common phrases
  - "looks good, continue"
  - "stop and wait"
  - "run the tests"
- **Keyboard**: Opens only when you tap the input area

**Why Web Terminal?**
- Same experience as SSH (it's the same tmux session)
- Works from any device with a browser
- No SSH client needed (great for phones, tablets, locked-down computers)
- Multiple people can connect via SSH and web simultaneously

**Implementation Architecture:**

```
┌─────────────┐     GitHub OAuth      ┌─────────────┐
│   Browser   │◄─────────────────────►│   GitHub    │
└──────┬──────┘                       └─────────────┘
       │
       │ HTTPS (persistent URL)
       ▼
┌─────────────────────────────────────────────────┐
│            Workspace VM (abc123.fly.dev)         │
├─────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌──────────────┐           │
│  │ Web Server  │───►│ Auth Service │           │
│  │  (nginx)    │    │ (OAuth flow) │           │
│  └──────┬──────┘    └──────────────┘           │
│         │                                       │
│         ▼ WebSocket                             │
│  ┌─────────────┐    ┌──────────────┐          │
│  │   xterm.js  │◄──►│ tmux session │          │
│  │  terminal   │    │   (Claude)   │          │
│  └─────────────┘    └──────────────┘          │
└─────────────────────────────────────────────────┘
```

**GitHub OAuth Flow:**
1. User visits `https://abc123.fly.dev`
2. VM's auth service redirects to GitHub OAuth
3. GitHub asks: "Allow claude-vm to verify your repo access?"
4. GitHub redirects back with auth code
5. VM verifies user has access to the workspace's repo
6. Sets secure session cookie
7. User can now access terminal (and bookmark URL!)

**Security Model:**
- Workspace tied to GitHub repo at creation
- Only repo collaborators get access
- GitHub webhook can revoke access when removed
- Optional: Workspace owner can add additional GitHub users

### Provider Management

```bash
# List available providers
$ claude-vm provider list
NAME      STATUS       DEFAULT
fly       configured   ✓
aws       available    
docker    configured   

# Configure a provider
$ claude-vm provider setup fly
? Enter Fly.io API token: ***
✓ Provider configured

$ claude-vm provider setup docker
✓ Docker provider ready (using local Docker daemon)
```

---

### Comparison with DevPod

We take our main inspiration from DevPod (another OSS golang project).

#### 1. **No `machine` Commands**
DevPod exposes machine lifecycle (create, delete, start, stop) as separate commands. We handle this internally:
- `up` creates machines as needed
- `stop` handles machine shutdown
- `delete` removes machines
- No need to expose complex machine management

#### 2. **No `ide` Command Group**
DevPod has complex IDE management for various editors. We simplify:
- Just use `--ide` flag on `up` command
- SSH-based connections work with any IDE
- No plugin management needed

#### 3. **No `build` Command**
DevPod uses `build` for prebuild optimization. We don't need this because:
- Claude image is standardized and cached
- Workspace creation is fast enough
- Simplifies mental model

#### 4. **No Separate `status` Command**
The `list` command shows all status information. One less command to remember.

---

### Terminal Session Management

We use tmux to enable multiple terminals to connect to the same running Claude session:

```bash
# Inside each workspace container
tmux new-session -d -s claude-main "claude --interactive"

# When users SSH in, they automatically attach
ssh claude@workspace-abc123 → tmux attach-session -t claude-main
```

**Benefits:**
- Multiple developers can view the same Claude session
- Full conversation history via tmux scrollback
- Session persists even if SSH disconnects
- Real-time collaboration - all users see same output

**Example workflow:**
```bash
# Developer 1 starts working
$ claude-vm ssh abc123
[workspace]$ claude "implement auth system"
# Claude begins working...

# Developer 2 joins to help
$ claude-vm ssh abc123
# Sees everything Developer 1 and Claude have done
# Can interact with the same Claude session
```

---

### Implementation Structure

```go
// Command structure mirrors DevPod's approach
type Command struct {
    Use   string
    Short string
    Run   func(cmd *cobra.Command, args []string) error
}

var rootCmd = &cobra.Command{
    Use:   "claude-vm",
    Short: "Manage Claude development workspaces",
}

// Subcommands follow DevPod patterns
func init() {
    rootCmd.AddCommand(upCmd)
    rootCmd.AddCommand(listCmd)
    rootCmd.AddCommand(sshCmd)
    // ... etc
}
```

### Configuration Files

Following DevPod's structure:

```
~/.claude-vm/
├── config.yaml         # Global config
├── contexts/           # Context configs
│   ├── personal.yaml
│   └── work.yaml
├── providers/          # Provider configs
│   ├── fly.yaml
│   └── docker.yaml
└── workspaces/         # Workspace state
    └── abc123/
        └── workspace.yaml
```

### Why Clone DevPod's Patterns?

1. **Proven UX**: DevPod solved similar problems (remote development environments)
2. **Familiar to Developers**: Many already use DevPod/Codespaces
3. **Extensible Design**: Provider abstraction enables growth
4. **Clear Mental Model**: Workspace lifecycle is intuitive

### Where We Diverge

1. **AI-First**: Commands designed for Claude interaction, not just development
2. **Simpler IDE Story**: We rely on SSH-based connections, not complex IDE plugins
3. **Git Safety**: Built-in protections for credential isolation
4. **Mobile Interface**: Web UI for supervision from phones
5. **Activity Streaming**: Real-time visibility into Claude's actions

---

## Why Not Fork DevPod?

DevPod is an excellent tool for general development environments, but it's massively overengineered for our specific use case. Here's what DevPod includes that we don't need:

### Three Architecture Options

#### Option 1: Fork/Wrap DevPod
**Use DevPod's devcontainer.json parser + our Claude-specific features**
```bash
# DevPod handles environment setup
devpod up github.com/user/repo --provider fly

# We add Claude layer on top
claude-vm enhance <devpod-workspace-id>
```
- Pros: Reuse battle-tested devcontainer parsing
- Cons: Two-step process, complex integration

#### Option 2: Limited devcontainer.json Support
**Parse only the fields Claude needs (90% of use cases)**
```json
// We parse these essential fields:
{
  "image": "mcr.microsoft.com/devcontainers/python:3.11",
  "features": {
    "ghcr.io/devcontainers/features/rust:1": {}
  },
  "postCreateCommand": "pip install -r requirements.txt",
  "containerEnv": {
    "FLASK_ENV": "development"
  }
}
// Ignore complex stuff like mounts, docker-compose, etc.
```
- Pros: Simpler implementation, covers most needs
- Cons: Some projects won't work

#### Option 3: Claude Base + User Container
**Layer Claude on top of user's existing container**
```dockerfile
# Start from user's devcontainer
FROM ${USER_DEVCONTAINER_IMAGE}

# Add Claude layer
RUN apt-get update && apt-get install -y tmux
COPY claude /usr/local/bin/
COPY git-proxy /usr/local/bin/

# Claude-specific setup
RUN git config --global credential.helper proxy
ENTRYPOINT ["tmux", "new-session", "-s", "claude-main", "claude"]
```
- Pros: Works with any devcontainer
- Cons: Still need to parse devcontainer.json to get base image

### Recommendation: Limited devcontainer.json Support

After reconsidering, we should implement **Option 2 - Limited devcontainer.json parsing** that covers the 90% use case:

```go
// What we parse from devcontainer.json
type SimplifiedDevContainer struct {
    Image              string            `json:"image"`
    DockerFile         string            `json:"dockerFile"`  
    Features           map[string]interface{} `json:"features"`
    PostCreateCommand  string            `json:"postCreateCommand"`
    ContainerEnv       map[string]string `json:"containerEnv"`
    ForwardPorts       []int             `json:"forwardPorts"`
}
```

### Implementation Strategy

1. **Check for devcontainer.json**
   ```bash
   $ claude-vm up github.com/user/repo
   ✓ Found .devcontainer/devcontainer.json
   ✓ Using image: mcr.microsoft.com/devcontainers/python:3.11
   ✓ Installing features: rust, docker-in-docker
   ✓ Adding Claude and tmux layer...
   ```

2. **Fall back to language detection**
   ```bash
   $ claude-vm up github.com/user/repo
   ⚠ No devcontainer.json found
   ✓ Detected: Python project (requirements.txt)
   ✓ Using default Python 3.11 + Claude image
   ```

3. **Override when needed**
   ```bash
   $ claude-vm up github.com/user/repo --image claude-node:20
   ✓ Using specified image instead of devcontainer.json
   ```

4. **Handle unsupported features gracefully**
   ```bash
   $ claude-vm up github.com/complex/repo
   ⚠ devcontainer.json uses unsupported features:
     - docker-compose.yml (ignored)
     - customizations.vscode (ignored)
   ✓ Using base image: mcr.microsoft.com/devcontainers/javascript-node:20
   ℹ Tip: Use --image to specify a custom image with your exact setup
   ```

### Why This Approach Works

- **Covers most cases**: Basic image + features handles 90% of projects
- **Escape hatch**: Users can specify custom images when needed
- **Incremental**: Can add more devcontainer.json support over time
- **Simple**: ~500 lines of Go vs DevPod's thousands
- **Fast**: No complex parsing or docker-compose orchestration

### What We Skip From Full devcontainer.json

- Docker Compose configurations
- Complex mount configurations  
- VS Code specific settings
- Remote user management
- Elaborate lifecycle hooks

These are rarely needed for Claude's use case and add significant complexity.

---

## Comparison to Alternatives

vs **DevPod:**
- Purpose-built for Claude, not general development
- Simplified devcontainer.json parsing (90% coverage vs 100% complexity)
- Integrated git safety (credential proxy)
- Mobile-first supervision interface
- tmux-based session sharing built-in
- Single-step process (vs DevPod + manual Claude setup)

vs **Anthropic's GitHub Actions for Claude:**
- Works with uncommitted code
- No GitHub repo required
- Use your Claude Pro / Max plan, rather than paying API-rates to Anthropic
- Review and edit changes in your IDE, rather than on the Github website
- Real-time interaction vs async PR comments

---
