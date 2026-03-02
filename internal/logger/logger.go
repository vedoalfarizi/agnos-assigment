package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
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
