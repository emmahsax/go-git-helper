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

func Test_AskOpenEndedQuestion_NoDefault(t *testing.T) {
	// Save original function
	originalFunc := AskOpenEndedQuestion
	t.Cleanup(func() {
		AskOpenEndedQuestion = originalFunc
	})

	// Mock the function
	AskOpenEndedQuestion = func(question, defaultVal string, secret bool) string {
		return "test response"
	}

	result := AskOpenEndedQuestion("Enter something", "", false)
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
	AskOpenEndedQuestion = func(question, defaultVal string, secret bool) string {
		if defaultVal != "" {
			return defaultVal
		}
		return "test response"
	}

	result := AskOpenEndedQuestion("Enter something", "default value", false)
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
	AskOpenEndedQuestion = func(question, defaultVal string, secret bool) string {
		if secret {
			return "******"
		}
		return "test response"
	}

	result := AskOpenEndedQuestion("Enter password", "", true)
	if result != "******" {
		t.Errorf("Expected '******', got '%s'", result)
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
