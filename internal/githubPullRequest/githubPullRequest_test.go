package githubPullRequest

import (
	"testing"
)

func TestNewPrBody(t *testing.T) {
	body := newPrBody()

	if body != "" {
		t.Fatalf(`Body was non-empty: %s`, body)
	}
}

func TestTemplateNameToApply(t *testing.T) {
	template := templateNameToApply()

	if template != "" {
		t.Fatalf(`Template was non-empty: %s`, template)
	}
}

func TestPrTemplateOptions(t *testing.T) {
	options := prTemplateOptions()

	if len(options) != 0 {
		t.Fatalf(`PR options should be 0 when there are no templates: %v`, options)
	}
}
