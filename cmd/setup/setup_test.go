package setup

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/testutil"
	"github.com/emmahsax/go-git-helper/plugins"
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
	Dir      string
}

func (mc *MockConfig) ConfigDir() string {
	return mc.Dir
}

func (mc *MockConfig) ConfigDirExists() bool {
	return true
}

func (mc *MockConfig) ConfigFile() string {
	return filepath.Join(mc.ConfigDir(), "config-test.yml")
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

func (mc *MockConfig) SpecialCapitalization() map[string]string {
	return map[string]string{}
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
		commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret bool) string {
			return "hello_world"
		}

		executor := &MockExecutor{Debug: true}
		configFile := &MockConfig{
			Dir:   t.TempDir(),
			Debug: true,
			Contents: map[string]string{
				"github_username": "test_user",
				"github_token":    "test_token",
			},
		}
		s := newSetup(true, executor, configFile)

		// The config file already exists
		if err := os.WriteFile(configFile.ConfigFile(), []byte("existing"), 0600); err != nil {
			t.Fatal(err)
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
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret bool) string {
		return "hello_world"
	}

	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{
		Dir:   t.TempDir(),
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

func Test_createOrUpdatePlugins(t *testing.T) {
	for _, goos := range []string{"darwin", "linux", "windows"} {
		t.Run(goos, func(t *testing.T) {
			configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
			s := newSetup(true, &MockExecutor{Debug: true}, configFile)
			s.GOOS = goos

			s.createOrUpdatePlugins()

			entries, err := plugins.FS.ReadDir(".")
			if err != nil || len(entries) == 0 {
				t.Fatalf("expected embedded plugins, got %d (err: %v)", len(entries), err)
			}

			for _, entry := range entries {
				name := entry.Name()
				embedded, _ := plugins.FS.ReadFile(name)

				installed, err := os.ReadFile(filepath.Join(configFile.ConfigDir(), "plugins", name))
				if err != nil {
					t.Fatalf("expected plugin %s to be installed: %s", name, err)
				}

				if string(installed) != string(lfLineEndings(embedded)) {
					t.Errorf("expected plugin %s to match the embedded script", name)
				}

				if runtime.GOOS != "windows" {
					info, _ := os.Stat(filepath.Join(configFile.ConfigDir(), "plugins", name))
					if info.Mode().Perm() != 0755 {
						t.Errorf("expected plugin %s to be executable, got %v", name, info.Mode().Perm())
					}
				}

				shim, shimErr := os.ReadFile(filepath.Join(configFile.ConfigDir(), "plugins", name+".cmd"))
				if goos == "windows" {
					expected := "@git-helper " + strings.TrimPrefix(name, "git-") + " %*\r\n"
					if shimErr != nil || string(shim) != expected {
						t.Errorf("expected shim %q for %s, got %q (err: %v)", expected, name, shim, shimErr)
					}
				} else if shimErr == nil {
					t.Errorf("did not expect a .cmd shim for %s on %s", name, goos)
				}
			}
		})
	}
}

// Each plugin must call the git-helper subcommand matching its file name, and
// pass arguments through safely (quoted).
func Test_embeddedPluginsCallMatchingCommand(t *testing.T) {
	entries, _ := plugins.FS.ReadDir(".")

	for _, entry := range entries {
		name := entry.Name()
		content, _ := plugins.FS.ReadFile(name)
		expected := "git-helper " + strings.TrimPrefix(name, "git-") + ` "$@"`

		if !strings.Contains(string(content), expected) {
			t.Errorf("expected %s to contain %q, got %q", name, expected, content)
		}
	}
}

func Test_lfLineEndings(t *testing.T) {
	got := string(lfLineEndings([]byte("#!/bin/sh\r\n\r\ngit-helper new-branch \"$@\"\r\n")))
	expected := "#!/bin/sh\n\ngit-helper new-branch \"$@\"\n"

	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func Test_createOrUpdateConfig_permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not enforced on Windows")
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool { return true }
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret bool) string { return "x" }

	configFile := &MockConfig{Dir: filepath.Join(t.TempDir(), "new-dir"), Debug: true}
	configFile.Dir, _ = filepath.Abs(configFile.Dir)

	// A config file left over from an older version with loose permissions
	if err := os.MkdirAll(configFile.ConfigDir(), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile.ConfigFile(), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	newSetup(true, &MockExecutor{}, configFile).createOrUpdateConfig()

	info, err := os.Stat(configFile.ConfigFile())
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("expected config file mode 0600, got %v", info.Mode().Perm())
	}
}

func Test_instructions(t *testing.T) {
	home := testutil.IsolateHome(t)

	windowsPath := pluginPathInstructions("windows", `C:\Users\me\.git-helper\plugins`)
	if !strings.Contains(windowsPath, `C:\Users\me\.git-helper\plugins`) || !strings.Contains(windowsPath, `GetEnvironmentVariable("Path", "User")`) {
		t.Errorf("unexpected Windows PATH instructions: %s", windowsPath)
	}

	// It must extend only the user's own PATH, never copy the merged process PATH into it
	if strings.Contains(windowsPath, "$env:Path") {
		t.Errorf("Windows PATH instructions must not use the process PATH: %s", windowsPath)
	}

	unixPath := pluginPathInstructions("linux", filepath.Join(home, ".git-helper", "plugins"))
	if !strings.Contains(unixPath, `export PATH="$HOME/.git-helper/plugins:$PATH"`) {
		t.Errorf("unexpected Unix PATH instructions: %s", unixPath)
	}

	outside := pluginPathInstructions("linux", "/opt/git-helper/plugins")
	if !strings.Contains(outside, `export PATH="/opt/git-helper/plugins:$PATH"`) {
		t.Errorf("expected a path outside the home directory to be unchanged: %s", outside)
	}

	windowsCompletion := completionInstructions("windows", "dir")
	if !strings.Contains(windowsCompletion, "$PROFILE") || strings.Contains(windowsCompletion, "zsh") {
		t.Errorf("unexpected Windows completion instructions: %s", windowsCompletion)
	}

	unixCompletion := completionInstructions("darwin", filepath.Join(home, "with space", "completions"))
	for _, sh := range []string{"zsh", "bash", "fish"} {
		if !strings.Contains(unixCompletion, `"$HOME/with space/completions/completion.`+sh+`"`) {
			t.Errorf("expected a quoted, $HOME-relative path for %s: %s", sh, unixCompletion)
		}
	}
}

func Test_shellPath(t *testing.T) {
	home := testutil.IsolateHome(t)

	tests := []struct {
		path     string
		expected string
	}{
		{filepath.Join(home, ".git-helper", "plugins"), "$HOME/.git-helper/plugins"},
		{home, home},
		{filepath.Dir(home), filepath.Dir(home)},
		{"/opt/elsewhere", "/opt/elsewhere"},
	}

	for _, test := range tests {
		if got := shellPath(test.path); got != test.expected {
			t.Errorf("shellPath(%q) = %q, expected %q", test.path, got, test.expected)
		}
	}
}

func Test_createOrUpdateConfig_tightensExistingDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not enforced on Windows")
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool { return true }
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret bool) string { return "x" }

	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	if err := os.Chmod(configFile.Dir, 0755); err != nil {
		t.Fatal(err)
	}

	newSetup(true, &MockExecutor{}, configFile).createOrUpdateConfig()

	info, err := os.Stat(configFile.Dir)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0700 {
		t.Errorf("expected config directory mode 0700, got %v", info.Mode().Perm())
	}
}

func Test_createOrUpdateCompletion(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{
		Dir:   t.TempDir(),
		Debug: true,
		Contents: map[string]string{
			"github_username": "test_user",
			"github_token":    "test_token",
		},
	}
	s := newSetup(true, executor, configFile)

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
