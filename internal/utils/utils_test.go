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
	got := PrependTicketLink("existing body", "no ticket reference here", "pull request", "https://example.com/browse", true)

	if got != "existing body" {
		t.Errorf("expected body unchanged, got %q", got)
	}
}

// Regression: without a configured base URL, the feature must do nothing
// at all -- no prompt, no bare-bracket fallback -- regardless of whether
// the title has a ticket reference. The destination platform/project is
// trusted to do its own linkification (e.g. GitHub autolink references).
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

	got := PrependTicketLink("existing body", "proj-123 something", "pull request", "", true)

	expected := "### [proj-123]\n\nexisting body"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
	if !asked {
		t.Error("expected the prompt even with no base URL configured")
	}
}

func Test_PrependTicketLink_NonInteractiveAlwaysIncludes(t *testing.T) {
	got := PrependTicketLink("existing body", "proj-123 something", "pull request", "https://example.com/browse", false)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\nexisting body"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func Test_PrependTicketLink_TrimsTrailingSlashFromBaseURL(t *testing.T) {
	got := PrependTicketLink("", "proj-123 something", "pull request", "https://example.com/browse/", false)

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

	got := PrependTicketLink("existing body", "proj-123 something", "merge request", "https://example.com/browse", true)

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

	got := PrependTicketLink("existing body", "proj-123 something", "pull request", "https://example.com/browse", true)

	if got != "existing body" {
		t.Errorf("expected body unchanged when declined, got %q", got)
	}
}

// Regression: the ticket link must be addable even with no template-derived
// body at all (an empty string), not just appended to existing content.
func Test_PrependTicketLink_WorksWithEmptyBody(t *testing.T) {
	got := PrependTicketLink("", "proj-123 something", "pull request", "https://example.com/browse", false)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
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
