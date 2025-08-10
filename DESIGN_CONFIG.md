## Configuration

claude-vm follows standard CLI configuration patterns with secure secret management.

### Configuration Hierarchy (highest to lowest priority)

1. **Command line flags**: `--provider fly`, `--token xxx`
2. **Environment variables**: `CLAUDE_VM_FLY_TOKEN=xxx`  
3. **Config files**: `~/.config/claude-vm/config.yaml`
4. **Defaults**: provider: docker, agent: claude

### File Structure

```bash
~/.config/claude-vm/config.yaml
```

### Explicit Provider Configuration

**(Following DevPod Pattern):**

**Method 1: Interactive CLI (Recommended)**
```bash
claude-vm provider add digitalocean
# Prompts for required options:
# → DigitalOcean Token: [hidden input]
# → Default Region (nyc1): 
# → Default Droplet Size (g-2vcpu-8gb):
```

**Method 2: Direct CLI with Options**
```bash
claude-vm provider add digitalocean \
  --option token=dop_v1_abc123 \
  --option region=nyc1 \
  --option droplet_size=g-2vcpu-8gb
```

**Method 3: Environment Variables**
```bash
export DIGITALOCEAN_TOKEN=dop_v1_abc123
export DIGITALOCEAN_REGION=nyc1

claude-vm provider add digitalocean \
  --option token='${DIGITALOCEAN_TOKEN}' \
  --option region='${DIGITALOCEAN_REGION}'
```

### Auto Provider Configuration

If the user has devpod installed, we can probably just fetch configurations directly from there for some providers. Maybe we can also get configurations from the provider's specific CLI? Like can we access the configuration for doctl for digitalocean?

---

**Provider-Specific Examples:**
```bash
# Fly.io
claude-vm provider add fly
# Auto-detects: flyctl auth token
# Prompts for: organization, region

# AWS  
claude-vm provider add aws
# Auto-detects: AWS CLI profile
# Prompts for: region, instance_type

# DigitalOcean
claude-vm provider add digitalocean  
# Prompts for: token, region, droplet_size
```

**Provider Management (CRUD Operations):**
```bash
# Create/Add provider
claude-vm provider add digitalocean          # Interactive mode
claude-vm provider add fly \
  --option token=fo1_xxx \
  --option organization=my-org

# Read/List providers  
claude-vm provider list                      # Show all configured providers
# Output:
# NAME           TYPE            STATUS
# docker         local           ✓ Ready
# fly            cloud           ✓ Authenticated
# digitalocean   cloud           ✓ Configured
# aws            cloud           ✗ Not configured

claude-vm provider list -v                   # Verbose: show options (secrets hidden)
# Output:
# NAME: digitalocean
#   TYPE: cloud
#   STATUS: ✓ Configured
#   OPTIONS:
#     token: dop_v1_****** (hidden)
#     region: nyc1
#     droplet_size: g-2vcpu-8gb

# Update provider options
claude-vm provider set-options digitalocean \
  --option region=sfo3 \
  --option droplet_size=s-4vcpu-8gb
  
# Delete provider
claude-vm provider delete digitalocean       # Remove provider configuration
```

### Secret Storage Strategy

Store secrets in the machine's local keychain. We can also use browser-based oauth flows for some services.

Devpod stores user-secrets in plaintext in JSON files; we can probably do better than this though.

---

### Config File Format

```yaml
# ~/.config/claude-vm/config.yaml

# Default settings
defaults:
  provider: "docker"               # Default cloud provider: docker, fly, aws, digitalocean
  agent: "claude"                  # Default coding agent for new-chat

# Coding agents configuration
agents:
  enabled:                         # All enabled agents installed in devcontainer
    - claude                       # Claude Code (Anthropic)
    - qwen                         # Qwen Coder
    - codex                        # OpenAI Codex
    - goose                        # Goose AI
    - gemini                       # Google Gemini
  
  # Agent-specific settings (non-secret)
  claude:
    model: "sonnet"                # Model: sonnet, opus, haiku
    
  qwen:
    base_url: "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
    model: "qwen3-coder-plus"      # qwen3-coder-plus, Qwen3-Coder-480B
    
  codex:
    model: "o4-mini"               # o3, o4-mini, gpt-4.1
    
  goose:
    provider: "anthropic"          # anthropic, openai, google, groq
    model: "claude-3-5-sonnet-20241022"  # For anthropic provider
    
  gemini:
    model: "gemini-1.5-pro"        # gemini-1.5-pro, gemini-1.5-flash

# Cloud provider configuration (non-secret)
providers:
  docker:
    host: "unix:///var/run/docker.sock"
    default_image: "mcr.microsoft.com/devcontainers/universal:2"
    
  fly:
    organization: "my-org"         # Default organization
    region: "iad"                  # Default region (iad, lax, ams, etc.)
    # Custom machine configuration (2 vCPUs, 8GB RAM)
    cpu_kind: "performance"        # performance or shared
    cpus: 2                        # Number of vCPUs
    memory_mb: 8192                # Memory in MB (8GB)
    # Alternative preset: performance-4x (4 vCPUs, 8GB RAM)
    
  digitalocean:
    region: "nyc1"
    droplet_size: "g-2vcpu-8gb"    # General purpose: 2 vCPUs, 8GB RAM
    # Alternatives: gd-2vcpu-8gb (dedicated CPU), m-2vcpu-16gb (memory-optimized)

# Storage configuration (S3-compatible)
storage:
  endpoint: "https://<account-id>.r2.cloudflarestorage.com"     # S3 endpoint URL
  bucket: "claude-vm-backups"                                   # S3 bucket name
  region: "auto"                                                # S3 region
  prefix: "workspaces/"                                         # Key prefix for organization

# Workspace defaults
workspace:
  auto_stop_minutes: 60            # Stop workspace after idle time
    
# NO SECRETS IN CONFIG FILES - ALL SECRETS STORED IN OS KEYCHAIN
```

---

### GitHub App Authentication Flow

claude-vm uses a GitHub App for secure, granular access to private repositories. This enables cloning repos into devcontainers and pushing commits/PRs back to the original repo.

**1. App Registration (One-time, by claude-vm developers):**
```bash
# We create a GitHub App at: https://github.com/settings/apps/new
App Name: "claude-vm"
Homepage: "https://claude-vm.dev"
Permissions:
  - Repository contents: Read & write    # Clone repos, read/write files
  - Pull requests: Write                # Create PRs from workspace changes  
  - Metadata: Read                      # Basic repo info
  - Actions: Read                       # Access to repository actions (optional)
```

**2. User Installation Flow:**
```bash
claude-vm auth github
# Opens: https://github.com/apps/claude-vm/installations/new
# User selects:
#   - Personal account OR organization
#   - All repositories OR selected repositories
#   - Installs app with chosen scope
```

**3. Token Generation & Storage:**
```bash
# After installation, GitHub provides:
Installation ID: 12345678
Repository Access: ["user/private-repo", "org/secret-project"] 

# We store in OS keychain:
CLAUDE_VM_GITHUB_INSTALLATION_ID: "12345678"
CLAUDE_VM_GITHUB_APP_ID: "our-app-id" 
CLAUDE_VM_GITHUB_PRIVATE_KEY: "-----BEGIN RSA PRIVATE KEY-----..."
```

**4. Runtime Token Generation:**
```python
# When workspace needs GitHub access:
installation_token = github_api.create_installation_access_token(
    app_id=CLAUDE_VM_GITHUB_APP_ID,
    private_key=CLAUDE_VM_GITHUB_PRIVATE_KEY, 
    installation_id=CLAUDE_VM_GITHUB_INSTALLATION_ID
)
# Returns: 1-hour token with access to user's selected repos
```

**5. Authentication Transfer (Local vs Remote):**

**Local Containers (Docker):**
```bash
# Mount credentials read-only
docker run -v ~/.claude:/workspace/.claude:ro \
           -v ~/.config/claude-vm:/workspace/.config/claude-vm:ro \
           devcontainer
```

**Remote Containers (Fly.io, DigitalOcean, AWS):**
```bash
# Extract and inject secrets during deployment
claude-vm workspace up https://github.com/user/repo --provider fly

# Behind the scenes:
# 1. Extract from local storage
claude_token=$(security find-generic-password -s "claude-vm-claude" -w)
github_key=$(security find-generic-password -s "claude-vm-github-key" -w)
github_install_id=$(security find-generic-password -s "claude-vm-github-install" -w)

# 2. Generate fresh GitHub token
github_token=$(curl -X POST \
  -H "Authorization: Bearer ${jwt_token}" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/app/installations/${github_install_id}/access_tokens")

# 3. Deploy with secrets as environment variables  
fly deploy --env CLAUDE_CREDENTIALS="${claude_token}" \
           --env GITHUB_TOKEN="${github_token}" \
           --env GITHUB_INSTALL_ID="${github_install_id}"
```

**Inside Remote Container:**
```bash
# Container startup script recreates local auth files
mkdir -p /workspace/.claude
echo "${CLAUDE_CREDENTIALS}" > /workspace/.claude/.credentials.json
chmod 600 /workspace/.claude/.credentials.json

# GitHub token setup with safety measures
export GITHUB_TOKEN="${GITHUB_TOKEN}"  # 1-hour installation token
git config --global credential.helper '!f() { echo "password=${GITHUB_TOKEN}"; }; f'
git config --global url."https://oauth:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Agent can now use git normally without seeing token
git clone https://github.com/user/private-repo.git  # Token injected automatically
git push origin feature-branch                      # Works seamlessly
gh pr create --title "AI changes"                   # gh CLI uses GITHUB_TOKEN env var
```

**Token Safety Measures:**

1. **Short-lived (1 hour)**: Installation tokens expire quickly
   - If leaked, window of vulnerability is minimal
   - Token auto-refreshes in background

2. **Repository-scoped**: Token only works for repos user explicitly granted
   - Can't access user's other repos
   - Can't access user's profile/settings

3. **Limited permissions**: Only what we requested in GitHub App
   - contents: read/write
   - pull_requests: write
   - No admin, no delete, no settings access

4. **Credential helper**: Agent never sees the actual token
   ```bash
   # Agent runs: git push
   # Git automatically injects token via credential helper
   # Agent output never shows the token value
   ```

5. **Token rotation**: Each workspace gets fresh token
   ```bash
   # Inside container: /usr/local/bin/github-token-refresh.sh
   # Runs via cron every 50 minutes (before 1-hour expiry)
   
   #!/bin/bash
   # Generate JWT for GitHub App authentication
   jwt_token=$(generate_jwt_token \
     --app-id="${GITHUB_APP_ID}" \
     --private-key="${GITHUB_PRIVATE_KEY}")
   
   # Request new installation token
   new_token=$(curl -X POST \
     -H "Authorization: Bearer ${jwt_token}" \
     -H "Accept: application/vnd.github.v3+json" \
     "https://api.github.com/app/installations/${GITHUB_INSTALL_ID}/access_tokens" \
     | jq -r '.token')
   
   # Update environment for all processes
   echo "export GITHUB_TOKEN='${new_token}'" > /etc/profile.d/github-token.sh
   
   # Update git credential helper
   git config --global credential.helper \
     "!f() { echo \"password=${new_token}\"; }; f"
   ```
   
   **Container crontab:**
   ```cron
   # Refresh GitHub token every 50 minutes
   */50 * * * * /usr/local/bin/github-token-refresh.sh >> /var/log/token-refresh.log 2>&1
   ```

**What Agent CAN do:**
- ✅ Clone repos
- ✅ Create branches
- ✅ Commit changes
- ✅ Push to branches
- ✅ Open pull requests
- ✅ Read issues

**What Agent CANNOT do:**
- ❌ Delete repos
- ❌ Change repo settings
- ❌ Access other repos not granted
- ❌ Access user account settings
- ❌ Create/delete GitHub Apps
- ❌ Access billing

**If Token is Leaked (worst case):**
- Attacker has 1 hour max to use it
- Can only access specific repos user granted
- All actions logged as "claude-vm[bot]"
- User can revoke GitHub App installation immediately
- No access to user's actual GitHub account

**Repository Access Workflow:**
1. **Clone**: Use installation token to clone private repos into `/workspace`
2. **Development**: AI agent modifies code within devcontainer
3. **Commit**: Push changes to new branch using same token
4. **Pull Request**: Open PR back to original repo via GitHub API

**Security Considerations for Remote Containers:**

**Token Lifecycle Management:**
- **GitHub tokens**: Generated fresh for each workspace (1-hour expiry)
- **Claude credentials**: Copied but never stored in provider infrastructure  
- **Environment isolation**: Each workspace gets its own token set
- **Automatic cleanup**: Tokens cleared when workspace deleted

**Provider-Specific Security:**
```bash
# Fly.io: Use secrets (encrypted at rest)
fly secrets set CLAUDE_CREDENTIALS="..." --app workspace-abc123

# AWS: Use Systems Manager Parameter Store  
aws ssm put-parameter --name "/claude-vm/workspace-abc123/claude-creds" \
                     --value "..." --type SecureString

# DigitalOcean: Environment variables (less secure)
# Best practice: Use their upcoming secrets management
```

**Benefits:**
- **Granular Access**: User chooses exactly which repos to grant access
- **Security**: 1-hour token expiry, fresh tokens per workspace
- **Audit Trail**: All actions appear as "claude-vm[bot]" in GitHub
- **Organization Friendly**: Org admins can see and control app installations
- **Remote Compatibility**: Works across all cloud providers
- **Zero Local Dependency**: Container runs independently after deployment

**Multi-Account Support:**
```bash
# Different installations for different contexts
claude-vm auth github --account personal    # Personal repositories
claude-vm auth github --account work-org    # Work organization repositories

# Workspace inherits from repo context
claude-vm workspace up https://github.com/work-org/private-repo  # Uses work-org installation
claude-vm workspace up https://github.com/user/personal-repo    # Uses personal installation
```

---

### Hosted Service Authentication (claude-vm.com)

Users can access their workspaces from anywhere through our hosted web interface, even after turning off their local machine.

**1. Account Setup:**
```bash
claude-vm auth login
# Opens: https://claude-vm.com/auth/login
# OAuth options: GitHub, Google, Email magic link
# Creates account linked to user's email
```

**2. Workspace Registration:**
```bash
# When creating remote workspace, it's registered with our service
claude-vm workspace up https://github.com/user/repo --provider fly

# Behind the scenes:
POST https://api.claude-vm.com/workspaces
{
  "workspace_id": "bold-fire-1234",
  "provider": "fly",
  "repo": "github.com/user/repo",
  "user_id": "user-uuid-123",
  "access_token": "encrypted-token",
  "api_endpoint": "https://bold-fire-1234.fly.dev"
}
```

**3. Web Interface Access:**
```bash
# Option 1: Local Web UI (when on same machine)
claude-vm web                           # Opens localhost:8080

# Option 2: Hosted Web UI (from anywhere)
https://claude-vm.com/workspaces        # Shows all user's workspaces
```

**Hosted Web Features:**
- **Workspace List**: All workspaces across all providers
- **Status Dashboard**: Running, stopped, errors
- **Direct Access**: Click to open workspace web interface
- **Chat Interface**: Talk to Claude in any workspace
- **File Browser**: Review and edit files
- **Terminal Access**: Web-based SSH terminal
- **Logs**: View Claude conversation history

**4. Security Model:**
```yaml
# Stored in our database (encrypted)
user_workspaces:
  user_id: "user-uuid-123"
  workspaces:
    - id: "bold-fire-1234"
      provider: "fly"
      access_url: "https://bold-fire-1234.fly.dev"
      encrypted_token: "..." # For workspace API access
    - id: "3f2504e0bb11"
      provider: "docker"
      access_url: "https://tunnel-abc.claude-vm.com" # Tunneled
```

**5. Authentication Flow:**
```mermaid
User → claude-vm.com → OAuth Provider → User Account
         ↓
   Workspace List
         ↓
   Select Workspace → Proxy to Workspace API
                     (using encrypted token)
```

**Benefits:**
- **Access from anywhere**: Phone, tablet, different computer
- **No local dependency**: Workspaces continue running
- **Centralized management**: See all workspaces in one place
- **Shared workspaces** (Pro): Teams can share workspace access
- **Mobile friendly**: Responsive web UI for supervision on the go

**Free vs Pro:**
```bash
# Free tier
- Unlimited local workspaces (Docker)
- 3 remote workspaces
- Basic web UI

# Pro tier ($X/month)
- Unlimited remote workspaces
- Team workspace sharing
- Priority support
- Advanced monitoring
- Workspace templates
```

### Configuration Status

```bash
claude-vm config status                  # Show current configuration
```

Output example:
```
✓ claude-vm.com: Logged in as user@example.com
✓ Default Provider: fly (authenticated)
✓ GitHub: App installed, 3 repositories  
✓ Claude: Credentials found, ready
- OpenAI: Not configured (optional)
- Google: Not configured (optional)

Remote Workspaces: 2 running, 1 stopped
Web Dashboard: https://claude-vm.com/workspaces
```

### Security & Best Practices

**Secret Storage:**
- **OS Keychain**: Primary storage for API keys, OAuth tokens
- **File Permissions**: Private keys stored with 600 permissions
- **No Plain Text**: Config files contain ZERO secrets

**Multi-Environment Support:**
```bash
# Development
claude-vm config set default_provider docker

# Production/CI
CLAUDE_VM_PROVIDER=fly CLAUDE_VM_FLY_TOKEN=xxx claude-vm workspace up

# Enterprise
claude-vm config set config_url https://company.com/claude-vm/config
```

**GitHub Apps vs PAT:**
- **GitHub Apps** (recommended): Repository-specific, 1-hour tokens, audit trail
- **Personal Access Tokens** (fallback): Broader permissions, no expiration

---

### DevPod Import

If the user has devpod installed, we can probably just fetch configurations directly from there for some providers.
