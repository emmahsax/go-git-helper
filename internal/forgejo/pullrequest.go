package forgejo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v4"
	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/utils"
)

type ForgejoPullRequest struct {
	BaseBranch      string
	Debug           bool
	Draft           string
	GitRootDir      string
	Host            string
	InteractiveMode bool
	LocalBranch     string
	LocalRepo       string
	NewPrTitle      string
	Username        string
}

func NewForgejoPullRequest(options map[string]string, debug, interactiveMode bool) *ForgejoPullRequest {
	return &ForgejoPullRequest{
		BaseBranch:      options["baseBranch"],
		Debug:           debug,
		Draft:           options["draft"],
		GitRootDir:      options["gitRootDir"],
		Host:            options["host"],
		InteractiveMode: interactiveMode,
		LocalBranch:     options["localBranch"],
		LocalRepo:       options["localRepo"],
		NewPrTitle:      options["newPrTitle"],
		Username:        options["username"],
	}
}

func (pr *ForgejoPullRequest) Create() {
	options := forgejo.CreatePullRequestOption{
		Base:  pr.BaseBranch,
		Body:  pr.newPrBody(),
		Head:  pr.LocalBranch,
		Title: pr.determineTitle(),
	}

	repo := strings.Split(pr.LocalRepo, "/")

	fmt.Println("Creating pull request:", options.Title)
	resp, err := NewForgejo(pr.Host, pr.Username, pr.Debug).CreatePullRequest(repo[0], repo[1], options)
	if err != nil {
		customErr := errors.New("could not create pull request: " + err.Error())
		utils.HandleError(customErr, pr.Debug, nil)
		return
	}

	fmt.Println("Pull request successfully created:", resp.HTMLURL)
}

// determineTitle prefixes the title with "WIP:" for a draft pull request.
// Forgejo (like Gitea before it) has no dedicated draft field on
// CreatePullRequestOption -- a PR is only treated as a draft/work-in-progress
// when its title starts with one of a few configurable prefixes, of which
// "WIP:" is the default.
func (pr *ForgejoPullRequest) determineTitle() string {
	if d, _ := strconv.ParseBool(pr.Draft); d {
		return "WIP: " + pr.NewPrTitle
	}

	return pr.NewPrTitle
}

func (pr *ForgejoPullRequest) newPrBody() string {
	templateName := pr.templateNameToApply()
	if templateName != "" {
		content, err := os.ReadFile(templateName)
		if err != nil {
			utils.HandleError(err, pr.Debug, nil)
			return ""
		}

		return string(content)
	}

	return ""
}

// templateNameToApply returns the repo's pull request template, after
// confirming with the user (when interactive) that it should be applied.
// Forgejo allows at most one pull request template per repo, so -- unlike
// GitHub/GitLab's templateNameToApply -- there's never a choice to make
// between several, only whether to use the one that exists.
func (pr *ForgejoPullRequest) templateNameToApply() string {
	options := pr.prTemplateOptions()
	if len(options) == 0 {
		return ""
	}

	template := options[0]

	var applyTemplate bool
	if pr.InteractiveMode {
		applyTemplate = commandline.AskYesNoQuestion(
			fmt.Sprintf("Apply the pull request template from %s?", strings.TrimPrefix(template, pr.GitRootDir+"/")),
		)
	} else {
		applyTemplate = true
	}

	if applyTemplate {
		return template
	}

	return ""
}

// prTemplateOptions returns every pull request template Forgejo itself would
// recognize for this repo. Forgejo allows at most one: it checks each of
// these directories, in order, for one of a few fixed filenames, and uses the
// first match -- there's no nested multi-template directory like GitHub's
// PULL_REQUEST_TEMPLATE/, and no root-level file.
func (pr *ForgejoPullRequest) prTemplateOptions() []string {
	dirs := []string{".forgejo", ".gitea", ".github", "docs"}
	names := []string{
		"PULL_REQUEST_TEMPLATE.md", "PULL_REQUEST_TEMPLATE.yaml", "PULL_REQUEST_TEMPLATE.yml",
		"pull_request_template.md", "pull_request_template.yaml", "pull_request_template.yml",
	}

	var found []string
	for _, dir := range dirs {
		for _, name := range names {
			path := filepath.Join(pr.GitRootDir, dir, name)
			if _, err := os.Stat(path); err == nil {
				found = append(found, path)
			}
		}
	}

	return found
}
