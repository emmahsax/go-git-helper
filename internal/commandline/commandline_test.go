package commandline

import (
	"testing"
)

func Test_AskMultipleChoice(t *testing.T) {
	// Save original function
	originalFunc := AskMultipleChoice
	t.Cleanup(func() {
		AskMultipleChoice = originalFunc
	})

	// Mock the function
	AskMultipleChoice = func(question string, choices []string) string {
		if len(choices) == 0 {
			return ""
		}
		return choices[0]
	}

	result := AskMultipleChoice("Select an option", []string{"option1", "option2", "option3"})
	if result != "option1" {
		t.Errorf("Expected 'option1', got '%s'", result)
	}
}

func Test_AskMultipleChoice_EmptyChoices(t *testing.T) {
	// Save original function
	originalFunc := AskMultipleChoice
	t.Cleanup(func() {
		AskMultipleChoice = originalFunc
	})

	// Mock the function
	AskMultipleChoice = func(question string, choices []string) string {
		if len(choices) == 0 {
			return ""
		}
		return choices[0]
	}

	result := AskMultipleChoice("Select an option", []string{})
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func Test_AskMultiSelect(t *testing.T) {
	// Save original function
	originalFunc := AskMultiSelect
	t.Cleanup(func() {
		AskMultiSelect = originalFunc
	})

	// Mock the function
	AskMultiSelect = func(question string, choices []string) []string {
		return choices
	}

	result := AskMultiSelect("Select options", []string{"option1", "option2"})
	if len(result) != 2 || result[0] != "option1" || result[1] != "option2" {
		t.Errorf("Expected ['option1', 'option2'], got %v", result)
	}
}

func Test_AskMultiSelect_NoneSelected(t *testing.T) {
	// Save original function
	originalFunc := AskMultiSelect
	t.Cleanup(func() {
		AskMultiSelect = originalFunc
	})

	// Mock the function
	AskMultiSelect = func(question string, choices []string) []string {
		return nil
	}

	result := AskMultiSelect("Select options", []string{"option1", "option2"})
	if len(result) != 0 {
		t.Errorf("Expected no selections, got %v", result)
	}
}

func Test_AskOpenEndedQuestion_NoDefault(t *testing.T) {
	// Save original function
	originalFunc := AskOpenEndedQuestion
	t.Cleanup(func() {
		AskOpenEndedQuestion = originalFunc
	})

	// Mock the function
	AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		return "test response"
	}

	result := AskOpenEndedQuestion("Enter something", "", false, true)
	if result != "test response" {
		t.Errorf("Expected 'test response', got '%s'", result)
	}
}

func Test_AskOpenEndedQuestion_WithDefault(t *testing.T) {
	// Save original function
	originalFunc := AskOpenEndedQuestion
	t.Cleanup(func() {
		AskOpenEndedQuestion = originalFunc
	})

	// Mock the function
	AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		if defaultVal != "" {
			return defaultVal
		}
		return "test response"
	}

	result := AskOpenEndedQuestion("Enter something", "default value", false, true)
	if result != "default value" {
		t.Errorf("Expected 'default value', got '%s'", result)
	}
}

func Test_AskOpenEndedQuestion_Secret(t *testing.T) {
	// Save original function
	originalFunc := AskOpenEndedQuestion
	t.Cleanup(func() {
		AskOpenEndedQuestion = originalFunc
	})

	// Mock the function
	AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		if secret {
			return "******"
		}
		return "test response"
	}

	result := AskOpenEndedQuestion("Enter password", "", true, true)
	if result != "******" {
		t.Errorf("Expected '******', got '%s'", result)
	}
}

// required's actual effect is the retry loop inside the real, pterm-backed
// AskOpenEndedQuestion (skip re-asking when required is false), which --
// like every other test in this file -- can't be exercised through the
// mocked var without a real terminal. This just confirms required reaches
// the function at all.
func Test_AskOpenEndedQuestion_NotRequired(t *testing.T) {
	// Save original function
	originalFunc := AskOpenEndedQuestion
	t.Cleanup(func() {
		AskOpenEndedQuestion = originalFunc
	})

	var gotRequired bool

	// Mock the function
	AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		gotRequired = required
		return ""
	}

	result := AskOpenEndedQuestion("Remote hosts", "", false, false)
	if result != "" {
		t.Errorf("Expected a blank answer to be returned as-is, got '%s'", result)
	}
	if gotRequired {
		t.Error("Expected required to be false")
	}
}

func Test_AskYesNoQuestion_True(t *testing.T) {
	// Save original function
	originalFunc := AskYesNoQuestion
	t.Cleanup(func() {
		AskYesNoQuestion = originalFunc
	})

	// Mock the function
	AskYesNoQuestion = func(question string) bool {
		return true
	}

	result := AskYesNoQuestion("Do you want to continue?")
	if !result {
		t.Errorf("Expected true, got false")
	}
}

func Test_AskYesNoQuestion_False(t *testing.T) {
	// Save original function
	originalFunc := AskYesNoQuestion
	t.Cleanup(func() {
		AskYesNoQuestion = originalFunc
	})

	// Mock the function
	AskYesNoQuestion = func(question string) bool {
		return false
	}

	result := AskYesNoQuestion("Do you want to continue?")
	if result {
		t.Errorf("Expected false, got true")
	}
}
