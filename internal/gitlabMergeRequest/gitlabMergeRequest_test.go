package gitlabMergeRequest

import (
	"testing"
)

func TestNewMrBody(t *testing.T) {
	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newMrTitle"] = "Example MR Title"
	options["localBranch"] = "feature-branch"
	options["localProject"] = "test-project"
	mr := NewGitLabMergeRequest(options, true)
	body := mr.newMrBody()

	if body != "" {
		t.Fatalf(`Body was non-empty: %s`, body)
	}
}

func TestTemplateNameToApply(t *testing.T) {
	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newMrTitle"] = "Example MR Title"
	options["localBranch"] = "feature-branch"
	options["localProject"] = "test-project"
	mr := NewGitLabMergeRequest(options, true)
	template := mr.templateNameToApply()

	if template != "" {
		t.Fatalf(`Template was non-empty: %s`, template)
	}
}

func TestMrTemplateOptions(t *testing.T) {
	options := make(map[string]string)
	options["baseBranch"] = "main"
	options["newMrTitle"] = "Example MR Title"
	options["localBranch"] = "feature-branch"
	options["localProject"] = "test-project"
	mr := NewGitLabMergeRequest(options, true)
	tempOptions := mr.mrTemplateOptions()

	if len(tempOptions) != 0 {
		t.Fatalf(`PR options should be 0 when there are no templates: %v`, tempOptions)
	}
}
