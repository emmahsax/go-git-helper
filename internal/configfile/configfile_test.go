package configfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/testutil"
)

func Test_NewConfigFile(t *testing.T) {
	cf := NewConfigFile(false)

	if cf == nil {
		t.Fatal("Expected NewConfigFile to return non-nil ConfigFile")
	}

	if cf.Debug != false {
		t.Errorf("Expected Debug to be false, got %v", cf.Debug)
	}
}

func Test_NewConfigFile_WithDebug(t *testing.T) {
	cf := NewConfigFile(true)

	if cf.Debug != true {
		t.Errorf("Expected Debug to be true, got %v", cf.Debug)
	}
}

func Test_ConfigDir(t *testing.T) {
	testutil.IsolateHome(t)
	cf := NewConfigFile(false)
	dir := cf.ConfigDir()

	if dir == "" {
		t.Errorf("Expected a directory, got an empty string")
	}

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".git-helper")
	if dir != expectedDir {
		t.Errorf("Expected directory '%s', got '%s'", expectedDir, dir)
	}
}

func Test_ConfigDirExists(t *testing.T) {
	testutil.IsolateHome(t)
	cf := NewConfigFile(false)

	if cf.ConfigDirExists() {
		t.Errorf("Expected ConfigDirExists to return false before the directory is created")
	}

	if err := os.MkdirAll(cf.ConfigDir(), 0755); err != nil {
		t.Fatal(err)
	}

	if !cf.ConfigDirExists() {
		t.Errorf("Expected ConfigDirExists to return true, got false")
	}
}

func Test_ConfigDirExists_FileInsteadOfDirectory(t *testing.T) {
	testutil.IsolateHome(t)
	cf := NewConfigFile(false)

	if err := os.WriteFile(cf.ConfigDir(), []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	if cf.ConfigDirExists() {
		t.Errorf("Expected ConfigDirExists to return false when the path is a file")
	}
}

func Test_ConfigFile(t *testing.T) {
	testutil.IsolateHome(t)
	cf := NewConfigFile(false)
	file := cf.ConfigFile()

	if file == "" {
		t.Errorf("Expected a file, got an empty string")
	}

	homeDir, _ := os.UserHomeDir()
	expectedFile := filepath.Join(homeDir, ".git-helper", "config.yml")
	if file != expectedFile {
		t.Errorf("Expected file '%s', got '%s'", expectedFile, file)
	}
}

func Test_ConfigFileExists(t *testing.T) {
	home := testutil.IsolateHome(t)
	cf := NewConfigFile(false)

	if cf.ConfigFileExists() {
		t.Errorf("Expected ConfigFileExists to return false before the file is created")
	}

	testutil.WriteConfig(t, home, "providers:\n  - name: github\n    username: testuser\n")

	if !cf.ConfigFileExists() {
		t.Errorf("Expected ConfigFileExists to return true, got false")
	}
}

func Test_ConfigFileExists_False(t *testing.T) {
	// Create a temporary config file that we can delete
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.yml")

	_, err := os.Stat(nonExistentFile)
	if err == nil {
		t.Error("Expected file to not exist")
	}
}

// Helper function to create a test config file
func createTestConfigFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yml")

	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Point the home directory at tempDir so ConfigDir() returns tempDir/.git-helper
	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir)

	// Create .git-helper directory in temp
	gitHelperDir := filepath.Join(tempDir, ".git-helper")
	os.MkdirAll(gitHelperDir, 0755)

	// Move config file to the right location
	finalConfigFile := filepath.Join(gitHelperDir, "config.yml")
	os.Rename(configFile, finalConfigFile)

	return tempDir
}

func Test_GitHubAndGitLab_BothConfigured(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    username: githubuser
    token: ghp_token123
  - name: gitlab
    username: gitlabuser
    token: glpat-token123
    remote_hosts:
      - gitlab.example.com
`)

	cf := NewConfigFile(false)

	github := ResolveAccount(cf.GitHubAccounts("github.com"), "")
	if github.Username != "githubuser" {
		t.Errorf("Expected GitHub username 'githubuser', got '%s'", github.Username)
	}
	if github.Token != "ghp_token123" {
		t.Errorf("Expected GitHub token 'ghp_token123', got '%s'", github.Token)
	}

	gitlab := ResolveAccount(cf.GitLabAccounts("gitlab.example.com"), "")
	if gitlab.Username != "gitlabuser" {
		t.Errorf("Expected GitLab username 'gitlabuser', got '%s'", gitlab.Username)
	}
	if gitlab.Token != "glpat-token123" {
		t.Errorf("Expected GitLab token 'glpat-token123', got '%s'", gitlab.Token)
	}
}

func Test_GitHubHosts(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    token: public_token
  - name: github
    token: enterprise_token
    remote_hosts:
      - github.example.com
      - git.internal.example.com
`)

	cf := NewConfigFile(false)
	hosts := cf.GitHubHosts()

	expected := []string{"github.example.com", "git.internal.example.com"}
	if len(hosts) != len(expected) {
		t.Fatalf("Expected hosts %v, got %v", expected, hosts)
	}
	for i := range expected {
		if hosts[i] != expected[i] {
			t.Errorf("Expected hosts %v, got %v", expected, hosts)
		}
	}
}

func Test_GitHubAccounts_PublicAndEnterprise(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    username: public_user
    token: public_token
  - name: github
    username: work_user
    token: work_token
    remote_hosts:
      - github.example.com
    api_host: github.example.com
`)

	cf := NewConfigFile(false)

	public := ResolveAccount(cf.GitHubAccounts("github.com"), "")
	if public.Username != "public_user" || public.Token != "public_token" {
		t.Errorf("Expected the public account for github.com, got (%q, %q)", public.Username, public.Token)
	}
	if public.APIHost != "" {
		t.Errorf("Expected an empty APIHost for the public account, got %q", public.APIHost)
	}

	enterprise := ResolveAccount(cf.GitHubAccounts("github.example.com"), "")
	if enterprise.Username != "work_user" || enterprise.Token != "work_token" {
		t.Errorf("Expected the Enterprise account for github.example.com, got (%q, %q)", enterprise.Username, enterprise.Token)
	}
	if enterprise.APIHost != "github.example.com" {
		t.Errorf("Expected APIHost 'github.example.com', got %q", enterprise.APIHost)
	}
}

func Test_GitHubAccounts_SSHAliasKeepsPublicAPIHost(t *testing.T) {
	// A host used only to tell two accounts apart in a remote (an SSH
	// config alias) must not imply a different API server: api_host is
	// absent here, even though Hosts is set.
	createTestConfigFile(t, `providers:
  - name: github
    username: personal_user
    token: personal_token
  - name: github
    username: work_user
    token: work_token
    remote_hosts:
      - github.com-work
`)

	cf := NewConfigFile(false)

	aliased := ResolveAccount(cf.GitHubAccounts("github.com-work"), "")
	if aliased.Username != "work_user" {
		t.Errorf("Expected the work account for the aliased remote, got %q", aliased.Username)
	}
	if aliased.APIHost != "" {
		t.Errorf("Expected an empty APIHost for an SSH-alias-only account, got %q", aliased.APIHost)
	}
}

func Test_GitHubAccounts_TwoAccountsSameHost(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    username: personal_user
    token: personal_token
  - name: github
    username: work_user
    token: work_token
`)

	cf := NewConfigFile(false)
	accounts := cf.GitHubAccounts("github.com")

	if len(accounts) != 2 {
		t.Fatalf("Expected 2 tied accounts, got %d", len(accounts))
	}

	if ResolveAccount(accounts, "work_user").Username != "work_user" {
		t.Errorf("Expected ResolveAccount to pick the account matching the given username")
	}
	if ResolveAccount(accounts, "").Username != "personal_user" {
		t.Errorf("Expected ResolveAccount to default to the first account when no username is given")
	}
}

func Test_GitHubAccounts_NoneConfigured(t *testing.T) {
	testutil.IsolateHome(t)

	cf := NewConfigFile(false)
	if accounts := cf.GitHubAccounts("github.com"); len(accounts) != 0 {
		t.Errorf("Expected no accounts, got %v", accounts)
	}
}

func Test_ResolveAccount_EmptyCandidates(t *testing.T) {
	if got := ResolveAccount(nil, "anything"); got != (Account{}) {
		t.Errorf("Expected the zero Account, got %+v", got)
	}
}

func Test_GitLabHosts(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: gitlab
    token: public_token
  - name: gitlab
    token: self_hosted_token
    remote_hosts:
      - gitlab.example.com
      - git.internal.example.com
`)

	cf := NewConfigFile(false)
	hosts := cf.GitLabHosts()

	expected := []string{"gitlab.example.com", "git.internal.example.com"}
	if len(hosts) != len(expected) {
		t.Fatalf("Expected hosts %v, got %v", expected, hosts)
	}
	for i := range expected {
		if hosts[i] != expected[i] {
			t.Errorf("Expected hosts %v, got %v", expected, hosts)
		}
	}
}

func Test_GitLabHosts_NoneConfigured(t *testing.T) {
	createTestConfigFile(t, "providers:\n  - name: gitlab\n    token: glpat-token123\n")

	cf := NewConfigFile(false)
	if hosts := cf.GitLabHosts(); len(hosts) != 0 {
		t.Errorf("Expected no hosts, got %v", hosts)
	}
}

func Test_GitLabAccounts_PublicAccount(t *testing.T) {
	createTestConfigFile(t, "providers:\n  - name: gitlab\n    username: public_user\n    token: public_token\n")

	cf := NewConfigFile(false)
	account := ResolveAccount(cf.GitLabAccounts("gitlab.com"), "")
	if account.Username != "public_user" || account.Token != "public_token" {
		t.Errorf("Expected ('public_user', 'public_token'), got (%q, %q)", account.Username, account.Token)
	}
}

func Test_GitLabAccounts_PublicAndSelfHosted(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: gitlab
    username: public_user
    token: public_token
  - name: gitlab
    username: work_user
    token: work_token
    remote_hosts:
      - gitlab.example.com
      - git.internal.example.com
    api_host: gitlab.example.com
`)

	cf := NewConfigFile(false)

	if account := ResolveAccount(cf.GitLabAccounts("gitlab.com"), ""); account.Username != "public_user" || account.Token != "public_token" {
		t.Errorf("Expected the public account for gitlab.com, got (%q, %q)", account.Username, account.Token)
	}

	if account := ResolveAccount(cf.GitLabAccounts("gitlab.example.com"), ""); account.Username != "work_user" || account.APIHost != "gitlab.example.com" {
		t.Errorf("Expected the work account with APIHost 'gitlab.example.com', got (%q, %q)", account.Username, account.APIHost)
	}

	if account := ResolveAccount(cf.GitLabAccounts("git.internal.example.com"), ""); account.Username != "work_user" {
		t.Errorf("Expected the work account for its second alias host, got %q", account.Username)
	}
}

func Test_GitLabAccounts_NoneConfigured(t *testing.T) {
	testutil.IsolateHome(t)

	cf := NewConfigFile(false)
	if accounts := cf.GitLabAccounts("gitlab.com"); len(accounts) != 0 {
		t.Errorf("Expected no accounts, got %v", accounts)
	}
}

func Test_GitLabHosts_ProviderNotConfigured(t *testing.T) {
	createTestConfigFile(t, "providers:\n  - name: github\n    token: ghp_token123\n")

	cf := NewConfigFile(false)
	if hosts := cf.GitLabHosts(); len(hosts) != 0 {
		t.Errorf("Expected no hosts, got %v", hosts)
	}
}

func Test_SpecialCapitalization(t *testing.T) {
	content := `special_capitalization:
  api: API
  aws: AWS
  github: GitHub
`
	createTestConfigFile(t, content)

	cf := NewConfigFile(false)
	caps := cf.SpecialCapitalization()

	if len(caps) != 3 {
		t.Errorf("Expected 3 capitalizations, got %d", len(caps))
	}

	if caps["api"] != "API" {
		t.Errorf("Expected 'API', got '%s'", caps["api"])
	}

	if caps["aws"] != "AWS" {
		t.Errorf("Expected 'AWS', got '%s'", caps["aws"])
	}

	if caps["github"] != "GitHub" {
		t.Errorf("Expected 'GitHub', got '%s'", caps["github"])
	}
}

func Test_SpecialCapitalization_Empty(t *testing.T) {
	createTestConfigFile(t, "providers:\n  - name: github\n    username: testuser\n")

	cf := NewConfigFile(false)
	caps := cf.SpecialCapitalization()

	if len(caps) != 0 {
		t.Errorf("Expected 0 capitalizations, got %d", len(caps))
	}
}

func Test_SpecialCapitalization_FileNotFound(t *testing.T) {
	// Use a temp directory without creating config file
	testutil.IsolateHome(t)

	cf := NewConfigFile(true) // Use debug mode to avoid exit
	caps := cf.SpecialCapitalization()

	if len(caps) != 0 {
		t.Errorf("Expected empty map when file not found, got %d items", len(caps))
	}
}

func Test_GitHubAccounts_InvalidYAML(t *testing.T) {
	createTestConfigFile(t, "providers: [this is not: valid: yaml\n")

	cf := NewConfigFile(false)
	if accounts := cf.GitHubAccounts("github.com"); len(accounts) != 0 {
		t.Errorf("Expected no accounts for invalid YAML, got %v", accounts)
	}
}

func Test_GitHubAccounts_FileNotFound(t *testing.T) {
	testutil.IsolateHome(t)

	cf := NewConfigFile(false)
	if accounts := cf.GitHubAccounts("github.com"); len(accounts) != 0 {
		t.Errorf("Expected no accounts when the config file doesn't exist, got %v", accounts)
	}
}

func Test_Usernames(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    username: personal_user
  - name: github
    username: work_user
    remote_hosts:
      - github.example.com
  - name: gitlab
    username: gitlab_user
`)

	cf := NewConfigFile(false)

	if got := cf.Usernames("github"); len(got) != 2 || got[0] != "personal_user" || got[1] != "work_user" {
		t.Errorf("Expected ['personal_user', 'work_user'], got %v", got)
	}

	if got := cf.Usernames("gitlab"); len(got) != 1 || got[0] != "gitlab_user" {
		t.Errorf("Expected ['gitlab_user'], got %v", got)
	}
}

func Test_Usernames_NoneConfigured(t *testing.T) {
	testutil.IsolateHome(t)

	cf := NewConfigFile(false)
	if got := cf.Usernames("github"); len(got) != 0 {
		t.Errorf("Expected no usernames, got %v", got)
	}
}

func Test_AddAccount_PreservesExistingEntries(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    username: existing_user
    token: existing_token
special_capitalization:
  api: API
`)

	cf := NewConfigFile(false)
	if err := cf.AddAccount("github", "new_user", "new_token", []string{"github.com-work"}, ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The pre-existing account must survive untouched.
	existing := ResolveAccount(cf.GitHubAccounts("github.com"), "")
	if existing.Username != "existing_user" || existing.Token != "existing_token" {
		t.Errorf("expected the existing account to be preserved, got %+v", existing)
	}

	// The new account must be reachable by its remote_hosts entry.
	added := ResolveAccount(cf.GitHubAccounts("github.com-work"), "")
	if added.Username != "new_user" || added.Token != "new_token" {
		t.Errorf("expected the new account to be added, got %+v", added)
	}

	// special_capitalization (an unrelated top-level key) must survive untouched.
	if caps := cf.SpecialCapitalization(); caps["api"] != "API" {
		t.Errorf("expected special_capitalization to be preserved, got %v", caps)
	}

	// The written file must match the rest of config.yml's 2-space indentation
	// (yaml.Marshal's own default is 4 spaces).
	data, err := os.ReadFile(cf.ConfigFile())
	if err != nil {
		t.Fatalf("expected to read back the config file, got error: %v", err)
	}
	if !strings.Contains(string(data), "\n  - name: github\n") {
		t.Errorf("expected 2-space-indented list items, got:\n%s", data)
	}
	if !strings.Contains(string(data), "\n    remote_hosts:\n      - github.com-work\n") {
		t.Errorf("expected 2-space-indented nested fields, got:\n%s", data)
	}
}

func Test_AddAccount_CreatesFileWhenMissing(t *testing.T) {
	testutil.IsolateHome(t)

	cf := NewConfigFile(false)
	if err := cf.AddAccount("gitlab", "first_user", "first_token", nil, ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	account := ResolveAccount(cf.GitLabAccounts("gitlab.com"), "")
	if account.Username != "first_user" || account.Token != "first_token" {
		t.Errorf("expected the new account to be added, got %+v", account)
	}
}

func Test_UpdateAccount_ReplacesFieldsInPlace(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    username: alice
    token: old_token
  - name: github
    username: bob
    token: bob_token
special_capitalization:
  api: API
`)

	cf := NewConfigFile(false)
	if err := cf.UpdateAccount("github", "alice", "alice2", "new_token", []string{"github.com-alice2"}, ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated := ResolveAccount(cf.GitHubAccounts("github.com-alice2"), "")
	if updated.Username != "alice2" || updated.Token != "new_token" {
		t.Errorf("expected the updated account, got %+v", updated)
	}

	// The old username must no longer resolve anything of its own --
	// alice's entry was replaced, not duplicated alongside a new one.
	if accounts := cf.GitHubAccounts("github.com"); len(accounts) != 1 || accounts[0].Username != "bob" {
		t.Errorf("expected only bob left on the public host, got %+v", accounts)
	}

	// bob, an unrelated account, must survive untouched.
	bob := ResolveAccount(cf.GitHubAccounts("github.com"), "bob")
	if bob.Username != "bob" || bob.Token != "bob_token" {
		t.Errorf("expected bob preserved, got %+v", bob)
	}

	if caps := cf.SpecialCapitalization(); caps["api"] != "API" {
		t.Errorf("expected special_capitalization preserved, got %v", caps)
	}
}

func Test_UpdateAccount_NotFound(t *testing.T) {
	createTestConfigFile(t, "providers:\n  - name: github\n    username: alice\n    token: tok\n")

	cf := NewConfigFile(false)
	err := cf.UpdateAccount("github", "nonexistent", "new", "tok", nil, "")
	if err == nil {
		t.Fatal("expected an error updating a nonexistent account")
	}
}

func Test_RemoveAccount_PreservesOtherEntries(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    username: alice
    token: alice_token
  - name: github
    username: bob
    token: bob_token
special_capitalization:
  api: API
`)

	cf := NewConfigFile(false)
	if err := cf.RemoveAccount("github", "bob"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if usernames := cf.Usernames("github"); len(usernames) != 1 || usernames[0] != "alice" {
		t.Errorf("expected only alice left, got %v", usernames)
	}

	if caps := cf.SpecialCapitalization(); caps["api"] != "API" {
		t.Errorf("expected special_capitalization preserved, got %v", caps)
	}
}

func Test_RemoveAccount_NotFound(t *testing.T) {
	createTestConfigFile(t, "providers:\n  - name: github\n    username: alice\n    token: tok\n")

	cf := NewConfigFile(false)
	err := cf.RemoveAccount("github", "nonexistent")
	if err == nil {
		t.Fatal("expected an error removing a nonexistent account")
	}
}
