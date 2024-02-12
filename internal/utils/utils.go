package utils

import (
	"runtime/debug"
)

type Logger interface {
	Fatal(v ...interface{})
}

func HandleError(err error, debugB bool, logger Logger) {
	if debugB {
		debug.PrintStack()
	}

	var customLogger Logger
	if logger != nil {
		customLogger = logger
	}
	customLogger.Fatal(err)
}
