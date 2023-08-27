package configfile

import (
	"log"
	"os"
	"runtime/debug"

	yaml "gopkg.in/yaml.v3"
)

func ConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
		return ""
	}

	return homeDir + "/.git_helper"
}

func ConfigDirExists() bool {
	info, err := os.Stat(ConfigDir())
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return info.IsDir()
}

func ConfigFile() string {
	return ConfigDir() + "/config.yml"
}

func ConfigFileExists() bool {
	_, err := os.Stat(ConfigFile())
	return err == nil
}

// TODO: pull from the values w/o the : at the beginning, as that's leftover from ruby to go migration

func GitHubUsername() string {
	configFile := configFileContents()
	if configFile["github_username"] != "" {
		return configFile["github_username"]
	} else {
		return configFile[":github_user"]
	}
}

func GitLabUsername() string {
	configFile := configFileContents()
	if configFile["gitlab_username"] != "" {
		return configFile["gitlab_username"]
	} else {
		return configFile[":gitlab_user"]
	}
}

func GitHubToken() string {
	configFile := configFileContents()
	if configFile["github_token"] != "" {
		return configFile["github_token"]
	} else {
		return configFile[":github_token"]
	}
}

func GitLabToken() string {
	configFile := configFileContents()
	if configFile["gitlab_token"] != "" {
		return configFile["gitlab_token"]
	} else {
		return configFile[":gitlab_token"]
	}
}

func configFileContents() map[string]string {
	data, err := os.ReadFile(ConfigFile())
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	var result map[string]string
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	return result
}
