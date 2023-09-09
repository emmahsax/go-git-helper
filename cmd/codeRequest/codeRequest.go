package codeRequest

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/emmahsax/go-git-helper/internal/githubPullRequest"
	"github.com/emmahsax/go-git-helper/internal/gitlabMergeRequest"
	"github.com/spf13/cobra"
)

type CodeRequest struct{
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
			codeRequest(debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func codeRequest(debug bool) *CodeRequest {
	return &CodeRequest{
		Debug: debug,
	}
}

func (cr *CodeRequest) execute() {
	if isGitHub(cr) && isGitLab(cr) {
		askForClarification(cr)
	} else if isGitHub(cr) {
		createGitHub(cr)
	} else if isGitLab(cr) {
		createGitLab(cr)
	} else {
		log.Fatal("Could not locate GitHub or GitLab remote URLs")
	}
}

func askForClarification(cr *CodeRequest) {
	answer := commandline.AskMultipleChoice("Found git remotes for both GitHub and GitLab. Which would you like to proceed with?", []string{"GitHub", "GitLab"})
	if answer == "GitHub" {
		createGitHub(cr)
	} else {
		createGitLab(cr)
	}
}

func createGitHub(cr *CodeRequest) {
	options := make(map[string]string)
	options["baseBranch"] = baseBranch(cr)
	options["newPrTitle"] = newPrTitle(cr)
	g := git.NewGitClient(cr.Debug)
	options["localBranch"] = g.CurrentBranch()
	options["localRepo"] = g.RepoName()
	githubPullRequest.NewGitHubPullRequest(options, cr.Debug).Create()
}

func createGitLab(cr *CodeRequest) {
	options := make(map[string]string)
	options["baseBranch"] = baseBranch(cr)
	options["newMrTitle"] = newMrTitle(cr)
	g := git.NewGitClient(cr.Debug)
	options["localBranch"] = g.CurrentBranch()
	options["localProject"] = g.RepoName()
	gitlabMergeRequest.NewGitLabMergeRequest(options, cr.Debug).Create()
}

func baseBranch(cr *CodeRequest) string {
	g := git.NewGitClient(cr.Debug)
	answer := commandline.AskYesNoQuestion("Is '" + g.DefaultBranch() + "' the correct base branch for your new code request?")

	if answer {
		return g.DefaultBranch()
	} else {
		return commandline.AskOpenEndedQuestion("Base branch?", false)
	}
}

func newMrTitle(cr *CodeRequest) string {
	return newPrTitle(cr)
}

func newPrTitle(cr *CodeRequest) string {
	answer := commandline.AskYesNoQuestion("Accept the autogenerated code request title '" + autogeneratedTitle(cr) + "'?")

	if answer {
		return autogeneratedTitle(cr)
	} else {
		return commandline.AskOpenEndedQuestion("Title?", false)
	}
}

func autogeneratedTitle(cr *CodeRequest) string {
	g := git.NewGitClient(cr.Debug)
	branchArr := strings.FieldsFunc(g.CurrentBranch(), func(r rune) bool {
		return r == '-' || r == '_'
	})

	if len(branchArr) == 0 {
		return ""
	}

	var result string

	if len(branchArr) == 1 {
		result = titleize(branchArr[0])
	} else if checkAllLetters(branchArr[0]) && checkAllNumbers(branchArr[1]) { // Branch includes jira_123 at beginning
		issue := fmt.Sprintf("%s-%s", strings.ToUpper(branchArr[0]), branchArr[1])
		description := strings.Join(branchArr[2:], " ")
		result = fmt.Sprintf("%s %s", issue, titleize(description))
	} else if matchesFullJiraPattern(branchArr[0]) { // Branch includes jira-123 at beginning
		issueSplit := strings.Split(branchArr[0], "-")
		issue := fmt.Sprintf("%s-%s", strings.ToUpper(issueSplit[0]), issueSplit[1])
		description := strings.Join(branchArr[2:], " ")
		result = fmt.Sprintf("%s %s", issue, titleize(description))
	} else {
		result = titleize(strings.Join(branchArr, " "))
	}

	return result
}

func checkAllLetters(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z]+$", s)
	return match
}

func checkAllNumbers(s string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}

func matchesFullJiraPattern(str string) bool {
	match, _ := regexp.MatchString(`^\w+-\d+$`, str)
	return match
}

func titleize(s string) string {
	if len(s) == 0 {
		return s
	}

	firstChar := strings.ToUpper(string(s[0]))
	return firstChar + s[1:]
}

func isGitHub(cr *CodeRequest) bool {
	return containsSubstring(git.NewGitClient(cr.Debug).Remotes(), "github")
}

func isGitLab(cr *CodeRequest) bool {
	return containsSubstring(git.NewGitClient(cr.Debug).Remotes(), "gitlab")
}

func containsSubstring(strs []string, substring string) bool {
	for _, str := range strs {
		if strings.Contains(str, substring) {
			return true
		}
	}
	return false
}
