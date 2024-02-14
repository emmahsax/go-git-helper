package utils

import (
	"log"
	"runtime/debug"
)

type Logger interface {
	Fatal(v ...interface{})
}

func HandleError(err error, debugB bool, logger Logger) {
	if debugB {
		debug.PrintStack()
	}

	if logger == nil {
		log.Fatal(err)
	} else {
		logger.Fatal(err)
	}
}
