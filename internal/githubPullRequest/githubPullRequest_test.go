package githubPullRequest

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
)

func Test_newPrBody(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "github")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	err = os.MkdirAll(filepath.Join(tempDir, ".github", "PULL_REQUEST_TEMPLATE"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	files := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE/template1.md",
		".github/PULL_REQUEST_TEMPLATE/template2.md",
		"pull_request_template.md",
	}
	for _, file := range files {
		err = os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	mr := NewGitHubPullRequest(
		map[string]string{
			"baseBranch":  "master",
			"localRepo":   "repo",
			"localBranch": "feature",
			"gitRootDir":  tempDir,
			"newPrTitle":  "hello world",
			"draft":       "false",
		},
		false,
	)

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() {
		commandline.AskMultipleChoice = originalAskMultipleChoice
	})
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return ""
	}

	expected := ""
	actual := mr.newPrBody()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_templateNameToApply(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "github")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	err = os.MkdirAll(filepath.Join(tempDir, ".github", "PULL_REQUEST_TEMPLATE"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	files := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE/template1.md",
		".github/PULL_REQUEST_TEMPLATE/template2.md",
		"pull_request_template.md",
	}
	for _, file := range files {
		err = os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	mr := &GitHubPullRequest{
		GitRootDir: tempDir,
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() {
		commandline.AskMultipleChoice = originalAskMultipleChoice
	})
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return "/path/to/repo/.github/pull_request_template.md"
	}

	expected := "/path/to/repo/.github/pull_request_template.md"
	actual := mr.templateNameToApply()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_determineTemplate(t *testing.T) {
	mr := &GitHubPullRequest{
		GitRootDir: "/path/to/repo",
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return true
	}

	originalAskMultipleChoice := commandline.AskMultipleChoice
	t.Cleanup(func() {
		commandline.AskMultipleChoice = originalAskMultipleChoice
	})
	commandline.AskMultipleChoice = func(question string, choices []string) string {
		return "/path/to/repo/.github/pull_request_template.md"
	}

	expected := "/path/to/repo/.github/pull_request_template.md"
	actual := mr.determineTemplate()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_prTemplateOptions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "github")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	err = os.MkdirAll(filepath.Join(tempDir, ".github", "PULL_REQUEST_TEMPLATE"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	files := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE/template1.md",
		".github/PULL_REQUEST_TEMPLATE/template2.md",
		"pull_request_template.md",
	}
	for _, file := range files {
		err = os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	mr := &GitHubPullRequest{
		GitRootDir: tempDir,
	}

	expected := []string{
		filepath.Join(tempDir, ".github/pull_request_template.md"),
		filepath.Join(tempDir, ".github/PULL_REQUEST_TEMPLATE/template1.md"),
		filepath.Join(tempDir, ".github/PULL_REQUEST_TEMPLATE/template2.md"),
		filepath.Join(tempDir, "pull_request_template.md"),
	}
	actual := mr.prTemplateOptions()

	if len(expected) != len(actual) {
		t.Fatalf("Expected and actual slices have different lengths: expected %v, got %v", len(expected), len(actual))
	}

	sort.Strings(expected)
	sort.Strings(actual)

	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf("Expected and actual slices do not match: expected %v, got %v", expected, actual)
		}
	}
}
