package cleanBranches

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
		name           string
		executorOutput []byte
		expectedArgs   []string
	}{
		{
			name:           "Git directory",
			executorOutput: []byte("refs/remotes/origin/main"),
			expectedArgs:   []string{"branch", "-vv"},
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: test.executorOutput,
		}

		cb := &CleanBranches{
			Debug:    true,
			Executor: executor,
		}

		cb.execute()

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
