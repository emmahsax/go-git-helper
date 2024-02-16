package setup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		originalAskYesNoQuestion := commandline.AskYesNoQuestion
		t.Cleanup(func() {
			commandline.AskYesNoQuestion = originalAskYesNoQuestion
		})
		commandline.AskYesNoQuestion = func(question string) bool {
			return test.replace
		}

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
		s := newSetup("owner", "repo", true, executor, configFile)

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
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

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
	s := newSetup("owner", "repo", true, executor, configFile)
	contents := s.generateConfigFileContents()

	expectedContents := "github_username: hello_world\ngithub_token: hello_world\ngitlab_username: hello_world\ngitlab_token: hello_world\n"
	if contents != expectedContents {
		t.Errorf("Expected '%s', got '%s'", expectedContents, contents)
	}
}

func Test_CceateOrUpdatePlugins(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{
		Debug: true,
		Contents: map[string]string{
			"github_username": "test_user",
			"github_token":    "test_token",
		},
	}
	s := newSetup("owner", "repo", true, executor, configFile)

	serverPlugin1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("plugin1 content"))
	}))
	defer serverPlugin1.Close()

	serverPlugin2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("plugin2 content"))
	}))
	defer serverPlugin2.Close()

	response := []map[string]string{
		{
			"name":         "plugin1",
			"download_url": serverPlugin1.URL,
		},
		{
			"name":         "plugin2",
			"download_url": serverPlugin2.URL,
		},
	}

	jsonData, _ := json.Marshal(response)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(jsonData))
	}))
	defer server.Close()
	defer os.RemoveAll(configFile.ConfigDir())

	s.createOrUpdatePlugins(server.URL)

	resPlugin1, _ := os.ReadFile(configFile.ConfigDir() + "/plugins/plugin1")

	if string(resPlugin1) != "plugin1 content" {
		t.Errorf("expected output to be '%s', but got '%s'", "plugin1 content", resPlugin1)
	}

	resPlugin2, _ := os.ReadFile(configFile.ConfigDir() + "/plugins/plugin2")

	if string(resPlugin2) != "plugin2 content" {
		t.Errorf("expected output to be '%s', but got '%s'", "plugin2 content", resPlugin2)
	}
}

func Test_createOrUpdateCompletion(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{
		Debug: true,
		Contents: map[string]string{
			"github_username": "test_user",
			"github_token":    "test_token",
		},
	}
	s := newSetup("owner", "repo", true, executor, configFile)
	defer os.RemoveAll(configFile.ConfigDir())

	s.createOrUpdateCompletion()

	shes := []string{"bash", "fish", "powershell", "zsh"}
	for _, sh := range shes {
		_, err := os.Stat(configFile.ConfigDir() + "/completions/completion." + sh)
		if err != nil {
			if os.IsNotExist(err) {
				t.Errorf("expected completion file %s to exist, but got error: %s", sh, err)
			} else {
				t.Errorf("unexpected error: %s", err)
			}
		}
	}
}
