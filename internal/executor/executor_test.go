package executor

import (
	"reflect"
	"testing"
)

func Test_Exec(t *testing.T) {
	executor := NewExecutor(false)

	output, err := executor.Exec("actionAndOutput", "echo", "hello")
	if err != nil {
		t.Errorf("expected nil error, got '%s'", err)
	}

	expectedOutput := []byte("hello\n")
	if !reflect.DeepEqual(output, expectedOutput) {
		t.Errorf("expected '%s', got '%s'", expectedOutput, output)
	}

	_, err = executor.Exec("waitAndStdout", "echo", "hello")
	if err != nil {
		t.Errorf("expected nil error, got '%s'", err)
	}

	_, err = executor.Exec("invalid", "echo", "hello")
	expectedError := "invalid exec type"
	if err == nil || err.Error() != expectedError {
		t.Errorf("expected '%s' error, got '%s'", expectedError, err)
	}
}
