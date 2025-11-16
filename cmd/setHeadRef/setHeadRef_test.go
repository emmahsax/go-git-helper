package setHeadRef

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

	if cmd.Use != "set-head-ref [defaultBranch]" {
		t.Errorf("Expected Use 'set-head-ref [defaultBranch]', got '%s'", cmd.Use)
	}

	if cmd.Short != "Sets the HEAD ref as a symbolic ref" {
		t.Errorf("Expected Short 'Sets the HEAD ref as a symbolic ref', got '%s'", cmd.Short)
	}
}

func Test_newSetHeadRef(t *testing.T) {
	executor := &MockExecutor{Debug: false}
	shr := newSetHeadRef("main", false, executor)

	if shr == nil {
		t.Fatal("Expected non-nil SetHeadRef")
	}

	if shr.Debug != false {
		t.Errorf("Expected Debug false, got %v", shr.Debug)
	}

	if shr.DefaultBranch != "main" {
		t.Errorf("Expected DefaultBranch 'main', got '%s'", shr.DefaultBranch)
	}

	if shr.Executor == nil {
		t.Error("Expected non-nil Executor")
	}
}

func Test_newSetHeadRef_DifferentBranch(t *testing.T) {
	executor := &MockExecutor{Debug: false}
	shr := newSetHeadRef("develop", false, executor)

	if shr.DefaultBranch != "develop" {
		t.Errorf("Expected DefaultBranch 'develop', got '%s'", shr.DefaultBranch)
	}
}

func Test_execute(t *testing.T) {
	tests := []struct {
		name          string
		defaultBranch string
		debug         bool
		expectedArgs  []string
	}{
		{
			name:          "sets HEAD ref to main",
			defaultBranch: "main",
			debug:         false,
			expectedArgs:  []string{"symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main"},
		},
		{
			name:          "sets HEAD ref to master",
			defaultBranch: "master",
			debug:         false,
			expectedArgs:  []string{"symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/master"},
		},
		{
			name:          "sets HEAD ref to develop with debug",
			defaultBranch: "develop",
			debug:         true,
			expectedArgs:  []string{"symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/develop"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &MockExecutor{Debug: test.debug}
			shr := newSetHeadRef(test.defaultBranch, test.debug, executor)
			shr.execute()

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
