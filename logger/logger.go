package logger

import (
	"bytes"
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// A simple hook to allow for logging to stdout and a file at the same time
type stdOutHook struct{}

func (hook *stdOutHook) Fire(entry *logrus.Entry) error {
	b := &bytes.Buffer{}
	b.WriteString(entry.Message)
	b.WriteByte('\n')
	os.Stdout.Write(b.Bytes())
	return nil
}

func (hook *stdOutHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func init() {
	logger = logrus.New()
	file, err := os.OpenFile("/var/log/osquery-sgt.log", os.O_CREATE|os.O_WRONLY, 0666)
	logger.SetLevel(logrus.InfoLevel)
	if err == nil {
		// Add a hook so stdout also contains log messages along with the output file
		// This is helpful for debugging in an active console
		logger.Hooks.Add(&stdOutHook{})
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func WithFields(args ...interface{}) *logrus.Entry {
	return logger.WithFields(logrus.Fields{})
}
