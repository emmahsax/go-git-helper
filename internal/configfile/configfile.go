package configfile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/emmahsax/go-git-helper/internal/utils"
	yaml "gopkg.in/yaml.v3"
)

// ConfigFileInterface covers the config-file-management methods that
// `setup` needs to inject a mock for. Provider credential lookups
// (GitHubAccounts, GitLabAccounts, etc.) are read straight off the concrete
// *ConfigFile elsewhere, so they're intentionally not part of this interface.
type ConfigFileInterface interface {
	ConfigDir() string
	ConfigDirExists() bool
	ConfigFile() string
	ConfigFileExists() bool
	AddAccount(name, username, token string, remoteHosts []string, apiHost string) error
	UpdateAccount(name, oldUsername, newUsername, newToken string, newRemoteHosts []string, newAPIHost string) error
	RemoveAccount(name, username string) error
	Usernames(name string) []string
}

type ConfigFile struct {
	Debug bool

	// cachedContents caches contents()'s result for this *ConfigFile's
	// lifetime, since the same instance is often asked for it more than
	// once per invocation (e.g. matching a provider, then resolving which
	// account). Any write (AddAccount/UpdateAccount/RemoveAccount) refreshes
	// it to the just-written state rather than leaving it stale.
	cachedContents *fileContents
}

func NewConfigFile(debug bool) *ConfigFile {
	return &ConfigFile{
		Debug: debug,
	}
}

// provider is one entry in the config file's "providers" list: an account's
// credentials, plus RemoteHosts -- the strings that identify this account in
// a git remote. RemoteHosts can have more than one entry (e.g. a real
// self-hosted domain reachable under more than one hostname, or just a
// local SSH config alias used to tell two accounts apart). A provider name
// (e.g. "gitlab") can appear more than once in the providers list, to
// configure multiple accounts -- a personal account and a work account,
// say, whether or not they're on the same real server.
//
// RemoteHosts is deliberately separate from APIHost: RemoteHosts only has to
// match something in a git remote, which for an SSH alias is a purely local
// name with no meaning to the provider's API. APIHost is the real server
// this account's API calls should hit, and is empty unless the account is
// genuinely self-hosted/Enterprise -- an entry with no APIHost always talks
// to the provider's real public API, even if RemoteHosts lists an alias.
type provider struct {
	Name        string   `yaml:"name"`
	Username    string   `yaml:"username"`
	Token       string   `yaml:"token"`
	RemoteHosts []string `yaml:"remote_hosts,omitempty"`
	APIHost     string   `yaml:"api_host,omitempty"`
}

// Account is one provider entry's resolved identity: the username and token
// to use, and APIHost -- the real server its API calls should hit (empty
// means the provider's own public API).
type Account struct {
	Username string
	Token    string
	APIHost  string
}

// ResolveAccount picks which of candidates (as returned by
// GitHubAccounts/GitLabAccounts) to use: the one whose Username matches
// username, or -- when username is "" or matches none of them -- the first
// candidate. Callers disambiguate ties (len(candidates) > 1) themselves,
// e.g. by asking the user which account they meant; ResolveAccount's job is
// just turning that choice (or the lack of one) into a single Account.
// Returns the zero Account if candidates is empty.
func ResolveAccount(candidates []Account, username string) Account {
	for _, a := range candidates {
		if a.Username == username {
			return a
		}
	}

	if len(candidates) > 0 {
		return candidates[0]
	}

	return Account{}
}

// fileContents mirrors the on-disk shape of config.yml.
type fileContents struct {
	Providers             []provider        `yaml:"providers,omitempty"`
	SpecialCapitalization map[string]string `yaml:"special_capitalization,omitempty"`
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

// GitHubHosts returns every self-hosted/alias remote host configured across
// all "github" provider entries (not the implicit public github.com).
func (cf *ConfigFile) GitHubHosts() []string {
	return cf.remoteHostsFor("github")
}

// GitHubAccounts returns every "github" account matching host. See
// GitLabAccounts for the exact matching rules, which are identical.
func (cf *ConfigFile) GitHubAccounts(host string) []Account {
	return cf.accountsFor("github", host)
}

// GitLabHosts returns every self-hosted/alias remote host configured across
// all "gitlab" provider entries (not the implicit public gitlab.com).
func (cf *ConfigFile) GitLabHosts() []string {
	return cf.remoteHostsFor("gitlab")
}

// GitLabAccounts returns every "gitlab" account matching host: entries
// whose RemoteHosts contains host, or -- when host matches no entry's
// RemoteHosts -- every entry with no RemoteHosts at all (each a separate
// account on the public gitlab.com). Zero results means nothing is
// configured for host; more than one means the caller must disambiguate
// (e.g. by asking which account), since config.yml allows more than one
// account per host (a personal and a work account, say).
func (cf *ConfigFile) GitLabAccounts(host string) []Account {
	return cf.accountsFor("gitlab", host)
}

// ForgejoHosts returns every self-hosted/alias remote host configured across
// all "forgejo" provider entries (not the implicit public codeberg.org).
func (cf *ConfigFile) ForgejoHosts() []string {
	return cf.remoteHostsFor("forgejo")
}

// ForgejoAccounts returns every "forgejo" account matching host. See
// GitLabAccounts for the exact matching rules, which are identical.
func (cf *ConfigFile) ForgejoAccounts(host string) []Account {
	return cf.accountsFor("forgejo", host)
}

// remoteHostsFor returns every self-hosted/alias remote host configured
// across all of name's provider entries (not whichever public URL an entry
// with no RemoteHosts implies).
func (cf *ConfigFile) remoteHostsFor(name string) []string {
	var hosts []string

	for _, p := range cf.providersNamed(name) {
		hosts = append(hosts, p.RemoteHosts...)
	}

	return hosts
}

// accountsFor returns every name-named account matching host: entries whose
// RemoteHosts contains host, or -- when host matches no entry's RemoteHosts
// -- every entry with no RemoteHosts at all (each a separate account on
// that provider's public URL).
func (cf *ConfigFile) accountsFor(name, host string) []Account {
	providers := cf.providersNamed(name)

	var matched []Account
	for _, p := range providers {
		if utils.ContainsFold(p.RemoteHosts, host) {
			matched = append(matched, Account{Username: p.Username, Token: p.Token, APIHost: p.APIHost})
		}
	}
	if len(matched) > 0 {
		return matched
	}

	var public []Account
	for _, p := range providers {
		if len(p.RemoteHosts) == 0 {
			public = append(public, Account{Username: p.Username, Token: p.Token, APIHost: p.APIHost})
		}
	}

	return public
}

// Usernames returns the username of every name-named account already
// configured (e.g. "github", "gitlab"), in file order -- including any that
// are blank. Used by `setup` to reject a new account's username if it would
// collide with one already taken, since code-request's "which account?"
// prompt tells tied accounts apart by username.
func (cf *ConfigFile) Usernames(name string) []string {
	var usernames []string

	for _, p := range cf.providersNamed(name) {
		usernames = append(usernames, p.Username)
	}

	return usernames
}

// AddAccount appends one account to the config file's providers list for
// the given provider name (e.g. "github", "gitlab"), preserving every
// existing entry and the special_capitalization block untouched. Creates
// the config directory and file if they don't exist yet.
func (cf *ConfigFile) AddAccount(name, username, token string, remoteHosts []string, apiHost string) error {
	updated := cf.contents()
	updated.Providers = append(updated.Providers, provider{
		Name:        name,
		Username:    username,
		Token:       token,
		RemoteHosts: remoteHosts,
		APIHost:     apiHost,
	})

	return cf.writeContents(updated)
}

// UpdateAccount replaces the fields of the first name-named account whose
// Username is oldUsername, preserving its position in the providers list
// and every other entry untouched. Returns an error if no such account
// exists.
func (cf *ConfigFile) UpdateAccount(name, oldUsername, newUsername, newToken string, newRemoteHosts []string, newAPIHost string) error {
	updated := cf.contents()

	idx := indexOfAccount(updated.Providers, name, oldUsername)
	if idx == -1 {
		return fmt.Errorf("no %s account found with username %q", name, oldUsername)
	}

	updated.Providers[idx] = provider{
		Name:        name,
		Username:    newUsername,
		Token:       newToken,
		RemoteHosts: newRemoteHosts,
		APIHost:     newAPIHost,
	}

	return cf.writeContents(updated)
}

// RemoveAccount deletes the first name-named account whose Username is
// username, preserving every other entry untouched. Returns an error if no
// such account exists.
func (cf *ConfigFile) RemoveAccount(name, username string) error {
	updated := cf.contents()

	idx := indexOfAccount(updated.Providers, name, username)
	if idx == -1 {
		return fmt.Errorf("no %s account found with username %q", name, username)
	}

	// Build a fresh slice rather than append(updated.Providers[:idx],
	// updated.Providers[idx+1:]...): that form writes into the backing
	// array shared with cf.cachedContents.Providers, corrupting the
	// cache in place even if writeContents below fails.
	remaining := make([]provider, 0, len(updated.Providers)-1)
	remaining = append(remaining, updated.Providers[:idx]...)
	remaining = append(remaining, updated.Providers[idx+1:]...)
	updated.Providers = remaining

	return cf.writeContents(updated)
}

func indexOfAccount(providers []provider, name, username string) int {
	for i, p := range providers {
		if p.Name == name && p.Username == username {
			return i
		}
	}

	return -1
}

// writeContents marshals updated and writes it to the config file, creating
// the config directory if it doesn't exist yet and keeping the file private
// to the user (it holds access tokens).
func (cf *ConfigFile) writeContents(updated fileContents) error {
	data, err := marshalYAML(updated)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cf.ConfigDir(), 0700); err != nil {
		return err
	}

	// MkdirAll above is a no-op if the directory already exists, so also
	// explicitly tighten a directory left over from an older version with
	// looser permissions (mirrors setup.go's createOrUpdateConfig).
	if err := os.Chmod(cf.ConfigDir(), 0700); err != nil {
		return err
	}

	if err := os.WriteFile(cf.ConfigFile(), data, 0600); err != nil {
		return err
	}

	if err := os.Chmod(cf.ConfigFile(), 0600); err != nil {
		return err
	}

	cf.cachedContents = &updated

	return nil
}

// marshalYAML renders v with 2-space indentation, matching the rest of
// config.yml (yaml.Marshal's own default is 4 spaces).
func marshalYAML(v any) ([]byte, error) {
	var buf bytes.Buffer

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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

// contents reads and parses the config file, treating a missing file as
// "nothing configured" rather than a fatal error. Cached for this
// *ConfigFile's lifetime (see cachedContents); any write refreshes the
// cache rather than invalidating it, so a reused instance never re-reads
// stale data after its own AddAccount/UpdateAccount/RemoveAccount call.
//
// providers and special_capitalization are parsed via two independent
// yaml.Unmarshal calls against the same bytes, each into a struct that only
// declares the field it cares about. yaml.v3 doesn't validate the type of a
// key a struct doesn't map, so a malformed value under one top-level key
// can't blank the other -- only a file-read error or a top-level YAML
// syntax error (which breaks parsing regardless of destination type)
// empties both.
func (cf *ConfigFile) contents() fileContents {
	if cf.cachedContents != nil {
		return *cf.cachedContents
	}

	result := cf.parseContents()
	cf.cachedContents = &result

	return result
}

func (cf *ConfigFile) parseContents() fileContents {
	data, err := os.ReadFile(cf.ConfigFile())
	if err != nil {
		return fileContents{}
	}

	var providers struct {
		Providers []provider `yaml:"providers"`
	}
	if err := yaml.Unmarshal(data, &providers); err != nil {
		providers.Providers = nil
	}

	var specialCap struct {
		SpecialCapitalization map[string]string `yaml:"special_capitalization"`
	}
	if err := yaml.Unmarshal(data, &specialCap); err != nil {
		specialCap.SpecialCapitalization = nil
	}

	return fileContents{
		Providers:             providers.Providers,
		SpecialCapitalization: specialCap.SpecialCapitalization,
	}
}
