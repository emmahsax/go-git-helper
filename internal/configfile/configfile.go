package configfile

import (
	"errors"
	"os"

	"github.com/emmahsax/go-git-helper/internal/utils"
	yaml "gopkg.in/yaml.v3"
)

type ConfigFileInterface interface {
	ConfigDir() string
	ConfigDirExists() bool
	ConfigFile() string
	ConfigFileExists() bool
	GitHubUsername() string
	GitLabUsername() string
	GitHubToken() string
	GitLabToken() string
}

type ConfigFile struct {
	Debug bool
}

func NewConfigFile(debug bool) *ConfigFile {
	return &ConfigFile{
		Debug: debug,
	}
}

func (cf *ConfigFile) ConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		utils.HandleError(err, cf.Debug, nil)
		return ""
	}

	return homeDir + "/.git-helper"
}

func (cf *ConfigFile) ConfigDirExists() bool {
	info, err := os.Stat(cf.ConfigDir())
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return info.IsDir()
}

func (cf *ConfigFile) ConfigFile() string {
	return cf.ConfigDir() + "/config.yml"
}

func (cf *ConfigFile) ConfigFileExists() bool {
	_, err := os.Stat(cf.ConfigFile())
	return err == nil
}

// TODO: pull from the values w/o the : at the beginning, as that's leftover from ruby to go migration

func (cf *ConfigFile) GitHubUsername() string {
	configFile := cf.configFileContents()
	if configFile["github_username"] != "" {
		return configFile["github_username"]
	} else {
		return configFile[":github_user"]
	}
}

func (cf *ConfigFile) GitLabUsername() string {
	configFile := cf.configFileContents()
	if configFile["gitlab_username"] != "" {
		return configFile["gitlab_username"]
	} else {
		return configFile[":gitlab_user"]
	}
}

func (cf *ConfigFile) GitHubToken() string {
	configFile := cf.configFileContents()
	if configFile["github_token"] != "" {
		return configFile["github_token"]
	} else {
		return configFile[":github_token"]
	}
}

func (cf *ConfigFile) GitLabToken() string {
	configFile := cf.configFileContents()
	if configFile["gitlab_token"] != "" {
		return configFile["gitlab_token"]
	} else {
		return configFile[":gitlab_token"]
	}
}

func (cf *ConfigFile) configFileContents() map[string]string {
	var result map[string]string
	data, err := os.ReadFile(cf.ConfigFile())
	if err != nil {
		customErr := errors.New("error reading file: " + err.Error())
		utils.HandleError(customErr, cf.Debug, nil)
		return result
	}

	err = yaml.Unmarshal(data, &result)
	if err != nil {
		customErr := errors.New("error unmarshaling YAML: " + err.Error())
		utils.HandleError(customErr, cf.Debug, nil)
		return result
	}

	return result
}
