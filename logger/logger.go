package logger

import (
	"fmt"
	"log"
)

//LogLevel Type to model log levels used across this app
type LogLevel string

const (
	//DEBUG Debug messages level
	DEBUG LogLevel = "DEBUG"
	//INFO Information messages to track application state
	INFO LogLevel = "INFO"
	//WARN Recoverable and Ignorable Errors
	WARN LogLevel = "WARN"
	//ERROR Recoverable but needs to be fixed
	ERROR LogLevel = "ERROR"
)

//IsValidLogLevel Validates if a log level name is valid
func IsValidLogLevel(name string) bool {
	isValid := false
	desiredLevel := LogLevel(name)

	validLevels := GetSupportedLogLevels()
	for _, level := range validLevels {
		if desiredLevel == level {
			isValid = true
			break
		}
	}

	return isValid
}

//GetSupportedLogLevels Returns log levels followed in this application
func GetSupportedLogLevels() []LogLevel {
	return []LogLevel{DEBUG, INFO, WARN, ERROR}
}

func logWithLabel(label LogLevel, v ...interface{}) {
	log.Print(fmt.Sprintf("[%v] %s", label, fmt.Sprint(v...)))
}

func logFormattedWithLabel(label LogLevel, format string, v ...interface{}) {
	log.Print(fmt.Sprintf("[%v] %s", label, fmt.Sprintf(format, v...)))
}

//Debug Debug messages from the logger
func Debug(v ...interface{}) {
	logWithLabel(DEBUG, v...)
}

//Info Info messages from the logger
func Info(v ...interface{}) {
	logWithLabel(INFO, v...)
}

//Warn Warning messages from the logger
func Warn(v ...interface{}) {
	logWithLabel(WARN, v...)
}

//Error Error messages from the logger
func Error(v ...interface{}) {
	logWithLabel(ERROR, v...)
}

//Debugf Debug messages from the logger
func Debugf(format string, v ...interface{}) {
	logFormattedWithLabel(DEBUG, format, v...)
}

//Infof Info messages from the logger
func Infof(format string, v ...interface{}) {
	logFormattedWithLabel(INFO, format, v...)
}

//Warnf Warning messages from the logger
func Warnf(format string, v ...interface{}) {
	logFormattedWithLabel(WARN, format, v...)
}

//Errorf Error messages from the logger
func Errorf(format string, v ...interface{}) {
	logFormattedWithLabel(ERROR, format, v...)
}
