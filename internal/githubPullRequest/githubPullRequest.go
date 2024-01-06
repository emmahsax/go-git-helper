package githubPullRequest

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/github"
)

type GitHubPullRequest struct {
	BaseBranch  string
	Debug       bool
	GitRootDir  string
	LocalBranch string
	LocalRepo   string
	NewPrTitle  string
}

func NewGitHubPullRequest(options map[string]string, debug bool) *GitHubPullRequest {
	return &GitHubPullRequest{
		BaseBranch:  options["baseBranch"],
		Debug:       debug,
		GitRootDir:  options["gitRootDir"],
		LocalBranch: options["localBranch"],
		LocalRepo:   options["localRepo"],
		NewPrTitle:  options["newPrTitle"],
	}
}

func (pr *GitHubPullRequest) Create() {
	optionsMap := map[string]string{
		"base":  pr.BaseBranch,
		"body":  pr.newPrBody(),
		"head":  pr.LocalBranch,
		"title": pr.NewPrTitle,
	}
	repo := strings.Split(pr.LocalRepo, "/")

	fmt.Println("Creating pull request:", pr.NewPrTitle)
	resp, err := pr.github().CreatePullRequest(repo[0], repo[1], optionsMap)
	if err != nil {
		if pr.Debug {
			debug.PrintStack()
		}
		log.Fatal("Could not create pull request: " + err.Error())
	}

	fmt.Println("Pull request successfully created:", *resp.HTMLURL)
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
			fmt.Sprintf("Apply the pull request template from %s?", strings.TrimPrefix(pr.prTemplateOptions()[0], pr.GitRootDir+"/")),
		)
		if applySingleTemplate {
			return pr.prTemplateOptions()[0]
		}
	} else {
		temp := []string{}
		for _, str := range pr.prTemplateOptions() {
			modifiedStr := strings.TrimPrefix(str, pr.GitRootDir+"/")
			temp = append(temp, modifiedStr)
		}

		response := commandline.AskMultipleChoice(
			"Choose a pull request template to be applied", append(temp, "None"),
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
		filepath.Join(pr.GitRootDir, identifiers["templateDir"], identifiers["nestedDirName"], "*.md"),
	)
	nonNestedTemplates, _ := filepath.Glob(
		filepath.Join(pr.GitRootDir, identifiers["templateDir"], identifiers["nonNestedFileName"]+".md"),
	)
	rootTemplates, _ := filepath.Glob(filepath.Join(pr.GitRootDir, identifiers["nonNestedFileName"]+".md"))

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
