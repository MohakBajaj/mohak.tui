package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[LogLevel]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

var levelColors = map[LogLevel]string{
	LevelDebug: "\033[36m", // cyan
	LevelInfo:  "\033[32m", // green
	LevelWarn:  "\033[33m", // yellow
	LevelError: "\033[31m", // red
}

const (
	colorReset = "\033[0m"
	colorDim   = "\033[2m"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// Logger is a structured logger
type Logger struct {
	minLevel   LogLevel
	service    string
	jsonFormat bool
}

// NewLogger creates a new logger instance
func NewLogger(service string) *Logger {
	minLevel := LevelInfo
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		switch strings.ToLower(lvl) {
		case "debug":
			minLevel = LevelDebug
		case "info":
			minLevel = LevelInfo
		case "warn":
			minLevel = LevelWarn
		case "error":
			minLevel = LevelError
		}
	}

	jsonFormat := os.Getenv("LOG_FORMAT") == "json"

	return &Logger{
		minLevel:   minLevel,
		service:    service,
		jsonFormat: jsonFormat,
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.minLevel
}

func (l *Logger) log(level LogLevel, message string, context map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     levelNames[level],
		Message:   message,
		Service:   l.service,
		Context:   context,
	}

	if l.jsonFormat {
		data, _ := json.Marshal(entry)
		fmt.Fprintln(os.Stderr, string(data))
	} else {
		// Pretty format
		var contextStr string
		if len(context) > 0 {
			data, _ := json.Marshal(context)
			contextStr = fmt.Sprintf(" %s%s%s", colorDim, string(data), colorReset)
		}

		fmt.Fprintf(os.Stderr, "%s%s%s %s%-5s%s %s[%s]%s %s%s\n",
			colorDim, entry.Timestamp, colorReset,
			levelColors[level], entry.Level, colorReset,
			colorDim, l.service, colorReset,
			message, contextStr)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context)
	l.log(LevelDebug, message, ctx)
}

// Info logs an info message
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context)
	l.log(LevelInfo, message, ctx)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context)
	l.log(LevelWarn, message, ctx)
}

// Error logs an error message
func (l *Logger) Error(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context)
	l.log(LevelError, message, ctx)
}

// With returns a new logger with additional context
func (l *Logger) With(context map[string]interface{}) *Logger {
	return &Logger{
		minLevel:   l.minLevel,
		service:    l.service,
		jsonFormat: l.jsonFormat,
	}
}

// Helper to merge context maps
func mergeContext(contexts []map[string]interface{}) map[string]interface{} {
	if len(contexts) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for _, ctx := range contexts {
		for k, v := range ctx {
			result[k] = v
		}
	}
	return result
}

// Ctx is a helper to create context maps
func Ctx(keyValues ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for i := 0; i < len(keyValues)-1; i += 2 {
		if key, ok := keyValues[i].(string); ok {
			result[key] = keyValues[i+1]
		}
	}
	return result
}
