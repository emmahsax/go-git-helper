package forgetLocalChanges

import (
	"testing"
)

type MockExecutor struct {
	Args    []string
	Command string
	Debug   bool
	Output  []byte
}

func (me *MockExecutor) Exec(execType string, command string, args ...string) ([]byte, error) {
	me.Command = command
	me.Args = args
	return me.Output, nil
}

func Test_execute(t *testing.T) {
	tests := []struct {
		name         string
		expectedArgs []string
	}{
		{
			name:         "Git directory",
			expectedArgs: []string{"commit", "stash", "drop"},
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		flc := newForgetLocalChanges(true, executor)
		flc.execute()

		if executor.Command != "git" {
			t.Errorf("Unexpected command received: expected %s, but got %s", "git", executor.Command)
		}

		if len(executor.Args) != len(test.expectedArgs) {
			t.Errorf("Unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
		}

		for i, v := range executor.Args {
			if v != test.expectedArgs[i] {
				t.Errorf("Unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
			}
		}
	}
}
