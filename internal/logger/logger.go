package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	ERROR LogLevel = iota
	WARN
	INFO
	DEBUG
	TRACE
)

var (
	logFile       *os.File
	Logger        *log.Logger
	currentLevel  LogLevel = INFO
	enableConsole bool     = false
)

// Init initializes the logger with optional verbosity level
func Init() error {
	return InitWithLevel(INFO, false)
}

// InitWithLevel initializes the logger with a specific verbosity level
func InitWithLevel(level LogLevel, console bool) error {
	currentLevel = level
	enableConsole = console
	
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

	// Create logger with more detailed format
	Logger = log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
	
	logMessage("INFO", fmt.Sprintf("Logger initialized. Log file: %s, Level: %s, Console: %t", 
		logPath, getLevelName(currentLevel), enableConsole))
	return nil
}

// getLevelName returns the string representation of log level
func getLevelName(level LogLevel) string {
	switch level {
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown:0"
	}
	// Get just the filename, not the full path
	filename := filepath.Base(file)
	return fmt.Sprintf("%s:%d", filename, line)
}

// logMessage is the core logging function
func logMessage(level string, message string) {
	if Logger == nil {
		return
	}
	
	caller := getCallerInfo()
	logLine := fmt.Sprintf("[%s] [%s] %s", level, caller, message)
	
	Logger.Println(logLine)
	
	// Also log to console if enabled
	if enableConsole {
		fmt.Fprintf(os.Stderr, "%s\n", logLine)
	}
}

// Error logs error level messages
func Error(message string) {
	if currentLevel >= ERROR {
		logMessage("ERROR", message)
	}
}

// Errorf logs formatted error level messages
func Errorf(format string, args ...interface{}) {
	if currentLevel >= ERROR {
		logMessage("ERROR", fmt.Sprintf(format, args...))
	}
}

// Warn logs warning level messages
func Warn(message string) {
	if currentLevel >= WARN {
		logMessage("WARN", message)
	}
}

// Warnf logs formatted warning level messages
func Warnf(format string, args ...interface{}) {
	if currentLevel >= WARN {
		logMessage("WARN", fmt.Sprintf(format, args...))
	}
}

// Info logs info level messages
func Info(message string) {
	if currentLevel >= INFO {
		logMessage("INFO", message)
	}
}

// Infof logs formatted info level messages
func Infof(format string, args ...interface{}) {
	if currentLevel >= INFO {
		logMessage("INFO", fmt.Sprintf(format, args...))
	}
}

// Debug logs debug level messages
func Debug(message string) {
	if currentLevel >= DEBUG {
		logMessage("DEBUG", message)
	}
}

// Debugf logs formatted debug level messages
func Debugf(format string, args ...interface{}) {
	if currentLevel >= DEBUG {
		logMessage("DEBUG", fmt.Sprintf(format, args...))
	}
}

// Trace logs trace level messages (most verbose)
func Trace(message string) {
	if currentLevel >= TRACE {
		logMessage("TRACE", message)
	}
}

// Tracef logs formatted trace level messages
func Tracef(format string, args ...interface{}) {
	if currentLevel >= TRACE {
		logMessage("TRACE", fmt.Sprintf(format, args...))
	}
}

// SetLevel changes the current logging level
func SetLevel(level LogLevel) {
	currentLevel = level
	Infof("Log level changed to: %s", getLevelName(level))
}

// SetConsoleOutput enables/disables console output
func SetConsoleOutput(enabled bool) {
	enableConsole = enabled
	Infof("Console output set to: %t", enabled)
}

// Close closes the log file
func Close() {
	if Logger != nil {
		Info("Logger shutting down")
	}
	if logFile != nil {
		logFile.Close()
	}
}
