package logg

import (
	"fmt"
	"log"

	"github.com/ondevice/ondevice/util"
)

const (
	// DEBUG -- debug log level
	DEBUG = iota
	// INFO -- info log level
	INFO = iota
	// WARNING -- warning log level
	WARNING = iota
	// ERROR -- error log level
	ERROR = iota
	// FATAL -- fatal log level
	FATAL = iota
)

var _level = INFO
var _levels = map[int]string{}

func init() {
	// Debug, Info, Warning, Error, Fatal
	_levels[DEBUG] = "Debug"
	_levels[INFO] = "Info"
	_levels[WARNING] = "Warning"
	_levels[ERROR] = "Error"
	_levels[FATAL] = "Fatal"
}

// Log -- log message with the given log level
func Log(level int, args ...interface{}) {
	prefix := fmt.Sprintf("%s: ", _getName(level))
	args = append([]interface{}{prefix}, args...)
	if level == FATAL {
		log.Fatal(args...)
	}
	if level < _level {
		return
	}

	log.Print(args...)
}

// Logf -- formatted log message with the given log level
func Logf(level int, format string, args ...interface{}) {
	format = fmt.Sprintf("%s: %s", _getName(level), format)

	if level == FATAL {
		log.Fatalf(format, args...)
	}
	if level < _level {
		return
	}

	log.Printf(format, args...)
}

// Fatal -- fail with the given error message
func Fatal(args ...interface{}) {
	Log(FATAL, args...)
}

// Fatalf -- fail with the given error message
func Fatalf(fmt string, args ...interface{}) {
	Logf(FATAL, fmt, args...)
}

// Error -- log error message
func Error(args ...interface{}) {
	Log(ERROR, args...)
}

// Errorf -- log formatted error message
func Errorf(fmt string, args ...interface{}) {
	Logf(ERROR, fmt, args...)
}

// Warning -- log warning message
func Warning(args ...interface{}) {
	Log(WARNING, args...)
}

// Warningf -- log formatted warning message
func Warningf(fmt string, args ...interface{}) {
	Logf(WARNING, fmt, args...)
}

// Info -- log info message
func Info(args ...interface{}) {
	Log(INFO, args...)
}

// Infof -- log formatted info message
func Infof(fmt string, args ...interface{}) {
	Logf(INFO, fmt, args...)
}

// Debug -- log debug message
func Debug(args ...interface{}) {
	Log(DEBUG, args...)
}

// Debugf -- log formatted debug message
func Debugf(fmt string, args ...interface{}) {
	Logf(DEBUG, fmt, args...)
}

// SetLevel -- Set the minimum log level
func SetLevel(level int) {
	_level = level
}

func _getName(level int) string {
	rc, _ := _levels[level]
	return rc
}

// FailWithAPIError -- call Fatal with a nice error message matching the API error we got
func FailWithAPIError(err util.APIError) {
	if err.Code() == util.NotFoundError {
		Fatal("Not found: ", err.Error())
	} else if err.Code() == util.AuthenticationError {
		Fatal("Authentication failed (try running ondevice login)")
	} else if err.Code() == util.ForbiddenError {
		Fatal("Access denied (are you sure your API key has the required roles?)")
	} else {
		Fatalf("Unexpected API error (code %d): %s", err.Code(), err.Error())
	}
}
