package configfile

import (
	"os"
	"testing"
)

func Test_ConfigDir(t *testing.T) {
	cf := NewConfigFile(false)
	dir := cf.ConfigDir()

	if dir == "" {
		t.Errorf("Expected a directory, got an empty string")
	}
}

func Test_ConfigDirExists(t *testing.T) {
	cf := NewConfigFile(false)

	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

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
}

func Test_ConfigFileExists(t *testing.T) {
	cf := NewConfigFile(false)

	tempFile, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile)

	if !cf.ConfigFileExists() {
		t.Errorf("Expected ConfigFileExists to return true, got false")
	}
}
