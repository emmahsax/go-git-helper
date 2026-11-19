package setup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/emmahsax/go-git-helper/plugins"
	"github.com/spf13/cobra"
)

type Setup struct {
	Debug    bool
	Executor executor.ExecutorInterface
	Config   configfile.ConfigFileInterface
	GOOS     string
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "setup",
		Short:                 "Creates a Git Helper config file at ~/.git-helper/config.yml",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newSetup(debug, executor.NewExecutor(debug), configfile.NewConfigFile(debug)).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newSetup(debug bool, executor executor.ExecutorInterface, config configfile.ConfigFileInterface) *Setup {
	return &Setup{
		Debug:    debug,
		Executor: executor,
		Config:   config,
		GOOS:     runtime.GOOS,
	}
}

func (s *Setup) execute() {
	s.setupConfig()
	s.setupPlugins()
	s.setupCompletion()
}

func (s *Setup) setupConfig() {
	if !s.Config.ConfigFileExists() {
		s.createOrUpdateConfig()
		return
	}

	switch commandline.AskMultipleChoice(
		"The "+s.Config.ConfigFile()+" file already exists. What do you want to do?",
		[]string{"Add a new account", "Edit or remove an account", "Replace everything", "Leave it as is"},
	) {
	case "Add a new account":
		s.addAccount()
	case "Edit or remove an account":
		s.editAccount()
	case "Replace everything":
		s.createOrUpdateConfig()
	}
}

// chooseProvider asks the user which supported provider they mean.
func (s *Setup) chooseProvider(question string) gitProviderCredential {
	names := make([]string, len(gitProviderCredentials))
	for i, p := range gitProviderCredentials {
		names[i] = p.Name
	}

	chosenName := commandline.AskMultipleChoice(question, names)

	for _, p := range gitProviderCredentials {
		if p.Name == chosenName {
			return p
		}
	}

	return gitProviderCredential{}
}

// addAccount appends exactly one new account, for a provider the user
// chooses, to the existing config file -- without touching any account
// already there. This is the normal way to add a second account (e.g. a
// work account alongside a personal one) without retyping everything.
func (s *Setup) addAccount() {
	provider := s.chooseProvider("Which provider do you want to add an account for?")

	answers := s.askAccount(provider, s.Config.Usernames(provider.Key))

	if err := s.Config.AddAccount(provider.Key, answers.Username, answers.Token, answers.RemoteHosts, answers.APIHost); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	fmt.Printf("\nAdded a new %s account to %s\n\n", provider.Name, s.Config.ConfigFile())
}

// editAccount lets the user update or remove one existing account, for a
// provider and username they pick, without touching any other account --
// the normal way to rotate a token, rename an account, or drop a stale one
// without retyping everything else.
func (s *Setup) editAccount() {
	provider := s.chooseProvider("Which provider's account do you want to edit or remove?")

	usernames := s.Config.Usernames(provider.Key)
	if len(usernames) == 0 {
		fmt.Println("--- No " + provider.Name + " accounts are configured yet. ---")
		return
	}

	username := commandline.AskMultipleChoice("Which "+provider.Name+" account?", usernames)

	switch commandline.AskMultipleChoice("What do you want to do with the "+provider.Name+" account "+username+"?", []string{"Update it", "Remove it", "Cancel"}) {
	case "Update it":
		s.updateAccount(provider, username, usernames)
	case "Remove it":
		if err := s.Config.RemoveAccount(provider.Key, username); err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}
		fmt.Printf("\nRemoved the %s account %s from %s\n\n", provider.Name, username, s.Config.ConfigFile())
	}
}

// updateAccount re-asks every field for username fresh and overwrites that
// one account in place. existingUsernames excludes username itself from the
// collision guard, so keeping the same username isn't treated as a
// collision with itself.
func (s *Setup) updateAccount(provider gitProviderCredential, username string, existingUsernames []string) {
	var taken []string
	for _, u := range existingUsernames {
		if u != username {
			taken = append(taken, u)
		}
	}

	answers := s.askAccount(provider, taken)

	if err := s.Config.UpdateAccount(provider.Key, username, answers.Username, answers.Token, answers.RemoteHosts, answers.APIHost); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	fmt.Printf("\nUpdated the %s account %s in %s\n\n", provider.Name, username, s.Config.ConfigFile())
}

func (s *Setup) createOrUpdateConfig() {
	content := s.generateConfigFileContents()

	if !s.Config.ConfigDirExists() {
		err := os.MkdirAll(s.Config.ConfigDir(), 0700)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}
	}

	// Also tighten a directory created by an older version with looser permissions.
	if err := os.Chmod(s.Config.ConfigDir(), 0700); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	// The config file holds access tokens, so keep it private to the user. The
	// explicit chmod also tightens a file left over from an older version.
	err := os.WriteFile(s.Config.ConfigFile(), []byte(content), 0600)
	if err == nil {
		err = os.Chmod(s.Config.ConfigFile(), 0600)
	}
	if err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	fmt.Printf("\nDone setting up %s\n\n", s.Config.ConfigFile())
}

// gitProviderCredential describes one supported git provider's setup
// prompts. Every provider can have multiple accounts (e.g. a public
// github.com account alongside a separate Enterprise one, or even two
// accounts that are both just the public URL, told apart only by an SSH
// remote alias) and each account can list multiple remote_hosts.
type gitProviderCredential struct {
	Key               string // matches the "name" written under providers, and read back in configfile.ConfigFile.providersNamed
	Name              string // display name used in setup prompts
	PublicHost        string // e.g. "github.com" -- what leaving remote_hosts blank already means, so it must never be entered there
	TokenPrompt       string
	RemoteHostsPrompt string
}

// gitProviderCredentials lists every supported git provider's setup
// prompts. Adding support for a new provider (e.g. Bitbucket) only requires
// a new entry here -- *configfile.ConfigFile's Accounts/Hosts already read
// any provider name generically.
var gitProviderCredentials = []gitProviderCredential{
	{
		Key:               "github",
		Name:              "GitHub",
		PublicHost:        "github.com",
		TokenPrompt:       "GitHub personal access token",
		RemoteHostsPrompt: "Remote host(s) (comma-separated if more than one) -- leave blank if this is the public github.com",
	},
	{
		Key:               "gitlab",
		Name:              "GitLab",
		PublicHost:        "gitlab.com",
		TokenPrompt:       "GitLab personal access token",
		RemoteHostsPrompt: "Remote host(s) (comma-separated if more than one) -- leave blank if this is the public gitlab.com",
	},
	{
		Key:               "forgejo",
		Name:              "Forgejo",
		PublicHost:        "codeberg.org",
		TokenPrompt:       "Forgejo access token",
		RemoteHostsPrompt: "Remote host(s) (comma-separated if more than one) -- leave blank if this is the public codeberg.org",
	},
}

// generateConfigFileContents writes the config file's "providers" list, one
// entry per provider/account the user opts into, e.g.:
//
//	providers:
//	  - name: github
//	    username: YOUR-USERNAME
//	    token: YOUR-TOKEN
//	  - name: github
//	    username: YOUR-WORK-USERNAME
//	    token: YOUR-WORK-TOKEN
//	    remote_hosts:
//	      - github.com-work # an SSH config alias, not a real GitHub server -- no api_host below, so this still calls the real github.com API
//	  - name: gitlab
//	    username: YOUR-USERNAME
//	    token: YOUR-TOKEN
//	  - name: gitlab
//	    username: YOUR-SELF-HOSTED-USERNAME
//	    token: YOUR-SELF-HOSTED-TOKEN
//	    remote_hosts:
//	      - gitlab.example.com
//	      - git.internal.example.com # an alternate hostname for the same self-hosted instance
//	    api_host: gitlab.example.com # a genuinely different server, so API calls go here too
//
// remote_hosts controls which git remotes recognize this account; api_host
// (only present for a genuinely different server) controls where its API
// calls go. They answer different questions and are never inferred from
// each other -- see askAccount.
func (s *Setup) generateConfigFileContents() string {
	var providerBlocks []string

	for _, provider := range s.chooseProviders() {
		providerBlocks = append(providerBlocks, s.providerAccountBlocks(provider)...)
	}

	if len(providerBlocks) == 0 {
		return "\n"
	}

	return "providers:\n" + strings.Join(providerBlocks, "")
}

// chooseProviders lets the user check off, in one multi-select, every
// provider they want to set up credentials for -- rather than asking a
// separate yes/no question per provider. Adding a new provider to
// gitProviderCredentials automatically appears here; nothing else needs to
// change.
func (s *Setup) chooseProviders() []gitProviderCredential {
	names := make([]string, len(gitProviderCredentials))
	for i, p := range gitProviderCredentials {
		names[i] = p.Name
	}

	selected := commandline.AskMultiSelect("Which providers do you want to set up credentials for?", names)

	var chosen []gitProviderCredential
	for _, p := range gitProviderCredentials {
		if slices.Contains(selected, p.Name) {
			chosen = append(chosen, p)
		}
	}

	return chosen
}

// providerAccountBlocks asks for one or more accounts for provider (already
// chosen via chooseProviders) -- e.g. a public github.com account alongside
// a separate Enterprise one, or two accounts both on the public URL but
// told apart by an SSH remote alias -- looping until the user is done
// adding accounts. taken starts out empty and accumulates each account's
// username as it's added, so two accounts asked about in the same loop
// can't collide with each other any more than they could with an
// already-existing account (see askAccount).
func (s *Setup) providerAccountBlocks(provider gitProviderCredential) []string {
	var blocks []string
	var taken []string

	for {
		answers := s.askAccount(provider, taken)
		taken = append(taken, answers.Username)
		blocks = append(blocks, answers.block(provider.Key))

		another := commandline.AskYesNoQuestion("Do you wish to set up another " + provider.Name + " account (e.g. a different self-hosted instance, or another account on the same host)?")
		if !another {
			return blocks
		}
	}
}

// accountAnswers holds one account's user-entered fields, before they're
// serialized either as a block of config.yml text (the bulk initial setup
// flow, via block) or appended structurally to an existing file (addAccount,
// via configfile.ConfigFile.AddAccount).
type accountAnswers struct {
	Username    string
	Token       string
	RemoteHosts []string
	APIHost     string
}

// askAccount asks for one account's username, token, remote_hosts, and --
// only if remote_hosts names a genuinely different server -- api_host.
// taken is every username already configured for this provider; the
// username prompt re-asks on a collision, since code-request's "which
// account?" prompt tells tied accounts apart by username, and two
// identically-labeled accounts can't be told apart there.
//
// Two separate questions ask about remote_hosts and api_host, deliberately
// not inferred from each other:
//  1. remote_hosts: what should git-helper look for in a git remote to
//     recognize this account? Accepts a comma-separated list, since a
//     single instance may be reachable under more than one hostname. This
//     alone never changes where API calls go -- an SSH alias here has no
//     meaning to any API.
//  2. Only if (1) was answered: is that actually a different real server
//     (Enterprise/self-hosted), as opposed to just a locally-made-up alias?
//     Only "yes" here writes api_host, which is what redirects this
//     account's API calls away from the public API.
func (s *Setup) askAccount(provider gitProviderCredential, taken []string) accountAnswers {
	answers := accountAnswers{
		Username:    s.askUsername(provider, taken),
		Token:       commandline.AskOpenEndedQuestion(provider.TokenPrompt, "", true, true),
		RemoteHosts: s.askRemoteHosts(provider),
	}

	if len(answers.RemoteHosts) == 0 {
		return answers
	}

	isDifferentServer := commandline.AskYesNoQuestion(
		"Is " + answers.RemoteHosts[0] + " a REAL, different " + provider.Name + " server (e.g. " + provider.Name + " Enterprise or a self-hosted instance), " +
			"as opposed to an SSH config alias to tell accounts apart on the public " + provider.Name + "?",
	)
	if !isDifferentServer {
		return answers
	}

	answers.APIHost = answers.RemoteHosts[0]
	if len(answers.RemoteHosts) > 1 {
		answers.APIHost = commandline.AskOpenEndedQuestion(provider.Name+" API host for this instance (the real server address, not an alias)", answers.RemoteHosts[0], false, true)
	}

	return answers
}

// askRemoteHosts asks for remote_hosts, re-asking if the answer includes
// provider's own public domain -- that's not a real self-hosted/alias host,
// and entering it would silently shadow whichever other account relies on
// an absent remote_hosts meaning "the public account" (leaving it blank is
// the correct way to mean that).
//
// Not required: leaving this blank is how an account stays on the
// provider's public URL, so Enter must be accepted immediately.
func (s *Setup) askRemoteHosts(provider gitProviderCredential) []string {
	for {
		hosts := s.splitHosts(commandline.AskOpenEndedQuestion(provider.RemoteHostsPrompt, "", false, false))

		if !utils.ContainsFold(hosts, provider.PublicHost) {
			return hosts
		}

		fmt.Println("--- " + provider.PublicHost + " is " + provider.Name + "'s own public host. Leave remote_hosts blank instead of listing it, or this may have unintended side effects. ---")
	}
}

// askUsername prompts for provider's account username, re-asking if it
// collides with one already taken by another account for the same
// provider.
func (s *Setup) askUsername(provider gitProviderCredential, taken []string) string {
	for {
		username := commandline.AskOpenEndedQuestion(provider.Name+" username", "", false, true)

		if !slices.Contains(taken, username) {
			return username
		}

		fmt.Println("--- " + username + " is already used by another " + provider.Name + " account. Choose a different username so code-request can tell them apart. ---")
	}
}

// block renders a as a config.yml providers list entry for the given
// provider key (e.g. "github", "gitlab").
func (a accountAnswers) block(key string) string {
	block := "  - name: " + key + "\n" +
		"    username: " + a.Username + "\n" +
		"    token: " + a.Token + "\n"

	if len(a.RemoteHosts) > 0 {
		block += "    remote_hosts:\n"
		for _, host := range a.RemoteHosts {
			block += "      - " + host + "\n"
		}
	}

	if a.APIHost != "" {
		block += "    api_host: " + a.APIHost + "\n"
	}

	return block
}

func (s *Setup) splitHosts(input string) []string {
	if input == "" {
		return nil
	}

	var hosts []string
	for _, host := range strings.Split(input, ",") {
		if host = strings.TrimSpace(host); host != "" {
			hosts = append(hosts, host)
		}
	}

	return hosts
}

func (s *Setup) setupPlugins() {
	setup := commandline.AskYesNoQuestion("Do you wish to set up the Git Helper plugins?")

	if setup {
		s.createOrUpdatePlugins()
	}
}

func (s *Setup) createOrUpdatePlugins() {
	pluginsDir := filepath.Join(s.Config.ConfigDir(), "plugins")

	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	entries, err := plugins.FS.ReadDir(".")
	if err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	for _, entry := range entries {
		name := entry.Name()

		content, err := plugins.FS.ReadFile(name)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}

		if err := writeExecutable(filepath.Join(pluginsDir, name), lfLineEndings(content)); err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}

		// cmd.exe and PowerShell can't run the sh script, so also provide a shim
		// that they will find through PATHEXT.
		if s.GOOS == "windows" {
			shim := "@git-helper " + strings.TrimPrefix(name, "git-") + " %*\r\n"
			if err := writeExecutable(filepath.Join(pluginsDir, name+".cmd"), []byte(shim)); err != nil {
				utils.HandleError(err, s.Debug, nil)
				return
			}
		}
	}

	fmt.Printf("\nDone setting up plugins at %s\n", pluginsDir)
	fmt.Printf("\n%s\n\n", pluginPathInstructions(s.GOOS, pluginsDir))
}

// lfLineEndings strips carriage returns so the sh scripts still run when the
// binary was built from a checkout that converted them to CRLF (as Git for
// Windows can), where `#!/bin/sh\r` would fail to execute.
func lfLineEndings(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
}

func writeExecutable(path string, content []byte) error {
	if err := os.WriteFile(path, content, 0755); err != nil {
		return err
	}

	return os.Chmod(path, 0755)
}

func pluginPathInstructions(goos, pluginsDir string) string {
	if goos == "windows" {
		return "Now add the plugins directory to your PATH. In PowerShell:\n" +
			"  $dir = \"" + pluginsDir + "\"\n" +
			"  $parts = @([Environment]::GetEnvironmentVariable(\"Path\", \"User\") -split \";\" | Where-Object { $_ })\n" +
			"  if ($parts -notcontains $dir) { [Environment]::SetEnvironmentVariable(\"Path\", (($parts + $dir) -join \";\"), \"User\") }\n" +
			"Then open a new terminal"
	}

	dir := shellPath(pluginsDir)

	return "Now add this line to your shell's startup file (e.g. ~/.zshrc or ~/.bashrc):\n  export PATH=\"" + dir + ":$PATH\"\nFor fish, run:\n  fish_add_path \"" + dir + "\""
}

// shellPath writes a path under the home directory as $HOME/..., so the line
// users copy into their shell startup files also works when those are shared
// between machines. Other paths are returned unchanged.
func shellPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}

	rel, err := filepath.Rel(home, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}

	return "$HOME/" + filepath.ToSlash(rel)
}

func (s *Setup) setupCompletion() {
	setup := commandline.AskYesNoQuestion("Do you wish to set up Git Helper completion?")

	if setup {
		s.createOrUpdateCompletion()
	}
}

func (s *Setup) createOrUpdateCompletion() {
	shes := []string{"bash", "fish", "powershell", "zsh"}
	completionsDir := filepath.Join(s.Config.ConfigDir(), "completions")
	if err := os.MkdirAll(completionsDir, 0755); err != nil {
		utils.HandleError(err, s.Debug, nil)
		return
	}

	for _, sh := range shes {
		output, err := s.Executor.Exec("actionAndOutput", "git-helper", "completion", sh)
		if err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}

		if err := os.WriteFile(filepath.Join(completionsDir, "completion."+sh), output, 0644); err != nil {
			utils.HandleError(err, s.Debug, nil)
			return
		}
	}

	fmt.Printf("\nCompletions (for bash, fish, powershell, and zsh) generated in %s\n\n%s\n\n", completionsDir, completionInstructions(s.GOOS, completionsDir))
}

func completionInstructions(goos, completionsDir string) string {
	if goos == "windows" {
		return "To activate completion in PowerShell, add this line to your profile (run `notepad $PROFILE` to open it):\n  . \"" + filepath.Join(completionsDir, "completion.powershell") + "\""
	}

	file := func(sh string) string {
		return "\"" + shellPath(filepath.Join(completionsDir, "completion."+sh)) + "\""
	}

	return "To activate completion, add the line for your shell to its startup file:\n" +
		"  zsh  (~/.zshrc):                   [ -f " + file("zsh") + " ] && source " + file("zsh") + "\n" +
		"  bash (~/.bashrc):                  [ -f " + file("bash") + " ] && source " + file("bash") + "\n" +
		"  fish (~/.config/fish/config.fish): test -f " + file("fish") + "; and source " + file("fish")
}
