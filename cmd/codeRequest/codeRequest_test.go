package codeRequest

import (
	"testing"
)

type MockExecutor struct {
	Args    []string
	Command string
	Debug   bool
	Output  []byte
}

func (me *MockExecutor) Exec(execType string, command string, args ...string) ([]byte, error) {
	me.Command = command
	me.Args = args
	return me.Output, nil
}

func TestCheckAllLetters(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	cr := newCodeRequest(true, executor)
	resp := cr.checkAllLetters("iekslkjasd")

	if resp == false {
		t.Fatalf(`String %v should be all letters`, resp)
	}

	resp = cr.checkAllLetters("iekslkjasd321")

	if resp == true {
		t.Fatalf(`String %v should not be all letters`, resp)
	}
}

func TestCheckAllNumbers(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	cr := newCodeRequest(true, executor)
	resp := cr.checkAllNumbers("284161")

	if resp == false {
		t.Fatalf(`String %v should be all numbers`, resp)
	}

	resp = cr.checkAllNumbers("39812k3jiksd9z")

	if resp == true {
		t.Fatalf(`String %v should not be all numbers`, resp)
	}
}

func TestMatchesFullJiraPattern(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	cr := newCodeRequest(true, executor)
	resp := cr.matchesFullJiraPattern("jira-29142")

	if resp == false {
		t.Fatalf(`String %v should match Jira pattern (e.g. jira-123)`, resp)
	}

	resp = cr.matchesFullJiraPattern("jIra*3291")

	if resp == true {
		t.Fatalf(`String %v should not match Jira pattern (e.g. jira-123)`, resp)
	}
}

func TestTitleize(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	cr := newCodeRequest(true, executor)
	resp := cr.titleize("mysTrInG")

	if resp != "MysTrInG" {
		t.Fatalf(`String %v was not properly titleized`, resp)
	}
}

func Test_isGitHub(t *testing.T) {
	output := `origin  git@github.com:emmahsax/go-git-helper.git (fetch)
origin  git@github.com:emmahsax/go-git-helper.git (push)`
	executor := &MockExecutor{
		Debug:  true,
		Output: []byte(output),
	}
	cr := newCodeRequest(true, executor)
	resp := cr.isGitHub()

	if resp != true {
		t.Fatalf(`Project should be GitHub, but was %v`, resp)
	}
}

func Test_isGitLab(t *testing.T) {
	output := `origin  git@gitlab.com:emmahsax/go-git-helper.git (fetch)
origin  git@gitlab.com:emmahsax/go-git-helper.git (push)`
	executor := &MockExecutor{
		Debug:  true,
		Output: []byte(output),
	}
	cr := newCodeRequest(true, executor)
	resp := cr.isGitLab()

	if resp != true {
		t.Fatalf(`Project should not be GitLab, but was %v`, resp)
	}
}

func TestContainsSubstring(t *testing.T) {
	executor := &MockExecutor{Debug: true}
	cr := newCodeRequest(true, executor)
	strs := []string{"string1", "string3", "string18"}
	resp := cr.containsSubstring(strs, "string3")

	if resp == false {
		t.Fatalf(`String %v should be present in %v`, "string3", strs)
	}

	resp = cr.containsSubstring(strs, "string2")

	if resp == true {
		t.Fatalf(`String %v should not be present in %v`, "string2", strs)
	}
}
