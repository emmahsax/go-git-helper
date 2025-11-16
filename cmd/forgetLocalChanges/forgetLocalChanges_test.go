package forgetLocalChanges

import (
	"testing"
)

type MockExecutor struct {
	Args      []string
	Command   string
	Debug     bool
	Output    []byte
	CallCount int
}

func (me *MockExecutor) Exec(execType string, command string, args ...string) ([]byte, error) {
	me.Command = command
	me.Args = args
	me.CallCount++
	return me.Output, nil
}

func Test_NewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Use != "forget-local-changes" {
		t.Errorf("Expected Use 'forget-local-changes', got '%s'", cmd.Use)
	}

	if cmd.Short != "Forget all changes that aren't committed" {
		t.Errorf("Expected Short 'Forget all changes that aren't committed', got '%s'", cmd.Short)
	}
}

func Test_newForgetLocalChanges(t *testing.T) {
	executor := &MockExecutor{Debug: false}
	flc := newForgetLocalChanges(false, executor)

	if flc == nil {
		t.Fatal("Expected non-nil ForgetLocalChanges")
	}

	if flc.Debug != false {
		t.Errorf("Expected Debug false, got %v", flc.Debug)
	}

	if flc.Executor == nil {
		t.Error("Expected non-nil Executor")
	}
}

func Test_newForgetLocalChanges_WithDebug(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	flc := newForgetLocalChanges(true, executor)

	if flc.Debug != true {
		t.Errorf("Expected Debug true, got %v", flc.Debug)
	}
}

func Test_execute(t *testing.T) {
	tests := []struct {
		name         string
		debug        bool
		expectedArgs []string
	}{
		{
			name:         "forgets local changes",
			debug:        false,
			expectedArgs: []string{"stash", "drop"},
		},
		{
			name:         "forgets local changes with debug",
			debug:        true,
			expectedArgs: []string{"stash", "drop"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &MockExecutor{Debug: test.debug}
			flc := newForgetLocalChanges(test.debug, executor)
			flc.execute()

			if executor.Command != "git" {
				t.Errorf("Expected command 'git', got '%s'", executor.Command)
			}

			if len(executor.Args) != len(test.expectedArgs) {
				t.Errorf("Expected %d args %v, got %d args %v", len(test.expectedArgs), test.expectedArgs, len(executor.Args), executor.Args)
			}

			for i, v := range executor.Args {
				if v != test.expectedArgs[i] {
					t.Errorf("Arg[%d]: expected '%s', got '%s'", i, test.expectedArgs[i], v)
				}
			}

			// Verify execute calls both Stash and StashDrop
			if executor.CallCount < 2 {
				t.Errorf("Expected at least 2 executor calls (Stash + StashDrop), got %d", executor.CallCount)
			}
		})
	}
}
