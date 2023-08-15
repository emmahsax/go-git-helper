package gitlabMergeRequest

import (
	"testing"
)

func TestNewMrBody(t *testing.T) {
	body := newMrBody()

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

func TestMrTemplateOptions(t *testing.T) {
	options := mrTemplateOptions()

	if len(options) != 0 {
		t.Fatalf(`PR options should be 0 when there are no templates: %v`, options)
	}
}
