package utils

import (
	"log"
	"runtime/debug"
	"strings"
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

// SplitLines splits command output into lines, accepting both "\n" and the
// "\r\n" endings that git can emit on Windows.
func SplitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}
