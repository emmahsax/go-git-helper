package git

import (
	"testing"
)

func TestDefaultBranch(t *testing.T) {
	branch := DefaultBranch()

	if branch != "main" {
		t.Fatalf(`Branch should match %s, was %s`, "main", branch)
	}
}
