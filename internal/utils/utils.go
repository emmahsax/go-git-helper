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

// ContainsFold reports whether s is present in list, ignoring case.
// Hostnames are case-insensitive, so comparisons against them should be too.
func ContainsFold(list []string, s string) bool {
	for _, v := range list {
		if strings.EqualFold(v, s) {
			return true
		}
	}

	return false
}
