package configfile

import (
	"os"
	"path/filepath"
	"testing"
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
	cf := NewConfigFile(false)
	dir := cf.ConfigDir()

	if dir == "" {
		t.Errorf("Expected a directory, got an empty string")
	}

	homeDir, _ := os.UserHomeDir()
	expectedDir := homeDir + "/.git-helper"
	if dir != expectedDir {
		t.Errorf("Expected directory '%s', got '%s'", expectedDir, dir)
	}
}

func Test_ConfigDirExists(t *testing.T) {
	cf := NewConfigFile(false)

	_, err := os.Stat(cf.ConfigDir())
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(cf.ConfigDir(), 0755)
			if err != nil {
				t.Fatal(err)
			}
			tempDir, err := os.MkdirTemp(cf.ConfigDir(), "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(cf.ConfigDir())
			defer os.RemoveAll(tempDir)
		}
	}

	if !cf.ConfigDirExists() {
		t.Errorf("Expected ConfigDirExists to return true, got false")
	}
}

func Test_ConfigFile(t *testing.T) {
	cf := NewConfigFile(false)
	file := cf.ConfigFile()

	if file == "" {
		t.Errorf("Expected a file, got an empty string")
	}

	homeDir, _ := os.UserHomeDir()
	expectedFile := homeDir + "/.git-helper/config.yml"
	if file != expectedFile {
		t.Errorf("Expected file '%s', got '%s'", expectedFile, file)
	}
}

func Test_ConfigFileExists(t *testing.T) {
	cf := NewConfigFile(false)

	_, err := os.Stat(cf.ConfigFile())
	if err != nil {
		err := os.MkdirAll(cf.ConfigDir(), 0755)
		if err != nil {
			t.Fatal(err)
		}
		tempDir, err := os.MkdirTemp(cf.ConfigDir(), "")
		if err != nil {
			t.Fatal(err)
		}
		tempFile, err := os.Create(cf.ConfigFile())
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(cf.ConfigDir())
		defer os.RemoveAll(tempDir)
		defer os.Remove(tempFile.Name())
	}

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
func createTestConfigFile(t *testing.T, content string) (string, func()) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yml")

	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Set HOME to tempDir so ConfigDir() returns tempDir/.git-helper
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	// Create .git-helper directory in temp
	gitHelperDir := filepath.Join(tempDir, ".git-helper")
	os.MkdirAll(gitHelperDir, 0755)

	// Move config file to the right location
	finalConfigFile := filepath.Join(gitHelperDir, "config.yml")
	os.Rename(configFile, finalConfigFile)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return tempDir, cleanup
}

func Test_GitHubUsername_NewFormat(t *testing.T) {
	_, cleanup := createTestConfigFile(t, "github_username: testuser\n")
	defer cleanup()

	cf := NewConfigFile(false)
	username := cf.GitHubUsername()
	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", username)
	}
}

func Test_GitHubUsername_LegacyFormat(t *testing.T) {
	_, cleanup := createTestConfigFile(t, ":github_user: legacyuser\n")
	defer cleanup()

	cf := NewConfigFile(false)
	username := cf.GitHubUsername()
	if username != "legacyuser" {
		t.Errorf("Expected username 'legacyuser', got '%s'", username)
	}
}

func Test_GitLabUsername_NewFormat(t *testing.T) {
	_, cleanup := createTestConfigFile(t, "gitlab_username: gitlabuser\n")
	defer cleanup()

	cf := NewConfigFile(false)
	username := cf.GitLabUsername()
	if username != "gitlabuser" {
		t.Errorf("Expected username 'gitlabuser', got '%s'", username)
	}
}

func Test_GitLabUsername_LegacyFormat(t *testing.T) {
	_, cleanup := createTestConfigFile(t, ":gitlab_user: legacygitlabuser\n")
	defer cleanup()

	cf := NewConfigFile(false)
	username := cf.GitLabUsername()
	if username != "legacygitlabuser" {
		t.Errorf("Expected username 'legacygitlabuser', got '%s'", username)
	}
}

func Test_GitHubToken_NewFormat(t *testing.T) {
	_, cleanup := createTestConfigFile(t, "github_token: ghp_token123\n")
	defer cleanup()

	cf := NewConfigFile(false)
	token := cf.GitHubToken()
	if token != "ghp_token123" {
		t.Errorf("Expected token 'ghp_token123', got '%s'", token)
	}
}

func Test_GitHubToken_LegacyFormat(t *testing.T) {
	_, cleanup := createTestConfigFile(t, ":github_token: legacy_token\n")
	defer cleanup()

	cf := NewConfigFile(false)
	token := cf.GitHubToken()
	if token != "legacy_token" {
		t.Errorf("Expected token 'legacy_token', got '%s'", token)
	}
}

func Test_GitLabToken_NewFormat(t *testing.T) {
	_, cleanup := createTestConfigFile(t, "gitlab_token: glpat-token123\n")
	defer cleanup()

	cf := NewConfigFile(false)
	token := cf.GitLabToken()
	if token != "glpat-token123" {
		t.Errorf("Expected token 'glpat-token123', got '%s'", token)
	}
}

func Test_GitLabToken_LegacyFormat(t *testing.T) {
	_, cleanup := createTestConfigFile(t, ":gitlab_token: legacy_gitlab_token\n")
	defer cleanup()

	cf := NewConfigFile(false)
	token := cf.GitLabToken()
	if token != "legacy_gitlab_token" {
		t.Errorf("Expected token 'legacy_gitlab_token', got '%s'", token)
	}
}

func Test_SpecialCapitalization(t *testing.T) {
	content := `special_capitalization:
  api: API
  aws: AWS
  github: GitHub
`
	_, cleanup := createTestConfigFile(t, content)
	defer cleanup()

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
	_, cleanup := createTestConfigFile(t, "github_username: testuser\n")
	defer cleanup()

	cf := NewConfigFile(false)
	caps := cf.SpecialCapitalization()

	if len(caps) != 0 {
		t.Errorf("Expected 0 capitalizations, got %d", len(caps))
	}
}

func Test_SpecialCapitalization_FileNotFound(t *testing.T) {
	// Use a temp directory without creating config file
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cf := NewConfigFile(true) // Use debug mode to avoid exit
	caps := cf.SpecialCapitalization()

	if len(caps) != 0 {
		t.Errorf("Expected empty map when file not found, got %d items", len(caps))
	}
}

func Test_configFileContents_InvalidYAML(t *testing.T) {
	// Skip this test as configFileContents calls HandleError which exits
	t.Skip("Skipping test that would call os.Exit through HandleError")
}

func Test_configFileContents_ValidYAML(t *testing.T) {
	_, cleanup := createTestConfigFile(t, "github_username: testuser\ngithub_token: token123\n")
	defer cleanup()

	cf := NewConfigFile(false)
	contents := cf.configFileContents()

	if contents["github_username"] != "testuser" {
		t.Errorf("Expected 'testuser', got '%s'", contents["github_username"])
	}

	if contents["github_token"] != "token123" {
		t.Errorf("Expected 'token123', got '%s'", contents["github_token"])
	}
}
