package setup

import (
	"os"
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

type MockConfig struct {
	Contents map[string]string
	Debug    bool
}

func (mc *MockConfig) ConfigDir() string {
	return "./git-helper-test"
}

func (mc *MockConfig) ConfigDirExists() bool {
	return true
}

func (mc *MockConfig) ConfigFile() string {
	return mc.ConfigDir() + "/config-test.yml"
}

func (mc *MockConfig) ConfigFileExists() bool {
	return true
}

func (mc *MockConfig) GitHubUsername() string {
	return "test-user-github"
}

func (mc *MockConfig) GitLabUsername() string {
	return "test-user-gitlab"
}

func (mc *MockConfig) GitHubToken() string {
	return "random-github-token"
}

func (mc *MockConfig) GitLabToken() string {
	return "random-gitlab-token"
}

// func captureStdout(f func()) string {
// 	old := os.Stdout // keep backup of the real stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w

// 	f()

// 	w.Close()
// 	out, _ := io.ReadAll(r)
// 	os.Stdout = old // restore the real stdout

// 	return string(out)
// }

func Test_createOrUpdateConfig(t *testing.T) {
	tests := []struct {
		name     string
		replace  bool
		expected string
	}{
		{
			name:    "file exists, replacing",
			replace: true,
			expected: `github_username: hello_world
github_token: hello_world
gitlab_username: hello_world
gitlab_token: hello_world
`,
		},
		{
			name:     "file exists, not replacing",
			replace:  false,
			expected: "\n",
		},
	}

	for _, test := range tests {
		// Mock the AskYesNoQuestion function to always return true
		originalAskYesNoQuestion := commandline.AskYesNoQuestion
		t.Cleanup(func() {
			commandline.AskYesNoQuestion = originalAskYesNoQuestion
		})
		commandline.AskYesNoQuestion = func(question string) bool {
			return test.replace
		}

		// Mock the AskOpenEndedQuestion function to always return "hello world"
		originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
		t.Cleanup(func() {
			commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
		})
		commandline.AskOpenEndedQuestion = func(question string, secret bool) string {
			return "hello_world"
		}

		executor := &MockExecutor{Debug: true}
		configFile := &MockConfig{
			Debug: true,
			Contents: map[string]string{
				"github_username": "test_user",
				"github_token":    "test_token",
			},
		}
		s := newSetup(true, executor, configFile)

		_, err := os.Stat(configFile.ConfigFile())
		if err != nil {
			err := os.MkdirAll(configFile.ConfigDir(), 0755)
			if err != nil {
				t.Fatal(err)
			}
			tempDir, err := os.MkdirTemp(configFile.ConfigDir(), "")
			if err != nil {
				t.Fatal(err)
			}
			tempFile, err := os.Create(configFile.ConfigFile())
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(configFile.ConfigDir())
			defer os.RemoveAll(tempDir)
			defer os.Remove(tempFile.Name())
		}

		s.createOrUpdateConfig()
		res, _ := os.ReadFile(configFile.ConfigFile())

		if string(res) != test.expected {
			t.Errorf("expected output to be '%s', but got '%s'", test.expected, res)
		}
	}
}

func Test_generateConfigFileContents(t *testing.T) {
	// Mock the AskYesNoQuestion function to always return true
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	// Mock the AskOpenEndedQuestion function to always return "hello world"
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskOpenEndedQuestion = func(question string, secret bool) string {
		return "hello_world"
	}

	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{
		Debug: true,
		Contents: map[string]string{
			"github_username": "test_user",
			"github_token":    "test_token",
		},
	}
	s := newSetup(true, executor, configFile)
	contents := s.generateConfigFileContents()

	expectedContents := "github_username: hello_world\ngithub_token: hello_world\ngitlab_username: hello_world\ngitlab_token: hello_world\n"
	if contents != expectedContents {
		t.Errorf("Expected '%s', got '%s'", expectedContents, contents)
	}
}
