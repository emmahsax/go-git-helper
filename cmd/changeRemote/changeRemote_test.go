package changeRemote

import (
	"os"
	"path/filepath"
	"reflect"
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

func Test_execute(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)

	if err != nil {
		t.Errorf("error was found: %s", err.Error())
	}

	cr := newChangeRemote("oldOwner", "newOwner", true, &MockExecutor{Debug: true})
	cr.execute()
}

func Test_processDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	executor := &MockExecutor{Debug: true}
	cr := newChangeRemote("oldOwner", "newOwner", true, executor)
	cr.processDir(tempDir, "")
	args := []string{"remote", "-v"}

	if cr.Executor.(*MockExecutor).Command != "git" {
		t.Errorf("unexpected command received: expected %s, but got %s", "git", cr.Executor.(*MockExecutor).Command)
	}

	if len(executor.Args) != len(args) {
		t.Errorf("unexpected args received: expected %v, but got %v", args, executor.Args)
	}

	for i, v := range executor.Args {
		if v != args[i] {
			t.Errorf("unexpected args received: expected %v, but got %v", args, executor.Args)
		}
	}
}

func Test_processGitRepository(t *testing.T) {
	tests := []struct {
		name           string
		executorOutput []byte
		expected       map[string]map[string]string
	}{
		{
			name: "SSH remote",
			executorOutput: []byte(`origin  git@github.com:oldOwner/repository.git (fetch)
origin  git@github.com:oldOwner/repository.git (push)
`),
			expected: map[string]map[string]string{
				"repository.git": {
					"plainRemote": "git@github.com:oldOwner/repository.git",
					"remoteName":  "origin",
					"host":        "git@github.com",
				},
			},
		},
		{
			name: "HTTP remote",
			executorOutput: []byte(`origin  https://github.com/oldOwner/repository.git (fetch)
origin  https://github.com/oldOwner/repository.git (push)
`),
			expected: map[string]map[string]string{
				"repository.git": {
					"plainRemote": "https://github.com/oldOwner/repository.git",
					"remoteName":  "origin",
					"host":        "https://github.com",
				},
			},
		},
		{
			"wrong owner",
			[]byte(`origin  https://github.com/randomOwner/repository.git (fetch)
origin  https://github.com/randomOwner/repository.git (push)
`),
			map[string]map[string]string{},
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: test.executorOutput,
		}

		cr := newChangeRemote("oldOwner", "newOwner", true, executor)
		fullRemoteInfo := cr.processGitRepository()

		if !reflect.DeepEqual(fullRemoteInfo, test.expected) {
			t.Errorf("expected %v, but got %v", test.expected, fullRemoteInfo)
		}
	}
}

func Test_processRemote(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		host         string
		repo         string
		remoteName   string
		expectedArgs []string
	}{
		{
			name:         "SSH remote",
			url:          "git@github.com:oldOwner/repo.git",
			host:         "git@github.com",
			repo:         "repo.git",
			remoteName:   "origin",
			expectedArgs: []string{"remote", "set-url", "origin", "git@github.com:newOwner/repo.git"},
		},
		{
			name:         "HTTP remote",
			url:          "https://github.com/oldOwner/repo.git",
			host:         "https://github.com",
			repo:         "repo.git",
			remoteName:   "origin",
			expectedArgs: []string{"remote", "set-url", "origin", "https://github.com/newOwner/repo.git"},
		},
	}

	executor := &MockExecutor{Debug: true}
	cr := newChangeRemote("oldOwner", "newOwner", true, executor)

	for _, test := range tests {
		cr.processRemote(test.url, test.host, test.repo, test.remoteName)

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

func Test_remoteInfo(t *testing.T) {
	tests := []struct {
		name     string
		remote   string
		expected []string
	}{
		{
			name:     "SSH remote",
			remote:   "git@github.com:oldOwner/repo.git",
			expected: []string{"git@github.com", "oldOwner", "repo.git"},
		},
		{
			name:     "HTTPS remote",
			remote:   "https://github.com/oldOwner/repo.git",
			expected: []string{"https://github.com", "oldOwner", "repo.git"},
		},
	}

	cr := newChangeRemote("oldOwner", "newOwner", false, &MockExecutor{Debug: false})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			host, owner, repo := cr.remoteInfo(test.remote)
			if host != test.expected[0] || owner != test.expected[1] || repo != test.expected[2] {
				t.Errorf("expected %v, but got %v, %v, %v", test.expected, host, owner, repo)
			}
		})
	}
}
