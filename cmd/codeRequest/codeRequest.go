package codeRequest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/configfile"
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/spf13/cobra"
)

type CodeRequest struct {
	Debug           bool
	Executor        executor.ExecutorInterface
	InteractiveMode bool

	// gitClient and configFile memoize git()/config() for this invocation's
	// lifetime: both are asked for more than once (detection, then again
	// when actually creating the request), and each already caches its own
	// underlying subprocess call / file read -- but only if the same
	// instance is reused rather than reconstructed each time.
	gitClient  *git.Git
	configFile *configfile.ConfigFile
}

// git returns a *git.Git shared across this invocation.
func (cr *CodeRequest) git() *git.Git {
	if cr.gitClient == nil {
		cr.gitClient = git.NewGit(cr.Debug, cr.Executor)
	}

	return cr.gitClient
}

// config returns a *configfile.ConfigFile shared across this invocation.
func (cr *CodeRequest) config() *configfile.ConfigFile {
	if cr.configFile == nil {
		cr.configFile = configfile.NewConfigFile(cr.Debug)
	}

	return cr.configFile
}

func NewCommand() *cobra.Command {
	var (
		debug           bool
		interactiveMode bool
	)

	cmd := &cobra.Command{
		Use:                   "code-request",
		Short:                 "Create a GitHub, GitLab, or Forgejo pull/merge request",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newCodeRequest(debug, interactiveMode, executor.NewExecutor(debug)).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")
	cmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", true, "interactive mode")

	return cmd
}

func newCodeRequest(debug, interactiveMode bool, executor executor.ExecutorInterface) *CodeRequest {
	return &CodeRequest{
		Debug:           debug,
		Executor:        executor,
		InteractiveMode: interactiveMode,
	}
}

// codeRequestProvider pairs a git host provider's remote-detection and
// request-creation logic. Supporting a new provider (e.g. Bitbucket) only
// requires a new entry in providers(), plus its own detection method and
// createX method (mirroring isGitHub/createGitHub) -- execute() and
// askForClarification() don't need to change.
type codeRequestProvider struct {
	Name    string
	Matches func(cr *CodeRequest) bool
	Create  func(cr *CodeRequest)
}

func (cr *CodeRequest) providers() []codeRequestProvider {
	return []codeRequestProvider{
		{Name: "GitHub", Matches: (*CodeRequest).isGitHub, Create: (*CodeRequest).createGitHub},
		{Name: "GitLab", Matches: (*CodeRequest).isGitLab, Create: (*CodeRequest).createGitLab},
		{Name: "Forgejo", Matches: (*CodeRequest).isForgejo, Create: (*CodeRequest).createForgejo},
	}
}

func (cr *CodeRequest) matchingProviders() []codeRequestProvider {
	var matched []codeRequestProvider

	for _, p := range cr.providers() {
		if p.Matches(cr) {
			matched = append(matched, p)
		}
	}

	return matched
}

func (cr *CodeRequest) execute() {
	matched := cr.matchingProviders()

	switch len(matched) {
	case 0:
		names := make([]string, len(cr.providers()))
		for i, p := range cr.providers() {
			names[i] = p.Name
		}
		err := fmt.Errorf("could not locate a git remote for a supported provider (%s)", strings.Join(names, " or "))
		utils.HandleError(err, cr.Debug, nil)
	case 1:
		matched[0].Create(cr)
	default:
		cr.askForClarification(matched)
	}
}

func (cr *CodeRequest) askForClarification(matched []codeRequestProvider) {
	names := make([]string, len(matched))
	for i, p := range matched {
		names[i] = p.Name
	}

	var answer string

	if cr.InteractiveMode {
		answer = commandline.AskMultipleChoice("Found git remotes for multiple supported providers. Choose one to proceed with:", names)
	} else {
		answer = names[0]
	}

	for _, p := range matched {
		if p.Name == answer {
			p.Create(cr)
			return
		}
	}
}

// chosenUsername disambiguates which account to use when more than one
// candidate (as returned by configfile.ConfigFile.Accounts) resolves to the
// same remote host -- e.g. a personal and a work account that are both just
// the public github.com/gitlab.com, told apart only by
// an SSH remote alias on one of them. Asks interactively when there's a
// real choice to make; otherwise (including non-interactive mode, where
// nothing can answer a prompt) returns "", leaving
// configfile.ResolveAccount's default of the first configured account.
func (cr *CodeRequest) chosenUsername(providerName string, candidates []configfile.Account) string {
	if len(candidates) < 2 || !cr.InteractiveMode {
		return ""
	}

	usernames := make([]string, len(candidates))
	for i, a := range candidates {
		usernames[i] = a.Username
	}

	return commandline.AskMultipleChoice("Found multiple "+providerName+" accounts for this remote. Choose one to use:", usernames)
}

func (cr *CodeRequest) baseBranch() string {
	defaultBranch := cr.git().DefaultBranch()

	if !cr.InteractiveMode {
		return defaultBranch
	}

	return commandline.AskOpenEndedQuestion("Base branch", defaultBranch, false, true)
}

func (cr *CodeRequest) draft() string {
	if !cr.InteractiveMode {
		return "true"
	}

	if commandline.AskYesNoQuestion("Create a draft code request?") {
		return "true"
	}

	return "false"
}

func (cr *CodeRequest) newMrTitle() string {
	return cr.newPrTitle()
}

func (cr *CodeRequest) newPrTitle() string {
	autogeneratedTitle := cr.autogeneratedTitle()

	if !cr.InteractiveMode {
		return autogeneratedTitle
	}

	return commandline.AskOpenEndedQuestion("Title", autogeneratedTitle, false, true)
}

func (cr *CodeRequest) autogeneratedTitle() string {
	g := cr.git()
	branchArr := strings.FieldsFunc(g.CurrentBranch(), func(r rune) bool {
		return r == '-' || r == '_'
	})

	if len(branchArr) == 0 {
		return ""
	}

	var result string

	if len(branchArr) == 1 {
		result = cr.titleize(branchArr[0])
	} else if cr.checkAllLetters(branchArr[0]) && cr.checkAllNumbers(branchArr[1]) { // Branch includes proj_123 at beginning
		issue := fmt.Sprintf("%s-%s", strings.ToUpper(branchArr[0]), branchArr[1])
		description := strings.Join(branchArr[2:], " ")
		result = fmt.Sprintf("%s %s", issue, cr.titleize(description))
	} else if cr.matchesFullProjectTicketPattern(branchArr[0]) { // Branch includes proj-123 at beginning
		issueSplit := strings.Split(branchArr[0], "-")
		issue := fmt.Sprintf("%s-%s", strings.ToUpper(issueSplit[0]), issueSplit[1])
		description := strings.Join(branchArr[2:], " ")
		result = fmt.Sprintf("%s %s", issue, cr.titleize(description))
	} else {
		result = cr.titleize(strings.Join(branchArr, " "))
	}

	return cr.applySpecialCapitalization(result)
}

func (cr *CodeRequest) checkAllLetters(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z]+$", s)
	return match
}

func (cr *CodeRequest) checkAllNumbers(s string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}

func (cr *CodeRequest) matchesFullProjectTicketPattern(str string) bool {
	match, _ := regexp.MatchString(`^\w+-\d+$`, str)
	return match
}

func (cr *CodeRequest) titleize(s string) string {
	if len(s) == 0 {
		return s
	}

	firstChar := strings.ToUpper(string(s[0]))
	return firstChar + s[1:]
}

func (cr *CodeRequest) applySpecialCapitalization(title string) string {
	cf := cr.config()
	capitalizationMap := cf.SpecialCapitalization()

	if len(capitalizationMap) == 0 {
		return title
	}

	words := strings.FieldsFunc(title, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '/' || r == '\\' || r == '.'
	})

	delimiters := []rune{}
	for _, char := range title {
		if char == ' ' || char == '-' || char == '_' || char == '/' || char == '\\' || char == '.' {
			delimiters = append(delimiters, char)
		}
	}

	for i, word := range words {
		if specialCap, ok := capitalizationMap[strings.ToLower(word)]; ok {
			words[i] = specialCap
		}
	}

	if len(words) == 0 {
		return title
	}

	result := words[0]
	for i := 1; i < len(words); i++ {
		if i-1 < len(delimiters) {
			result += string(delimiters[i-1]) + words[i]
		} else {
			result += " " + words[i]
		}
	}

	return result
}

// matchedHost returns whichever of publicHost or configuredHosts (a
// provider's self-hosted/alias hosts from config.yml, there can be more
// than one, e.g. a public account alongside a separate self-hosted one)
// matches one of remoteHosts, or "" if none do. The result identifies which
// configured account to use, since more than one may be configured.
// Matching is exact (case-insensitively) and not substring-based:
// "github.company.com" must not match "github.com", nor "git.example.com"
// match "sub.git.example.com".
func matchedHost(remoteHosts []string, publicHost string, configuredHosts []string) string {
	for _, host := range append([]string{publicHost}, configuredHosts...) {
		if utils.ContainsFold(remoteHosts, host) {
			return host
		}
	}

	return ""
}
