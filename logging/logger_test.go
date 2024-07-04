package logging

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		logFunc  func(*Logger)
		wantLog  string
		dontWant string
	}{
		{
			name:  "Debug log at Debug level",
			level: LogLevelDebug,
			logFunc: func(l *Logger) {
				l.Debug("test %s", "message")
			},
			wantLog: "DEBUG: test message",
		},
		{
			name:  "Info log at Info level",
			level: LogLevelInfo,
			logFunc: func(l *Logger) {
				l.Info("test %s", "message")
			},
			wantLog: "INFO: test message",
		},
		{
			name:  "Warn log at Warn level",
			level: LogLevelWarn,
			logFunc: func(l *Logger) {
				l.Warn("test %s", "message")
			},
			wantLog: "WARN: test message",
		},
		{
			name:  "Error log at Error level",
			level: LogLevelError,
			logFunc: func(l *Logger) {
				l.Error("test %s", "message")
			},
			wantLog: "ERROR: test message",
		},
		{
			name:  "Debug log at Info level",
			level: LogLevelInfo,
			logFunc: func(l *Logger) {
				l.Debug("test %s", "message")
			},
			dontWant: "DEBUG: test message",
		},
		{
			name:  "Debugln log at Debug level",
			level: LogLevelDebug,
			logFunc: func(l *Logger) {
				l.Debugln("test", "message")
			},
			wantLog: "DEBUG: test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := &Logger{
				logger: log.New(&buf, "", 0), // Remove time prefix for easier testing
				level:  tt.level,
			}

			tt.logFunc(l)

			got := buf.String()
			if tt.wantLog != "" && !strings.Contains(got, tt.wantLog) {
				t.Errorf("log output = %q, want to contain %q", got, tt.wantLog)
			}
			if tt.dontWant != "" && strings.Contains(got, tt.dontWant) {
				t.Errorf("log output = %q, don't want to contain %q", got, tt.dontWant)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{
		logger: log.New(&buf, "", 0),
		level:  LogLevelWarn,
	}

	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")

	got := buf.String()
	if strings.Contains(got, "DEBUG") || strings.Contains(got, "INFO") {
		t.Errorf("log contains DEBUG or INFO message when level is WARN")
	}
	if !strings.Contains(got, "WARN") || !strings.Contains(got, "ERROR") {
		t.Errorf("log doesn't contain WARN or ERROR message when level is WARN")
	}
}
