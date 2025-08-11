## Configuration

claude-vm stores all configuration in a single YAML file with secrets in the OS keychain.

### Storage Locations

```bash
# Configuration (non-secrets)
~/.claude-vm/config.yaml

# Secrets (OS keychain)
macOS: Keychain Access
Linux: Secret Service API (gnome-keyring/KWallet)
Windows: Credential Manager
```

### Provider Configuration

Providers are any external service we interact with - VM providers (AWS, DigitalOcean, Fly), API providers (Anthropic, OpenAI), or both (Google provides GCP VMs and Gemini API).

**Configuration Priority (Highest to Lowest):**
1. **CLI flags**: `--option api_key=xxx`
2. **Environment variables**: `AWS_ACCESS_KEY_ID`, `ANTHROPIC_API_KEY`
3. **Stored config**: `~/.claude-vm/config.yaml` (non-secrets) + OS keychain (secrets)
4. **Existing CLI tools**: 
   - AWS: `~/.aws/credentials`
   - Google Cloud: `~/.config/gcloud/application_default_credentials.json`
   - DigitalOcean: `~/.config/doctl/config.yaml` (Linux) or `~/Library/Application Support/doctl/config.yaml` (macOS)
   - Fly.io: `~/.fly/config.yml`
   - Claude Code: `~/.claude.json` or macOS Keychain
5. **Interactive prompt**: Ask user for required values (disabled by --non-interactive)

**Interactive Configuration:**

By default, claude-vm uses interactive prompts when configuration is missing:

```bash
# Example: Missing provider configuration
claude-vm workspace up . --agent goose
# ✗ Goose agent has no configured provider

┌─ Goose Agent Setup Required ─────────────────────────────────┐
│ Goose requires a provider configuration.                       │
│                                                              │
│ Available providers:                                         │
│ 1) Anthropic (claude-3.5-sonnet) - ✓ configured             │
│ 2) OpenAI (gpt-4o) - ✓ configured                          │
│ 3) OpenRouter (anthropic/claude-3.5-sonnet) - setup needed  │
│                                                              │
│ Select provider [1-3]: 1                                     │
└──────────────────────────────────────────────────────────────┘

✓ Configured goose agent to use anthropic provider
❯ Continuing with workspace creation...
```

**Non-Interactive Mode:**

Use `--non-interactive` flag to disable prompts and fail with actionable errors:

```bash
claude-vm workspace up . --agent goose --non-interactive
# ERROR: Agent 'goose' has no configured provider
#   Run: claude-vm agent set-config goose --option provider=anthropic
#   Available providers: anthropic (configured), openai (configured)
```

**Cloud Platform Commands:**

```bash
# Set cloud platform options (configure or update)
claude-vm cloud set-config <name> [--option key=value ...]

# List all cloud platforms with their configuration status
claude-vm cloud list

# Clear cloud platform options (remove configuration)
claude-vm cloud clear-config <name>
```

**LLM API Provider Commands:**

```bash
# Set LLM provider options (configure or update)
claude-vm provider set-config <name> [--option key=value ...]

# List all LLM providers with their configuration status
claude-vm provider list

# Clear LLM provider options (remove configuration)
claude-vm provider clear-config <name>
```

**Configuration Examples by Type:**

#### Cloud Platforms (for workspaces)
```yaml
docker:
  # No auth needed, uses local Docker daemon
  auto-detection: Checks /var/run/docker.sock

digitalocean:
  options:
    - api_key (required, password → keychain)
    - region (default: nyc1 → config.yaml)
    - size (default: s-2vcpu-8gb → config.yaml)
  env: DIGITALOCEAN_ACCESS_TOKEN
  cli-tool: ~/.config/doctl/config.yaml (Linux) or ~/Library/Application Support/doctl/config.yaml (macOS)

fly:
  options:
    - api_key (required, password → keychain)
    - region (default: iad → config.yaml)
  env: FLY_API_TOKEN, FLY_ACCESS_TOKEN
  cli-tool: ~/.fly/config.yml

aws:
  options:
    - access_key_id (required → keychain)
    - secret_access_key (required, password → keychain)
    - region (default: us-east-1 → config.yaml)
  env: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
  cli-tool: ~/.aws/credentials (default profile)
```

#### LLM API Providers (for agents)
```yaml
anthropic:
  options:
    - api_key (required, password → keychain)
  env: ANTHROPIC_API_KEY
  cli-tool: ~/.claude.json or macOS Keychain (OAuth)

openai:
  options:
    - api_key (required, password → keychain)
    - org_id (optional → config.yaml)
  env: OPENAI_API_KEY, OPENAI_ORG_ID

google:
  options:
    - api_key (required, password → keychain)
    - project (optional → config.yaml)
  env: GOOGLE_API_KEY, GOOGLE_CLOUD_PROJECT
  cli-tool: ~/.config/gcloud/application_default_credentials.json
  # Note: Google provider works for both GCP VMs and Gemini API
```

**Usage Examples:**

```bash
# Method 1: Environment variables (auto-detected)
export DIGITALOCEAN_ACCESS_TOKEN=dop_xxx
export ANTHROPIC_API_KEY=sk-ant-xxx
claude-vm cloud set-config digitalocean  # Picks up from env
claude-vm provider set-config anthropic     # Picks up from env

# Method 2: Existing CLI tools (auto-detected)
aws configure  # Set up AWS credentials
gcloud auth login  # Set up Google credentials
claude-vm cloud set-config aws     # Uses ~/.aws/credentials
claude-vm cloud set-config google     # Uses gcloud auth

# Method 3: Direct CLI options
claude-vm cloud set-config fly --option api_key=$FLY_TOKEN --option region=lax

# Method 4: Interactive (when required options missing)
claude-vm provider set-config openai
# Prompts: "Enter api_key (required):"

# List shows all configured providers
claude-vm provider list
# PROVIDER       STATUS       OPTIONS
# docker         configured   (local)
# digitalocean   configured   api_key=dop_*** region=nyc1
# anthropic      configured   api_key=sk-ant***
# aws            configured   (using ~/.aws/credentials)
# google         configured   (using gcloud auth)
```

### Complete Workflow Example (DevPod Pattern)

```bash
# Step 1: Configure providers (one-time setup)
# Method A: Use environment variables (recommended for CI/CD)
export DIGITALOCEAN_ACCESS_TOKEN=dop_xxx
export ANTHROPIC_API_KEY=sk-ant-xxx
export OPENAI_API_KEY=sk-xxx

# Method B: Interactive configuration
claude-vm cloud set-config digitalocean
# Checks env → existing tools → prompts for api_key, region

claude-vm provider set-config anthropic
# Checks env → ~/.claude-code/credentials → prompts if needed

# Step 2: Configure agents (one-time setup)
claude-vm agent set-config claude --option model=opus
# Auto-detects auth from: OAuth (~/.claude-code/credentials) or anthropic provider

claude-vm agent set-config codex --option model=gpt-4.1
# Auto-detects auth from: openai provider config

# Step 3: Verify configuration
claude-vm provider list
# PROVIDER       STATUS       OPTIONS
# docker         configured   (local)
# digitalocean   configured   api_key=dop_*** region=nyc1
# anthropic      configured   api_key=sk-ant***
# openai         configured   api_key=sk-xxx***

claude-vm agent list
# AGENT    MODEL       STATUS       AUTH
# claude   opus        configured   OAuth (~/.claude-code/credentials)
# codex    gpt-4.1     configured   API key (openai provider)
# qwen     default     available    Not configured
# goose    default     available    Multiple providers available
# gemini   default     not-available   Missing google provider

# Step 4: Create workspaces (clean, no configuration)
claude-vm workspace up .                      # Auto-detect .devcontainer/ or generate
claude-vm workspace up . --cloud digitalocean
claude-vm workspace up https://github.com/user/repo --cloud fly
claude-vm workspace up . --devcontainer .devcontainer/ai-dev.json  # Custom devcontainer
claude-vm workspace up . --image node:18-alpine  # Simple Docker image override

# Step 5: Use agents in workspaces (only available if baked-in during creation)
claude-vm new-chat workspace-123 claude       # Uses baked-in Claude with Opus
claude-vm new-chat workspace-123 codex        # ERROR: codex not available (not baked into workspace-123)
claude-vm chat conv-456                       # Continue existing conversation
```

### Configuration Auto-Detection Example

```bash
# Example: Setting up DigitalOcean provider
claude-vm cloud set-config digitalocean

# Auto-detection flow:
# 1. CLI flags: (none provided)
# 2. Environment: Checks DIGITALOCEAN_ACCESS_TOKEN ✓ Found!
# 3. CLI tool: Would check doctl auth list
# 4. Stored: Would check ~/.claude-vm/providers/digitalocean/config.json
# 5. Interactive: Would prompt "Enter api_key (required):"

# Result: Automatically configured from environment variable
```

**Cost Reference:**
- **Docker**: Free (local resources)
- **DigitalOcean s-2vcpu-8gb**: ~$56/month ($0.08/hour)
- **Fly.io shared-cpu-2x@2048**: ~$30/month ($0.04/hour)

---

### Agent Configuration

Agents share common API provider credentials. Multiple agents can use the same underlying API keys from providers like Anthropic, OpenAI, Google, etc.

**API Provider Credentials (Shared by Multiple Agents):**

| API Provider | Authentication Methods | Used by Agents |
|--------------|----------------------|----------------|
| Anthropic    | `ANTHROPIC_API_KEY`, OAuth (unlimited) | claude, goose(anthropic), qwen-coder(via proxy) |
| OpenAI       | `OPENAI_API_KEY`, ChatGPT Login (limited credits) | codex, goose(openai), qwen-coder(via base_url) |
| Google       | `GOOGLE_API_KEY` | gemini, goose(google) |  
| Groq         | `GROQ_API_KEY` | goose(groq) |
| Databricks   | `DATABRICKS_API_KEY` | goose(databricks) |
| OpenRouter   | `OPENROUTER_API_KEY` | goose(openrouter), qwen-coder |
| Alibaba      | `DASHSCOPE_API_KEY` | qwen-coder(primary) |

**Note:** Qwen-Coder uses OpenAI-compatible interface with configurable base URLs, enabling support for multiple providers including DashScope, OpenRouter, ModelScope, and any OpenAI-compatible API.

### Agent Configuration (Following DevPod Pattern)

Agents use the same configuration priority as providers:

**Configuration Priority (Highest to Lowest):**
1. **CLI flags**: `--option model=opus`
2. **Provider API keys**: Auto-detected from configured providers (keychain + env)
3. **Stored config**: `~/.claude-vm/config.yaml` (agent settings section)
4. **Agent-specific tools**: Check agent's native config files
5. **Interactive prompt**: Ask for required options

**Agent Commands:**

```bash
# Set agent options (configure or update)
claude-vm agent set-config <name> [--option key=value ...]

# List all agents with their configuration status
claude-vm agent list

# Clear agent options (remove configuration)
claude-vm agent clear-config <name>
```

**Supported Agents:**

#### 1. Claude Code
```yaml
Agent: claude
Provider: anthropic

Options:
  - auth: oauth or api-key (default: oauth)
  - model: default, sonnet, or opus (default: default)

Auto-detection:
  1. OAuth: ~/.claude-code/credentials
  2. API key: anthropic provider configuration
  3. Environment: ANTHROPIC_API_KEY

Models:
  - default/sonnet: claude-sonnet-4-20250514
  - opus: claude-opus-4-20250514 (requires Max plan)

Example:
  claude-vm agent set-config claude --option model=opus
  claude-vm agent set-config claude --option auth_preference=api-key  # Use API key instead of OAuth
```

#### 2. Codex CLI
```yaml
Agent: codex  
Provider: openai

Options:
  - model: codex-mini-latest, gpt-4.1, gpt-4-turbo, o4-mini (default: codex-mini-latest)

Auto-detection:
  1. API key: openai provider configuration
  2. Environment: OPENAI_API_KEY

Models:
  - codex-mini-latest: CLI-optimized, 200k context
  - gpt-4.1, gpt-4-turbo, o4-mini

Example:
  claude-vm agent set-config codex --option model=gpt-4.1
```

#### 3. Qwen-Coder
```yaml
Agent: qwen
Provider: dashscope (or openai, openrouter)

Options:
  - model: qwen3-coder-plus, qwen-coder-7b, qwen-coder-32b (default: qwen3-coder-plus)
  - temperature: 0.0-2.0 (default: 0.7)
  - base_url: Custom API endpoint

Auto-detection:
  1. API key: dashscope provider configuration
  2. Environment: DASHSCOPE_API_KEY or OPENAI_API_KEY (with base_url)

Models:
  - qwen3-coder-plus: 256K context, flagship
  - Qwen3-Coder-480B: Most capable
  - qwen-coder-7b/32b: Smaller variants

Example:
  claude-vm agent set-config qwen --option model=qwen3-coder-plus
```

#### 4. Goose AI
```yaml
Agent: goose
Provider: Multiple (anthropic, openai, google, groq, etc.)

Options:
  - provider: anthropic, openai, google, groq, databricks, ollama, openrouter
  - model: Provider-specific model name
  - temperature: 0.0-2.0 (default: 0.7)

Auto-detection:
  1. Checks configured providers for API keys
  2. Uses first available provider (anthropic → openai → google)

Example:
  claude-vm agent set-config goose --option provider=anthropic --option model=claude-3.5-sonnet
  claude-vm agent set-config goose --option provider=groq  # Uses groq's default model
```

#### 5. Gemini CLI
```yaml
Agent: gemini
Provider: google

Options:
  - model: gemini-2.5-pro, gemini-2.0-flash, gemini-1.5-pro, gemini-1.5-flash (default: gemini-2.5-pro)
  - temperature: 0.0-2.0 (default: 0.7)
  - thinking_budget: Tokens for reasoning (Gemini 2.5+ only, default: 1024)
  - max_output_tokens: Maximum response tokens (optional)
  - safety_settings: Content safety level (optional)

Auto-detection:
  1. API key: google provider configuration
  2. Environment: GOOGLE_API_KEY
  3. CLI tool: gcloud auth print-access-token

Models:
  - gemini-2.5-pro: Latest with thinking capabilities
  - gemini-2.0-flash: Fast, lightweight version
  - gemini-1.5-pro: Stable, high-performance
  - gemini-1.5-flash: Fast responses, good quality

Examples:
  claude-vm agent set-config gemini --option model=gemini-2.5-pro --option thinking_budget=2048
  claude-vm agent set-config gemini --option model=gemini-2.0-flash --option temperature=0.8
  claude-vm agent set-config gemini --option max_output_tokens=8192
```

**Final CLI Architecture - Complete Separation of Concerns (Option C):**

**Design Principle:**
Keep workspace management completely separate from agent configuration. Agents are configured once through environment variables or setup commands, then workspaces "just work".

#### 1. Workspace Management (Clean, No Agent Flags)
```bash
# Create workspace - just infrastructure
claude-vm workspace up .                    # Local Docker workspace
claude-vm workspace up . --cloud fly     # Remote Fly.io workspace
claude-vm workspace up https://github.com/user/repo --cloud digitalocean

# Workspace lifecycle commands
claude-vm workspace list
claude-vm workspace down workspace-123
claude-vm workspace delete workspace-123
claude-vm ssh workspace-123

# NO agent configuration flags here - keeps it simple
```

#### 2. Agent Configuration (Simple: setup, list, clear)
```bash
# Configure providers first (stores API keys)
claude-vm provider set-config anthropic --option api-key=sk-ant-xxx
claude-vm provider set-config openai --option api-key=sk-xxx
claude-vm provider set-config google --option api-key=AIza-xxx

# Configure agent preferences
claude-vm agent set-config claude --option model=opus --option auth_preference=oauth     # Use OAuth + Opus
claude-vm agent set-config claude --option model=sonnet --option auth_preference=api-key # Use API key + Sonnet  
claude-vm agent set-config codex --option model=gpt-4.1
claude-vm agent set-config qwen --option model=qwen3-coder-plus
claude-vm agent set-config gemini --option model=gemini-2.5-pro
claude-vm agent set-config goose --option provider=anthropic --option model=claude-3.5-sonnet

# Agent management (only 3 commands)
claude-vm agent list                    # Show all agents with their configuration
claude-vm agent clear-config claude    # Clear Claude configuration
```

#### 3. Agent Interaction (Runtime)
```bash
# Start new conversations (must specify agent)
claude-vm new-chat workspace-123 claude     # Start new conversation with Claude
claude-vm new-chat workspace-123 codex      # Start new conversation with Codex  
claude-vm new-chat workspace-123 qwen       # Start new conversation with Qwen
claude-vm new-chat workspace-123            # Interactive: pick agent for new conversation

# Continue existing conversations (agent implicit from conversation)
claude-vm chat conv-456                     # Continue conversation conv-456
claude-vm chat --workspace workspace-123    # Interactive: pick existing conversation

# Conversation management
claude-vm conversations workspace-123       # List all conversations in workspace
claude-vm conversations workspace-123 --agent claude # List Claude conversations only
```

**Benefits of Separated Architecture:**
- ✅ **Single responsibility**: Each command does one thing well
- ✅ **Reusable configuration**: Set up agents once, use across workspaces
- ✅ **Clear mental model**: Workspace = infrastructure, Agent = tools, Chat = interaction
- ✅ **Easier troubleshooting**: "Is my workspace running?" vs "Is my agent configured?"
- ✅ **Better defaults**: `workspace up` just creates infrastructure with sensible defaults

**Why Complete Separation (Option C)?**

For scripting and CI/CD:
- ✅ Environment variables are standard for credentials
- ✅ Workspace commands stay simple and predictable  
- ✅ No flag namespace pollution (no --claude-*, --codex-*, --qwen-*)
- ✅ Configuration is reusable across all workspaces
- ✅ Clear separation makes debugging easier

Example CI/CD workflow:
```bash
# CI setup (once)
export ANTHROPIC_API_KEY=${{ secrets.ANTHROPIC_KEY }}
export OPENAI_API_KEY=${{ secrets.OPENAI_KEY }}

# Create workspaces (simple, repeatable)
claude-vm workspace up repo1/ --cloud fly
claude-vm workspace up repo2/ --cloud fly
claude-vm workspace up repo3/ --cloud fly

# Use agents (already configured from environment)
claude-vm new-chat workspace-1 claude
claude-vm new-chat workspace-2 codex
```

**Backward Compatibility Option:**
```bash
# For users who want the old "everything in one command" approach
claude-vm up . --agent claude              # Shorthand that:
                                          # 1. Creates workspace
                                          # 2. Auto-configures claude agent if needed
                                          # 3. Starts new conversation with Claude

# Equivalent to:
claude-vm workspace up .                   # Creates workspace-123
claude-vm agent set-config claude --option auth_preference=oauth  # (if not already configured)
claude-vm new-chat workspace-123 claude    # Starts new Claude conversation
```

**Agent Installation Strategy:**

**DevContainer Features (Baked-In at Creation):**

Agents are baked into the workspace container at creation time using devcontainer features:

```bash
# User specifies agents when creating workspace
claude-vm workspace up . --agent claude,goose,gemini
```

**Generated DevContainer:**
```json
// claude-vm generates this devcontainer.json
{
  "name": "claude-vm-workspace",
  "features": {
    // Only requested agents get installed
    "ghcr.io/anthropics/devcontainer-features/claude-code:latest": {},
    "ghcr.io/block/devcontainer-features/goose:latest": {},
    "ghcr.io/google/devcontainer-features/gemini:latest": {},
    
    "ghcr.io/claude-vm/devcontainer-features/workspace-manager:latest": {
      "enabledAgents": ["claude", "goose", "gemini"]
    }
  }
}
```

**Key Principles:**
- **Fixed Agent Set**: Agents specified at creation time cannot be changed later
- **Build-Time Installation**: Agents installed during container build, not runtime
- **Immediate Availability**: All agents ready when container starts
- **No Dynamic Addition**: Cannot add/remove agents from existing workspaces

**Agent Configuration Examples:**

```bash
# Environment Variable Approach (Recommended) - Provider-Based
export ANTHROPIC_API_KEY=sk-ant-xxx  # Enables claude + goose-anthropic
export OPENAI_API_KEY=sk-xxx         # Enables codex + goose-openai
export DASHSCOPE_API_KEY=sk-xxx      # Enables qwen
export GOOGLE_API_KEY=AIza-xxx       # Enables gemini + goose-google

# Agents specified at workspace creation (baked into devcontainer)
claude-vm workspace up . --agent claude        # Claude baked in, uses ANTHROPIC_API_KEY
claude-vm workspace up . --agent codex         # Codex baked in, uses OPENAI_API_KEY  
claude-vm workspace up . --agent claude,goose  # Both baked in, share anthropic provider
claude-vm workspace up . --agent claude,codex,gemini  # All three baked in

# Goose uses pre-configured provider (set via agent config)
claude-vm agent set-config goose --option provider=anthropic    # Configure goose to use Anthropic
claude-vm workspace up . --agent goose                         # Uses pre-configured provider

# Pre-configure providers, then use agents
claude-vm provider set-config anthropic --option api-key=sk-ant-xxx
claude-vm provider set-config openai --option api-key=sk-xxx
claude-vm provider set-config alibaba --option api-key=sk-xxx
claude-vm provider set-config google --option api-key=AIza-xxx

# Then use agents (they automatically use their configured providers)
claude-vm workspace up . --agent claude      # Uses anthropic provider
claude-vm workspace up . --agent codex       # Uses openai provider
claude-vm workspace up . --agent qwen        # Uses alibaba provider
claude-vm workspace up . --agent gemini      # Uses google provider

# Claude Code OAuth (Auto-detected, preferred)
claude-vm workspace up . --agent claude
# → Auto-detects ~/.claude-code/credentials, no additional config needed

# Claude Code OAuth (Explicit file path)
claude-vm workspace up . --agent claude --claude-oauth-file ~/.my-claude-creds.json

# Claude Code OAuth (Custom location)
claude-vm workspace up . --agent claude --claude-oauth-file /path/to/project/.claude-credentials

# Claude Code API Key (Fallback)
claude-vm workspace up . --agent claude --anthropic-key sk-ant-xxx
export ANTHROPIC_API_KEY=sk-ant-xxx && claude-vm workspace up . --agent claude

# OpenAI Codex (ChatGPT Login - Recommended for existing ChatGPT users)
claude-vm workspace up . --agent codex
# → Prompts for "Sign in with ChatGPT" if no API key found
# → Auto-generates API key, provides promotional credits

# OpenAI Codex (API Key Method with Model Selection)
claude-vm workspace up . --agent codex --openai-key sk-xxx --codex-model gpt-4.1
export OPENAI_API_KEY=sk-xxx && claude-vm workspace up . --agent codex --codex-model codex-mini-latest

# Agent Configuration Examples with Advanced Settings
claude-vm workspace up . --agent claude --claude-model opus
claude-vm workspace up . --agent qwen-coder --qwen-model qwen3-coder-plus --qwen-temperature 0.7
claude-vm workspace up . --agent goose --goose-provider anthropic --goose-model claude-3.5-sonnet --goose-temperature 0.8
claude-vm workspace up . --agent gemini --gemini-model gemini-2.5-pro --gemini-temperature 0.6 --gemini-thinking-budget 1024
```

**Complete Multi-Agent Setup:**

```bash
# Set up API provider credentials once (enables multiple agents per provider)
export ANTHROPIC_API_KEY=sk-ant-xxx    # Enables: claude, goose-anthropic
export OPENAI_API_KEY=sk-xxx           # Enables: codex, goose-openai
export GOOGLE_API_KEY=AIza-xxx         # Enables: gemini, goose-google
export DASHSCOPE_API_KEY=sk-xxx        # Enables: qwen

# Option 1: Auto-detect all available agents
claude-vm workspace up .
# → Enables: claude, codex, gemini, qwen (+ goose variants)

# Option 2: Choose specific agents (share provider credentials)  
claude-vm workspace up . --agent claude,codex,gemini
# Both agents use their configured providers
claude-vm workspace up . --agent claude,goose    # Claude uses anthropic, goose uses its configured provider

# Option 3: Mixed authentication methods
# Each agent uses its configured provider
claude-vm workspace up . --agent claude,codex,goose
# → claude uses OAuth (auto-detected), codex+goose use OPENAI_API_KEY

# Option 4: Explicit OAuth + API keys
claude-vm workspace up . --agent claude,codex \
  --claude-oauth-file ~/.claude-code/credentials \
  --openai-key sk-xxx
```

**Configuration Persistence (Provider-Based):**

```bash
# First run - API provider credentials saved to keychain
export ANTHROPIC_API_KEY=sk-ant-xxx
export OPENAI_API_KEY=sk-xxx  
claude-vm workspace up . --agent claude,codex

# Future runs - uses saved provider credentials
# Each agent uses its configured provider  
claude-vm workspace up . --agent claude,codex,goose
# → claude uses OAuth (if available) or saved Anthropic key, codex uses saved OpenAI key

# OAuth credentials are NOT persisted by claude-vm (uses existing Claude Code files)
claude-vm workspace up . --agent claude  
# → Always reads fresh OAuth data from ~/.claude-code/credentials
```

**Agent Selection in Runtime:**

Once workspace is created with multiple agents, users can work with different agents through conversations:

```bash
# List conversations and see which agent each uses
claude-vm conversations workspace-123
# Shows: 
# Conversations:
# - conv-001 (claude): "Fix authentication bug"  
# - conv-002 (gemini): "Add new API endpoint"
# - conv-003 (qwen): "Optimize database queries"

# Create new conversations with specific agents
claude-vm new-chat workspace-123 codex
claude-vm new-chat workspace-123 qwen  
claude-vm new-chat workspace-123 goose

# Continue existing conversations (agent implicit from conversation metadata)
claude-vm chat conv-001  # Uses claude (stored in conversation)
claude-vm chat conv-002  # Uses gemini (stored in conversation)
claude-vm chat conv-003  # Uses qwen (stored in conversation)

# Interactive conversation picker
claude-vm chat --workspace workspace-123
# Shows conversation list, user picks one to continue
```

**Cost Reference:**
- **Claude Code**: $15/million tokens (Sonnet), $75/million (Opus), $0.80/million (Haiku)
- **OpenAI Codex**: $10/million tokens (GPT-4), $2/million (GPT-3.5), $20/million (O1)
- **Qwen Coder**: $2/million tokens (Plus), $0.50/million (7B), $1/million (32B)
- **Goose AI**: Varies by provider (uses same rates as Claude/OpenAI/Gemini)
- **Google Gemini**: $3/million tokens (Pro), $0.50/million (Flash), $8/million (1.0 Pro)

---

**Configuration Management:**

We add commands so that users can view the configuration being used, and also clear out credentials they no longer want stored.

```bash
# Show configurations by category
claude-vm cloud list
claude-vm provider list
claude-vm agent list

# Clear specific cloud platform credentials
claude-vm cloud clear-config digitalocean
# Output: Cleared DigitalOcean credentials from keychain

# Clear specific LLM provider credentials (affects multiple agents)
claude-vm provider clear-config anthropic
# Output: Cleared Anthropic API key (affects claude, goose)

claude-vm provider clear-config openai  
# Output: Cleared OpenAI API key (affects codex, goose)

# Clear all credentials (run individual commands)
# No single command - use specific commands as needed
```

### Secret Storage Strategy

**API Provider Credentials (shared by multiple agents):**

Store API provider credentials in OS keychain with shared keys:

```bash
# Keychain Storage (Secure) - Provider-Based
claude-vm.anthropic.api_key      # API key for Anthropic provider
claude-vm.anthropic.oauth_token  # OAuth JSON token for Anthropic provider
claude-vm.openai.api_key         # API key for OpenAI provider
claude-vm.openai.oauth_token     # OAuth JSON token for OpenAI provider
claude-vm.google.api_key         # API key for Google provider
claude-vm.alibaba.api_key        # API key for Alibaba/DashScope provider
claude-vm.groq.api_key           # API key for Groq provider
claude-vm.databricks.api_key     # API key for Databricks provider
claude-vm.openrouter.api_key     # API key for OpenRouter provider

# Agent → Provider Mapping:
# claude:  uses anthropic provider (prefers oauth_token over api_key)
# codex:   uses openai provider (prefers oauth_token over api_key)
# qwen:    uses alibaba provider
# gemini:  uses google provider
# goose:   uses configured provider (anthropic, openai, google, groq, etc.)
```

**Benefits of Shared Storage:**
- One API key enables multiple agents (e.g., ANTHROPIC_API_KEY → claude + goose-anthropic)
- Cleaner credential management - fewer keys to maintain
- Natural grouping by API provider rather than individual tools

---

### Config File Format

```yaml
# ~/.claude-vm/config.yaml
# IMPORTANT: This file contains NO secrets - all secrets are stored in OS keychain

# Default settings
defaults:
  cloud: "docker"                 # Default cloud platform: docker, fly, aws, digitalocean
  agent: "claude"                  # Default coding agent for new-chat

# Cloud platform configuration (non-secrets only)
clouds:
  docker:
    # No configuration needed - uses local Docker daemon
    
  digitalocean:
    region: "nyc1"                 # Default region
    size: "s-2vcpu-8gb"           # Default machine size
    # API key stored in keychain: claude-vm.digitalocean.api_key
    
  fly:
    region: "iad"                  # Default region  
    size: "shared-cpu-2x@2048"     # Default machine size
    # API key stored in keychain: claude-vm.fly.api_key
    
  aws:
    region: "us-east-1"            # Default region
    instance_type: "t3.medium"     # Default instance type
    # Access keys stored in keychain: 
    # - claude-vm.aws.access_key_id
    # - claude-vm.aws.secret_access_key

  google:
    project: ""                    # Optional GCP project for VM hosting
    region: "us-central1"          # Default region for Compute Engine
    # Credentials stored in keychain:
    # - claude-vm.google.gcp_credentials (JSON service account key)

# LLM API provider configuration (non-secrets only)  
providers:
  anthropic:
    # Credentials stored in keychain:
    # - claude-vm.anthropic.api_key (API key authentication)
    # - claude-vm.anthropic.oauth_token (OAuth JSON token - preferred by claude)
    
  openai:
    organization: ""               # Optional org ID
    # Credentials stored in keychain:
    # - claude-vm.openai.api_key (API key authentication)
    # - claude-vm.openai.oauth_token (OAuth JSON token - preferred by codex)
    
  google:
    project: ""                    # Optional GCP project for AI APIs
    # Credentials stored in keychain:
    # - claude-vm.google.api_key (API key for Gemini)
    
  alibaba:
    # Credentials stored in keychain:
    # - claude-vm.alibaba.api_key (DashScope API key for qwen)
    
  groq:
    # Credentials stored in keychain:
    # - claude-vm.groq.api_key (Groq API key)

# Agent configuration (agents use their provider's credentials)
# NOTE: All agents support --option key=value pattern for flexible configuration
agents:
  claude:
    provider: "anthropic"          # Uses anthropic provider credentials
    model: "sonnet"                # opus, sonnet, haiku
    auth_preference: "oauth"       # oauth, api-key (prefers OAuth)
    temperature: 0.7               # 0.0-2.0 (default for all agents)
    
  codex:
    provider: "openai"             # Uses openai provider credentials
    model: "gpt-4o"                # gpt-4o, gpt-4-turbo, o1-mini, o1-preview
    auth_preference: "oauth"       # oauth, api-key (prefers OAuth/ChatGPT login)
    temperature: 0.7               # 0.0-2.0
    organization: ""               # Optional OpenAI org ID
    
  qwen:
    provider: "alibaba"            # Uses alibaba provider credentials
    model: "qwen3-coder-plus"     # qwen3-coder-plus, qwen-coder-32b, qwen-coder-7b
    temperature: 0.7               # 0.0-2.0
    base_url: ""                   # Custom API endpoint (optional)
    max_tokens: 4096               # Maximum response tokens (optional)
    
  goose:
    provider: "anthropic"          # anthropic, openai, google, groq, databricks, openrouter
    model: "claude-3.5-sonnet"     # Provider-specific model name
    planner_provider: ""           # Separate provider for planning tasks (optional)
    planner_model: ""              # Model for planning tasks (optional)
    temperature: 0.7               # 0.0-2.0
    max_turns: 1000                # Maximum conversation turns
    
  gemini:
    provider: "google"             # Uses google provider credentials
    model: "gemini-2.5-pro"        # gemini-2.5-pro, gemini-2.0-flash, gemini-1.5-pro
    temperature: 0.7               # 0.0-2.0 (standardized default)
    thinking_budget: 1024          # Tokens for reasoning (Gemini 2.5+)
    max_output_tokens: 8192        # Maximum response tokens (optional)
    safety_settings: "default"     # Content safety level (optional)

# Storage configuration (S3-compatible)
storage:
  endpoint: "https://r2.cloudflarestorage.com"  # S3 endpoint
  bucket: "claude-vm-backups"                   # Bucket name
  # Access keys stored in keychain:
  # - claude-vm.storage.access_key_id
  # - claude-vm.storage.secret_access_key

# Workspace defaults
workspace:
  auto_stop_minutes: 60            # Stop workspace after idle time
```

**Keychain Storage Pattern:**
```
Service: claude-vm
Account: <type>.<name>.<field>
Password: <secret_value>

Cloud Platform Examples:
- claude-vm.cloud.digitalocean.api_key
- claude-vm.cloud.fly.api_key
- claude-vm.cloud.aws.access_key_id
- claude-vm.cloud.aws.secret_access_key
- claude-vm.cloud.google.gcp_credentials

LLM Provider Examples:
- claude-vm.provider.anthropic.api_key
- claude-vm.provider.anthropic.oauth_token
- claude-vm.provider.openai.api_key
- claude-vm.provider.openai.oauth_token
- claude-vm.provider.google.api_key
- claude-vm.provider.alibaba.api_key
```
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
claude-vm workspace up https://github.com/user/repo --cloud fly

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

5. **GitHub Access: Three Modes**

   **Mode 1: Managed (Enterprise/Recommended)**
   ```bash
   # User installs claude-vm GitHub App
   claude-vm workspace up . --github-mode=managed
   
   # Our service (claude-vm.com) handles token refresh
   # Container polls our API for fresh tokens every 45 minutes
   # Private key NEVER enters container
   ```
   
   **Mode 2: GitHub CLI Auth (Secure/Self-Contained)**
   ```bash
   # User authenticates locally with GitHub CLI
   gh auth login
   # Selects: HTTPS, authenticate with web browser
   # Token stored securely in OS keychain
   
   # Deploy workspace with gh auth forwarding
   claude-vm workspace up . --github-mode=gh-cli
   
   # Container setup:
   # 1. gh CLI pre-installed in container
   # 2. Secure credential helper configured
   # 3. Token injected into gh's secure storage (not env var)
   # 4. Agent CANNOT see raw token (built-in security)
   
   # Inside container:
   gh auth setup-git  # Configures git to use gh for auth
   # Now git operations work without exposing token:
   git clone https://github.com/user/private-repo  # Works!
   gh pr create --title "Changes"                  # Works!
   # But agent cannot access the actual token value
   ```
   
   **Mode 3: PAT (Simple/Legacy)**
   ```bash
   # User provides Personal Access Token
   claude-vm workspace up . --github-token=ghp_xxx
   
   # WARNING displayed:
   # ⚠️  GitHub PAT will be accessible to AI agents
   # ⚠️  Agent can read/write all repos the PAT can access
   # ⚠️  Consider using a limited-scope PAT
   # ⚠️  Recommended: Use --github-mode=gh-cli instead
   
   # Container gets PAT as environment variable
   export GITHUB_TOKEN=ghp_xxx
   # Agent CAN see this token (trade-off accepted)
   ```

### GitHub CLI Auth Implementation Strategy (Mode 2 Details)

**How gh Stores Credentials Securely:**
```yaml
Local Storage:
  macOS: Keychain (via security command)
  Linux: Secret Service API or encrypted file
  Windows: Windows Credential Manager
  
File Locations:
  Config: ~/.config/gh/config.yml        # Non-sensitive settings
  Hosts: ~/.config/gh/hosts.yml          # Encrypted token storage
  Token: Never in plaintext, always encrypted
  
Security Properties:
  - Token encrypted at rest
  - Only gh binary can decrypt
  - Other processes cannot read (including Claude Code)
  - No environment variable exposure
```

**Implementation: Secure Token Transfer to Container**

**Step 1: Extract Token from Local gh Installation**
```bash
# claude-vm extracts token using gh's own API (not reading files)
gh_token=$(gh auth token)
# This is the ONLY time we have raw token in memory

# Alternative: Use gh's OAuth device flow for fresh token
gh_token=$(gh auth login --with-token < /dev/null 2>&1 | 
           gh auth token)
```

**Step 2: Simplified Token Transfer (Industry Standard)**

All providers use the same simple approach following industry standards from GitHub Codespaces, GitPod, and DevPod:

```bash
# Extract token from local gh CLI  
gh_token=$(gh auth token)

# Pass as environment variable during container creation (universal for all providers)
export GH_TOKEN_TEMP="$gh_token"
claude-vm workspace up . --cloud <any-cloud>
```

**Step 3: Universal Container Setup Process**
```bash
#!/bin/bash
# Container entrypoint script (same for all providers)

# 1. Install gh CLI if not present
if ! command -v gh &> /dev/null; then
    curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
      gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | \
      tee /etc/apt/sources.list.d/github-cli.list > /dev/null
    apt update && apt install gh -y
fi

# 2. Configure gh secure storage location
export GH_CONFIG_DIR=/workspace/.config/gh
mkdir -p $GH_CONFIG_DIR
chmod 700 $GH_CONFIG_DIR

# 3. Import token (universal method)
if [ -n "$GH_TOKEN_TEMP" ]; then
    echo "$GH_TOKEN_TEMP" | gh auth login --with-token
    unset GH_TOKEN_TEMP  # Remove from environment immediately
    # Token now safely stored in gh's encrypted storage
fi

# 4. Setup git to use gh for authentication
gh auth setup-git

# 5. Verify setup
gh auth status

# 6. Security wrapper to prevent token extraction
cat > /usr/local/bin/gh-wrapper.sh << 'WRAPPER'
#!/bin/bash
if [[ "$1" == "auth" && "$2" == "token" ]]; then
    echo "Error: Direct token access disabled for security" >&2
    exit 1
fi
exec /usr/bin/gh "$@"
WRAPPER
chmod +x /usr/local/bin/gh-wrapper.sh
alias gh='/usr/local/bin/gh-wrapper.sh'
```

**Step 4: Security Guarantees & Token Refresh**
```yaml
Token Protection:
  - Brief environment variable exposure during container startup only
  - Immediately moved to gh's encrypted storage 
  - Environment variable cleared after import
  - Wrapper prevents token extraction via 'gh auth token'
  
Agent Restrictions:
  - Cannot read ~/.config/gh/hosts.yml (encrypted by gh CLI)
  - Cannot run 'gh auth token' (wrapper blocks it)
  - Can only use git/gh commands with stored auth
  
Token Refresh:
  - Manual: claude-vm workspace refresh-auth <workspace-id>
  - Automatic: gh auth login --web (OAuth device flow in container)
```

**Advantages of gh CLI Auth Mode:**
- ✅ Token never visible to AI agents
- ✅ Same security as local development
- ✅ No token in environment variables
- ✅ Supports all gh and git operations
- ✅ Token encrypted at rest in container
- ✅ Works with GitHub Enterprise
- ✅ Automatic git credential helper setup

**Limitations:**
- ⚠️ Requires gh CLI in container (adds ~50MB)
- ⚠️ Token expires (need refresh mechanism)
- ⚠️ Initial transfer briefly has token in memory
- ⚠️ Complex implementation vs simple PAT

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
claude-vm workspace up https://github.com/user/repo --cloud fly

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

**GitHub Apps vs PAT:**
- **GitHub Apps** (recommended): Repository-specific, 1-hour tokens, audit trail
- **Personal Access Tokens** (fallback): Broader permissions, no expiration

---

### DevPod Import

If the user has devpod installed, we can probably just fetch configurations directly from there for some providers.
