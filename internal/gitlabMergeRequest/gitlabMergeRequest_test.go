package gitlabMergeRequest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/stretchr/testify/assert"
)

func Test_determineTitle(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		draft    string
		expected string
	}{
		{
			name:     "draft",
			title:    "hello world",
			draft:    "true",
			expected: "Draft: hello world",
		},
		{
			name:     "no draft",
			title:    "hello world",
			draft:    "false",
			expected: "hello world",
		},
	}

	for _, test := range tests {
		mr := NewGitLabMergeRequest(
			map[string]string{
				"baseBranch":   "master",
				"localProject": "project",
				"localBranch":  "feature",
				"gitRootDir":   "/path/to/repo",
				"newMrTitle":   test.title,
				"draft":        test.draft,
			},
			false,
		)
		actual := mr.determineTitle()

		if test.expected != actual {
			t.Errorf("expected '%s', got '%s'", test.expected, actual)
		}
	}
}

func Test_newMrBody(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitlab")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	err = os.MkdirAll(filepath.Join(tempDir, ".gitlab", "merge_request_templates"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	files := []string{
		".gitlab/merge_request_template.md",
		".gitlab/merge_request_templates/template1.md",
		".gitlab/merge_request_templates/template2.md",
		"merge_request_template.md",
	}
	for _, file := range files {
		err = os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	mr := &GitLabMergeRequest{
		GitRootDir: tempDir,
	}

	// Mock the AskYesNoQuestion function to always return true
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	// Mock the AskMultipleChoice function to always return "hello world"
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() {
		commandline.AskMultipleChoice = originalAskMultipleChoice
	})
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return ""
	}

	expected := ""
	actual := mr.newMrBody()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_templateNameToApply(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitlab")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	err = os.MkdirAll(filepath.Join(tempDir, ".gitlab", "merge_request_templates"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	files := []string{
		".gitlab/merge_request_template.md",
		".gitlab/merge_request_templates/template1.md",
		".gitlab/merge_request_templates/template2.md",
		"merge_request_template.md",
	}
	for _, file := range files {
		err = os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	mr := &GitLabMergeRequest{
		GitRootDir: tempDir,
	}

	// Mock the AskYesNoQuestion function to always return true
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	// Mock the AskMultipleChoice function to always return "hello world"
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() {
		commandline.AskMultipleChoice = originalAskMultipleChoice
	})
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return "/path/to/repo/.gitlab/merge_request_template.md"
	}

	expected := "/path/to/repo/.gitlab/merge_request_template.md"
	actual := mr.templateNameToApply()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_determineTemplate(t *testing.T) {
	mr := &GitLabMergeRequest{
		GitRootDir: "/path/to/repo",
	}

	// Mock the AskYesNoQuestion function to always return true
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	// Mock the AskMultipleChoice function to always return "hello world"
	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() {
		commandline.AskMultipleChoice = originalAskMultipleChoice
	})
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return "/path/to/repo/.gitlab/merge_request_template.md"
	}

	expected := "/path/to/repo/.gitlab/merge_request_template.md"
	actual := mr.determineTemplate()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_mrTemplateOptions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitlab")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	err = os.MkdirAll(filepath.Join(tempDir, ".gitlab", "merge_request_templates"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	files := []string{
		".gitlab/merge_request_template.md",
		".gitlab/merge_request_templates/template1.md",
		".gitlab/merge_request_templates/template2.md",
		"merge_request_template.md",
	}
	for _, file := range files {
		err = os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	mr := &GitLabMergeRequest{
		GitRootDir: tempDir,
	}

	expected := []string{
		filepath.Join(tempDir, ".gitlab/merge_request_template.md"),
		filepath.Join(tempDir, ".gitlab/merge_request_templates/template1.md"),
		filepath.Join(tempDir, ".gitlab/merge_request_templates/template2.md"),
		filepath.Join(tempDir, "merge_request_template.md"),
	}
	actual := mr.mrTemplateOptions()
	assert.ElementsMatch(t, expected, actual)
}
