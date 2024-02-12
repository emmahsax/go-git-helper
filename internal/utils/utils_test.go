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
	// Create a mock logger
	logger := &MockLogger{}

	// Call HandleError with the mock logger
	HandleError(errors.New("test error"), false, logger)

	// Check if Fatal was called on the logger
	if !logger.FatalCalled {
		t.Errorf("Expected Fatal to be called on logger")
	}
}
