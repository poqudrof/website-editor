# Environment Variables Reference

## Overview

The backend server uses environment variables for configuration. This allows you to customize behavior without modifying code.

---

## Available Environment Variables

### `CLAUDE_WORKSPACE_DIR`

**Purpose:** Specifies the working directory where Claude CLI will execute commands.

**Default:** `/workspace/code`

**Usage:**
```bash
export CLAUDE_WORKSPACE_DIR=/path/to/your/project
```

**Example:**
```bash
# Set workspace to your Next.js project
export CLAUDE_WORKSPACE_DIR=/home/user/my-nextjs-app

# Start server
cd backend && go run .
```

**When to use:**
- You want Claude to work in a specific project directory
- You have multiple projects and want to switch between them
- You're running in a Docker container with a custom mount path

**Notes:**
- Must be an absolute path
- Directory must exist before starting the server
- Claude will have access to all files in this directory and subdirectories

---

### `LOG_LEVEL`

**Purpose:** Controls the verbosity of logging output. When set to `HIGH`, logs detailed information about Claude CLI calls including full command details, all environment variables, and complete stdout/stderr output.

**Default:** Normal logging (not set)

**Valid Values:**
- `HIGH` - Enables detailed logging with full Claude API call details
- (unset) - Normal logging level

**Usage:**
```bash
export LOG_LEVEL=HIGH
```

**Example:**
```bash
# Enable high-level logging
export LOG_LEVEL=HIGH

# Start server
cd backend && go run .
```

**What gets logged with LOG_LEVEL=HIGH:**
1. **Startup Banner:**
   ```
   🔍 [HIGH LOG] ================================
   🔍 [HIGH LOG] HIGH LOGGING ENABLED
   🔍 [HIGH LOG] All Claude API calls and responses will be logged in detail
   🔍 [HIGH LOG] ================================
   ```

2. **Full Request Body:**
   ```json
   🔍 [HIGH LOG] Full Request Body:
   {
     "prompt": "Create a new component",
     "scope": "current-page",
     "context": {
       "page": "/home",
       "timestamp": "2025-10-20T10:00:00Z"
     }
   }
   ```

3. **Claude CLI Command Details:**
   ```
   🔍 [HIGH LOG] ================================
   🔍 [HIGH LOG] CLAUDE CLI COMMAND DETAILS
   🔍 [HIGH LOG] ================================
   🔍 [HIGH LOG] Command ID: cmd_1729456789_abc123
   🔍 [HIGH LOG] Executable: claude
   🔍 [HIGH LOG] Arguments: [Scope: current-page | Page: /home | Task: Create a new component]
   🔍 [HIGH LOG] Working Directory: /workspace/code
   🔍 [HIGH LOG] Full Command: claude Scope: current-page | Page: /home | Task: Create a new component
   🔍 [HIGH LOG] Original Prompt: Create a new component
   🔍 [HIGH LOG] Scope: current-page
   🔍 [HIGH LOG] Page: /home
   🔍 [HIGH LOG] Environment Variables:
   🔍 [HIGH LOG]   PATH=/usr/local/bin:/usr/bin
   🔍 [HIGH LOG]   HOME=/home/user
   🔍 [HIGH LOG]   ... (all environment variables)
   🔍 [HIGH LOG] ================================
   ```

4. **Complete stdout/stderr Output:**
   ```
   🔍 [HIGH LOG] Claude stdout: Analyzing request...
   🔍 [HIGH LOG] Claude stdout: Creating component file...
   🔍 [HIGH LOG] Claude stderr: Warning: File exists, overwriting...
   ```

5. **Full Result:**
   ```
   🔍 [HIGH LOG] ================================
   🔍 [HIGH LOG] COMMAND COMPLETED SUCCESSFULLY
   🔍 [HIGH LOG] ================================
   🔍 [HIGH LOG] Command ID: cmd_1729456789_abc123
   🔍 [HIGH LOG] Execution Time: 2.34s
   🔍 [HIGH LOG] Status: completed
   🔍 [HIGH LOG] Result:
   🔍 [HIGH LOG] {
   🔍 [HIGH LOG]   "action": "Executed command for /home",
   🔍 [HIGH LOG]   "affectedPages": ["/home"],
   🔍 [HIGH LOG]   "changes": [...]
   🔍 [HIGH LOG] }
   🔍 [HIGH LOG] ================================
   ```

**When to use:**
- Debugging Claude CLI integration issues
- Monitoring AI command execution in detail
- Troubleshooting why commands aren't working as expected
- Development and testing environments
- Auditing AI operations

**Security Warning:**
⚠️ **HIGH logging mode outputs all environment variables including potentially sensitive information (API keys, secrets, etc.). Only use in secure, trusted environments. Never use in production unless you understand the security implications.**

**Notes:**
- All logs go to stdout
- High logging mode can produce very verbose output
- Useful for debugging but not recommended for production
- Does not affect client-facing responses, only server logs

---

## Setting Environment Variables

### Method 1: Export in Shell

**Temporary (current session only):**
```bash
export CLAUDE_WORKSPACE_DIR=/path/to/project
go run .
```

**Permanent (add to ~/.bashrc or ~/.zshrc):**
```bash
echo 'export CLAUDE_WORKSPACE_DIR=/path/to/project' >> ~/.bashrc
source ~/.bashrc
```

### Method 2: `.env` File (Future Enhancement)

Currently not implemented, but you could add support:

```bash
# .env file
CLAUDE_WORKSPACE_DIR=/path/to/project
```

### Method 3: Docker Compose

```yaml
version: '3.8'
services:
  backend:
    build: ./backend
    environment:
      - CLAUDE_WORKSPACE_DIR=/app/workspace
    volumes:
      - ./workspace:/app/workspace
```

### Method 4: Systemd Service

```ini
[Service]
Environment="CLAUDE_WORKSPACE_DIR=/var/www/myapp"
ExecStart=/usr/local/bin/site-editor
```

---

## Verification

Check if the environment variable is set:

```bash
echo $CLAUDE_WORKSPACE_DIR
```

Check from server logs when starting:

```bash
go run .
# Look for log line:
🤖 Calling Claude CLI with prompt: ... | Workspace: /your/path
```

---

## Examples

### Example 1: Development Setup

```bash
# Terminal 1: Set workspace and start server
export CLAUDE_WORKSPACE_DIR=/home/dev/my-project
cd backend
go run .

# Terminal 2: Test command
curl -X POST http://localhost:9000/api/ai/command \
  -H "Content-Type: application/json" \
  -d '{"prompt":"List files","scope":"current-page","context":{"page":"/","timestamp":"2025-10-20T10:00:00Z"}}'
```

### Example 2: Multiple Projects

```bash
# Script to switch projects
#!/bin/bash

case $1 in
  "project1")
    export CLAUDE_WORKSPACE_DIR=/home/user/project1
    ;;
  "project2")
    export CLAUDE_WORKSPACE_DIR=/home/user/project2
    ;;
  *)
    export CLAUDE_WORKSPACE_DIR=/workspace/code
    ;;
esac

cd backend && go run .
```

Usage:
```bash
./run-server.sh project1
```

### Example 3: Docker

```dockerfile
# Dockerfile
FROM golang:1.21
WORKDIR /app
COPY . .
RUN go build -o site-editor

# Set default workspace
ENV CLAUDE_WORKSPACE_DIR=/app/workspace

# Create workspace directory
RUN mkdir -p /app/workspace

CMD ["./site-editor"]
```

### Example 4: Debugging with High Logging

```bash
# Enable high-level logging for debugging
export LOG_LEVEL=HIGH
export CLAUDE_WORKSPACE_DIR=/home/user/my-project

# Start server
cd backend && go run .

# In another terminal, send a test command
curl -X POST http://localhost:9000/api/ai/command \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "List all files in the current directory",
    "scope": "current-page",
    "context": {
      "page": "/debug",
      "timestamp": "2025-10-20T10:00:00Z"
    }
  }'
```

The server output will include:
```
🔍 [HIGH LOG] ================================
🔍 [HIGH LOG] HIGH LOGGING ENABLED
🔍 [HIGH LOG] All Claude API calls and responses will be logged in detail
🔍 [HIGH LOG] ================================
Server started on :9000
📥 AI Command Received: "List all files in the current directory" | Scope: current-page | Page: /debug
🔍 [HIGH LOG] Full Request Body:
{
  "prompt": "List all files in the current directory",
  "scope": "current-page",
  "context": {
    "page": "/debug",
    "timestamp": "2025-10-20T10:00:00Z"
  }
}
🔄 Processing Command [cmd_1729456789_abc123]: "List all files in the current directory" | Scope: current-page | Page: /debug
🤖 Calling Claude CLI with prompt: Scope: current-page | Page: /debug | Task: List all files in the current directory | Workspace: /home/user/my-project
🔍 [HIGH LOG] ================================
🔍 [HIGH LOG] CLAUDE CLI COMMAND DETAILS
🔍 [HIGH LOG] ================================
🔍 [HIGH LOG] Command ID: cmd_1729456789_abc123
🔍 [HIGH LOG] Executable: claude
🔍 [HIGH LOG] Arguments: [Scope: current-page | Page: /debug | Task: List all files in the current directory]
🔍 [HIGH LOG] Working Directory: /home/user/my-project
... (continues with full output)
```

---

## Troubleshooting

### Issue: "no such file or directory"

**Error:**
```
❌ Error [cmd_123]: chdir /path/to/project: no such file or directory
```

**Solution:**
Create the directory before starting:
```bash
mkdir -p $CLAUDE_WORKSPACE_DIR
```

### Issue: Permission Denied

**Error:**
```
❌ Error [cmd_123]: permission denied
```

**Solution:**
Ensure the directory has proper permissions:
```bash
chmod 755 $CLAUDE_WORKSPACE_DIR
```

### Issue: Environment Variable Not Set

**Symptom:** Claude runs in `/workspace/code` even though you set the variable

**Solution:**
Make sure you export in the same shell where you run the server:
```bash
export CLAUDE_WORKSPACE_DIR=/your/path
go run .  # In the same terminal session
```

---

## Future Environment Variables

Potential additions for future development:

- `CLAUDE_CLI_PATH` - Custom path to Claude CLI executable
- `CLAUDE_MODEL` - Default Claude model to use
- `CLAUDE_MAX_TOKENS` - Token limit for responses
- `CLAUDE_TIMEOUT` - Command timeout in seconds
- `AI_PROVIDER` - Switch between Claude, GPT, etc.

---

## Code Reference

Environment variable handling is in `backend/ai_command.go`:

```go
// Line 85-92
func getWorkspaceDir() string {
    if dir := os.Getenv("CLAUDE_WORKSPACE_DIR"); dir != "" {
        return dir
    }
    return "/workspace/code"
}

// Line 261-266
workspaceDir := getWorkspaceDir()
log.Printf("🤖 Calling Claude CLI with prompt: %s | Workspace: %s", prompt, workspaceDir)
cmd := exec.CommandContext(session.Context, "claude", prompt)
cmd.Dir = workspaceDir
```

---

## Best Practices

1. **Use absolute paths** - Relative paths may cause issues
2. **Create directory first** - Ensure it exists before starting server
3. **Set permissions properly** - Ensure Claude can read/write
4. **Document your setup** - Add to project README
5. **Use per-environment configs** - Different values for dev/staging/prod

---

**Quick Reference:**

```bash
# Set workspace
export CLAUDE_WORKSPACE_DIR=/path/to/project

# Enable high logging (optional, for debugging)
export LOG_LEVEL=HIGH

# Verify
echo $CLAUDE_WORKSPACE_DIR
echo $LOG_LEVEL

# Run server
cd backend && go run .
```
