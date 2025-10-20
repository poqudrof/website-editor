package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AICommandRequest represents the request to execute an AI command
type AICommandRequest struct {
	Prompt  string         `json:"prompt"`
	Scope   string         `json:"scope"` // current-page, new-page, global
	Context CommandContext `json:"context"`
}

// CommandContext provides context about the command execution environment
type CommandContext struct {
	Page      string `json:"page"`
	Timestamp string `json:"timestamp"`
	UserID    string `json:"userId,omitempty"`
	ProjectID string `json:"projectId,omitempty"`
}

// AICommand represents a stored command in the database
type AICommand struct {
	ID            string `gorm:"primaryKey"`
	Prompt        string `gorm:"type:text"`
	Scope         string
	Page          string
	UserID        string
	ProjectID     string
	Status        string // queued, processing, completed, failed, interrupted
	Result        string `gorm:"type:text"` // JSON-encoded result
	ErrorMessage  string `gorm:"type:text"`
	CreatedAt     int64
	CompletedAt   int64
	ProcessingLog string `gorm:"type:text"` // Stream of progress updates
}

// AICommandSession manages an active AI command execution
type AICommandSession struct {
	ID            string
	Command       *AICommand
	Context       context.Context
	Cancel        context.CancelFunc
	Status        string
	StartTime     time.Time
	mu            sync.RWMutex
	isProcessing  bool
	progressQueue chan ProgressUpdate
}

// ProgressUpdate represents a real-time progress update
type ProgressUpdate struct {
	Type      string      `json:"type"` // status, thinking, output, tool_use, result, error, complete
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
}

// WebSocket message types
const (
	WSMsgTypeStatus   = "status"
	WSMsgTypeThinking = "thinking"
	WSMsgTypeOutput   = "output"
	WSMsgTypeToolUse  = "tool_use"
	WSMsgTypeResult   = "result"
	WSMsgTypeError    = "error"
	WSMsgTypeComplete = "complete"
	WSMsgTypePing     = "ping"
)

// getWorkspaceDir returns the workspace directory from environment variable
// Falls back to /workspace/code if CLAUDE_WORKSPACE_DIR is not set
func getWorkspaceDir() string {
	if dir := os.Getenv("CLAUDE_WORKSPACE_DIR"); dir != "" {
		return dir
	}
	return "/workspace/code"
}

// isHighLogLevel returns true if LOG_LEVEL is set to HIGH
func isHighLogLevel() bool {
	return os.Getenv("LOG_LEVEL") == "HIGH"
}

// Global command sessions
var (
	commandSessions = make(map[string]*AICommandSession)
	commandMu       sync.RWMutex
)

// ExecuteAICommand handles the POST endpoint for executing AI commands
func ExecuteAICommand(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req AICommandRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INVALID_REQUEST",
					"message": "Invalid request body",
					"details": err.Error(),
				},
			})
		}

		// Validate request
		if req.Prompt == "" {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "MISSING_PROMPT",
					"message": "Prompt is required",
				},
			})
		}

		if req.Scope != "current-page" && req.Scope != "new-page" && req.Scope != "global" {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INVALID_SCOPE",
					"message": "Invalid scope value provided",
					"details": "Scope must be one of: current-page, new-page, global",
				},
			})
		}

		// Log incoming command
		log.Printf("📥 AI Command Received: \"%s\" | Scope: %s | Page: %s", req.Prompt, req.Scope, req.Context.Page)

		// High-level logging: log full request
		if isHighLogLevel() {
			reqJSON, _ := json.MarshalIndent(req, "", "  ")
			log.Printf("🔍 [HIGH LOG] Full Request Body:\n%s", string(reqJSON))
		}

		// Create command record
		commandID := fmt.Sprintf("cmd_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
		command := &AICommand{
			ID:        commandID,
			Prompt:    req.Prompt,
			Scope:     req.Scope,
			Page:      req.Context.Page,
			UserID:    req.Context.UserID,
			ProjectID: req.Context.ProjectID,
			Status:    "queued",
			CreatedAt: time.Now().Unix(),
		}

		// Save to database
		if err := db.Create(command).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "DATABASE_ERROR",
					"message": "Failed to create command",
					"details": err.Error(),
				},
			})
		}

		// Return immediate response with command ID
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Command queued successfully",
			"data": fiber.Map{
				"commandId": commandID,
				"status":    "queued",
				"message":   "Connect to WebSocket to receive real-time updates",
				"wsUrl":     fmt.Sprintf("ws://localhost:9000/api/ai/command/%s/stream", commandID),
			},
		})
	}
}

// StreamAICommand handles WebSocket streaming for AI command execution
func StreamAICommand(db *gorm.DB) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		commandID := conn.Params("commandId")

		// Retrieve command from database
		var command AICommand
		if err := db.First(&command, "id = ?", commandID).Error; err != nil {
			sendWSError(conn, "COMMAND_NOT_FOUND", "Command not found", err.Error())
			return
		}

		// Create session
		ctx, cancel := context.WithCancel(context.Background())
		session := &AICommandSession{
			ID:            commandID,
			Command:       &command,
			Context:       ctx,
			Cancel:        cancel,
			Status:        "processing",
			StartTime:     time.Now(),
			isProcessing:  true,
			progressQueue: make(chan ProgressUpdate, 100),
		}

		// Store session
		commandMu.Lock()
		commandSessions[commandID] = session
		commandMu.Unlock()

		// Send initial status
		sendWSMessage(conn, ProgressUpdate{
			Type:      WSMsgTypeStatus,
			Timestamp: time.Now().Format(time.RFC3339),
			Data: fiber.Map{
				"commandId": commandID,
				"status":    "connected",
				"message":   "WebSocket connected, starting AI processing",
			},
		})

		// Start AI processing in background
		go processAICommand(session, db)

		// Handle incoming messages (for interrupt/ping)
		go handleWSMessages(conn, session)

		// Stream progress updates to client
		streamProgressUpdates(conn, session)

		// Cleanup
		cleanup(session)
	})
}

// processAICommand executes the AI command using Claude CLI
func processAICommand(session *AICommandSession, db *gorm.DB) {
	defer func() {
		session.mu.Lock()
		session.isProcessing = false
		session.mu.Unlock()
		close(session.progressQueue)
	}()

	command := session.Command

	// Log processing start
	log.Printf("🔄 Processing Command [%s]: \"%s\" | Scope: %s | Page: %s", command.ID, command.Prompt, command.Scope, command.Page)

	// Update status to processing
	command.Status = "processing"
	db.Save(command)

	// Send status update
	session.progressQueue <- ProgressUpdate{
		Type:      WSMsgTypeStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Starting Claude CLI...",
	}

	// Build the prompt for Claude
	prompt := buildClaudePrompt(command)
	workspaceDir := getWorkspaceDir()
	log.Printf("🤖 Calling Claude CLI with prompt: %s | Workspace: %s", prompt, workspaceDir)

	// Create command with context for cancellation
	cmd := exec.CommandContext(session.Context, "claude", prompt)
	cmd.Dir = workspaceDir // Set working directory from environment variable

	// High-level logging: log full Claude command details
	if isHighLogLevel() {
		log.Printf("🔍 [HIGH LOG] ================================")
		log.Printf("🔍 [HIGH LOG] CLAUDE CLI COMMAND DETAILS")
		log.Printf("🔍 [HIGH LOG] ================================")
		log.Printf("🔍 [HIGH LOG] Command ID: %s", command.ID)
		log.Printf("🔍 [HIGH LOG] Executable: claude")
		log.Printf("🔍 [HIGH LOG] Arguments: [%s]", prompt)
		log.Printf("🔍 [HIGH LOG] Working Directory: %s", workspaceDir)
		log.Printf("🔍 [HIGH LOG] Full Command: claude %s", prompt)
		log.Printf("🔍 [HIGH LOG] Original Prompt: %s", command.Prompt)
		log.Printf("🔍 [HIGH LOG] Scope: %s", command.Scope)
		log.Printf("🔍 [HIGH LOG] Page: %s", command.Page)
		log.Printf("🔍 [HIGH LOG] Environment Variables:")
		for _, env := range os.Environ() {
			log.Printf("🔍 [HIGH LOG]   %s", env)
		}
		log.Printf("🔍 [HIGH LOG] ================================")
	}

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		handleCommandError(session, command, db, fmt.Errorf("failed to create stdout pipe: %w", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		handleCommandError(session, command, db, fmt.Errorf("failed to create stderr pipe: %w", err))
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		handleCommandError(session, command, db, fmt.Errorf("failed to start Claude CLI: %w", err))
		return
	}

	log.Printf("✅ Claude CLI process started")

	// Read stdout and stderr concurrently
	var wg sync.WaitGroup

	// Read stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			// Log to stdout
			if isHighLogLevel() {
				log.Printf("🔍 [HIGH LOG] Claude stdout: %s", line)
			} else {
				log.Printf("📤 Claude: %s", line)
			}

			// Stream output to client
			select {
			case session.progressQueue <- ProgressUpdate{
				Type:      WSMsgTypeOutput,
				Timestamp: time.Now().Format(time.RFC3339),
				Data:      line,
			}:
			case <-session.Context.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			log.Printf("❌ Error reading stdout: %v", err)
		}
	}()

	// Read stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()

			// Log to stdout
			if isHighLogLevel() {
				log.Printf("🔍 [HIGH LOG] Claude stderr: %s", line)
			} else {
				log.Printf("⚠️ Claude stderr: %s", line)
			}

			// Stream to client as output
			select {
			case session.progressQueue <- ProgressUpdate{
				Type:      WSMsgTypeOutput,
				Timestamp: time.Now().Format(time.RFC3339),
				Data:      fmt.Sprintf("[stderr] %s", line),
			}:
			case <-session.Context.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			log.Printf("❌ Error reading stderr: %v", err)
		}
	}()

	// Wait for command to complete
	cmdErr := cmd.Wait()
	wg.Wait()

	// Handle completion
	executionTime := time.Since(session.StartTime).Seconds()

	if cmdErr != nil {
		if session.Context.Err() == context.Canceled {
			// Interrupted by user
			log.Printf("⚠️ Command Interrupted [%s]", command.ID)
			command.Status = "interrupted"
			db.Save(command)

			session.progressQueue <- ProgressUpdate{
				Type:      WSMsgTypeStatus,
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Command was interrupted",
			}
		} else {
			// Error occurred
			log.Printf("❌ Command Failed [%s]: %v", command.ID, cmdErr)
			handleCommandError(session, command, db, cmdErr)
		}
		return
	}

	// Success
	log.Printf("✅ Command Completed [%s]: %.2fs", command.ID, executionTime)

	command.Status = "completed"
	command.CompletedAt = time.Now().Unix()

	// Create result
	result := fiber.Map{
		"action":        fmt.Sprintf("Executed command for %s", command.Page),
		"affectedPages": []string{command.Page},
		"changes": []fiber.Map{
			{
				"type":        "update",
				"target":      command.Page,
				"description": "Applied changes via Claude CLI",
			},
		},
	}
	resultJSON, _ := json.Marshal(result)
	command.Result = string(resultJSON)
	db.Save(command)

	// High-level logging: log full result
	if isHighLogLevel() {
		log.Printf("🔍 [HIGH LOG] ================================")
		log.Printf("🔍 [HIGH LOG] COMMAND COMPLETED SUCCESSFULLY")
		log.Printf("🔍 [HIGH LOG] ================================")
		log.Printf("🔍 [HIGH LOG] Command ID: %s", command.ID)
		log.Printf("🔍 [HIGH LOG] Execution Time: %.2fs", executionTime)
		log.Printf("🔍 [HIGH LOG] Status: %s", command.Status)
		resultPretty, _ := json.MarshalIndent(result, "🔍 [HIGH LOG] ", "  ")
		log.Printf("🔍 [HIGH LOG] Result:\n🔍 [HIGH LOG] %s", string(resultPretty))
		log.Printf("🔍 [HIGH LOG] ================================")
	}

	// Send result
	session.progressQueue <- ProgressUpdate{
		Type:      WSMsgTypeResult,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      result,
	}

	// Send completion
	session.progressQueue <- ProgressUpdate{
		Type:      WSMsgTypeComplete,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Command completed successfully",
		Data: fiber.Map{
			"commandId":     command.ID,
			"status":        "completed",
			"executionTime": executionTime,
		},
	}
}

// buildClaudePrompt builds the prompt for Claude CLI based on the command
func buildClaudePrompt(command *AICommand) string {
	// Build a prompt that includes scope and page context
	prompt := command.Prompt

	// Add context about the scope and page
	if command.Scope != "" && command.Page != "" {
		prompt = fmt.Sprintf("Scope: %s | Page: %s | Task: %s", command.Scope, command.Page, command.Prompt)
	}

	return prompt
}

// handleCommandError handles errors during command execution
func handleCommandError(session *AICommandSession, command *AICommand, db *gorm.DB, err error) {
	errMsg := err.Error()
	log.Printf("❌ Error [%s]: %s", command.ID, errMsg)

	command.Status = "failed"
	command.ErrorMessage = errMsg
	db.Save(command)

	session.progressQueue <- ProgressUpdate{
		Type:      WSMsgTypeError,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   errMsg,
		Data: fiber.Map{
			"error": errMsg,
		},
	}

	session.progressQueue <- ProgressUpdate{
		Type:      WSMsgTypeComplete,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Command failed",
		Data: fiber.Map{
			"commandId": command.ID,
			"status":    "failed",
		},
	}
}

// handleWSMessages handles incoming WebSocket messages from the client
func handleWSMessages(conn *websocket.Conn, session *AICommandSession) {
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			return
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "interrupt":
			session.Cancel()
			sendWSMessage(conn, ProgressUpdate{
				Type:      WSMsgTypeStatus,
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Interrupt signal received",
			})

		case "ping":
			sendWSMessage(conn, ProgressUpdate{
				Type:      WSMsgTypePing,
				Timestamp: time.Now().Format(time.RFC3339),
			})
		}
	}
}

// streamProgressUpdates streams progress updates from the queue to the WebSocket
func streamProgressUpdates(conn *websocket.Conn, session *AICommandSession) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case update, ok := <-session.progressQueue:
			if !ok {
				// Channel closed, processing complete
				return
			}
			if err := sendWSMessage(conn, update); err != nil {
				return
			}

		case <-ticker.C:
			// Send keep-alive ping
			sendWSMessage(conn, ProgressUpdate{
				Type:      WSMsgTypePing,
				Timestamp: time.Now().Format(time.RFC3339),
			})

		case <-session.Context.Done():
			return
		}
	}
}

// Helper functions

func sendWSMessage(conn *websocket.Conn, update ProgressUpdate) error {
	return conn.WriteJSON(update)
}

func sendWSError(conn *websocket.Conn, code, message, details string) {
	conn.WriteJSON(fiber.Map{
		"type":  WSMsgTypeError,
		"error": fiber.Map{
			"code":    code,
			"message": message,
			"details": details,
		},
	})
	conn.Close()
}

func cleanup(session *AICommandSession) {
	session.Cancel()
	commandMu.Lock()
	delete(commandSessions, session.ID)
	commandMu.Unlock()
}

// GetAICommandStatus returns the status of a command
func GetAICommandStatus(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandID := c.Params("commandId")

		var command AICommand
		if err := db.First(&command, "id = ?", commandID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "COMMAND_NOT_FOUND",
					"message": "Command not found",
				},
			})
		}

		response := fiber.Map{
			"success": true,
			"data": fiber.Map{
				"commandId":  command.ID,
				"status":     command.Status,
				"prompt":     command.Prompt,
				"scope":      command.Scope,
				"createdAt":  command.CreatedAt,
				"completedAt": command.CompletedAt,
			},
		}

		if command.Result != "" {
			var result map[string]interface{}
			json.Unmarshal([]byte(command.Result), &result)
			response["data"].(fiber.Map)["result"] = result
		}

		if command.ErrorMessage != "" {
			response["data"].(fiber.Map)["error"] = command.ErrorMessage
		}

		return c.JSON(response)
	}
}

// InterruptAICommand interrupts a running command
func InterruptAICommand() fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandID := c.Params("commandId")

		commandMu.RLock()
		session, exists := commandSessions[commandID]
		commandMu.RUnlock()

		if !exists {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "SESSION_NOT_FOUND",
					"message": "Command session not found or already completed",
				},
			})
		}

		session.Cancel()

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Command interrupted successfully",
			"data": fiber.Map{
				"commandId": commandID,
				"status":    "interrupted",
			},
		})
	}
}
