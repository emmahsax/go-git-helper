package gitlabMergeRequest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/gitlab"
	"github.com/emmahsax/go-git-helper/internal/utils"
	go_gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

type GitLabMergeRequest struct {
	BaseBranch      string
	Debug           bool
	Draft           string
	GitRootDir      string
	LocalBranch     string
	InteractiveMode bool
	LocalProject    string
	NewMrTitle      string
}

func NewGitLabMergeRequest(options map[string]string, debug, interactiveMode bool) *GitLabMergeRequest {
	return &GitLabMergeRequest{
		BaseBranch:      options["baseBranch"],
		Debug:           debug,
		Draft:           options["draft"],
		GitRootDir:      options["gitRootDir"],
		LocalBranch:     options["localBranch"],
		InteractiveMode: interactiveMode,
		LocalProject:    options["localProject"],
		NewMrTitle:      options["newMrTitle"],
	}
}

func (mr *GitLabMergeRequest) Create() {
	t := mr.determineTitle()

	options := go_gitlab.CreateMergeRequestOptions{
		Description:        go_gitlab.Ptr(mr.newMrBody()),
		RemoveSourceBranch: go_gitlab.Ptr(true),
		SourceBranch:       go_gitlab.Ptr(mr.LocalBranch),
		Squash:             go_gitlab.Ptr(true),
		TargetBranch:       go_gitlab.Ptr(mr.BaseBranch),
		Title:              go_gitlab.Ptr(t),
	}

	fmt.Println("Creating merge request:", t)
	resp, err := mr.gitlab().CreateMergeRequest(mr.LocalProject, &options)
	if err != nil {
		customErr := errors.New("could not create merge request: " + err.Error())
		utils.HandleError(customErr, mr.Debug, nil)
		return
	}

	fmt.Println("Merge request successfully created:", resp.WebURL)
}

func (mr *GitLabMergeRequest) determineTitle() string {
	var t string
	if d, _ := strconv.ParseBool(mr.Draft); d {
		t = "Draft: " + mr.NewMrTitle
	} else {
		t = mr.NewMrTitle
	}

	return t
}

func (mr *GitLabMergeRequest) newMrBody() string {
	templateName := mr.templateNameToApply()
	if templateName != "" {
		content, err := os.ReadFile(templateName)
		if err != nil {
			utils.HandleError(err, mr.Debug, nil)
			return ""
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
		var applySingleTemplate bool

		if mr.InteractiveMode {
			applySingleTemplate = commandline.AskYesNoQuestion(
				fmt.Sprintf("Apply the merge request template from %s?", strings.TrimPrefix(mr.mrTemplateOptions()[0], mr.GitRootDir+"/")),
			)
		} else {
			applySingleTemplate = true
		}

		if applySingleTemplate {
			return mr.mrTemplateOptions()[0]
		}
	} else {
		temp := []string{}
		for _, str := range mr.mrTemplateOptions() {
			modifiedStr := strings.TrimPrefix(str, mr.GitRootDir+"/")
			temp = append(temp, modifiedStr)
		}

		if mr.InteractiveMode {
			response := commandline.AskMultipleChoice("Choose a merge request template to be applied", append(temp, "None"))

			if response != "None" {
				return response
			}
		} else {
			return mr.mrTemplateOptions()[0]
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
