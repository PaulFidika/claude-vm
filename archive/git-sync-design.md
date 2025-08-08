# Git-Based Sync for claude-vm

## The Pattern: VM as Git Remote

```bash
# Setup (happens automatically)
git remote add claude-vm-abc123 git@vm-abc123:/workspace/.git

# Your workflow:
$ claude-vm start
Starting session abc123...
Pushing current state to VM...
[claude-vm automatically does: git push claude-vm-abc123 HEAD:refs/heads/claude-work]

$ claude-vm status abc123
Claude made 3 commits:
  - abc1234 Implemented user authentication
  - def5678 Added error handling
  - ghi9012 Updated tests

$ git fetch claude-vm-abc123
$ git log claude-vm-abc123/claude-work  # See what Claude did
$ git diff HEAD..claude-vm-abc123/claude-work  # See all changes
$ git merge claude-vm-abc123/claude-work  # Accept changes
```

## Key Features

1. **Automatic WIP commits**: When you start claude-vm, it creates a WIP commit of your current state
2. **Branch isolation**: Claude works on `claude-work` branch, you stay on your branch
3. **Real-time sync**: Post-commit hooks push Claude's commits back immediately
4. **Conflict resolution**: Use git's merge strategies when both edited same files
5. **Cherry-pick specific changes**: Take only some of Claude's commits if needed

## Implementation Details

### VM Setup Script
```bash
#!/bin/bash
# Runs when VM starts

# Setup git repo
git init --bare /repos/session.git
git clone /repos/session.git /workspace
cd /workspace

# Configure hooks for auto-push
cat > .git/hooks/post-commit << 'EOF'
#!/bin/bash
git push origin HEAD:claude-work --force-with-lease
EOF
chmod +x .git/hooks/post-commit

# Start Claude with auto-commit wrapper
alias claude='claude-with-commit'
claude-with-commit() {
    command claude "$@"
    # After Claude runs, check for changes
    if [[ -n $(git status -s) ]]; then
        git add -A
        git commit -m "Claude: $1" --author="Claude <claude@vm>"
        # Hook auto-pushes
    fi
}
```

### Benefits Over Custom Sync

1. **No new concepts**: Developers already know git
2. **Audit trail**: Every change is a commit with author
3. **Selective integration**: Cherry-pick, rebase, or merge as needed
4. **Conflict resolution**: Git's three-way merge handles conflicts
5. **Works offline**: Can fetch changes and review later

### Advanced: Real-time Collaborative Editing

For true CRDT-style collaboration while Claude is working:

```bash
# Use git worktree for live view
git worktree add ~/claude-live claude-vm-abc123/claude-work

# FSWatch + auto-pull for near real-time
fswatch ~/claude-live | xargs -n1 -I{} git -C ~/claude-live pull

# Or use git-sync tool
git-sync --repo ~/claude-live --ref claude-vm-abc123/claude-work --period 1s
```

This gives you a live view of Claude's changes without affecting your working directory.