package ddd

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Log levels
const (
	LevelInfo  = "[\033[94mINFO\033[0m] "
	LevelWarn  = "[\033[93mWARN\033[0m] "
	LevelError = "[\033[91mERROR\033[0m] "
)

// Logger wraps the standard library logger with additional formatting
type Logger struct {
	*log.Logger
	dateFormat string
}

func NewLogger() *Logger {
	return &Logger{
		Logger:     log.New(os.Stdout, "", 0),
		dateFormat: "2006-01-02 15:04:05.000 -07:00",
	}
}

// getCallerInfo returns the file name and line number of the caller
func (l *Logger) getCallerInfo(skipFrames int) string {
	_, file, line, ok := runtime.Caller(skipFrames)
	if !ok {
		return "???:0"
	}

	// Extract just the filename from the full path
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]

	return file + ":" + strconv.Itoa(line)
}

// formatLogEntry creates a formatted log entry
func (l *Logger) formatLogEntry(level, caller, format string, args ...interface{}) string {
	timestamp := time.Now().Format(l.dateFormat)
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] %s %s: %s", timestamp, level, caller, message)
}

// SetOutput changes the output destination for the default logger
func (l *Logger) SetOutput(w *os.File) {
	l.Logger.SetOutput(w)
}

// SetDateFormat changes the date format for log timestamps
func (l *Logger) SetDateFormat(format string) {
	l.dateFormat = format
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...any) {
	l.Println(l.formatLogEntry(LevelInfo, l.getCallerInfo(2), format, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...any) {
	l.Println(l.formatLogEntry(LevelWarn, l.getCallerInfo(2), format, args...))
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	l.Println(l.formatLogEntry(LevelError, l.getCallerInfo(2), format, args...))
}
