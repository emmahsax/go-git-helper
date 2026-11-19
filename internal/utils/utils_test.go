package utils

import (
	"errors"
	"strings"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
)

type MockLogger struct {
	FatalCalled bool
}

func (m *MockLogger) Fatal(v ...interface{}) {
	m.FatalCalled = true
}

func TestHandleError(t *testing.T) {
	logger := &MockLogger{}

	HandleError(errors.New("test error"), false, logger)

	if !logger.FatalCalled {
		t.Errorf("Expected Fatal to be called on logger")
	}
}

func Test_SplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"LF", "a\nb\n", []string{"a", "b", ""}},
		{"CRLF", "a\r\nb\r\n", []string{"a", "b", ""}},
		{"empty", "", []string{""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := SplitLines(test.input)
			if len(got) != len(test.expected) {
				t.Fatalf("expected %q, got %q", test.expected, got)
			}

			for i := range got {
				if got[i] != test.expected[i] {
					t.Errorf("expected %q, got %q", test.expected, got)
				}
			}
		})
	}
}

func Test_ContainsFold(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		s        string
		expected bool
	}{
		{"exact match", []string{"github.com", "gitlab.com"}, "github.com", true},
		{"case-insensitive match", []string{"GitHub.com"}, "github.com", true},
		{"case-insensitive match, reversed case", []string{"github.com"}, "GitHub.Com", true},
		{"no match", []string{"github.com"}, "gitlab.com", false},
		{"empty list", nil, "github.com", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ContainsFold(test.list, test.s); got != test.expected {
				t.Errorf("expected %v, got %v", test.expected, got)
			}
		})
	}
}

func Test_FindTicketMatch(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{"match", "proj-123 something", "proj-123"},
		{"no match", "no ticket reference here", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FindTicketMatch(test.title); got != test.expected {
				t.Errorf("expected %q, got %q", test.expected, got)
			}
		})
	}
}

func Test_TicketPrefix(t *testing.T) {
	if got := TicketPrefix("proj-123"); got != "proj" {
		t.Errorf("expected %q, got %q", "proj", got)
	}
}

func Test_PrependTicketLink_NoMatch(t *testing.T) {
	got := PrependTicketLink("existing body", "no ticket reference here", "pull request", []string{"https://example.com/browse"}, true)

	if got != "existing body" {
		t.Errorf("expected body unchanged, got %q", got)
	}
}

// Regression: with no base URL configured, the feature still asks and
// still inserts the bare reference -- a base URL only upgrades that bare
// reference into a real link, it never gates whether the reference (or the
// prompt about it) appears at all. Without this fallback, anyone relying
// on their platform's own autolinking (e.g. GitHub autolink references)
// would stop getting any reference inserted the moment they didn't also
// configure ticket_urls.
func Test_PrependTicketLink_NoBaseURL_FallsBackToBareReference(t *testing.T) {
	original := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = original })

	asked := false
	commandline.AskYesNoQuestion = func(question string) bool {
		asked = true
		return true
	}

	got := PrependTicketLink("existing body", "proj-123 something", "pull request", nil, true)

	expected := "### [proj-123]\n\nexisting body"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
	if !asked {
		t.Error("expected the prompt even with no base URL configured")
	}
}

func Test_PrependTicketLink_NonInteractiveAlwaysIncludes(t *testing.T) {
	got := PrependTicketLink("existing body", "proj-123 something", "pull request", []string{"https://example.com/browse"}, false)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\nexisting body"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func Test_PrependTicketLink_TrimsTrailingSlashFromBaseURL(t *testing.T) {
	got := PrependTicketLink("", "proj-123 something", "pull request", []string{"https://example.com/browse/"}, false)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func Test_PrependTicketLink_InteractiveAccepts(t *testing.T) {
	original := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = original })

	var askedQuestion string
	commandline.AskYesNoQuestion = func(question string) bool {
		askedQuestion = question
		return true
	}

	got := PrependTicketLink("existing body", "proj-123 something", "merge request", []string{"https://example.com/browse"}, true)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\nexisting body"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
	if askedQuestion == "" || !containsAll(askedQuestion, "proj-123", "merge request") {
		t.Errorf("expected the prompt to mention the match and kind, got %q", askedQuestion)
	}
}

func Test_PrependTicketLink_InteractiveDeclines(t *testing.T) {
	original := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = original })

	commandline.AskYesNoQuestion = func(question string) bool { return false }

	got := PrependTicketLink("existing body", "proj-123 something", "pull request", []string{"https://example.com/browse"}, true)

	if got != "existing body" {
		t.Errorf("expected body unchanged when declined, got %q", got)
	}
}

// Regression: the ticket link must be addable even with no template-derived
// body at all (an empty string), not just appended to existing content.
func Test_PrependTicketLink_WorksWithEmptyBody(t *testing.T) {
	got := PrependTicketLink("", "proj-123 something", "pull request", []string{"https://example.com/browse"}, false)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

// Regression: when more than one base URL is configured for the same
// reference, "which URL?" must not be asked if the user declines the link
// entirely -- that would waste a choice they never needed to make.
func Test_PrependTicketLink_AmbiguousCandidates_NotAskedIfDeclined(t *testing.T) {
	original := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = original })
	commandline.AskYesNoQuestion = func(question string) bool { return false }

	originalMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalMultipleChoice })
	askedWhichURL := false
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		askedWhichURL = true
		return choices[0]
	}

	candidates := []string{"https://company.atlassian.net/browse", "https://linear.app/company/issue"}
	got := PrependTicketLink("existing body", "ops-123 something", "pull request", candidates, true)

	if askedWhichURL {
		t.Error("expected no 'which URL' prompt when the link itself was declined")
	}
	if got != "existing body" {
		t.Errorf("expected body unchanged when declined, got %q", got)
	}
}

// Regression: when more than one base URL is configured and the user
// confirms they want the link, "which URL?" must be asked (only) then.
func Test_PrependTicketLink_AmbiguousCandidates_AskedAfterAccepted(t *testing.T) {
	original := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = original })
	commandline.AskYesNoQuestion = func(question string) bool { return true }

	originalMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalMultipleChoice })
	var offeredChoices []string
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		offeredChoices = choices
		return choices[1]
	}

	candidates := []string{"https://company.atlassian.net/browse", "https://linear.app/company/issue"}
	got := PrependTicketLink("", "ops-123 something", "pull request", candidates, true)

	if len(offeredChoices) != 2 {
		t.Fatalf("expected to be asked which of 2 candidates, got %v", offeredChoices)
	}
	expected := "### [ops-123](https://linear.app/company/issue/ops-123)\n\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

// Regression: non-interactively, nothing can answer a prompt, so the first
// configured candidate must be used instead of asking.
func Test_PrependTicketLink_AmbiguousCandidates_NonInteractiveUsesFirst(t *testing.T) {
	candidates := []string{"https://company.atlassian.net/browse", "https://linear.app/company/issue"}
	got := PrependTicketLink("", "ops-123 something", "pull request", candidates, false)

	expected := "### [ops-123](https://company.atlassian.net/browse/ops-123)\n\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func Test_EncodeDecodeTicketBaseURLCandidates(t *testing.T) {
	tests := [][]string{
		nil,
		{"https://example.com/browse"},
		{"https://example.com/browse", "https://linear.app/company/issue"},
	}

	for _, candidates := range tests {
		encoded := EncodeTicketBaseURLCandidates(candidates)
		decoded := DecodeTicketBaseURLCandidates(encoded)

		if len(decoded) != len(candidates) {
			t.Fatalf("expected %v, got %v", candidates, decoded)
		}
		for i := range candidates {
			if decoded[i] != candidates[i] {
				t.Errorf("expected %v, got %v", candidates, decoded)
			}
		}
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			return false
		}
	}

	return true
}
