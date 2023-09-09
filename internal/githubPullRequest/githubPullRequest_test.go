package githubPullRequest

import (
	"testing"
)

func TestNewPrBody(t *testing.T) {
	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newPrTitle"] = "Example PR Title"
	options["localBranch"] = "feature-branch"
	options["localRepo"] = "test-repo"
	pr := NewGitHubPullRequest(options, true)
	body := pr.newPrBody()

	if body != "" {
		t.Fatalf(`Body was non-empty: %s`, body)
	}
}

func TestTemplateNameToApply(t *testing.T) {
	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newPrTitle"] = "Example PR Title"
	options["localBranch"] = "feature-branch"
	options["localRepo"] = "test-repo"
	pr := NewGitHubPullRequest(options, true)
	template := pr.templateNameToApply()

	if template != "" {
		t.Fatalf(`Template was non-empty: %s`, template)
	}
}

func TestPrTemplateOptions(t *testing.T) {
	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newPrTitle"] = "Example PR Title"
	options["localBranch"] = "feature-branch"
	options["localRepo"] = "test-repo"
	pr := NewGitHubPullRequest(options, true)
	tempOptions := pr.prTemplateOptions()

	if len(tempOptions) != 0 {
		t.Fatalf(`PR options should be 0 when there are no templates: %v`, tempOptions)
	}
}
