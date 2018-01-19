package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	file, err := os.OpenFile("/var/log/osquery-sgt.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}
	logger.Level = logrus.InfoLevel
}

func Info(args ...interface{}) {
	logger.Info(args...)
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

//func Warnf(args ...interface{}) {
//logger.Warnf(args...)
//}
