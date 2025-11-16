package checkoutDefault

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

	if cmd.Use != "checkout-default" {
		t.Errorf("Expected Use 'checkout-default', got '%s'", cmd.Use)
	}

	if cmd.Short != "Switches to the default branch" {
		t.Errorf("Expected Short 'Switches to the default branch', got '%s'", cmd.Short)
	}
}

func Test_newCheckoutDefault(t *testing.T) {
	executor := &MockExecutor{Debug: false}
	cd := newCheckoutDefault(false, executor)

	if cd == nil {
		t.Fatal("Expected non-nil CheckoutDefault")
	}

	if cd.Debug != false {
		t.Errorf("Expected Debug false, got %v", cd.Debug)
	}

	if cd.Executor == nil {
		t.Error("Expected non-nil Executor")
	}
}

func Test_newCheckoutDefault_WithDebug(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	cd := newCheckoutDefault(true, executor)

	if cd.Debug != true {
		t.Errorf("Expected Debug true, got %v", cd.Debug)
	}
}

func Test_execute(t *testing.T) {
	tests := []struct {
		name           string
		executorOutput []byte
		debug          bool
		expectedArgs   []string
	}{
		{
			name:           "checkouts main branch",
			executorOutput: []byte("refs/remotes/origin/main"),
			debug:          false,
			expectedArgs:   []string{"checkout", "main"},
		},
		{
			name:           "checkouts master branch",
			executorOutput: []byte("refs/remotes/origin/master"),
			debug:          false,
			expectedArgs:   []string{"checkout", "master"},
		},
		{
			name:           "checkouts develop branch with debug",
			executorOutput: []byte("refs/remotes/origin/develop"),
			debug:          true,
			expectedArgs:   []string{"checkout", "develop"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &MockExecutor{
				Debug:  test.debug,
				Output: test.executorOutput,
			}

			cd := newCheckoutDefault(test.debug, executor)
			cd.execute()

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

			// Verify execute calls both DefaultBranch and Checkout
			if executor.CallCount < 2 {
				t.Errorf("Expected at least 2 executor calls (DefaultBranch + Checkout), got %d", executor.CallCount)
			}
		})
	}
}
