# AI Agent API - Final Frontend Implementation Guide

**Version:** 2.0
**Last Updated:** 2025-10-20
**Protocol:** REST + WebSocket

This document provides the complete API specification for frontend implementation of the AI Command feature with real-time WebSocket streaming.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [API Endpoints](#api-endpoints)
4. [WebSocket Protocol](#websocket-protocol)
5. [Message Types](#message-types)
6. [Frontend Implementation Guide](#frontend-implementation-guide)
7. [Error Handling](#error-handling)
8. [TypeScript Types](#typescript-types)
9. [Complete Examples](#complete-examples)

---

## Overview

The AI Agent API allows users to send natural language commands through the Command+K modal to modify their website. The system provides:

- ✅ **Real-time streaming** via WebSocket for progress updates
- ✅ **Three scope modes**: current-page, new-page, global
- ✅ **Interrupt capability** to stop long-running operations
- ✅ **Persistent command history** stored in database
- ✅ **Structured progress updates** (thinking, tool usage, results)
- ✅ **Automatic reconnection** support

### Base URL

```
HTTP:  http://localhost:9000
WebSocket: ws://localhost:9000
```

---

## Architecture

### Flow Diagram

```
User → Command Modal → POST /api/ai/command
                            ↓
                     Command Queued (returns commandId)
                            ↓
Frontend → WebSocket connect: ws://localhost:9000/api/ai/command/{commandId}/stream
                            ↓
Backend → Processes command → Streams updates
                            ↓
Frontend ← Receives: thinking → tool_use → result → complete
```

---

## API Endpoints

### 1. Execute AI Command

**POST** `/api/ai/command`

Submit a new AI command for execution. Returns immediately with a command ID.

#### Request

```typescript
interface AICommandRequest {
  prompt: string;                    // User's natural language instruction
  scope: 'current-page' | 'new-page' | 'global'; // Scope of changes
  context: {
    page: string;                    // Current page path (e.g., "/about")
    timestamp: string;               // ISO 8601 timestamp
    userId?: string;                 // Optional user identifier
    projectId?: string;              // Optional project identifier
  };
}
```

**Example Request:**
```json
{
  "prompt": "Add a contact form with name, email, and message fields",
  "scope": "current-page",
  "context": {
    "page": "/contact",
    "timestamp": "2025-10-20T15:30:00Z",
    "userId": "user_123",
    "projectId": "proj_abc"
  }
}
```

#### Response

**Success (200 OK):**
```json
{
  "success": true,
  "message": "Command queued successfully",
  "data": {
    "commandId": "cmd_1729435800_a1b2c3d4",
    "status": "queued",
    "message": "Connect to WebSocket to receive real-time updates",
    "wsUrl": "ws://localhost:9000/api/ai/command/cmd_1729435800_a1b2c3d4/stream"
  }
}
```

**Error (400/500):**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_SCOPE",
    "message": "Invalid scope value provided",
    "details": "Scope must be one of: current-page, new-page, global"
  }
}
```

#### Error Codes

| Code | Description |
|------|-------------|
| `INVALID_REQUEST` | Malformed request body |
| `MISSING_PROMPT` | Prompt field is required |
| `INVALID_SCOPE` | Invalid scope value |
| `DATABASE_ERROR` | Failed to store command |

---

### 2. Stream Command Progress (WebSocket)

**WebSocket** `/api/ai/command/:commandId/stream`

Connect to this WebSocket endpoint to receive real-time updates about command execution.

**Connection:**
```javascript
const ws = new WebSocket('ws://localhost:9000/api/ai/command/cmd_1729435800_a1b2c3d4/stream');
```

**See [WebSocket Protocol](#websocket-protocol) section for detailed message formats.**

---

### 3. Get Command Status

**GET** `/api/ai/command/:commandId/status`

Retrieve the current status of a command (useful for reconnection or polling).

#### Response

```json
{
  "success": true,
  "data": {
    "commandId": "cmd_1729435800_a1b2c3d4",
    "status": "completed",
    "prompt": "Add a contact form",
    "scope": "current-page",
    "createdAt": 1729435800,
    "completedAt": 1729435810,
    "result": {
      "action": "Updated /contact based on your request",
      "affectedPages": ["/contact"],
      "changes": [
        {
          "type": "update",
          "target": "/contact",
          "description": "Applied requested changes to page content"
        }
      ]
    }
  }
}
```

**Status Values:**
- `queued` - Command is waiting to be processed
- `processing` - Currently executing
- `completed` - Successfully finished
- `failed` - Execution failed
- `interrupted` - User interrupted the command

---

### 4. Interrupt Command

**POST** `/api/ai/command/:commandId/interrupt`

Stop a running command gracefully.

#### Response

```json
{
  "success": true,
  "message": "Command interrupted successfully",
  "data": {
    "commandId": "cmd_1729435800_a1b2c3d4",
    "status": "interrupted"
  }
}
```

---

## WebSocket Protocol

### Connection Lifecycle

1. **Connect** to WebSocket using command ID from initial POST
2. **Receive** `status` message confirming connection
3. **Stream** progress updates as command executes
4. **Complete** when final `complete` message received
5. **Close** connection automatically

### Sending Messages (Client → Server)

The client can send these message types:

#### Interrupt Command
```json
{
  "type": "interrupt"
}
```

#### Ping (Keep-Alive)
```json
{
  "type": "ping"
}
```

### Receiving Messages (Server → Client)

All messages follow this structure:

```typescript
interface ProgressUpdate {
  type: 'status' | 'thinking' | 'output' | 'tool_use' | 'result' | 'error' | 'complete' | 'ping';
  timestamp: string;        // ISO 8601 format
  message?: string;         // Human-readable message
  data?: any;              // Type-specific data
}
```

---

## Message Types

### 1. Status Message

Sent when connection is established or status changes.

```json
{
  "type": "status",
  "timestamp": "2025-10-20T15:30:00Z",
  "data": {
    "commandId": "cmd_1729435800_a1b2c3d4",
    "status": "connected",
    "message": "WebSocket connected, starting AI processing"
  }
}
```

---

### 2. Thinking Message

Indicates AI is analyzing or planning. Shows what phase of processing is active.

```json
{
  "type": "thinking",
  "timestamp": "2025-10-20T15:30:01Z",
  "message": "Planning the changes to implement",
  "data": {
    "phase": "planning"
  }
}
```

**Phases:**
- `analyzing` - Understanding the request
- `planning` - Planning implementation
- `executing` - Making changes
- `validating` - Verifying results

---

### 3. Tool Use Message

Indicates AI is using a specific tool (reading files, writing content, etc.).

```json
{
  "type": "tool_use",
  "timestamp": "2025-10-20T15:30:02Z",
  "message": "Using tool: read_file",
  "data": {
    "name": "read_file",
    "action": "Reading /contact"
  }
}
```

**Common Tools by Scope:**

**current-page:**
- `read_file` - Reading page file
- `parse_html` - Parsing HTML structure
- `update_content` - Updating content
- `write_file` - Saving changes

**new-page:**
- `generate_content` - Creating content
- `create_file` - Creating new file
- `update_navigation` - Updating nav

**global:**
- `scan_project` - Scanning structure
- `find_files` - Finding affected files
- `update_multiple` - Updating files
- `rebuild_index` - Rebuilding index

---

### 4. Output Message

Raw output from execution (if applicable).

```json
{
  "type": "output",
  "timestamp": "2025-10-20T15:30:03Z",
  "data": "Processing item 1 of 5..."
}
```

---

### 5. Result Message

Final result of the command execution.

```json
{
  "type": "result",
  "timestamp": "2025-10-20T15:30:05Z",
  "data": {
    "action": "Updated /contact based on your request",
    "affectedPages": ["/contact"],
    "changes": [
      {
        "type": "update",
        "target": "/contact",
        "description": "Applied requested changes to page content"
      }
    ],
    "newPageUrl": null  // Only present for scope: "new-page"
  }
}
```

**Result Structure:**

```typescript
interface CommandResult {
  action: string;              // Summary of what was done
  affectedPages: string[];     // List of modified pages
  changes: Change[];           // Detailed change list
  newPageUrl?: string;         // New page URL (only for new-page scope)
}

interface Change {
  type: 'create' | 'update' | 'delete';
  target: string;              // What was changed
  description: string;         // Description of change
}
```

---

### 6. Error Message

Indicates an error occurred during processing.

```json
{
  "type": "error",
  "timestamp": "2025-10-20T15:30:05Z",
  "message": "Failed to parse HTML",
  "data": {
    "code": "PARSE_ERROR",
    "details": "Invalid HTML structure at line 42"
  }
}
```

---

### 7. Complete Message

Final message indicating command has finished (success or failure).

```json
{
  "type": "complete",
  "timestamp": "2025-10-20T15:30:06Z",
  "message": "Command completed successfully",
  "data": {
    "commandId": "cmd_1729435800_a1b2c3d4",
    "status": "completed",
    "executionTime": 6.2
  }
}
```

---

### 8. Ping Message

Keep-alive message sent every 30 seconds.

```json
{
  "type": "ping",
  "timestamp": "2025-10-20T15:30:30Z"
}
```

---

## Frontend Implementation Guide

### Step 1: Execute Command

```typescript
async function executeAICommand(
  prompt: string,
  scope: 'current-page' | 'new-page' | 'global',
  page: string
): Promise<string> {
  const response = await fetch('http://localhost:9000/api/ai/command', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      prompt,
      scope,
      context: {
        page,
        timestamp: new Date().toISOString(),
      },
    }),
  });

  const data = await response.json();

  if (!data.success) {
    throw new Error(data.error.message);
  }

  return data.data.commandId; // Return command ID
}
```

---

### Step 2: Connect to WebSocket

```typescript
function connectToCommandStream(
  commandId: string,
  callbacks: {
    onStatus?: (data: any) => void;
    onThinking?: (message: string, data?: any) => void;
    onToolUse?: (tool: string, action: string) => void;
    onOutput?: (output: string) => void;
    onResult?: (result: CommandResult) => void;
    onError?: (error: any) => void;
    onComplete?: (data: any) => void;
  }
): WebSocket {
  const ws = new WebSocket(
    `ws://localhost:9000/api/ai/command/${commandId}/stream`
  );

  ws.onmessage = (event) => {
    const update: ProgressUpdate = JSON.parse(event.data);

    switch (update.type) {
      case 'status':
        callbacks.onStatus?.(update.data);
        break;
      case 'thinking':
        callbacks.onThinking?.(update.message!, update.data);
        break;
      case 'tool_use':
        callbacks.onToolUse?.(
          update.data.name,
          update.data.action
        );
        break;
      case 'output':
        callbacks.onOutput?.(update.data);
        break;
      case 'result':
        callbacks.onResult?.(update.data);
        break;
      case 'error':
        callbacks.onError?.(update);
        break;
      case 'complete':
        callbacks.onComplete?.(update.data);
        ws.close();
        break;
      case 'ping':
        // Keep-alive, no action needed
        break;
    }
  };

  ws.onerror = (error) => {
    callbacks.onError?.(error);
  };

  return ws;
}
```

---

### Step 3: Interrupt Command (Optional)

```typescript
async function interruptCommand(commandId: string): Promise<void> {
  // Option 1: Send via WebSocket
  ws.send(JSON.stringify({ type: 'interrupt' }));

  // Option 2: Send via HTTP POST
  await fetch(`http://localhost:9000/api/ai/command/${commandId}/interrupt`, {
    method: 'POST',
  });
}
```

---

## Error Handling

### Connection Errors

```typescript
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
  // Show user-friendly error message
  showError('Connection lost. Please try again.');
};

ws.onclose = (event) => {
  if (!event.wasClean) {
    // Unexpected disconnection
    showError('Connection closed unexpectedly.');
    // Optionally: implement reconnection logic
  }
};
```

### Command Errors

When `type: "error"` message is received:

```typescript
onError: (error) => {
  // Display error to user
  showError(error.message);
  // Log for debugging
  console.error('Command error:', error.data);
}
```

### HTTP Errors

```typescript
if (!response.ok) {
  const error = await response.json();
  throw new Error(error.error.message);
}
```

---

## TypeScript Types

Complete TypeScript definitions for your frontend:

```typescript
// Request Types
interface AICommandRequest {
  prompt: string;
  scope: CommandScope;
  context: CommandContext;
}

type CommandScope = 'current-page' | 'new-page' | 'global';

interface CommandContext {
  page: string;
  timestamp: string;
  userId?: string;
  projectId?: string;
}

// Response Types
interface AICommandResponse {
  success: boolean;
  message?: string;
  data?: {
    commandId: string;
    status: CommandStatus;
    message: string;
    wsUrl: string;
  };
  error?: APIError;
}

type CommandStatus =
  | 'queued'
  | 'processing'
  | 'completed'
  | 'failed'
  | 'interrupted';

interface APIError {
  code: string;
  message: string;
  details?: string;
}

// WebSocket Message Types
interface ProgressUpdate {
  type: MessageType;
  timestamp: string;
  message?: string;
  data?: any;
}

type MessageType =
  | 'status'
  | 'thinking'
  | 'output'
  | 'tool_use'
  | 'result'
  | 'error'
  | 'complete'
  | 'ping';

// Result Types
interface CommandResult {
  action: string;
  affectedPages: string[];
  changes: Change[];
  newPageUrl?: string;
}

interface Change {
  type: 'create' | 'update' | 'delete';
  target: string;
  description: string;
}

// Tool Use
interface ToolUse {
  name: string;
  action: string;
}

// Status Data
interface StatusData {
  commandId: string;
  status: string;
  message: string;
}

// Complete Data
interface CompleteData {
  commandId: string;
  status: CommandStatus;
  executionTime: number;
}
```

---

## Complete Examples

### React Hook Example

```typescript
import { useState, useEffect, useRef } from 'react';

interface UseAICommandOptions {
  onThinking?: (message: string) => void;
  onToolUse?: (tool: string, action: string) => void;
  onResult?: (result: CommandResult) => void;
  onError?: (error: any) => void;
  onComplete?: () => void;
}

export function useAICommand(options: UseAICommandOptions = {}) {
  const [isExecuting, setIsExecuting] = useState(false);
  const [progress, setProgress] = useState<string>('');
  const [currentCommandId, setCurrentCommandId] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const executeCommand = async (
    prompt: string,
    scope: CommandScope,
    page: string
  ) => {
    try {
      setIsExecuting(true);
      setProgress('Sending command...');

      // Step 1: Submit command
      const response = await fetch('http://localhost:9000/api/ai/command', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          prompt,
          scope,
          context: {
            page,
            timestamp: new Date().toISOString(),
          },
        }),
      });

      const data = await response.json();
      if (!data.success) {
        throw new Error(data.error.message);
      }

      const commandId = data.data.commandId;
      setCurrentCommandId(commandId);

      // Step 2: Connect to WebSocket
      const ws = new WebSocket(
        `ws://localhost:9000/api/ai/command/${commandId}/stream`
      );

      ws.onmessage = (event) => {
        const update: ProgressUpdate = JSON.parse(event.data);

        switch (update.type) {
          case 'status':
            setProgress(update.data.message);
            break;

          case 'thinking':
            setProgress(update.message!);
            options.onThinking?.(update.message!);
            break;

          case 'tool_use':
            setProgress(`Using: ${update.data.name}`);
            options.onToolUse?.(update.data.name, update.data.action);
            break;

          case 'result':
            options.onResult?.(update.data);
            break;

          case 'error':
            options.onError?.(update);
            setProgress('Error occurred');
            setIsExecuting(false);
            break;

          case 'complete':
            setProgress('Completed');
            setIsExecuting(false);
            options.onComplete?.();
            ws.close();
            break;
        }
      };

      ws.onerror = (error) => {
        options.onError?.(error);
        setIsExecuting(false);
      };

      wsRef.current = ws;

    } catch (error) {
      options.onError?.(error);
      setIsExecuting(false);
    }
  };

  const interrupt = () => {
    if (wsRef.current && currentCommandId) {
      wsRef.current.send(JSON.stringify({ type: 'interrupt' }));
      setProgress('Interrupting...');
    }
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  return {
    executeCommand,
    interrupt,
    isExecuting,
    progress,
    commandId: currentCommandId,
  };
}
```

### Usage in Component

```typescript
function CommandModal() {
  const [prompt, setPrompt] = useState('');
  const [scope, setScope] = useState<CommandScope>('current-page');
  const [logs, setLogs] = useState<string[]>([]);

  const { executeCommand, interrupt, isExecuting, progress } = useAICommand({
    onThinking: (message) => {
      setLogs(prev => [...prev, `💭 ${message}`]);
    },
    onToolUse: (tool, action) => {
      setLogs(prev => [...prev, `🔧 ${tool}: ${action}`]);
    },
    onResult: (result) => {
      setLogs(prev => [...prev, `✅ ${result.action}`]);
      console.log('Result:', result);
    },
    onError: (error) => {
      setLogs(prev => [...prev, `❌ Error: ${error.message}`]);
    },
    onComplete: () => {
      setLogs(prev => [...prev, '🎉 Complete!']);
    },
  });

  const handleSubmit = () => {
    executeCommand(prompt, scope, window.location.pathname);
  };

  return (
    <div className="command-modal">
      <input
        value={prompt}
        onChange={(e) => setPrompt(e.target.value)}
        placeholder="What would you like to do?"
      />

      <select value={scope} onChange={(e) => setScope(e.target.value as CommandScope)}>
        <option value="current-page">Current Page</option>
        <option value="new-page">New Page</option>
        <option value="global">Global</option>
      </select>

      <button onClick={handleSubmit} disabled={isExecuting}>
        {isExecuting ? progress : 'Execute'}
      </button>

      {isExecuting && (
        <button onClick={interrupt}>Interrupt</button>
      )}

      <div className="logs">
        {logs.map((log, i) => (
          <div key={i}>{log}</div>
        ))}
      </div>
    </div>
  );
}
```

---

### Vue 3 Composition API Example

```typescript
import { ref, onUnmounted } from 'vue';

export function useAICommand() {
  const isExecuting = ref(false);
  const progress = ref('');
  const logs = ref<string[]>([]);
  let ws: WebSocket | null = null;

  const executeCommand = async (
    prompt: string,
    scope: CommandScope,
    page: string
  ) => {
    try {
      isExecuting.value = true;
      progress.value = 'Sending command...';
      logs.value = [];

      const response = await fetch('http://localhost:9000/api/ai/command', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          prompt,
          scope,
          context: {
            page,
            timestamp: new Date().toISOString(),
          },
        }),
      });

      const data = await response.json();
      if (!data.success) throw new Error(data.error.message);

      const commandId = data.data.commandId;

      ws = new WebSocket(
        `ws://localhost:9000/api/ai/command/${commandId}/stream`
      );

      ws.onmessage = (event) => {
        const update: ProgressUpdate = JSON.parse(event.data);

        switch (update.type) {
          case 'thinking':
            progress.value = update.message!;
            logs.value.push(`💭 ${update.message}`);
            break;
          case 'tool_use':
            logs.value.push(`🔧 ${update.data.name}: ${update.data.action}`);
            break;
          case 'result':
            logs.value.push(`✅ ${update.data.action}`);
            break;
          case 'complete':
            isExecuting.value = false;
            progress.value = 'Completed';
            ws?.close();
            break;
          case 'error':
            logs.value.push(`❌ ${update.message}`);
            isExecuting.value = false;
            break;
        }
      };

    } catch (error: any) {
      logs.value.push(`❌ Error: ${error.message}`);
      isExecuting.value = false;
    }
  };

  const interrupt = () => {
    ws?.send(JSON.stringify({ type: 'interrupt' }));
  };

  onUnmounted(() => {
    ws?.close();
  });

  return {
    executeCommand,
    interrupt,
    isExecuting,
    progress,
    logs,
  };
}
```

---

## Testing

### Manual Testing with curl & wscat

```bash
# 1. Submit command
COMMAND_ID=$(curl -s -X POST http://localhost:9000/api/ai/command \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Add a hero section",
    "scope": "current-page",
    "context": {
      "page": "/",
      "timestamp": "2025-10-20T10:00:00Z"
    }
  }' | jq -r '.data.commandId')

echo "Command ID: $COMMAND_ID"

# 2. Connect to WebSocket (requires wscat: npm install -g wscat)
wscat -c "ws://localhost:9000/api/ai/command/$COMMAND_ID/stream"

# 3. Check status
curl http://localhost:9000/api/ai/command/$COMMAND_ID/status | jq

# 4. Interrupt (if needed)
curl -X POST http://localhost:9000/api/ai/command/$COMMAND_ID/interrupt
```

---

## Best Practices

### 1. Connection Management

- ✅ Always close WebSocket when component unmounts
- ✅ Implement reconnection logic for network failures
- ✅ Handle keep-alive pings automatically

### 2. User Experience

- ✅ Show progress indicator during execution
- ✅ Display thinking phases to user
- ✅ Provide interrupt button for long operations
- ✅ Show detailed logs in debug mode

### 3. Error Handling

- ✅ Display user-friendly error messages
- ✅ Log technical details for debugging
- ✅ Implement retry logic for network errors
- ✅ Validate inputs before submission

### 4. Performance

- ✅ Debounce rapid command submissions
- ✅ Cache command results if needed
- ✅ Clean up old WebSocket connections
- ✅ Use command ID for deduplication

---

## Security Considerations

1. **Input Validation**: Sanitize prompts before submission
2. **Rate Limiting**: Limit commands per user/session
3. **Authentication**: Add auth headers if implementing user system
4. **CORS**: Configure appropriate CORS policies
5. **WebSocket Security**: Use WSS in production

---

## Appendix: Scope Behavior Reference

| Scope | Behavior | Returns newPageUrl | Affects Multiple Pages |
|-------|----------|-------------------|----------------------|
| `current-page` | Modifies only the current page | No | No |
| `new-page` | Creates a new page | Yes | No |
| `global` | Site-wide changes | No | Yes |

---

**Questions or Issues?**
Report at: https://github.com/anthropics/claude-code/issues

**End of Document**
