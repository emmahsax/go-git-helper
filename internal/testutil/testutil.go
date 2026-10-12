package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// IsolateHome points the home directory at a fresh temporary directory so tests
// never read or modify the real ~/.git-helper. It sets both HOME and
// USERPROFILE, since os.UserHomeDir reads HOME on Unix and USERPROFILE on Windows.
func IsolateHome(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	return home
}

// WriteConfig writes content to <home>/.git-helper/config.yml.
func WriteConfig(t *testing.T, home, content string) {
	t.Helper()

	dir := filepath.Join(home, ".git-helper")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
