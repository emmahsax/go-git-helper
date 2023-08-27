package gitlabMergeRequest

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/gitlab"
)

type GitLabMergeRequest struct {
	baseBranch   string
	localBranch  string
	localProject string
	newMrTitle   string
}

func NewGitLabMergeRequest(options map[string]string) *GitLabMergeRequest {
	return &GitLabMergeRequest{
		baseBranch:   options["baseBranch"],
		localBranch:  options["localBranch"],
		localProject: options["localProject"],
		newMrTitle:   options["newMrTitle"],
	}
}

func (mr *GitLabMergeRequest) Create() {
	body := newMrBody()
	optionsMap := map[string]string{
		"description":          body,
		"remove_source_branch": "true",
		"squash":               "true",
		"source_branch":        mr.localBranch,
		"target_branch":        mr.baseBranch,
		"title":                mr.newMrTitle,
	}

	fmt.Println("Creating merge request:", mr.newMrTitle)
	mrResponse := gitlabClient().CreateMergeRequest(mr.localProject, optionsMap).(gitlab.Response)

	if mrResponse.WebURL == "" {
		errorMessage := mrResponse.Message[0]
		debug.PrintStack()
		log.Fatal("Could not create merge request: " + errorMessage)
	} else {
		fmt.Println("Merge request successfully created:", mrResponse.WebURL)
	}
}

func newMrBody() string {
	templateName := templateNameToApply()
	if templateName != "" {
		content, err := os.ReadFile(templateName)
		if err != nil {
			debug.PrintStack()
			log.Fatal(err)
		}

		return string(content)
	}
	return ""
}

func templateNameToApply() string {
	templateName := ""
	if len(mrTemplateOptions()) > 0 {
		templateName = determineTemplate()
	}

	return templateName
}

func determineTemplate() string {
	if len(mrTemplateOptions()) == 1 {
		applySingleTemplate := commandline.AskYesNoQuestion(
			fmt.Sprintf("Apply the merge request template from %s?", mrTemplateOptions()[0]),
		)
		if applySingleTemplate {
			return mrTemplateOptions()[0]
		}
	} else {
		response := commandline.AskMultipleChoice(
			"Which merge request template should be applied?", append(mrTemplateOptions(), "None"),
		)
		if response != "None" {
			return response
		}
	}

	return ""
}

func mrTemplateOptions() []string {
	identifiers := map[string]string{
		"templateDir":       ".gitlab",
		"nestedDirName":     "merge_request_templates",
		"nonNestedFileName": "merge_request_template",
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

func gitlabClient() *gitlab.GitLabClient {
	return gitlab.NewGitLabClient()
}