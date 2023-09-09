package configfile

import (
	"os"
	"testing"
)

var (
	home, _ = os.UserHomeDir()
)

func TestConfigDir(t *testing.T) {
	cf := NewConfigFileClient(false)
	output := cf.ConfigDir()

	if output != home+"/.git_helper" {
		t.Fatalf(`ConfigDir should be %s, not %s`, home+"/.git_helper", output)
	}
}

func TestConfigFile(t *testing.T) {
	cf := NewConfigFileClient(false)
	output := cf.ConfigFile()

	if output != home+"/.git_helper/config.yml" {
		t.Fatalf(`ConfigFile should be %s, not %s`, home+"/.git_helper/config.yml", output)
	}
}
