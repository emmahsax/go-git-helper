package checkoutDefault

import (
	"testing"
)

func TestGetDefaultBranch(t *testing.T) {
	branch := getDefaultBranch()

	if branch != "main" {
		t.Fatalf(`Branch should match %s, was %s`, "main", branch)
	}
}
