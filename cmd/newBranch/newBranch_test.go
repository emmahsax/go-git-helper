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
		{args: []string{}, branch: "hello-something-or-other"},
		{args: []string{"hello-world"}, branch: ""},
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})

	for _, test := range tests {
		commandline.AskOpenEndedQuestion = func(question string, secret bool) string {
			return test.branch
		}

		o := determineBranch(test.args)

		if o == test.branch {
			continue
		}

		if len(test.args) > 0 && o == test.args[0] {
			continue
		}

		t.Errorf("branch should be %s, but was %s", test.branch, o)
	}
}

func Test_askForBranch(t *testing.T) {
	tests := []struct {
		branch string
	}{
		{branch: "hello-world"},
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})

	for _, test := range tests {
		commandline.AskOpenEndedQuestion = func(question string, secret bool) string {
			return test.branch
		}

		o := askForBranch()

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
