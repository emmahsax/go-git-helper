package githubPullRequest

import (
	"os"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/git"
)

func TestNewPrBody(t *testing.T) {
	g := git.NewGit(true)
	rootDir := g.GetGitRootDir()

	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newPrTitle"] = "Example PR Title"
	options["gitRootDir"] = rootDir
	options["localBranch"] = "feature-branch"
	options["localRepo"] = "test-repo"
	pr := NewGitHubPullRequest(options, true)
	body := pr.newPrBody()

	realTemplate := rootDir + "/.github/pull_request_template.md"
	content, _ := os.ReadFile(realTemplate)
	realBody := string(content)

	if body != realBody {
		t.Fatalf(`Body was not the real repo template: %s`, body)
	}
}

func TestTemplateNameToApply(t *testing.T) {
	g := git.NewGit(true)
	rootDir := g.GetGitRootDir()

	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newPrTitle"] = "Example PR Title"
	options["gitRootDir"] = rootDir
	options["localBranch"] = "feature-branch"
	options["localRepo"] = "test-repo"
	pr := NewGitHubPullRequest(options, true)
	template := pr.templateNameToApply()
	realTemplate := rootDir + "/.github/pull_request_template.md"

	if template != realTemplate {
		t.Fatalf(`Template was not the real repo template: %s`, template)
	}
}

func TestPrTemplateOptions(t *testing.T) {
	g := git.NewGit(true)
	rootDir := g.GetGitRootDir()

	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newPrTitle"] = "Example PR Title"
	options["gitRootDir"] = rootDir
	options["localBranch"] = "feature-branch"
	options["localRepo"] = "test-repo"
	pr := NewGitHubPullRequest(options, true)
	tempOptions := pr.prTemplateOptions()

	if len(tempOptions) != 1 {
		t.Fatalf(`PR options should be 1 when there is a single real repo template: %v`, tempOptions)
	}
}
