package logger

import "github.com/gogf/gf/os/glog"

var defaultLogger = glog.DefaultLogger()

func IsDebug() bool {
	return defaultLogger.GetLevel()&glog.LEVEL_DEBU > 0
}

// Debug prints the logging content with [DEBU] header and newline.
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

// Debugf prints the logging content with [DEBU] header, custom format and newline.
func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

func SetDebug(debug bool) {
	defaultLogger.SetDebug(debug)
}
