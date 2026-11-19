package utils

import (
	"fmt"
	"log"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
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

// ticketPattern matches a PROJ-123-style ticket reference, the convention
// shared by Jira, Linear, Trello, and most other issue trackers.
var ticketPattern = regexp.MustCompile(`[A-Za-z]+-\d+`)

// FindTicketMatch returns the first PROJ-123-style ticket reference found
// in title (e.g. "proj-123" in "proj-123-fix-the-thing"), or "" if none.
func FindTicketMatch(title string) string {
	return ticketPattern.FindString(title)
}

// TicketPrefix returns match's letters-only prefix (e.g. "proj" from
// "proj-123"). match must be non-empty, as returned by FindTicketMatch.
func TicketPrefix(match string) string {
	return strings.SplitN(match, "-", 2)[0]
}

// PrependTicketLink optionally prepends a "### [PROJ-123]" header to body,
// naming the first ticket reference found in title (see FindTicketMatch).
// kind names what's being created for the confirmation prompt (e.g. "pull
// request", "merge request"). Works independently of whether body is empty
// (no PR/MR template configured) -- the ticket reference is worth
// surfacing either way.
//
// When baseURL is configured (non-""), the reference becomes a real link:
// "### [PROJ-123](baseURL/PROJ-123)". With no baseURL, it falls back to a
// bare "### [PROJ-123]", trusting whatever the destination platform/project
// already does with it (e.g. GitHub's own autolink references, or a
// GitLab/Jira integration) rather than guessing at a URL -- baseURL only
// upgrades that bare reference into an explicit link, it never gates
// whether the reference appears at all.
//
// Interactively, asks before including it; non-interactively, always
// includes it when title has a reference. Returns body unchanged if title
// has no such reference, or if the user declines when asked.
func PrependTicketLink(body, title, kind, baseURL string, interactiveMode bool) string {
	match := FindTicketMatch(title)
	if match == "" {
		return body
	}

	include := true
	if interactiveMode {
		include = commandline.AskYesNoQuestion(
			fmt.Sprintf("Include a link to the project ticket (%s) in the beginning of the %s body?", match, kind),
		)
	}

	if !include {
		return body
	}

	if baseURL == "" {
		return "### [" + match + "]\n\n" + body
	}

	url := strings.TrimSuffix(baseURL, "/") + "/" + match
	return "### [" + match + "](" + url + ")\n\n" + body
}
