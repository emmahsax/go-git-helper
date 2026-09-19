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
	SpecialCapitalization() map[string]string
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

func (cf *ConfigFile) GitHubUsername() string {
	configFile := cf.configFileContents()
	return configFile["github_username"]
}

func (cf *ConfigFile) GitLabUsername() string {
	configFile := cf.configFileContents()
	return configFile["gitlab_username"]
}

func (cf *ConfigFile) GitHubToken() string {
	configFile := cf.configFileContents()
	return configFile["github_token"]
}

func (cf *ConfigFile) GitLabToken() string {
	configFile := cf.configFileContents()
	return configFile["gitlab_token"]
}

func (cf *ConfigFile) SpecialCapitalization() map[string]string {
	var result map[string]interface{}
	data, err := os.ReadFile(cf.ConfigFile())
	if err != nil {
		return map[string]string{}
	}

	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return map[string]string{}
	}

	if specialCap, ok := result["special_capitalization"].(map[string]interface{}); ok {
		capitalizations := make(map[string]string)
		for key, value := range specialCap {
			if strValue, ok := value.(string); ok {
				capitalizations[key] = strValue
			}
		}
		return capitalizations
	}

	return map[string]string{}
}

func (cf *ConfigFile) configFileContents() map[string]string {
	var rawResult map[string]interface{}
	data, err := os.ReadFile(cf.ConfigFile())
	if err != nil {
		customErr := errors.New("error reading file: " + err.Error())
		utils.HandleError(customErr, cf.Debug, nil)
		return map[string]string{}
	}

	err = yaml.Unmarshal(data, &rawResult)
	if err != nil {
		customErr := errors.New("error unmarshaling YAML: " + err.Error())
		utils.HandleError(customErr, cf.Debug, nil)
		return map[string]string{}
	}

	// Convert to map[string]string, skipping non-string values
	result := make(map[string]string)
	for key, value := range rawResult {
		if strValue, ok := value.(string); ok {
			result[key] = strValue
		}
	}

	return result
}
