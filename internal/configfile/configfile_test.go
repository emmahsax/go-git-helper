package configfile

import (
	"fmt"
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
	// created := false

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
			// created = true
			defer os.RemoveAll(cf.ConfigDir())
			defer os.RemoveAll(tempDir)
		}
	}

	fmt.Println(cf.ConfigDir())

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
