package githubPullRequest

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/github"
)

type GitHubPullRequest struct {
	BaseBranch  string
	Debug       bool
	LocalBranch string
	LocalRepo   string
	NewPrTitle  string
}

func NewGitHubPullRequest(options map[string]string, debug bool) *GitHubPullRequest {
	return &GitHubPullRequest{
		BaseBranch:  options["baseBranch"],
		Debug:       debug,
		LocalBranch: options["localBranch"],
		LocalRepo:   options["localRepo"],
		NewPrTitle:  options["newPrTitle"],
	}
}

func (pr *GitHubPullRequest) Create() {
	body := pr.newPrBody()
	optionsMap := map[string]interface{}{
		"base":  pr.BaseBranch,
		"body":  body,
		"head":  pr.LocalBranch,
		"title": pr.NewPrTitle,
	}

	fmt.Println("Creating pull request:", pr.NewPrTitle)
	prResponse := pr.github().CreatePullRequest(pr.LocalRepo, optionsMap).(github.Response)

	if prResponse.HtmlURL == "" {
		errorMessage := prResponse.Errors[0].Message
		if pr.Debug {
			debug.PrintStack()
		}
		log.Fatal("Could not create pull request: " + errorMessage)
	} else {
		fmt.Println("Pull request successfully created:", prResponse.HtmlURL)
	}
}

func (pr *GitHubPullRequest) newPrBody() string {
	templateName := pr.templateNameToApply()
	if templateName != "" {
		content, err := os.ReadFile(templateName)
		if err != nil {
			if pr.Debug {
				debug.PrintStack()
			}
			log.Fatal(err)
		}

		return string(content)
	}
	return ""
}

func (pr *GitHubPullRequest) templateNameToApply() string {
	templateName := ""
	if len(pr.prTemplateOptions()) > 0 {
		templateName = pr.determineTemplate()
	}

	return templateName
}

func (pr *GitHubPullRequest) determineTemplate() string {
	if len(pr.prTemplateOptions()) == 1 {
		applySingleTemplate := commandline.AskYesNoQuestion(
			fmt.Sprintf("Apply the pull request template from %s?", pr.prTemplateOptions()[0]),
		)
		if applySingleTemplate {
			return pr.prTemplateOptions()[0]
		}
	} else {
		response := commandline.AskMultipleChoice(
			"Which pull request template should be applied?", append(pr.prTemplateOptions(), "None"),
		)
		if response != "None" {
			return response
		}
	}

	return ""
}

func (pr *GitHubPullRequest) prTemplateOptions() []string {
	identifiers := map[string]string{
		"templateDir":       ".github",
		"nestedDirName":     "PULL_REQUEST_TEMPLATE",
		"nonNestedFileName": "pull_request_template",
	}

	nestedTemplates, _ := filepath.Glob(
		filepath.Join(identifiers["templateDir"], identifiers["nestedDirName"], "*.md"),
	)
	nonNestedTemplates, _ := filepath.Glob(
		filepath.Join(identifiers["templateDir"], identifiers["nonNestedFileName"]+".md"),
	)
	rootTemplates, _ := filepath.Glob(filepath.Join(".", identifiers["nonNestedFileName"]+".md"))

	allTemplates := append(append(nestedTemplates, nonNestedTemplates...), rootTemplates...)
	uniqueTemplates := make(map[string]bool)
	for _, template := range allTemplates {
		uniqueTemplates[template] = true
	}

	templateList := []string{}
	for template := range uniqueTemplates {
		templateList = append(templateList, template)
	}

	return templateList
}

func (pr *GitHubPullRequest) github() *github.GitHub {
	return github.NewGitHub(pr.Debug)
}
