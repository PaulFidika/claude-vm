# Claude VM Design Approaches Comparison

## Approach 1: Binary Wrapper (claude-vm.go)
**How it works:** Single binary that intercepts all commands and decides whether to handle locally or remotely.

**Pros:**
- Single binary distribution
- Works everywhere (no shell dependencies)
- Can intercept all commands transparently

**Cons:**
- Complex argument parsing (need to separate our flags from Claude's)
- Potential conflicts with Claude's future flags
- Always runs through our wrapper (slight overhead)

## Approach 2: SSH-Style Session Manager
**How it works:** `claude-vm` only manages sessions. When connected, you get a modified shell.

**Pros:**
- Clean separation: session management vs command execution
- Familiar model (like SSH)
- No need to parse Claude's arguments
- Clear indication when you're in a VM session

**Cons:**
- Requires shell manipulation
- User needs to understand session concept

## Approach 3: Shell Function Override
**How it works:** Shell functions override the `claude` command when in VM context.

**Pros:**
- Minimal overhead when not using VM
- Very transparent to user
- Easy to implement

**Cons:**
- Shell-specific (needs bash/zsh)
- Requires modifying user's shell config
- Can be confusing (claude behavior changes based on environment)

## Recommendation

I recommend **Approach 2 (SSH-Style)** for these reasons:

1. **Clear mental model**: Users understand "connect to remote, then work normally"
2. **No conflicts**: We don't interfere with Claude's CLI at all
3. **Flexibility**: Can add features like session sharing, read-only access, etc.
4. **Clean implementation**: No complex argument parsing needed

Usage would be:
```bash
# Start new session with repo
$ claude-vm github.com/user/repo
Connected to session abc123
[claude-vm:abc123] $ claude "implement the TODO list"
[claude-vm:abc123] $ claude --resume  # This resumes within the VM
[claude-vm:abc123] $ exit

# Resume later
$ claude-vm abc123
[claude-vm:abc123] $ 
```

This matches how developers already think about remote development (SSH, Docker exec, kubectl exec, etc.).