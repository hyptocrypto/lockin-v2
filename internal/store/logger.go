package store

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger *log.Logger
var logFile *os.File

// InitLogger initializes the file logger in the .lockin directory
func InitLogger() error {
	logPath := filepath.Join(GetConfigDir(), "lockin.log")

	// Ensure directory exists
	if err := os.MkdirAll(GetConfigDir(), 0700); err != nil {
		return err
	}

	// Open log file in append mode
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	logFile = f
	logger = log.New(f, "", 0)

	LogInfo("=== LockIn started ===")
	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		LogInfo("=== LockIn stopped ===")
		logFile.Close()
	}
}

// timestamp returns current timestamp for logs
func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// LogInfo logs an info message
func LogInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[%s] INFO: %s", timestamp(), msg)
	}
}

// LogError logs an error message
func LogError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[%s] ERROR: %s", timestamp(), msg)
	}
}

// LogDebug logs a debug message
func LogDebug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[%s] DEBUG: %s", timestamp(), msg)
	}
}
