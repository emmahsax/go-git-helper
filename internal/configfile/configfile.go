package configfile

import (
	"errors"
	"os"
	"path/filepath"

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

	return filepath.Join(homeDir, ".git-helper")
}

func (cf *ConfigFile) ConfigDirExists() bool {
	info, err := os.Stat(cf.ConfigDir())
	if err != nil {
		return false
	}

	return info.IsDir()
}

func (cf *ConfigFile) ConfigFile() string {
	return filepath.Join(cf.ConfigDir(), "config.yml")
}

func (cf *ConfigFile) ConfigFileExists() bool {
	_, err := os.Stat(cf.ConfigFile())
	return err == nil
}

func (cf *ConfigFile) GitHubUsername() string {
	return cf.providerUsername("github")
}

func (cf *ConfigFile) GitLabUsername() string {
	return cf.providerUsername("gitlab")
}

func (cf *ConfigFile) GitHubToken() string {
	return cf.providerToken("github")
}

func (cf *ConfigFile) GitLabToken() string {
	return cf.providerToken("gitlab")
}

// providerUsername and providerToken read a git provider's credentials out of
// the config file. Keys follow the "<provider>_username"/"<provider>_token"
// convention (e.g. "github_username", "gitlab_token"), so supporting a new
// provider is a one-line wrapper method rather than a new lookup.
func (cf *ConfigFile) providerUsername(provider string) string {
	return cf.configFileContents()[provider+"_username"]
}

func (cf *ConfigFile) providerToken(provider string) string {
	return cf.configFileContents()[provider+"_token"]
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
