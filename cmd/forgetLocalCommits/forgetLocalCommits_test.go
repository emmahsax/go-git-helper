package forgetLocalCommits

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

	if cmd.Use != "forget-local-commits" {
		t.Errorf("Expected Use 'forget-local-commits', got '%s'", cmd.Use)
	}

	if cmd.Short != "Forget all commits that aren't pushed to remote" {
		t.Errorf("Expected Short 'Forget all commits that aren't pushed to remote', got '%s'", cmd.Short)
	}
}

func Test_newForgetLocalCommits(t *testing.T) {
	executor := &MockExecutor{Debug: false}
	flc := newForgetLocalCommits(false, executor)

	if flc == nil {
		t.Fatal("Expected non-nil ForgetLocalCommits")
	}

	if flc.Debug != false {
		t.Errorf("Expected Debug false, got %v", flc.Debug)
	}

	if flc.Executor == nil {
		t.Error("Expected non-nil Executor")
	}
}

func Test_newForgetLocalCommits_WithDebug(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	flc := newForgetLocalCommits(true, executor)

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
			name:         "forgets local commits",
			debug:        false,
			expectedArgs: []string{"reset", "--hard", "origin/HEAD"},
		},
		{
			name:         "forgets local commits with debug",
			debug:        true,
			expectedArgs: []string{"reset", "--hard", "origin/HEAD"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &MockExecutor{Debug: test.debug}
			flc := newForgetLocalCommits(test.debug, executor)
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

			// Verify execute calls both Pull and Reset
			if executor.CallCount < 2 {
				t.Errorf("Expected at least 2 executor calls (Pull + Reset), got %d", executor.CallCount)
			}
		})
	}
}
