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
