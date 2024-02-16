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
