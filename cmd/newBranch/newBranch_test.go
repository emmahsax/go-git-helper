package newBranch

import (
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
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

func Test_determineBranch(t *testing.T) {
	tests := []struct {
		args   []string
		branch string
	}{
		{args: []string{}, branch: "hello-world"},
		{args: []string{"hello-world"}, branch: "hello-world"},
		{args: []string{"hello_world"}, branch: "hello_world"},
		{args: []string{"hello world"}, branch: "hello-world"},
		{args: []string{"hello_world!"}, branch: "hello-world"},
		{args: []string{"hello_world?"}, branch: "hello-world"},
		{args: []string{"#HelloWorld"}, branch: "hello-world"},
		{args: []string{"hello-world*"}, branch: "hello-world"},
	}

	// Mock the AskOpenEndedQuestion function to always return "hello-world"
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})

	for _, test := range tests {
		commandline.AskOpenEndedQuestion = func(question string, secret bool) string {
			return test.branch
		}

		o := determineBranch(test.args, true)

		if o != test.branch {
			t.Errorf("branch should be %s, but was %s", test.branch, o)
		}
	}
}

func Test_isValidBranch(t *testing.T) {
	tests := []struct {
		branch string
		valid  bool
	}{
		{branch: "hello-world", valid: true},
		{branch: "hello_world", valid: true},
		{branch: "hello world", valid: false},
		{branch: "hello_world!", valid: false},
		{branch: "hello_world?", valid: false},
		{branch: "#HelloWorld", valid: false},
		{branch: "hello-world*", valid: false},
	}

	for _, test := range tests {
		o := isValidBranch(test.branch)

		if o != test.valid {
			t.Errorf("branch %s should be %v, but wasn't", test.branch, test.valid)
		}
	}
}

func Test_getValidBranch(t *testing.T) {
	tests := []struct {
		branch string
		valid  bool
	}{
		{branch: "hello-world", valid: true},
	}

	// Mock the AskOpenEndedQuestion function to always return "hello-world"
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})

	for _, test := range tests {
		commandline.AskOpenEndedQuestion = func(question string, secret bool) string {
			return test.branch
		}

		o := getValidBranch()

		if o != test.branch {
			t.Errorf("branch should be %s, but was %s", "hello-world", o)
		}
	}
}

func Test_execute(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"push", "--set-upstream", "origin", "hello-world"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}
		nb := newNewBranch("hello-world", true, executor)
		nb.execute()

		if executor.Command != "git" {
			t.Errorf("unexpected command received: expected %s, but got %s", "git", executor.Command)
		}

		if len(executor.Args) != len(test.expectedArgs) {
			t.Errorf("unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
		}

		for i, v := range executor.Args {
			if v != test.expectedArgs[i] {
				t.Errorf("unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
			}
		}
	}
}
