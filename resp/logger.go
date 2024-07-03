package resp

import (
	"fmt"
	"log"
	"os"
)

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger struct to wrap the standard log package
type Logger struct {
	logger *log.Logger
	level  int
}

// NewLogger creates a new Logger instance
func NewLogger(level int) *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds),
		level:  level,
	}
}

// Log methods
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= LogLevelError {
		l.log("ERROR", format, v...)
	}
}

func (l *Logger) log(level, format string, v ...interface{}) {
	l.logger.Printf("%s: %s", level, fmt.Sprintf(format, v...))
}
