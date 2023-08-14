package configfile

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func ConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error:", err)
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

func GitHubUsername() string {
	configFile := configFileContents()
	return configFile["github_username"]
}

func GitLabUsername() string {
	configFile := configFileContents()
	return configFile["gitlab_username"]
}

func GitHubToken() string {
	configFile := configFileContents()
	return configFile["github_token"]
}

func GitLabToken() string {
	configFile := configFileContents()
	return configFile["gitlab_token"]
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
