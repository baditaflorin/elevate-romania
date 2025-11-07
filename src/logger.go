package main

import (
	"fmt"
	"time"
)

// SimpleLogger implements a simple structured logger
type SimpleLogger struct {
	prefix string
}

// NewLogger creates a new logger instance
func NewLogger(prefix string) *SimpleLogger {
	return &SimpleLogger{
		prefix: prefix,
	}
}

// Info logs an informational message
func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	l.log("INFO", msg, args...)
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	l.log("WARN", msg, args...)
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	l.log("ERROR", msg, args...)
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	l.log("DEBUG", msg, args...)
}

// log is the internal logging function
func (l *SimpleLogger) log(level, msg string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := ""
	if l.prefix != "" {
		prefix = fmt.Sprintf("[%s] ", l.prefix)
	}
	
	message := fmt.Sprintf(msg, args...)
	fmt.Printf("%s [%s] %s%s\n", timestamp, level, prefix, message)
}
