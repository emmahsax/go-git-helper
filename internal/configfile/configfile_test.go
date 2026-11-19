package configfile

import (
	"os"
	"path/filepath"
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
    hosts:
      - gitlab.example.com
`)

	cf := NewConfigFile(false)

	githubUsername, githubToken := cf.GitHubCredentials("github.com")
	if githubUsername != "githubuser" {
		t.Errorf("Expected GitHub username 'githubuser', got '%s'", githubUsername)
	}
	if githubToken != "ghp_token123" {
		t.Errorf("Expected GitHub token 'ghp_token123', got '%s'", githubToken)
	}

	gitlabUsername, gitlabToken := cf.GitLabCredentials("gitlab.example.com")
	if gitlabUsername != "gitlabuser" {
		t.Errorf("Expected GitLab username 'gitlabuser', got '%s'", gitlabUsername)
	}
	if gitlabToken != "glpat-token123" {
		t.Errorf("Expected GitLab token 'glpat-token123', got '%s'", gitlabToken)
	}
}

func Test_GitHubHosts(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    token: public_token
  - name: github
    token: enterprise_token
    hosts:
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

func Test_GitHubCredentials_PublicAndSelfHosted(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: github
    username: public_user
    token: public_token
  - name: github
    username: work_user
    token: work_token
    hosts:
      - github.example.com
`)

	cf := NewConfigFile(false)

	if username, token := cf.GitHubCredentials("github.com"); username != "public_user" || token != "public_token" {
		t.Errorf("Expected the public account for github.com, got ('%s', '%s')", username, token)
	}

	if username, token := cf.GitHubCredentials("github.example.com"); username != "work_user" || token != "work_token" {
		t.Errorf("Expected the Enterprise account for github.example.com, got ('%s', '%s')", username, token)
	}
}

func Test_GitHubCredentials_NoneConfigured(t *testing.T) {
	testutil.IsolateHome(t)

	cf := NewConfigFile(false)
	username, token := cf.GitHubCredentials("github.com")
	if username != "" || token != "" {
		t.Errorf("Expected ('', ''), got ('%s', '%s')", username, token)
	}
}

func Test_GitLabHosts(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: gitlab
    token: public_token
  - name: gitlab
    token: self_hosted_token
    hosts:
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

func Test_GitLabCredentials_PublicAccount(t *testing.T) {
	createTestConfigFile(t, "providers:\n  - name: gitlab\n    username: public_user\n    token: public_token\n")

	cf := NewConfigFile(false)
	username, token := cf.GitLabCredentials("gitlab.com")
	if username != "public_user" || token != "public_token" {
		t.Errorf("Expected ('public_user', 'public_token'), got ('%s', '%s')", username, token)
	}
}

func Test_GitLabCredentials_PublicAndSelfHosted(t *testing.T) {
	createTestConfigFile(t, `providers:
  - name: gitlab
    username: public_user
    token: public_token
  - name: gitlab
    username: work_user
    token: work_token
    hosts:
      - gitlab.example.com
      - git.internal.example.com
`)

	cf := NewConfigFile(false)

	if username, token := cf.GitLabCredentials("gitlab.com"); username != "public_user" || token != "public_token" {
		t.Errorf("Expected the public account for gitlab.com, got ('%s', '%s')", username, token)
	}

	if username, token := cf.GitLabCredentials("gitlab.example.com"); username != "work_user" || token != "work_token" {
		t.Errorf("Expected the work account for gitlab.example.com, got ('%s', '%s')", username, token)
	}

	if username, token := cf.GitLabCredentials("git.internal.example.com"); username != "work_user" || token != "work_token" {
		t.Errorf("Expected the work account for its second alias host, got ('%s', '%s')", username, token)
	}
}

func Test_GitLabCredentials_NoneConfigured(t *testing.T) {
	testutil.IsolateHome(t)

	cf := NewConfigFile(false)
	username, token := cf.GitLabCredentials("gitlab.com")
	if username != "" || token != "" {
		t.Errorf("Expected ('', ''), got ('%s', '%s')", username, token)
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

func Test_GitHubCredentials_InvalidYAML(t *testing.T) {
	createTestConfigFile(t, "providers: [this is not: valid: yaml\n")

	cf := NewConfigFile(false)
	_, token := cf.GitHubCredentials("github.com")
	if token != "" {
		t.Errorf("Expected an empty token for invalid YAML, got '%s'", token)
	}
}

func Test_GitHubCredentials_FileNotFound(t *testing.T) {
	testutil.IsolateHome(t)

	cf := NewConfigFile(false)
	_, token := cf.GitHubCredentials("github.com")
	if token != "" {
		t.Errorf("Expected an empty token when the config file doesn't exist, got '%s'", token)
	}
}
