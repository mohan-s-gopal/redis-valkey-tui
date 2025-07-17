package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	logFile *os.File
	Logger  *log.Logger
)

// Init initializes the logger
func Init() error {
	// Create logs directory in user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(home, ".redis-cli-dashboard", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logPath := filepath.Join(logDir, "app.log")
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create logger
	Logger = log.New(logFile, "", log.LstdFlags)
	Logger.Printf("Logger initialized. Log file: %s", logPath)
	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
