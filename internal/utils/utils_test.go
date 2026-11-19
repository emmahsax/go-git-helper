package utils

import (
	"errors"
	"testing"
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
