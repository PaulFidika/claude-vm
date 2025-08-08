# Comparison: claude-vm vs Similar Projects

## Project Overview

### 1. **Claude Code Sandbox** (textcortex)
- **Approach**: Local Docker + Web UI
- **Key Features**:
  - Browser-based terminal with real-time streaming
  - Automatic credential discovery and forwarding
  - Commit monitoring with notifications
  - Multiple container support
  - Interactive git operations menu
- **Architecture**: Docker containers on local machine with web interface

### 2. **TSK** (Task Manager and Sandbox)
- **Approach**: Rust CLI for AI agent task delegation
- **Key Features**:
  - Agents work in isolated Docker containers
  - Returns git branches for review
  - Parallel task execution (up to 4)
  - Task templates for different types
  - Server mode for continuous processing
- **Architecture**: Local Docker orchestration with task queue

### 3. **Claude Squad** (smtg-ai)
- **Approach**: Terminal multiplexer for multiple AI agents
- **Key Features**:
  - Manages multiple agents (Claude Code, Aider, Codex, etc.)
  - Uses tmux for terminal isolation
  - Git worktrees for branch isolation
  - Background task completion
  - Simple TUI interface
- **Architecture**: Local tmux sessions with git worktrees

### 4. **Dagger Container Use**
- **Approach**: MCP server for agent development environments
- **Key Features**:
  - Fresh container per agent with own git branch
  - Real-time visibility into agent actions
  - Direct terminal intervention capability
  - Universal compatibility (any agent/model)
  - Powered by Dagger and git worktrees
- **Architecture**: MCP protocol with Dagger containers

### 5. **Our claude-vm**
- **Approach**: Remote VM with IDE Remote-SSH
- **Key Features**:
  - Fly.io cloud deployment
  - Remote-SSH for Cursor/Windsurf/VS Code
  - Tar upload for local code
  - Git master on VM
  - Safe from local damage
- **Architecture**: Cloud VM with SSH access

## Comparison Matrix

| Feature | Claude Sandbox | TSK | Claude Squad | Dagger Use | **claude-vm** |
|---------|---------------|-----|--------------|------------|---------------|
| **Execution** | Local Docker | Local Docker | Local Process | Local Docker | **Remote VM** |
| **Multi-agent** | ✅ Multiple containers | ✅ Parallel tasks | ✅ Multiple agents | ✅ Per-agent containers | ❌ Single Claude |
| **Interface** | Web UI | CLI | TUI (tmux) | CLI/MCP | **IDE Remote-SSH** |
| **Git Strategy** | Monitor commits | Return branches | Git worktrees | Branch per agent | **VM as master** |
| **Isolation** | Docker | Docker | tmux + worktrees | Docker + branches | **Full VM** |
| **Remote Access** | ❌ Local only | ❌ Local only | ❌ Local only | ❌ Local only | ✅ **SSH from anywhere** |
| **IDE Integration** | ❌ Web terminal | ❌ CLI only | ❌ Terminal only | ❌ MCP protocol | ✅ **Native IDE** |
| **Resource Usage** | Local CPU/RAM | Local CPU/RAM | Local CPU/RAM | Local CPU/RAM | **Cloud resources** |
| **Offline Work** | ✅ Fully local | ✅ Fully local | ✅ Fully local | ✅ Fully local | ❌ **Needs internet** |

## Key Differentiators

### What Makes claude-vm Unique:

1. **Remote-First Design**
   - Others are local Docker solutions
   - claude-vm runs on cloud infrastructure
   - Can close laptop and Claude keeps working

2. **IDE-Native Experience**
   - Others use CLI/TUI/Web interfaces
   - claude-vm uses native IDE features (Cursor/Windsurf)
   - Full debugging, extensions, etc.

3. **No Local Resources**
   - Others consume local CPU/RAM
   - claude-vm offloads to cloud
   - Better for long-running tasks

4. **True Isolation**
   - Others use Docker (still on your machine)
   - claude-vm uses separate VM (can't touch local)
   - Nuclear option: delete VM if Claude goes rogue

### What Others Do Better:

1. **Multi-Agent Orchestration**
   - TSK, Claude Squad, Dagger excel at multiple agents
   - claude-vm focuses on single Claude instance
   - Could be future enhancement

2. **Offline Development**
   - All others work fully offline
   - claude-vm requires internet connection
   - Trade-off for remote execution

3. **Simpler Setup**
   - Others just need Docker
   - claude-vm needs cloud account, SSH setup
   - More complex initial configuration

## Use Case Alignment

| Project | Best For |
|---------|----------|
| **Claude Sandbox** | Quick local experiments with web UI preference |
| **TSK** | Delegating specific tasks to AI, reviewing PRs |
| **Claude Squad** | Managing multiple AI agents simultaneously |
| **Dagger Use** | Team environments with MCP protocol |
| **claude-vm** | Long-running tasks, remote development, IDE users |

## Hybrid Possibilities

claude-vm could adopt features from others:
- **From TSK**: Task queue system for multiple jobs
- **From Claude Squad**: Multiple VM management
- **From Dagger**: MCP protocol support
- **From Sandbox**: Web UI for monitoring

## Conclusion

While all projects solve "safe Claude execution," they target different workflows:
- **Local multi-agent**: Claude Squad, TSK, Dagger
- **Local single-agent**: Claude Sandbox  
- **Remote single-agent**: claude-vm (our approach)

claude-vm is uniquely positioned for developers who:
1. Want to use their favorite IDE (Cursor/Windsurf)
2. Need long-running tasks without local resources
3. Prefer cloud-based development
4. Value true isolation from local system