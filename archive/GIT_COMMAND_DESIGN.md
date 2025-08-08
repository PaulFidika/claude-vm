# Git Command Design for claude-vm

## Option 1: Direct CLI Proxy
Standalone commands that proxy to remote VM:

```bash
# Usage
claude-vm git status
claude-vm git commit -m "Fixed auth"
claude-vm git push

# Or shortened aliases
claude-vm status
claude-vm commit "Fixed auth"
claude-vm push

# Implementation
func handleGitCommand(args []string) {
    result := sshClient.Exec("git " + strings.Join(args, " "))
    fmt.Print(result)
}
```

## Option 2: Inside Claude's Terminal
Special commands while connected to Claude session:

```bash
# While in claude-vm session
Connected to Claude session...
> /status              # Shows git status
> /commit Fixed auth bug
> /push
> /diff README.md

# Implementation
func handleSessionInput(input string) {
    if strings.HasPrefix(input, "/") {
        cmd := parseSlashCommand(input)
        executeRemoteGit(cmd)
    } else {
        sendToClaude(input)
    }
}
```

## Option 3: SSH Terminal (Recommended)
Direct SSH connection where ALL commands run remotely:

```bash
# Connect to VM
$ claude-vm
Connecting to session-123...
[session-123]$ pwd
/workspace

# Now you're IN the remote VM
[session-123]$ git status       # Runs on VM
[session-123]$ git commit -m "Fixed auth"
[session-123]$ ls -la          # All commands are remote
[session-123]$ claude "continue working on auth"

# Meanwhile, files are synced locally via Mutagen
# Edit in your IDE, changes appear in VM instantly
```

### Implementation
```go
func connectSession(sessionID string) {
    // Start Mutagen sync
    mutagen.Create(
        "vm:/workspace",
        "~/claude-sessions/session-123",
        "--ignore-vcs",
    )
    
    // Open SSH session
    ssh := exec.Command("ssh", 
        "-t",  // Allocate PTY
        fmt.Sprintf("vm-%s", sessionID),
        "cd /workspace && bash",
    )
    ssh.Stdin = os.Stdin
    ssh.Stdout = os.Stdout
    ssh.Stderr = os.Stderr
    ssh.Run()
}
```

## Architecture

```
┌─[Your Machine]──────────────┐     ┌─[Remote VM]────────────┐
│                             │     │                        │
│ Terminal: claude-vm ────────┼─SSH─┼→ Bash shell           │
│                             │     │  └→ git commands      │
│                             │     │  └→ claude commands   │
│                             │     │                        │
│ ~/claude-sessions/123/ ←────┼─────┼─ /workspace           │
│ (Mutagen sync)              │     │  (Source of truth)     │
│                             │     │                        │
│ Your IDE edits these files  │     │  .git/ lives here      │
└─────────────────────────────┘     └────────────────────────┘
```

## Why Option 3 is Best

1. **Natural workflow** - Just type `git commit` like normal
2. **No command parsing** - SSH handles everything
3. **Full access** - Any command works (ls, cat, npm, etc)
4. **Claude integration** - Can run `claude` commands directly
5. **File editing** - Edit locally via Mutagen sync

## Workflow Example

```bash
# Start session
$ claude-vm start --repo github.com/user/repo
Creating session-123...
Starting Mutagen sync to ~/claude-sessions/session-123
Connecting...

[session-123]$ git status
On branch claude-session-123
nothing to commit

[session-123]$ claude "implement the auth TODO"
[Claude starts working...]

# In another terminal/IDE
$ code ~/claude-sessions/session-123
# Edit files, they sync to VM instantly

[session-123]$ git add .
[session-123]$ git commit -m "Implemented auth with Claude"
[session-123]$ git push origin claude-session-123

[session-123]$ exit
Stopping Mutagen sync...
Session suspended. Resume with: claude-vm connect session-123
```