# Claude CLI Integration - What Changed

## Summary

Replaced the **simulated AI command execution** with **real Claude CLI integration**. Commands are now executed by Claude CLI running in `/workspace/code`, with real-time output streaming back to clients.

---

## Major Changes

### 1. **Real AI Execution** ✨

**Before:**
- Simulated processing phases
- Fake tool executions (read_file, parse_html, etc.)
- Generated mock results
- No actual AI involved

**After:**
- Real Claude CLI process execution
- Actual file modifications in `/workspace/code`
- Real AI responses streamed live
- Genuine code changes

---

### 2. **Process Management**

**New in `ai_command.go`:**

```go
// Spawn Claude CLI process
cmd := exec.CommandContext(session.Context, "claude", prompt)
cmd.Dir = "/workspace/code"

// Stream stdout/stderr concurrently
go readStdout()
go readStderr()

// Handle interrupts
cmd.Wait() // Blocks until Claude finishes or is killed
```

**Key Features:**
- Context-based cancellation
- Concurrent output reading
- Real-time streaming to WebSocket
- Process cleanup on interrupt

---

### 3. **Removed Simulation Code**

**Deleted Functions:**
- `getToolsForScope()` - Generated fake tools
- `generateResult()` - Created mock results
- Simulated processing phases
- Fake delays (`time.Sleep`)

**Why:**
These are no longer needed because Claude handles everything.

---

### 4. **Working Directory**

All Claude CLI commands run in:
```
/workspace/code
```

This is where Claude can:
- Read project files
- Create new files
- Modify existing code
- Run commands

---

### 5. **Logging Updates**

**Console Output:**

```
📥 AI Command Received: "Add a contact form" | Scope: current-page | Page: /contact
🔄 Processing Command [cmd_123]: "Add a contact form" | Scope: current-page | Page: /contact
🤖 Calling Claude CLI with prompt: Scope: current-page | Page: /contact | Task: Add a contact form
✅ Claude CLI process started
📤 Claude: [Real Claude output streams here]
✅ Command Completed [cmd_123]: 5.32s
```

**What You See:**
- User's original command
- Formatted prompt sent to Claude
- Claude's actual responses (line by line)
- Final execution time

---

### 6. **WebSocket Messages**

**Message Types:**

| Type | Before | After |
|------|--------|-------|
| `status` | "Analyzing..." | "Starting Claude CLI..." |
| `thinking` | Simulated phases | _(Not used currently)_ |
| `tool_use` | Fake tools | _(Not used currently)_ |
| `output` | _(Not used)_ | **Real Claude output!** |
| `result` | Mock data | Actual result summary |
| `error` | Simulated errors | Real errors from Claude |
| `complete` | Always success | Based on actual exit code |

---

### 7. **Error Handling**

**New Error Types:**

```go
// Claude not installed
"exec: \"claude\": executable file not found in $PATH"

// Authentication required
"Please run 'claude' to authenticate"

// Working directory doesn't exist
"chdir /workspace/code: no such file or directory"

// Claude execution error
"exit status 1"
```

All errors are:
- Logged to console
- Saved to database
- Sent to client via WebSocket

---

## Configuration

### Current Settings

| Setting | Value | Location |
|---------|-------|----------|
| CLI Command | `claude` | `ai_command.go:265` |
| Working Dir | `CLAUDE_WORKSPACE_DIR` env variable | `ai_command.go:87-92` |
| Default Dir | `/workspace/code` | `ai_command.go:91` |
| Prompt Format | `Scope: X \| Page: Y \| Task: Z` | `ai_command.go:415-425` |

### Easy to Change

**Working Directory (Recommended):**
```bash
export CLAUDE_WORKSPACE_DIR=/path/to/your/project
```

**Code Changes:**
```go
// Use different AI tool
cmd := exec.CommandContext(ctx, "gpt", prompt)  // Instead of "claude"

// Change default directory (in getWorkspaceDir function)
return "/your/custom/path"  // Change default

// Add CLI arguments
cmd := exec.CommandContext(ctx, "claude", "--model", "opus", prompt)
```

---

## Requirements

### To Use This Feature:

1. **Install Claude CLI**
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```

2. **Authenticate**
   ```bash
   claude  # Opens browser for login
   ```

3. **Create Workspace**
   ```bash
   mkdir -p /workspace/code
   ```

4. **Have Claude Subscription**
   - Claude Pro or Claude Max required

---

## Testing

### Quick Test

```bash
# 1. Start server
cd backend && go run .

# 2. Send command (in another terminal)
curl -X POST http://localhost:9000/api/ai/command \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "List the files in the workspace",
    "scope": "current-page",
    "context": {"page": "/", "timestamp": "2025-10-20T10:00:00Z"}
  }'

# 3. Watch server logs - you'll see real Claude output!
```

### Expected Output

```
📥 AI Command Received: "List the files in the workspace" | ...
🔄 Processing Command [cmd_...]: "List the files in the workspace" | ...
🤖 Calling Claude CLI with prompt: ...
✅ Claude CLI process started
📤 Claude: Listing files in /workspace/code:
📤 Claude: - file1.txt
📤 Claude: - file2.js
📤 Claude: - README.md
✅ Command Completed [cmd_...]: 2.15s
```

---

## What Stays the Same

### Unchanged Features:

- ✅ WebSocket streaming architecture
- ✅ Database persistence
- ✅ Interrupt capability
- ✅ Session management
- ✅ Frontend API (same endpoints)
- ✅ Client libraries (React/Vue examples)
- ✅ Web UI (`ai-command-client.html`)

The **API contract remains identical** - frontend doesn't need changes!

---

## Migration Notes

### For Developers:

**If you were relying on simulated tool logs:**
- `command-summary.md` is no longer populated by simulated tools
- Instead, watch Claude's actual output via:
  - Server console logs (`📤 Claude: ...`)
  - WebSocket `output` messages
  - Database `result` field

**If you customized `getToolsForScope()`:**
- This function is removed
- Claude now decides which tools/actions to use
- You can influence this via the prompt format

---

## Benefits

### Why This Is Better:

1. **Real AI Capabilities** - Actual Claude intelligence, not fake responses
2. **Accurate Results** - Claude modifies real files, not simulated actions
3. **Streaming Output** - See Claude "think" in real-time
4. **Interrupt Works** - Actually kills the Claude process
5. **Extensible** - Easy to swap Claude for other AI tools later

---

## Future Enhancements

### Coming Soon:

- **Tool Support** - Let Claude use tools/MCPs
- **Custom Instructions** - Add project context
- **Session Persistence** - Resume conversations
- **Multiple AI Providers** - GPT, Gemini, etc.
- **Agent Support** - Custom Claude agents

---

## Files Changed

| File | Status | Changes |
|------|--------|---------|
| `ai_command.go` | **Modified** | Replaced simulation with CLI execution |
| `command_logger.go` | Unchanged | Kept for future tool tracking |
| `main.go` | Unchanged | Routes work the same |
| `db.go` | Unchanged | Database schema unchanged |

---

## Summary

**In One Sentence:**
Commands now execute real Claude CLI instead of being simulated, with all output streamed live to clients.

**The Best Part:**
The frontend doesn't change at all - same API, same WebSocket messages, just with **real AI** now! 🎉

---

For complete details, see **`CLAUDE-CLI-INTEGRATION.md`**
