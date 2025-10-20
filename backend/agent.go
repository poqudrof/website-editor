package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AgentSession represents an active AI agent process
type AgentSession struct {
	ID        string
	Command   string
	Args      []string
	Process   *exec.Cmd
	Context   context.Context
	Cancel    context.CancelFunc
	Output    chan string
	Error     chan error
	StartTime time.Time
	mu        sync.Mutex
	isRunning bool
}

// Global session manager
var (
	sessions = make(map[string]*AgentSession)
	sessMu   sync.RWMutex
)

// AgentRunRequest represents the request to start an AI agent
type AgentRunRequest struct {
	Command string   `json:"command"` // The CLI command to run
	Args    []string `json:"args"`    // Command arguments
}

// RunAgent starts a new AI agent process
func RunAgent() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req AgentRunRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Command == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Command is required",
			})
		}

		// Create session ID
		sessionID := uuid.New().String()

		// Create context with cancel
		ctx, cancel := context.WithCancel(context.Background())

		// Create session
		session := &AgentSession{
			ID:        sessionID,
			Command:   req.Command,
			Args:      req.Args,
			Context:   ctx,
			Cancel:    cancel,
			Output:    make(chan string, 100),
			Error:     make(chan error, 10),
			StartTime: time.Now(),
			isRunning: true,
		}

		// Store session
		sessMu.Lock()
		sessions[sessionID] = session
		sessMu.Unlock()

		// Start the process in a goroutine
		go startAgentProcess(session)

		return c.JSON(fiber.Map{
			"session_id": sessionID,
			"status":     "started",
			"command":    req.Command,
			"args":       req.Args,
		})
	}
}

// startAgentProcess spawns and manages the AI agent process
func startAgentProcess(session *AgentSession) {
	defer func() {
		session.mu.Lock()
		session.isRunning = false
		session.mu.Unlock()
		close(session.Output)
		close(session.Error)
	}()

	// Create command with context for cancellation
	cmd := exec.CommandContext(session.Context, session.Command, session.Args...)
	session.Process = cmd

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		session.Error <- fmt.Errorf("failed to create stdout pipe: %w", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		session.Error <- fmt.Errorf("failed to create stderr pipe: %w", err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		session.Error <- fmt.Errorf("failed to start command: %w", err)
		return
	}

	// Read stdout and stderr concurrently
	var wg sync.WaitGroup

	// Read stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case session.Output <- line:
			case <-session.Context.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			session.Error <- fmt.Errorf("stdout error: %w", err)
		}
	}()

	// Read stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case session.Output <- fmt.Sprintf("[STDERR] %s", line):
			case <-session.Context.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			session.Error <- fmt.Errorf("stderr error: %w", err)
		}
	}()

	// Wait for command to complete
	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		if session.Context.Err() == context.Canceled {
			session.Output <- "[INTERRUPTED] Process was interrupted by user"
		} else {
			session.Error <- fmt.Errorf("command failed: %w", err)
		}
	} else {
		session.Output <- "[COMPLETED] Process finished successfully"
	}
}

// StreamAgent streams the output of a running AI agent using Server-Sent Events
func StreamAgent() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("sessionId")

		sessMu.RLock()
		session, exists := sessions[sessionID]
		sessMu.RUnlock()

		if !exists {
			return c.Status(404).JSON(fiber.Map{
				"error": "Session not found",
			})
		}

		// Set headers for SSE
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			// Send initial connection message
			fmt.Fprintf(w, "data: {\"type\":\"connected\",\"session_id\":\"%s\"}\n\n", sessionID)
			w.Flush()

			// Create ticker for keep-alive
			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case line, ok := <-session.Output:
					if !ok {
						// Channel closed, send completion and exit
						fmt.Fprintf(w, "data: {\"type\":\"closed\"}\n\n")
						w.Flush()
						return
					}
					// Send output line
					fmt.Fprintf(w, "data: {\"type\":\"output\",\"data\":%q}\n\n", line)
					w.Flush()

				case err := <-session.Error:
					// Send error
					fmt.Fprintf(w, "data: {\"type\":\"error\",\"error\":%q}\n\n", err.Error())
					w.Flush()

				case <-ticker.C:
					// Send keep-alive ping
					fmt.Fprintf(w, ": keep-alive\n\n")
					w.Flush()

				case <-c.Context().Done():
					// Client disconnected
					return
				}
			}
		})

		return nil
	}
}

// InterruptAgent stops a running AI agent process
func InterruptAgent() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("sessionId")

		sessMu.RLock()
		session, exists := sessions[sessionID]
		sessMu.RUnlock()

		if !exists {
			return c.Status(404).JSON(fiber.Map{
				"error": "Session not found",
			})
		}

		session.mu.Lock()
		isRunning := session.isRunning
		session.mu.Unlock()

		if !isRunning {
			return c.JSON(fiber.Map{
				"status":  "already_stopped",
				"message": "Process is not running",
			})
		}

		// Cancel the context, which will kill the process
		session.Cancel()

		return c.JSON(fiber.Map{
			"status":     "interrupted",
			"session_id": sessionID,
		})
	}
}

// GetAgentStatus returns the status of an AI agent session
func GetAgentStatus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("sessionId")

		sessMu.RLock()
		session, exists := sessions[sessionID]
		sessMu.RUnlock()

		if !exists {
			return c.Status(404).JSON(fiber.Map{
				"error": "Session not found",
			})
		}

		session.mu.Lock()
		isRunning := session.isRunning
		session.mu.Unlock()

		return c.JSON(fiber.Map{
			"session_id": session.ID,
			"command":    session.Command,
			"args":       session.Args,
			"is_running": isRunning,
			"start_time": session.StartTime,
			"uptime":     time.Since(session.StartTime).Seconds(),
		})
	}
}

// CleanupSessions removes old completed sessions
func CleanupSessions() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessMu.Lock()
		defer sessMu.Unlock()

		cleaned := 0
		for id, session := range sessions {
			session.mu.Lock()
			isRunning := session.isRunning
			session.mu.Unlock()

			// Remove sessions that have been completed for more than 1 hour
			if !isRunning && time.Since(session.StartTime) > time.Hour {
				delete(sessions, id)
				cleaned++
			}
		}

		return c.JSON(fiber.Map{
			"cleaned": cleaned,
			"active":  len(sessions),
		})
	}
}
