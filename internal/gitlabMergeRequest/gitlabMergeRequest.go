package gitlabMergeRequest

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/gitlab"
)

type GitLabMergeRequest struct {
	BaseBranch   string
	Debug        bool
	GitRootDir   string
	LocalBranch  string
	LocalProject string
	NewMrTitle   string
}

func NewGitLabMergeRequest(options map[string]string, debug bool) *GitLabMergeRequest {
	return &GitLabMergeRequest{
		BaseBranch:   options["baseBranch"],
		Debug:        debug,
		GitRootDir:   options["gitRootDir"],
		LocalBranch:  options["localBranch"],
		LocalProject: options["localProject"],
		NewMrTitle:   options["newMrTitle"],
	}
}

func (mr *GitLabMergeRequest) Create() {
	optionsMap := map[string]string{
		"description":          mr.newMrBody(),
		"remove_source_branch": "true",
		"squash":               "true",
		"source_branch":        mr.LocalBranch,
		"target_branch":        mr.BaseBranch,
		"title":                mr.NewMrTitle,
	}

	fmt.Println("Creating merge request:", mr.NewMrTitle)
	resp, err := mr.gitlab().CreateMergeRequest(mr.LocalProject, optionsMap)
	if err != nil {
		if mr.Debug {
			debug.PrintStack()
		}
		log.Fatal("Could not create merge request: " + err.Error())
	}

	fmt.Println("Merge request successfully created:", resp.WebURL)
}

func (mr *GitLabMergeRequest) newMrBody() string {
	templateName := mr.templateNameToApply()
	if templateName != "" {
		content, err := os.ReadFile(templateName)
		if err != nil {
			if mr.Debug {
				debug.PrintStack()
			}
			log.Fatal(err)
		}

		return string(content)
	}
	return ""
}

func (mr *GitLabMergeRequest) templateNameToApply() string {
	templateName := ""
	if len(mr.mrTemplateOptions()) > 0 {
		templateName = mr.determineTemplate()
	}

	return templateName
}

func (mr *GitLabMergeRequest) determineTemplate() string {
	if len(mr.mrTemplateOptions()) == 1 {
		applySingleTemplate := commandline.AskYesNoQuestion(
			fmt.Sprintf("Apply the merge request template from %s?", strings.TrimPrefix(mr.mrTemplateOptions()[0], mr.GitRootDir+"/")),
		)
		if applySingleTemplate {
			return mr.mrTemplateOptions()[0]
		}
	} else {
		temp := []string{}
		for _, str := range mr.mrTemplateOptions() {
			modifiedStr := strings.TrimPrefix(str, mr.GitRootDir+"/")
			temp = append(temp, modifiedStr)
		}

		response := commandline.AskMultipleChoice(
			"Choose a merge request template to be applied", append(temp, "None"),
		)
		if response != "None" {
			return response
		}
	}

	return ""
}

func (mr *GitLabMergeRequest) mrTemplateOptions() []string {
	identifiers := map[string]string{
		"templateDir":       ".gitlab",
		"nestedDirName":     "merge_request_templates",
		"nonNestedFileName": "merge_request_template",
	}

	nestedTemplates, _ := filepath.Glob(
		filepath.Join(mr.GitRootDir, identifiers["templateDir"], identifiers["nestedDirName"], "*.md"),
	)
	nonNestedTemplates, _ := filepath.Glob(
		filepath.Join(mr.GitRootDir, identifiers["templateDir"], identifiers["nonNestedFileName"]+".md"),
	)
	rootTemplates, _ := filepath.Glob(filepath.Join(mr.GitRootDir, identifiers["nonNestedFileName"]+".md"))

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

func (mr *GitLabMergeRequest) gitlab() *gitlab.GitLab {
	return gitlab.NewGitLab(mr.Debug)
}
