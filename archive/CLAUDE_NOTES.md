# Claude Code Session History Storage

## Overview
Claude Code stores session history and related data in a structured directory within the user's home folder.

## Storage Locations

### Primary Session Storage
- **Location**: `~/.claude-code/sessions/`
- **Format**: Each session is stored in a timestamped directory
- **Example**: `~/.claude-code/sessions/2025-08-04-123456/`

### Session Directory Structure
```
~/.claude-code/sessions/<session-id>/
├── conversation.json      # Full conversation history
├── metadata.json         # Session metadata (start time, model, etc.)
├── files/               # Copies of files accessed/modified
├── git-state.json       # Git repository state snapshots
└── tool-calls.log       # Log of all tool invocations
```

### Additional Storage Locations
- **Config**: `~/.claude-code/config.json` - User preferences and settings
- **Cache**: `~/.claude-code/cache/` - Cached responses and file contents
- **Logs**: `~/.claude-code/logs/` - Application logs and debug info

## Session Resume Functionality
When using `claude --resume`, Claude Code:
1. Finds the most recent session in `~/.claude-code/sessions/`
2. Loads the conversation history from `conversation.json`
3. Restores the working directory and git state
4. Continues the conversation with full context

## Session History Format
The `conversation.json` file contains:
```json
{
  "messages": [
    {
      "role": "user",
      "content": "...",
      "timestamp": "2025-08-04T12:34:56Z"
    },
    {
      "role": "assistant",
      "content": "...",
      "tools_used": ["Read", "Edit"],
      "timestamp": "2025-08-04T12:35:10Z"
    }
  ],
  "working_directory": "/path/to/project",
  "environment": {
    "platform": "linux",
    "claude_version": "1.0.0"
  }
}
```

## Integration Considerations for Remote Sessions
For the claude-vm implementation, session data could be:
1. Stored in a centralized database (PostgreSQL/Redis)
2. Synchronized between local and remote environments
3. Accessed via API endpoints for session listing/resumption
4. Streamed in real-time to connected clients