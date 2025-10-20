# Claude CLI Integration

## Overview

The AI Command API now integrates with **Claude CLI** to execute real AI commands instead of simulating them. When a user submits a command through the API, it is passed directly to the Claude CLI running in the `/workspace/code` directory.

## Requirements

### 1. Install Claude CLI

```bash
npm install -g @anthropic-ai/claude-code
```

### 2. Login to Claude

```bash
claude
```

The first time you run `claude`, it will open a browser window for authentication. You need an active Claude Pro or Claude Max subscription.

### 3. Configure Workspace Directory

Set the `CLAUDE_WORKSPACE_DIR` environment variable to specify where Claude should run:

```bash
# Set workspace directory
export CLAUDE_WORKSPACE_DIR=/path/to/your/project

# Or use default (/workspace/code)
mkdir -p /workspace/code
cd /workspace/code
# Initialize your project here
```

**Default:** If `CLAUDE_WORKSPACE_DIR` is not set, it defaults to `/workspace/code`

## How It Works

### 1. User Submits Command

```bash
POST /api/ai/command
{
  "prompt": "Add a contact form",
  "scope": "current-page",
  "context": {
    "page": "/contact",
    "timestamp": "2025-10-20T10:00:00Z"
  }
}
```

### 2. Server Processes Request

The server:
1. Receives the command and creates a database record
2. Builds a Claude-formatted prompt with context
3. Gets workspace directory from `CLAUDE_WORKSPACE_DIR` env variable (or uses default)
4. Spawns Claude CLI process in the workspace directory
5. Streams output in real-time via WebSocket

### 3. Claude CLI Execution

```go
workspaceDir := getWorkspaceDir() // From CLAUDE_WORKSPACE_DIR or default
cmd := exec.CommandContext(ctx, "claude", prompt)
cmd.Dir = workspaceDir
```

The Claude CLI:
- Runs in the configured workspace directory
- Has access to all project files in that directory
- Can read, edit, create files
- Streams output back to the server

### 4. Output Streaming

All Claude output is streamed to the client:
- Stdout → WebSocket as `output` messages
- Stderr → WebSocket as `output` messages (prefixed with `[stderr]`)
- Server logs → Console for debugging

## Command Format

### Basic Prompt

User sends: `"Add a contact form"`

Claude receives:
```
Scope: current-page | Page: /contact | Task: Add a contact form
```

This gives Claude context about:
- **Scope**: Where to apply changes (current-page, new-page, global)
- **Page**: Which page to modify
- **Task**: What to do

### Customizing Prompts

You can modify the `buildClaudePrompt()` function in `ai_command.go` to change how prompts are formatted:

```go
func buildClaudePrompt(command *AICommand) string {
    // Custom prompt formatting
    return fmt.Sprintf("...", command.Prompt, command.Page)
}
```

## Logging

### Console Logs (stdout)

```
📥 AI Command Received: "Add a contact form" | Scope: current-page | Page: /contact
🔄 Processing Command [cmd_123]: "Add a contact form" | Scope: current-page | Page: /contact
🤖 Calling Claude CLI with prompt: Scope: current-page | Page: /contact | Task: Add a contact form | Workspace: /workspace/code
✅ Claude CLI process started
📤 Claude: [Output from Claude appears here line by line]
✅ Command Completed [cmd_123]: 5.32s
```

### WebSocket Stream

Clients receive real-time updates:

```javascript
{
  "type": "output",
  "timestamp": "2025-10-20T10:00:05Z",
  "data": "Creating contact form component..."
}
```

## Interruption

Users can interrupt long-running Claude sessions:

### Via WebSocket
```javascript
ws.send(JSON.stringify({ type: 'interrupt' }));
```

### Via HTTP
```bash
curl -X POST http://localhost:9000/api/ai/command/{commandId}/interrupt
```

The interrupt will:
1. Cancel the Go context
2. Kill the Claude CLI process
3. Mark command as "interrupted" in database
4. Send status update to client

## Error Handling

### Claude CLI Not Found

```
❌ Command Failed [cmd_123]: exec: "claude": executable file not found in $PATH
```

**Solution:** Install Claude CLI globally:
```bash
npm install -g @anthropic-ai/claude-code
```

### Authentication Required

```
⚠️ Claude stderr: Please run 'claude' to authenticate
```

**Solution:** Run `claude` manually once to authenticate:
```bash
claude
# Follow browser authentication flow
```

### Working Directory Not Found

```
❌ Error [cmd_123]: chdir /workspace/code: no such file or directory
```

**Solution:** Create the workspace directory:
```bash
mkdir -p /workspace/code
```

## Customization

### Change Working Directory

**Recommended:** Use the `CLAUDE_WORKSPACE_DIR` environment variable:

```bash
export CLAUDE_WORKSPACE_DIR=/path/to/your/project
```

**Alternative:** Edit the default in `ai_command.go` function `getWorkspaceDir()`:

```go
func getWorkspaceDir() string {
    if dir := os.Getenv("CLAUDE_WORKSPACE_DIR"); dir != "" {
        return dir
    }
    return "/your/custom/path"  // Change default here
}
```

### Change Claude Command

To use different AI tools in the future:

```go
// Current:
cmd := exec.CommandContext(ctx, "claude", prompt)

// For other tools:
cmd := exec.CommandContext(ctx, "gpt", prompt)  // Example
cmd := exec.CommandContext(ctx, "aider", prompt)  // Example
```

### Add Command Arguments

```go
cmd := exec.CommandContext(ctx, "claude", "--model", "claude-3-opus-20240229", prompt)
```

## Testing

### 1. Start Server

```bash
# Set workspace directory (optional - defaults to /workspace/code)
export CLAUDE_WORKSPACE_DIR=/path/to/your/project

cd backend
go run .
```

### 2. Send Test Command

```bash
curl -X POST http://localhost:9000/api/ai/command \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "List files in the workspace",
    "scope": "current-page",
    "context": {
      "page": "/",
      "timestamp": "2025-10-20T10:00:00Z"
    }
  }'
```

### 3. Watch Server Logs

You'll see:
```
📥 AI Command Received: "List files in the workspace" | Scope: current-page | Page: /
🔄 Processing Command [cmd_...]: "List files in the workspace" | Scope: current-page | Page: /
🤖 Calling Claude CLI with prompt: Scope: current-page | Page: / | Task: List files in the workspace
✅ Claude CLI process started
📤 Claude: [Claude's response appears here]
✅ Command Completed [cmd_...]: 3.45s
```

### 4. Use Web Client

Open `ai-command-client.html` and test interactively:
- Real-time output streaming
- See Claude's responses as they happen
- Interrupt if needed

## Current Limitations

### What's NOT Included Yet (By Design)

1. **No Agents** - Claude CLI runs without custom agents
2. **No Tools** - No tool/MCP configuration yet
3. **No Custom Instructions** - Using default Claude behavior

These will be added in future iterations.

### What IS Included

- ✅ Real Claude CLI execution
- ✅ Working directory set to `/workspace/code`
- ✅ Real-time output streaming
- ✅ Interrupt capability
- ✅ Error handling
- ✅ Logging to stdout
- ✅ WebSocket streaming to clients

## Next Steps

Future enhancements to consider:

1. **Add Tool Support** - Enable Claude to use tools
2. **Add MCP Servers** - Connect external services
3. **Custom Instructions** - Add project-specific context
4. **Multiple AI Providers** - Support GPT, Gemini, etc.
5. **Session Persistence** - Resume Claude conversations
6. **File Watching** - Detect when Claude modifies files

## Architecture

```
User Request
    ↓
POST /api/ai/command
    ↓
Create DB Record
    ↓
Build Claude Prompt
    ↓
Spawn Claude CLI Process
    ├─ Working Dir: /workspace/code
    ├─ Stdout → WebSocket Stream
    ├─ Stderr → WebSocket Stream
    └─ Exit Code → Result Status
    ↓
Update DB with Result
    ↓
Send Completion to Client
```

## File Locations

- **Integration Code**: `backend/ai_command.go`
  - `processAICommand()` - Main execution function
  - `buildClaudePrompt()` - Prompt formatting
  - `handleCommandError()` - Error handling

- **Configuration**:
  - Working Directory: `getWorkspaceDir()` function (line 87-92) + `CLAUDE_WORKSPACE_DIR` env variable
  - Claude Command: Line 265
  - Prompt Format: Line 415-425

- **Environment Variables**:
  - `CLAUDE_WORKSPACE_DIR` - Working directory (default: `/workspace/code`)

## FAQ

**Q: Can I use a different AI CLI tool?**
A: Yes! Change line 265 to use a different command (e.g., `gpt`, `aider`).

**Q: How do I add custom Claude arguments?**
A: Modify line 265 to add arguments: `exec.CommandContext(ctx, "claude", "--arg", "value", prompt)`

**Q: Where does Claude save modified files?**
A: In the directory specified by `CLAUDE_WORKSPACE_DIR` (or `/workspace/code` by default).

**Q: Can I change the working directory?**
A: Yes, set the `CLAUDE_WORKSPACE_DIR` environment variable: `export CLAUDE_WORKSPACE_DIR=/your/path`

**Q: How do I see what Claude is doing?**
A: Watch the server console logs or the WebSocket stream in the client.

---

**Ready to use!** Start the server and send commands through the API or web client. Claude will execute them in `/workspace/code` and stream the results back in real-time.
