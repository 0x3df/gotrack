package db

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogEvent appends a timestamped event line to workspace/events.log.
// Format: 2006-01-02T15:04:05Z07:00 | event_type | detail
// Failures are silently ignored — logging must never block the caller.
func LogEvent(eventType, detail string) {
	workspace, err := GetWorkspacePath()
	if err != nil || workspace == "" {
		return
	}
	path := filepath.Join(workspace, "events.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	ts := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("%s | %-22s | %s\n", ts, eventType, detail)
	_, _ = f.WriteString(line)
}
