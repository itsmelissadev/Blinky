package logger

import (
	"fmt"
	"time"
)

var (
	isDebug = false
)

func Init(debug bool) {
	if debug {
		isDebug = true
	}
}

func IsDebug() bool {
	return isDebug
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

func format(level, color, msg string) string {
	timestamp := time.Now().Format("15:04:05")
	return fmt.Sprintf("%s[%s]%s %s%-5s%s %s\n", colorGray, timestamp, colorReset, color, level, colorReset, msg)
}

func logmsg(level, color, msg string, args ...interface{}) {
	output := format(level, color, fmt.Sprintf(msg, args...))
	fmt.Print(output)
}

func Error(msg string, args ...interface{}) {
	logmsg("ERROR", colorRed, msg, args...)
}

func Success(msg string, args ...interface{}) {
	logmsg("SUCCESS", colorGreen, msg, args...)
}

func Info(msg string, args ...interface{}) {
	logmsg("INFO", colorCyan, msg, args...)
}

func Debug(msg string, args ...interface{}) {
	if isDebug {
		logmsg("DEBUG", colorBlue, msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	logmsg("WARN", colorYellow, msg, args...)
}
