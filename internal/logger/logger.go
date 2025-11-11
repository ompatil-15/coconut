package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogLevel int

const (
	InfoLevel LogLevel = iota
	WarnLevel
	ErrorLevel
)

func (l LogLevel) String() string {
	switch l {
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	file *os.File
	mu   sync.Mutex
}

func New() (*Logger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	logDir := filepath.Join(home, ".coconut", "logs")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create log dir: %w", err)
	}

	logPath := filepath.Join(logDir, "coconut.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{file: f}, nil
}

func (lg *Logger) log(level LogLevel, format string, args ...interface{}) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if lg.file == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	fmt.Fprintf(lg.file, "%s [%s] %s\n", timestamp, level.String(), message)
}

func (lg *Logger) Info(format string, args ...interface{})  { lg.log(InfoLevel, format, args...) }
func (lg *Logger) Warn(format string, args ...interface{})  { lg.log(WarnLevel, format, args...) }
func (lg *Logger) Error(format string, args ...interface{}) { lg.log(ErrorLevel, format, args...) }

func (lg *Logger) Close() {
	if lg.file != nil {
		_ = lg.file.Close()
	}
}
