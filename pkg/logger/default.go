package logger

import "fmt"

func Info(msg string, args ...any) {
	mLogger.Info(msg, args...)
}

func Infof(format string, args ...any) {
	mLogger.Info(fmt.Sprintf(format, args...))
}

func Debug(msg string, args ...any) {
	mLogger.Debug(msg, args...)
}

func Debugf(format string, args ...any) {
	mLogger.Debug(fmt.Sprintf(format, args...))
}

func Warn(msg string, args ...any) {
	mLogger.Warn(msg, args...)
}

func Warnf(format string, args ...any) {
	mLogger.Warn(fmt.Sprintf(format, args...))
}

func Error(msg string, args ...any) {
	mLogger.Error(msg, args...)
}

func Errorf(format string, args ...any) {
	mLogger.Error(fmt.Sprintf(format, args...))
}
