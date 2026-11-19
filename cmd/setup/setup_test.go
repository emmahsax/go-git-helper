package setup

import (
	"fmt"
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
	// ExistingUsernames seeds Usernames(), keyed by provider name, so tests
	// can simulate an existing config file's accounts. AddAccount,
	// UpdateAccount, and RemoveAccount all keep this in sync, so a test
	// chaining multiple calls sees consistent state.
	ExistingUsernames map[string][]string
	// Added/Updated/Removed record every call, for assertions.
	Added   []AddedAccount
	Updated []UpdatedAccount
	Removed []RemovedAccount
}

type AddedAccount struct {
	Name        string
	Username    string
	Token       string
	RemoteHosts []string
	APIHost     string
}

type UpdatedAccount struct {
	Name           string
	OldUsername    string
	NewUsername    string
	NewToken       string
	NewRemoteHosts []string
	NewAPIHost     string
}

type RemovedAccount struct {
	Name     string
	Username string
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

func (mc *MockConfig) Usernames(name string) []string {
	return mc.ExistingUsernames[name]
}

func (mc *MockConfig) AddAccount(name, username, token string, remoteHosts []string, apiHost string) error {
	mc.Added = append(mc.Added, AddedAccount{
		Name:        name,
		Username:    username,
		Token:       token,
		RemoteHosts: remoteHosts,
		APIHost:     apiHost,
	})

	if mc.ExistingUsernames == nil {
		mc.ExistingUsernames = map[string][]string{}
	}
	mc.ExistingUsernames[name] = append(mc.ExistingUsernames[name], username)

	return nil
}

func (mc *MockConfig) UpdateAccount(name, oldUsername, newUsername, newToken string, newRemoteHosts []string, newAPIHost string) error {
	mc.Updated = append(mc.Updated, UpdatedAccount{
		Name:           name,
		OldUsername:    oldUsername,
		NewUsername:    newUsername,
		NewToken:       newToken,
		NewRemoteHosts: newRemoteHosts,
		NewAPIHost:     newAPIHost,
	})

	for i, u := range mc.ExistingUsernames[name] {
		if u == oldUsername {
			mc.ExistingUsernames[name][i] = newUsername
			return nil
		}
	}

	return fmt.Errorf("no %s account found with username %q", name, oldUsername)
}

func (mc *MockConfig) RemoveAccount(name, username string) error {
	mc.Removed = append(mc.Removed, RemovedAccount{Name: name, Username: username})

	usernames := mc.ExistingUsernames[name]
	for i, u := range usernames {
		if u == username {
			mc.ExistingUsernames[name] = append(usernames[:i], usernames[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("no %s account found with username %q", name, username)
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
			expected: `providers:
  - name: github
    username: hello_world
    token: hello_world
    remote_hosts:
      - hello_world
  - name: gitlab
    username: hello_world
    token: hello_world
    remote_hosts:
      - hello_world
`,
		},
		{
			name:     "file exists, not replacing",
			replace:  false,
			expected: "\n",
		},
	}

	for _, test := range tests {
		originalAskMultiSelect := commandline.AskMultiSelect
		t.Cleanup(func() {
			commandline.AskMultiSelect = originalAskMultiSelect
		})
		commandline.AskMultiSelect = func(question string, choices []string) []string {
			if test.replace {
				return choices
			}
			return nil
		}

		originalAskYesNoQuestion := commandline.AskYesNoQuestion
		t.Cleanup(func() {
			commandline.AskYesNoQuestion = originalAskYesNoQuestion
		})
		commandline.AskYesNoQuestion = func(question string) bool {
			if strings.Contains(question, "another") || strings.Contains(question, "REAL, different") {
				return false
			}
			return test.replace
		}

		originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
		t.Cleanup(func() {
			commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
		})
		commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
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

func Test_chooseProviders(t *testing.T) {
	tests := []struct {
		name     string
		selected []string
		expected []string
	}{
		{name: "only GitHub picked", selected: []string{"GitHub"}, expected: []string{"github"}},
		{name: "only GitLab picked", selected: []string{"GitLab"}, expected: []string{"gitlab"}},
		{name: "both picked", selected: []string{"GitHub", "GitLab"}, expected: []string{"github", "gitlab"}},
		{name: "none picked", selected: nil, expected: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			originalAskMultiSelect := commandline.AskMultiSelect
			t.Cleanup(func() { commandline.AskMultiSelect = originalAskMultiSelect })
			commandline.AskMultiSelect = func(question string, choices []string) []string {
				return test.selected
			}

			s := newSetup(true, &MockExecutor{Debug: true}, &MockConfig{Dir: t.TempDir(), Debug: true})
			chosen := s.chooseProviders()

			if len(chosen) != len(test.expected) {
				t.Fatalf("expected %v, got %+v", test.expected, chosen)
			}
			for i, key := range test.expected {
				if chosen[i].Key != key {
					t.Errorf("expected %v, got %+v", test.expected, chosen)
				}
			}
		})
	}
}

func Test_generateConfigFileContents(t *testing.T) {
	originalAskMultiSelect := commandline.AskMultiSelect
	t.Cleanup(func() {
		commandline.AskMultiSelect = originalAskMultiSelect
	})
	commandline.AskMultiSelect = func(question string, choices []string) []string {
		return choices
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return !strings.Contains(question, "another") && !strings.Contains(question, "REAL, different")
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
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

	expectedContents := `providers:
  - name: github
    username: hello_world
    token: hello_world
    remote_hosts:
      - hello_world
  - name: gitlab
    username: hello_world
    token: hello_world
    remote_hosts:
      - hello_world
`
	if contents != expectedContents {
		t.Errorf("Expected '%s', got '%s'", expectedContents, contents)
	}
}

func Test_generateConfigFileContents_WithDifferentServer(t *testing.T) {
	originalAskMultiSelect := commandline.AskMultiSelect
	t.Cleanup(func() {
		commandline.AskMultiSelect = originalAskMultiSelect
	})
	commandline.AskMultiSelect = func(question string, choices []string) []string {
		return choices
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return !strings.Contains(question, "another")
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		return "hello_world"
	}

	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, executor, configFile)
	contents := s.generateConfigFileContents()

	expectedContents := `providers:
  - name: github
    username: hello_world
    token: hello_world
    remote_hosts:
      - hello_world
    api_host: hello_world
  - name: gitlab
    username: hello_world
    token: hello_world
    remote_hosts:
      - hello_world
    api_host: hello_world
`
	if contents != expectedContents {
		t.Errorf("Expected '%s', got '%s'", expectedContents, contents)
	}
}

func Test_generateConfigFileContents_BlankHostOmitted(t *testing.T) {
	originalAskMultiSelect := commandline.AskMultiSelect
	t.Cleanup(func() {
		commandline.AskMultiSelect = originalAskMultiSelect
	})
	commandline.AskMultiSelect = func(question string, choices []string) []string {
		return choices
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return !strings.Contains(question, "another")
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		if !required {
			return ""
		}
		return "hello_world"
	}

	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, executor, configFile)
	contents := s.generateConfigFileContents()

	expectedContents := `providers:
  - name: github
    username: hello_world
    token: hello_world
  - name: gitlab
    username: hello_world
    token: hello_world
`
	if contents != expectedContents {
		t.Errorf("Expected '%s', got '%s'", expectedContents, contents)
	}
}

func Test_generateConfigFileContents_MultipleGitLabAccounts(t *testing.T) {
	originalAskMultiSelect := commandline.AskMultiSelect
	t.Cleanup(func() {
		commandline.AskMultiSelect = originalAskMultiSelect
	})
	commandline.AskMultiSelect = func(question string, choices []string) []string {
		return []string{"GitLab"}
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	yesNoCall := 0
	commandline.AskYesNoQuestion = func(question string) bool {
		yesNoCall++
		switch yesNoCall {
		// Account 1's remote_hosts answer is blank, so no "is this a real,
		// different server?" follow-up gets asked for it.
		case 1: // Set up another GitLab account?
			return true
		case 2: // Is gitlab.example.com a REAL, different GitLab server...? -> no, just an alias
			return false
		default: // Set up another GitLab account? (asked again, stop here)
			return false
		}
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	openEndedCall := 0
	responses := []string{
		"public-user", "public-token", "", // first (public) account
		"work-user", "work-token", "gitlab.example.com, git.internal.example.com", // second account: remote_hosts answered, but not a different server (just aliases)
	}
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { openEndedCall++ }()
		return responses[openEndedCall]
	}

	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, executor, configFile)
	contents := s.generateConfigFileContents()

	expectedContents := `providers:
  - name: gitlab
    username: public-user
    token: public-token
  - name: gitlab
    username: work-user
    token: work-token
    remote_hosts:
      - gitlab.example.com
      - git.internal.example.com
`
	if contents != expectedContents {
		t.Errorf("Expected:\n%s\nGot:\n%s", expectedContents, contents)
	}
}

func Test_generateConfigFileContents_MultipleGitHubAccounts(t *testing.T) {
	originalAskMultiSelect := commandline.AskMultiSelect
	t.Cleanup(func() {
		commandline.AskMultiSelect = originalAskMultiSelect
	})
	commandline.AskMultiSelect = func(question string, choices []string) []string {
		return []string{"GitHub"}
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	yesNoCall := 0
	commandline.AskYesNoQuestion = func(question string) bool {
		yesNoCall++
		switch yesNoCall {
		// Account 1's remote_hosts answer is blank, so no "is this a real,
		// different server?" follow-up gets asked for it.
		case 1: // Set up another GitHub account?
			return true
		case 2: // Is github.example.com a REAL, different GitHub server...? -> no, just an alias
			return false
		default: // Set up another GitHub account? (asked again, stop here)
			return false
		}
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	openEndedCall := 0
	responses := []string{
		"public-user", "public-token", "", // first (public) account
		"work-user", "work-token", "github.example.com", // second account: remote_hosts answered, but not a different server (just an alias)
	}
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { openEndedCall++ }()
		return responses[openEndedCall]
	}

	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, executor, configFile)
	contents := s.generateConfigFileContents()

	expectedContents := `providers:
  - name: github
    username: public-user
    token: public-token
  - name: github
    username: work-user
    token: work-token
    remote_hosts:
      - github.example.com
`
	if contents != expectedContents {
		t.Errorf("Expected:\n%s\nGot:\n%s", expectedContents, contents)
	}
}

// A genuine Enterprise account: the user says "yes, this is a real,
// different server" to the follow-up question, and gives a distinct
// api_host (different from the remote_hosts entries), proving api_host
// isn't just copied from remote_hosts[0] but actually asked for.
func Test_generateConfigFileContents_GenuineEnterpriseAccount(t *testing.T) {
	originalAskMultiSelect := commandline.AskMultiSelect
	t.Cleanup(func() {
		commandline.AskMultiSelect = originalAskMultiSelect
	})
	commandline.AskMultiSelect = func(question string, choices []string) []string {
		return []string{"GitHub"}
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	yesNoCall := 0
	commandline.AskYesNoQuestion = func(question string) bool {
		yesNoCall++
		switch yesNoCall {
		case 1: // Is github.example.com a REAL, different GitHub server...? -> yes
			return true
		default: // Set up another GitHub account?
			return false
		}
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	openEndedCall := 0
	responses := []string{
		"work-user", "work-token",
		"github.example.com, git.internal.example.com", // remote_hosts (more than one -> API host asked separately)
		"api.github.example.com",                       // the distinct API host
	}
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { openEndedCall++ }()
		return responses[openEndedCall]
	}

	executor := &MockExecutor{Debug: true}
	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, executor, configFile)
	contents := s.generateConfigFileContents()

	expectedContents := `providers:
  - name: github
    username: work-user
    token: work-token
    remote_hosts:
      - github.example.com
      - git.internal.example.com
    api_host: api.github.example.com
`
	if contents != expectedContents {
		t.Errorf("Expected:\n%s\nGot:\n%s", expectedContents, contents)
	}
}

// gitlabProviderCredential finds the "gitlab" entry in gitProviderCredentials,
// for tests that only want to exercise one provider's prompts directly.
func gitlabProviderCredential(t *testing.T) gitProviderCredential {
	t.Helper()
	for _, p := range gitProviderCredentials {
		if p.Key == "gitlab" {
			return p
		}
	}
	t.Fatal("no gitlab entry in gitProviderCredentials")
	return gitProviderCredential{}
}

func Test_askRemoteHosts_RejectsPublicDomain(t *testing.T) {
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	// First attempt types the public domain itself -- rejected and retried.
	responses := []string{"gitlab.com", "gitlab.example.com"}
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	s := newSetup(true, &MockExecutor{Debug: true}, &MockConfig{Dir: t.TempDir(), Debug: true})
	hosts := s.askRemoteHosts(gitlabProviderCredential(t))

	if len(hosts) != 1 || hosts[0] != "gitlab.example.com" {
		t.Errorf("expected ['gitlab.example.com'] after the retry, got %v", hosts)
	}
}

func Test_askRemoteHosts_RejectsPublicDomainAmongMultiple(t *testing.T) {
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	// First attempt lists the public domain alongside a real alias -- the
	// whole answer is rejected (not just the offending entry) and retried.
	responses := []string{"gitlab.example.com, gitlab.com", "gitlab.example.com"}
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	s := newSetup(true, &MockExecutor{Debug: true}, &MockConfig{Dir: t.TempDir(), Debug: true})
	hosts := s.askRemoteHosts(gitlabProviderCredential(t))

	if len(hosts) != 1 || hosts[0] != "gitlab.example.com" {
		t.Errorf("expected ['gitlab.example.com'] after the retry, got %v", hosts)
	}
}

// Regression: the public-domain guard must be case-insensitive.
func Test_askRemoteHosts_RejectsPublicDomainCaseInsensitively(t *testing.T) {
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	responses := []string{"GitLab.Com", "gitlab.example.com"}
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	s := newSetup(true, &MockExecutor{Debug: true}, &MockConfig{Dir: t.TempDir(), Debug: true})
	hosts := s.askRemoteHosts(gitlabProviderCredential(t))

	if len(hosts) != 1 || hosts[0] != "gitlab.example.com" {
		t.Errorf("expected a mixed-case public domain to be rejected and retried, got %v", hosts)
	}
}

func Test_providerAccountBlocks_DuplicateUsernameRejected(t *testing.T) {
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })
	yesNoCall := 0
	commandline.AskYesNoQuestion = func(question string) bool {
		yesNoCall++
		// call 1: "another GitLab account?" (after account 1) -> yes
		// call 2: "another GitLab account?" (after account 2) -> stop
		return yesNoCall <= 1
	}

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	responses := []string{
		"dup-user", "token1", "", // account 1
		"dup-user", "work-user", "token2", "", // account 2: first attempt collides with account 1, retried
	}
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	s := newSetup(true, &MockExecutor{Debug: true}, &MockConfig{Dir: t.TempDir(), Debug: true})
	blocks := s.providerAccountBlocks(gitlabProviderCredential(t))

	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %v", len(blocks), blocks)
	}
	if !strings.Contains(blocks[0], "username: dup-user") {
		t.Errorf("expected the first account to keep username 'dup-user', got:\n%s", blocks[0])
	}
	if !strings.Contains(blocks[1], "username: work-user") {
		t.Errorf("expected the second account to end up with username 'work-user' after the collision retry, got:\n%s", blocks[1])
	}
	if strings.Contains(blocks[1], "username: dup-user") {
		t.Errorf("expected the rejected 'dup-user' attempt not to appear in the second block, got:\n%s", blocks[1])
	}
}

func Test_addAccount(t *testing.T) {
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })
	commandline.AskYesNoQuestion = func(question string) bool { return false }

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	responses := []string{"new-user", "new-token", ""}
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return "GitHub"
	}

	configFile := &MockConfig{
		Dir:               t.TempDir(),
		Debug:             true,
		ExistingUsernames: map[string][]string{"github": {"existing-user"}},
	}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.addAccount()

	if len(configFile.Added) != 1 {
		t.Fatalf("expected 1 AddAccount call, got %d", len(configFile.Added))
	}

	added := configFile.Added[0]
	if added.Name != "github" || added.Username != "new-user" || added.Token != "new-token" {
		t.Errorf("unexpected AddAccount call: %+v", added)
	}
}

func Test_addAccount_RejectsUsernameAlreadyInConfig(t *testing.T) {
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })
	commandline.AskYesNoQuestion = func(question string) bool { return false }

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	// First attempt collides with the account already in config; retried.
	responses := []string{"existing-user", "new-user", "new-token", ""}
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return "GitHub"
	}

	configFile := &MockConfig{
		Dir:               t.TempDir(),
		Debug:             true,
		ExistingUsernames: map[string][]string{"github": {"existing-user"}},
	}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.addAccount()

	if len(configFile.Added) != 1 || configFile.Added[0].Username != "new-user" {
		t.Fatalf("expected the retried username 'new-user' to be added, got %+v", configFile.Added)
	}
}

func Test_setupConfig_AddAccount(t *testing.T) {
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	choiceCall := 0
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		defer func() { choiceCall++ }()
		if choiceCall == 0 {
			return "Add a new account"
		}
		return "GitLab"
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })
	commandline.AskYesNoQuestion = func(question string) bool { return false }

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	responses := []string{"gl-user", "gl-token", ""}
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.setupConfig()

	if len(configFile.Added) != 1 || configFile.Added[0].Name != "gitlab" {
		t.Fatalf("expected a gitlab account to be added, got %+v", configFile.Added)
	}
}

func Test_setupConfig_ReplaceEverything(t *testing.T) {
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return "Replace everything"
	}

	originalAskMultiSelect := commandline.AskMultiSelect
	t.Cleanup(func() { commandline.AskMultiSelect = originalAskMultiSelect })
	commandline.AskMultiSelect = func(question string, choices []string) []string {
		return nil
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })
	commandline.AskYesNoQuestion = func(question string) bool { return false }

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		return "unused"
	}

	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.setupConfig()

	if len(configFile.Added) != 0 {
		t.Errorf("expected no AddAccount calls when replacing everything, got %+v", configFile.Added)
	}

	res, err := os.ReadFile(configFile.ConfigFile())
	if err != nil {
		t.Fatalf("expected the config file to be (re)written, got error: %v", err)
	}
	if string(res) != "\n" {
		t.Errorf("expected an empty providers list (every credential question declined), got %q", res)
	}
}

func Test_setupConfig_LeaveAsIs(t *testing.T) {
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return "Leave it as is"
	}

	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.setupConfig()

	if len(configFile.Added) != 0 {
		t.Errorf("expected no AddAccount calls, got %+v", configFile.Added)
	}
	if _, err := os.Stat(configFile.ConfigFile()); err == nil {
		t.Error("expected no config file to be written")
	}
}

func Test_setupConfig_EditAccount_Update(t *testing.T) {
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	choiceCall := 0
	choices := []string{
		"Edit or remove an account", // setupConfig's menu
		"GitHub",                    // chooseProvider
		"old-user",                  // which account
		"Update it",                 // what to do with it
	}
	commandline.AskMultipleChoice = func(question string, options []string) string {
		defer func() { choiceCall++ }()
		return choices[choiceCall]
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })
	commandline.AskYesNoQuestion = func(question string) bool { return false }

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	responses := []string{"old-user", "new-token", ""} // username kept the same, token rotated, no remote_hosts
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	configFile := &MockConfig{
		Dir:               t.TempDir(),
		Debug:             true,
		ExistingUsernames: map[string][]string{"github": {"old-user"}},
	}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.setupConfig()

	if len(configFile.Updated) != 1 {
		t.Fatalf("expected 1 UpdateAccount call, got %d", len(configFile.Updated))
	}

	updated := configFile.Updated[0]
	if updated.Name != "github" || updated.OldUsername != "old-user" || updated.NewUsername != "old-user" || updated.NewToken != "new-token" {
		t.Errorf("unexpected UpdateAccount call: %+v", updated)
	}
}

func Test_setupConfig_EditAccount_UpdateAllowsKeepingSameUsername(t *testing.T) {
	// Regression: updating an account while keeping its own username must
	// not trip the duplicate-username guard against itself.
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	choiceCall := 0
	choices := []string{"Edit or remove an account", "GitHub", "alice", "Update it"}
	commandline.AskMultipleChoice = func(question string, options []string) string {
		defer func() { choiceCall++ }()
		return choices[choiceCall]
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })
	commandline.AskYesNoQuestion = func(question string) bool { return false }

	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() { commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion })
	responses := []string{"alice", "alice-new-token", ""}
	call := 0
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string {
		defer func() { call++ }()
		return responses[call]
	}

	configFile := &MockConfig{
		Dir:               t.TempDir(),
		Debug:             true,
		ExistingUsernames: map[string][]string{"github": {"alice", "bob"}},
	}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.setupConfig()

	if len(configFile.Updated) != 1 || configFile.Updated[0].NewUsername != "alice" {
		t.Fatalf("expected alice to be updated without a collision, got %+v", configFile.Updated)
	}
}

func Test_setupConfig_EditAccount_Remove(t *testing.T) {
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	choiceCall := 0
	choices := []string{"Edit or remove an account", "GitLab", "stale-user", "Remove it"}
	commandline.AskMultipleChoice = func(question string, options []string) string {
		defer func() { choiceCall++ }()
		return choices[choiceCall]
	}

	configFile := &MockConfig{
		Dir:               t.TempDir(),
		Debug:             true,
		ExistingUsernames: map[string][]string{"gitlab": {"stale-user"}},
	}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.setupConfig()

	if len(configFile.Removed) != 1 || configFile.Removed[0].Name != "gitlab" || configFile.Removed[0].Username != "stale-user" {
		t.Fatalf("expected stale-user to be removed, got %+v", configFile.Removed)
	}
}

func Test_setupConfig_EditAccount_Cancel(t *testing.T) {
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	choiceCall := 0
	choices := []string{"Edit or remove an account", "GitHub", "some-user", "Cancel"}
	commandline.AskMultipleChoice = func(question string, options []string) string {
		defer func() { choiceCall++ }()
		return choices[choiceCall]
	}

	configFile := &MockConfig{
		Dir:               t.TempDir(),
		Debug:             true,
		ExistingUsernames: map[string][]string{"github": {"some-user"}},
	}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.setupConfig()

	if len(configFile.Updated) != 0 || len(configFile.Removed) != 0 {
		t.Errorf("expected cancel to make no changes, got updated=%+v removed=%+v", configFile.Updated, configFile.Removed)
	}
}

func Test_setupConfig_EditAccount_NoneConfigured(t *testing.T) {
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() { commandline.AskMultipleChoice = originalAskMultipleChoice })
	choiceCall := 0
	choices := []string{"Edit or remove an account", "GitHub"}
	commandline.AskMultipleChoice = func(question string, options []string) string {
		defer func() { choiceCall++ }()
		return choices[choiceCall]
	}

	configFile := &MockConfig{Dir: t.TempDir(), Debug: true}
	s := newSetup(true, &MockExecutor{Debug: true}, configFile)
	s.setupConfig()

	if len(configFile.Updated) != 0 || len(configFile.Removed) != 0 {
		t.Errorf("expected no changes when no accounts are configured, got updated=%+v removed=%+v", configFile.Updated, configFile.Removed)
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

	originalAskMultiSelect := commandline.AskMultiSelect
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskMultiSelect = originalAskMultiSelect
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskMultiSelect = func(question string, choices []string) []string { return choices }
	commandline.AskYesNoQuestion = func(question string) bool { return !strings.Contains(question, "another") }
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string { return "x" }

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

	originalAskMultiSelect := commandline.AskMultiSelect
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	originalAskOpenEndedQuestion := commandline.AskOpenEndedQuestion
	t.Cleanup(func() {
		commandline.AskMultiSelect = originalAskMultiSelect
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
		commandline.AskOpenEndedQuestion = originalAskOpenEndedQuestion
	})
	commandline.AskMultiSelect = func(question string, choices []string) []string { return choices }
	commandline.AskYesNoQuestion = func(question string) bool { return !strings.Contains(question, "another") }
	commandline.AskOpenEndedQuestion = func(question, defaultVal string, secret, required bool) string { return "x" }

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
