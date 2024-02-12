package update

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

func Test_downloadGitHelper(t *testing.T) {

}

func Test_buildReleaseURL(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	u := newUpdate(true, executor)
	o := u.buildReleaseURL()
	output := "https://api.github.com/repos/emmahsax/go-git-helper/releases/latest"

	if o != output {
		t.Errorf("unexpected output received: expected %s, but got %s", output, o)
	}
}

func Test_moveGitHelper(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"mv", "./" + asset, "/usr/local/bin/git-helper"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}
		u := newUpdate(true, executor)
		u.moveGitHelper()

		if executor.Command != "sudo" {
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

func Test_setPermissions(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"chmod", "+x", "/usr/local/bin/git-helper"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}
		u := newUpdate(true, executor)
		u.setPermissions()

		if executor.Command != "sudo" {
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

func Test_outputNewVersion(t *testing.T) {
	tests := []struct {
		executorOutput string
		expectedArgs   []string
	}{
		{
			executorOutput: "Installed git-helper version",
			expectedArgs:   []string{"version"},
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: []byte(test.executorOutput),
		}

		u := newUpdate(true, executor)
		u.outputNewVersion()

		if executor.Command != "git-helper" {
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
