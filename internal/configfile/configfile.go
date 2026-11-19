package configfile

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/emmahsax/go-git-helper/internal/utils"
	yaml "gopkg.in/yaml.v3"
)

// ConfigFileInterface covers only the config-file-management methods that
// `setup` actually needs to inject a mock for. Provider credential lookups
// (GitHubToken, GitLabCredentials, etc.) are read straight off the concrete
// *ConfigFile elsewhere, so they're intentionally not part of this interface.
type ConfigFileInterface interface {
	ConfigDir() string
	ConfigDirExists() bool
	ConfigFile() string
	ConfigFileExists() bool
}

type ConfigFile struct {
	Debug bool
}

func NewConfigFile(debug bool) *ConfigFile {
	return &ConfigFile{
		Debug: debug,
	}
}

// provider is one entry in the config file's "providers" list: a git host's
// credentials, plus Hosts for a self-hosted instance (e.g. a self-hosted
// GitLab, or a future Forgejo/Gitea provider). An entry with no Hosts means
// the provider's public URL (e.g. gitlab.com). Hosts can have more than one
// entry when the same instance is reachable under more than one hostname
// (e.g. an internal DNS alias alongside a public one). A provider name (e.g.
// "gitlab") can also appear more than once in the providers list, to
// configure multiple accounts -- a public gitlab.com account alongside a
// separate self-hosted one, say.
type provider struct {
	Name     string   `yaml:"name"`
	Username string   `yaml:"username"`
	Token    string   `yaml:"token"`
	Hosts    []string `yaml:"hosts"`
}

// fileContents mirrors the on-disk shape of config.yml.
type fileContents struct {
	Providers             []provider        `yaml:"providers"`
	SpecialCapitalization map[string]string `yaml:"special_capitalization"`
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

// GitHubHosts returns every self-hosted host configured across all "github"
// provider entries (not the implicit public github.com).
func (cf *ConfigFile) GitHubHosts() []string {
	return cf.hostsFor("github")
}

// GitHubCredentials returns the username and token for the "github" provider
// entry whose Hosts contains host. See GitLabCredentials for the exact
// matching rules, which are identical.
func (cf *ConfigFile) GitHubCredentials(host string) (username, token string) {
	return cf.credentialsFor("github", host)
}

// GitLabHosts returns every self-hosted host configured across all "gitlab"
// provider entries (not the implicit public gitlab.com).
func (cf *ConfigFile) GitLabHosts() []string {
	return cf.hostsFor("gitlab")
}

// GitLabCredentials returns the username and token for the "gitlab" provider
// entry whose Hosts contains host. An empty host, or one that doesn't match
// any entry's Hosts, resolves to the entry with no Hosts configured (the
// public gitlab.com account). Returns "", "" if nothing matches.
func (cf *ConfigFile) GitLabCredentials(host string) (username, token string) {
	return cf.credentialsFor("gitlab", host)
}

// hostsFor returns every self-hosted host configured across all of name's
// provider entries (not whichever public URL an entry with no Hosts implies).
func (cf *ConfigFile) hostsFor(name string) []string {
	var hosts []string

	for _, p := range cf.providersNamed(name) {
		hosts = append(hosts, p.Hosts...)
	}

	return hosts
}

// credentialsFor returns the username and token for the name-named provider
// entry whose Hosts contains host. An empty host, or one that doesn't match
// any entry's Hosts, resolves to the entry with no Hosts configured (that
// provider's public account). Returns "", "" if nothing matches.
func (cf *ConfigFile) credentialsFor(name, host string) (username, token string) {
	providers := cf.providersNamed(name)

	for _, p := range providers {
		if slices.Contains(p.Hosts, host) {
			return p.Username, p.Token
		}
	}

	for _, p := range providers {
		if len(p.Hosts) == 0 {
			return p.Username, p.Token
		}
	}

	return "", ""
}

// providersNamed returns every entry in the config file's providers list
// matching name (e.g. "github", "gitlab"), in file order. A missing config
// file, or no entries matching name, both return nil: detection and lookups
// run before anything knows whether the user has ever run `git-helper
// setup`, so neither case can be a fatal error here.
func (cf *ConfigFile) providersNamed(name string) []provider {
	var result []provider

	for _, p := range cf.contents().Providers {
		if p.Name == name {
			result = append(result, p)
		}
	}

	return result
}

func (cf *ConfigFile) SpecialCapitalization() map[string]string {
	if m := cf.contents().SpecialCapitalization; m != nil {
		return m
	}

	return map[string]string{}
}

// contents reads and parses the config file, treating a missing file or
// invalid YAML as "nothing configured" rather than a fatal error.
func (cf *ConfigFile) contents() fileContents {
	var result fileContents

	data, err := os.ReadFile(cf.ConfigFile())
	if err != nil {
		return result
	}

	if err := yaml.Unmarshal(data, &result); err != nil {
		return result
	}

	return result
}
