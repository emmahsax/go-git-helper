package emptyCommit

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

func Test_NewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Use != "empty-commit" {
		t.Errorf("Expected Use 'empty-commit', got '%s'", cmd.Use)
	}

	if cmd.Short != "Creates an empty commit" {
		t.Errorf("Expected Short 'Creates an empty commit', got '%s'", cmd.Short)
	}
}

func Test_newEmptyCommit(t *testing.T) {
	executor := &MockExecutor{Debug: false}
	ec := newEmptyCommit(false, executor)

	if ec == nil {
		t.Fatal("Expected non-nil EmptyCommit")
	}

	if ec.Debug != false {
		t.Errorf("Expected Debug false, got %v", ec.Debug)
	}

	if ec.Executor == nil {
		t.Error("Expected non-nil Executor")
	}
}

func Test_newEmptyCommit_WithDebug(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	ec := newEmptyCommit(true, executor)

	if ec.Debug != true {
		t.Errorf("Expected Debug true, got %v", ec.Debug)
	}
}

func Test_execute(t *testing.T) {
	tests := []struct {
		name         string
		debug        bool
		expectedArgs []string
	}{
		{
			name:         "creates empty commit",
			debug:        false,
			expectedArgs: []string{"commit", "--allow-empty", "-m", "Empty commit"},
		},
		{
			name:         "creates empty commit with debug",
			debug:        true,
			expectedArgs: []string{"commit", "--allow-empty", "-m", "Empty commit"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &MockExecutor{Debug: test.debug}
			ec := newEmptyCommit(test.debug, executor)
			ec.execute()

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
		})
	}
}
