package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const commandLogFile = "command-summary.md"
const maxLogLines = 20

var (
	logMutex sync.Mutex
)

// LogInternalCommand logs an internal command/tool execution to command-summary.md
func LogInternalCommand(commandName, action, target string) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	// Read existing file
	content, err := os.ReadFile(commandLogFile)
	var lines []string

	if err == nil {
		// File exists, parse existing lines
		lines = strings.Split(strings.TrimSpace(string(content)), "\n")
	} else {
		// File doesn't exist, create header
		lines = []string{
			"# AI Command Log - Last 20 Internal Commands",
			"",
		}
	}

	// Create new entry
	timestamp := time.Now().Format("15:04:05")
	newEntry := fmt.Sprintf("- `[%s]` **%s** → %s | Target: `%s`",
		timestamp, commandName, action, target)

	// Add new entry at the top (after header)
	if len(lines) >= 2 {
		lines = append(lines[:2], append([]string{newEntry}, lines[2:]...)...)
	} else {
		lines = append(lines, newEntry)
	}

	// Keep only maxLogLines entries (plus 2 header lines)
	if len(lines) > maxLogLines+2 {
		lines = lines[:maxLogLines+2]
	}

	// Write back to file
	output := strings.Join(lines, "\n") + "\n"

	return os.WriteFile(commandLogFile, []byte(output), 0644)
}
