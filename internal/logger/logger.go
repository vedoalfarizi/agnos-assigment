package logger

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Context keys for request metadata
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	ClientIPKey  contextKey = "client_ip"
	PathKey      contextKey = "path"
	MethodKey    contextKey = "method"
)

type Logger struct {
	*logrus.Logger
}

// global logger instance
var log *Logger

// Init initializes the global logger with the specified level and format
func Init(level, format string) {
	log = New(level, format)
}

// Get returns the global logger instance
func Get() *Logger {
	if log == nil {
		// Fallback to a default logger if Init wasn't called
		log = New("info", "text")
	}
	return log
}

// Convenience functions that use the global logger
func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

func New(level, format string) *Logger {
	l := logrus.New()

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	l.SetLevel(logLevel)

	l.SetOutput(os.Stdout)

	if format == "json" {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		l.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05",
			FullTimestamp:   true,
			ForceColors:     true,
		})
	}

	return &Logger{l}
}

func (l *Logger) WithOutput(w io.Writer) *Logger {
	l.SetOutput(w)
	return l
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// extractRequestContext pulls request metadata from context into logrus Fields
func extractRequestContext(ctx context.Context) logrus.Fields {
	fields := logrus.Fields{}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		fields["request_id"] = requestID
	}

	if clientIP, ok := ctx.Value(ClientIPKey).(string); ok && clientIP != "" {
		fields["client_ip"] = clientIP
	}

	if path, ok := ctx.Value(PathKey).(string); ok && path != "" {
		fields["path"] = path
	}

	if method, ok := ctx.Value(MethodKey).(string); ok && method != "" {
		fields["method"] = method
	}

	return fields
}

// Context-aware logging functions that automatically include request metadata
func InfofWithContext(ctx context.Context, format string, args ...interface{}) {
	entry := Get().WithFields(extractRequestContext(ctx))
	entry.Infof(format, args...)
}

func WarnfWithContext(ctx context.Context, format string, args ...interface{}) {
	entry := Get().WithFields(extractRequestContext(ctx))
	entry.Warnf(format, args...)
}

func ErrorfWithContext(ctx context.Context, format string, args ...interface{}) {
	entry := Get().WithFields(extractRequestContext(ctx))
	entry.Errorf(format, args...)
}

func DebugfWithContext(ctx context.Context, format string, args ...interface{}) {
	entry := Get().WithFields(extractRequestContext(ctx))
	entry.Debugf(format, args...)
}
