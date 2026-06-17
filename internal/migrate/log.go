package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Logger writes structured log entries to a dated file under %APPDATA%\MoveKit\logs\.
type Logger struct {
	mu   sync.Mutex
	file *os.File
}

// NewLogger creates a Logger that writes to %APPDATA%\MoveKit\logs\movekit-{date}.log.
// The log directory is created if it doesn't exist.
func NewLogger() (*Logger, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}

	logDir := filepath.Join(appData, "MoveKit", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("movekit-%s.log", date))

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return &Logger{file: f}, nil
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.write("INFO", format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.write("WARN", format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.write("ERROR", format, args...)
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) write(level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return
	}

	ts := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s [%s] %s\n", ts, level, msg)
	l.file.WriteString(line)
}

// maskPath replaces the user's profile directory with C:\Users\*** to avoid
// leaking the actual username in log output.
func maskPath(path string) string {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return path
	}
	if strings.HasPrefix(path, userProfile) {
		return `C:\Users\***` + path[len(userProfile):]
	}
	return path
}
