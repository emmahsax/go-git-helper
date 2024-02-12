package codeRequest

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/emmahsax/go-git-helper/internal/githubPullRequest"
	"github.com/emmahsax/go-git-helper/internal/gitlabMergeRequest"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/spf13/cobra"
)

type CodeRequest struct {
	Debug bool
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "code-request",
		Short:                 "Create either a GitHub pull request or a GitLab merge request",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newCodeRequest(debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newCodeRequest(debug bool) *CodeRequest {
	return &CodeRequest{
		Debug: debug,
	}
}

func (cr *CodeRequest) execute() {
	if cr.isGitHub() && cr.isGitLab() {
		cr.askForClarification()
	} else if cr.isGitHub() {
		cr.createGitHub()
	} else if cr.isGitLab() {
		cr.createGitLab()
	} else {
		err := errors.New("could not locate GitHub or GitLab remote URLs")
		utils.HandleError(err, cr.Debug, nil)
	}
}

func (cr *CodeRequest) askForClarification() {
	answer := commandline.AskMultipleChoice("Found git remotes for both GitHub and GitLab. Choose one to proceed with", []string{"GitHub", "GitLab"})
	if answer == "GitHub" {
		cr.createGitHub()
	} else {
		cr.createGitLab()
	}
}

func (cr *CodeRequest) createGitHub() {
	options := make(map[string]string)
	options["baseBranch"] = cr.baseBranch()
	options["draft"] = cr.draft()
	options["newPrTitle"] = cr.newPrTitle()
	g := git.NewGit(cr.Debug, nil)
	options["gitRootDir"] = g.GetGitRootDir()
	options["localBranch"] = g.CurrentBranch()
	options["localRepo"] = g.RepoName()
	githubPullRequest.NewGitHubPullRequest(options, cr.Debug).Create()
}

func (cr *CodeRequest) createGitLab() {
	options := make(map[string]string)
	options["baseBranch"] = cr.baseBranch()
	options["draft"] = cr.draft()
	options["newMrTitle"] = cr.newMrTitle()
	g := git.NewGit(cr.Debug, nil)
	options["gitRootDir"] = g.GetGitRootDir()
	options["localBranch"] = g.CurrentBranch()
	options["localProject"] = g.RepoName()
	gitlabMergeRequest.NewGitLabMergeRequest(options, cr.Debug).Create()
}

func (cr *CodeRequest) baseBranch() string {
	g := git.NewGit(cr.Debug, nil)
	answer := commandline.AskYesNoQuestion("Is '" + g.DefaultBranch() + "' the correct base branch for your new code request?")

	if answer {
		return g.DefaultBranch()
	} else {
		return commandline.AskOpenEndedQuestion("Base branch", false)
	}
}

func (cr *CodeRequest) draft() string {
	answer := commandline.AskYesNoQuestion("Create a draft code request?")

	if answer {
		return "true"
	} else {
		return "false"
	}
}

func (cr *CodeRequest) newMrTitle() string {
	return cr.newPrTitle()
}

func (cr *CodeRequest) newPrTitle() string {
	answer := commandline.AskYesNoQuestion("Accept the autogenerated code request title '" + cr.autogeneratedTitle() + "'?")

	if answer {
		return cr.autogeneratedTitle()
	} else {
		return commandline.AskOpenEndedQuestion("Title", false)
	}
}

func (cr *CodeRequest) autogeneratedTitle() string {
	g := git.NewGit(cr.Debug, nil)
	branchArr := strings.FieldsFunc(g.CurrentBranch(), func(r rune) bool {
		return r == '-' || r == '_'
	})

	if len(branchArr) == 0 {
		return ""
	}

	var result string

	if len(branchArr) == 1 {
		result = cr.titleize(branchArr[0])
	} else if cr.checkAllLetters(branchArr[0]) && cr.checkAllNumbers(branchArr[1]) { // Branch includes jira_123 at beginning
		issue := fmt.Sprintf("%s-%s", strings.ToUpper(branchArr[0]), branchArr[1])
		description := strings.Join(branchArr[2:], " ")
		result = fmt.Sprintf("%s %s", issue, cr.titleize(description))
	} else if cr.matchesFullJiraPattern(branchArr[0]) { // Branch includes jira-123 at beginning
		issueSplit := strings.Split(branchArr[0], "-")
		issue := fmt.Sprintf("%s-%s", strings.ToUpper(issueSplit[0]), issueSplit[1])
		description := strings.Join(branchArr[2:], " ")
		result = fmt.Sprintf("%s %s", issue, cr.titleize(description))
	} else {
		result = cr.titleize(strings.Join(branchArr, " "))
	}

	return result
}

func (cr *CodeRequest) checkAllLetters(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z]+$", s)
	return match
}

func (cr *CodeRequest) checkAllNumbers(s string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}

func (cr *CodeRequest) matchesFullJiraPattern(str string) bool {
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

func (cr *CodeRequest) isGitHub() bool {
	return cr.containsSubstring(git.NewGit(cr.Debug, nil).Remotes(), "github")
}

func (cr *CodeRequest) isGitLab() bool {
	return cr.containsSubstring(git.NewGit(cr.Debug, nil).Remotes(), "gitlab")
}

func (cr *CodeRequest) containsSubstring(strs []string, substring string) bool {
	for _, str := range strs {
		if strings.Contains(str, substring) {
			return true
		}
	}
	return false
}
