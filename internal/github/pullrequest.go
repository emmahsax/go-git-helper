package github

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/utils"
	"github.com/google/go-github/v91/github"
)

type GitHubPullRequest struct {
	BaseBranch      string
	Debug           bool
	Draft           string
	GitRootDir      string
	Host            string
	InteractiveMode bool
	LocalBranch     string
	LocalRepo       string
	NewPrTitle      string
	TicketBaseURL   string
	Username        string
}

func NewGitHubPullRequest(options map[string]string, debug, interactiveMode bool) *GitHubPullRequest {
	return &GitHubPullRequest{
		BaseBranch:      options["baseBranch"],
		Debug:           debug,
		Draft:           options["draft"],
		GitRootDir:      options["gitRootDir"],
		Host:            options["host"],
		InteractiveMode: interactiveMode,
		LocalBranch:     options["localBranch"],
		LocalRepo:       options["localRepo"],
		NewPrTitle:      options["newPrTitle"],
		TicketBaseURL:   options["ticketBaseURL"],
		Username:        options["username"],
	}
}

func (pr *GitHubPullRequest) Create() {
	d, _ := strconv.ParseBool(pr.Draft)
	options := github.CreatePullRequest{
		Base:                pr.BaseBranch,
		Body:                new(pr.newPrBody()),
		Draft:               new(d),
		Head:                pr.LocalBranch,
		MaintainerCanModify: new(true),
		Title:               new(pr.NewPrTitle),
	}

	repo := strings.Split(pr.LocalRepo, "/")

	fmt.Println("Creating pull request:", pr.NewPrTitle)
	resp, err := NewGitHub(pr.Host, pr.Username, pr.Debug).CreatePullRequest(repo[0], repo[1], &options)
	if err != nil {
		customErr := errors.New("could not create pull request: " + err.Error())
		utils.HandleError(customErr, pr.Debug, nil)
		return
	}

	fmt.Println("Pull request successfully created:", *resp.HTMLURL)
}

func (pr *GitHubPullRequest) newPrBody() string {
	var content string

	if templateName := pr.templateNameToApply(); templateName != "" {
		raw, err := os.ReadFile(templateName)
		if err != nil {
			utils.HandleError(err, pr.Debug, nil)
			return ""
		}

		content = string(raw)
	}

	return utils.PrependTicketLink(content, pr.NewPrTitle, "pull request", pr.TicketBaseURL, pr.InteractiveMode)
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
		var applySingleTemplate bool

		if pr.InteractiveMode {
			applySingleTemplate = commandline.AskYesNoQuestion(
				fmt.Sprintf("Apply the pull request template from %s?", strings.TrimPrefix(pr.prTemplateOptions()[0], pr.GitRootDir+"/")),
			)
		} else {
			applySingleTemplate = true
		}

		if applySingleTemplate {
			return pr.prTemplateOptions()[0]
		}
	} else {
		temp := []string{}
		for _, str := range pr.prTemplateOptions() {
			modifiedStr := strings.TrimPrefix(str, pr.GitRootDir+"/")
			temp = append(temp, modifiedStr)
		}

		if pr.InteractiveMode {
			response := commandline.AskMultipleChoice("Choose a pull request template to be applied", append(temp, "None"))

			if response != "None" {
				return response
			}
		} else {
			return pr.prTemplateOptions()[0]
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
