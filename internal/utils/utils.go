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
// baseURLCandidates is usually 0 or 1 URLs; 2+ means the same prefix/repo
// genuinely maps to more than one tracker (e.g. "OPS" meaning both Jira
// and Linear). Deliberately asked about only *after* confirming a link
// should be included at all -- asking "which URL?" first would waste that
// answer if the user then declines the link entirely. With no candidate at
// all, it falls back to a bare "### [PROJ-123]", trusting whatever the
// destination platform/project already does with it (e.g. GitHub's own
// autolink references, or a GitLab/Jira integration) rather than guessing
// at a URL -- a configured candidate only upgrades that bare reference
// into an explicit link, it never gates whether the reference appears at
// all.
//
// Interactively, asks before including it; non-interactively, always
// includes it when title has a reference. Returns body unchanged if title
// has no such reference, or if the user declines when asked.
func PrependTicketLink(body, title, kind string, baseURLCandidates []string, interactiveMode bool) string {
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

	baseURL := chosenTicketBaseURL(baseURLCandidates, match, interactiveMode)
	if baseURL == "" {
		return "### [" + match + "]\n\n" + body
	}

	url := strings.TrimSuffix(baseURL, "/") + "/" + match
	return "### [" + match + "](" + url + ")\n\n" + body
}

// chosenTicketBaseURL picks which of candidates to use: "" if there are
// none, the only one if there's just one, or -- when there's more than
// one -- whichever the user picks interactively (asked "for match", e.g.
// "PROJ-123"), or the first one listed when non-interactive.
func chosenTicketBaseURL(candidates []string, match string, interactiveMode bool) string {
	if len(candidates) == 0 {
		return ""
	}

	if len(candidates) == 1 || !interactiveMode {
		return candidates[0]
	}

	return commandline.AskMultipleChoice(
		fmt.Sprintf("Found multiple ticket URLs configured for %s. Choose one to use:", match),
		candidates,
	)
}

// ticketBaseURLCandidatesSeparator joins/splits a ticket base URL
// candidate list when it travels through the string-only options map each
// provider's NewXPullRequest/NewXMergeRequest constructor already uses
// (the same convention "draft" already uses, as "true"/"false"). "\n" is
// safe since a URL never contains a literal newline.
const ticketBaseURLCandidatesSeparator = "\n"

// EncodeTicketBaseURLCandidates and DecodeTicketBaseURLCandidates convert
// a ticket base URL candidate list to and from the single string that
// travels through the options map.
func EncodeTicketBaseURLCandidates(candidates []string) string {
	return strings.Join(candidates, ticketBaseURLCandidatesSeparator)
}

func DecodeTicketBaseURLCandidates(s string) []string {
	if s == "" {
		return nil
	}

	return strings.Split(s, ticketBaseURLCandidatesSeparator)
}
