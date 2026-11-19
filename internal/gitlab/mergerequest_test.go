package gitlab

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
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
			true,
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
		GitRootDir:      tempDir,
		InteractiveMode: true,
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
		return ""
	}

	expected := ""
	actual := mr.newMrBody()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

// Regression: a ticket reference in the title must be linked even with no
// template at all -- the feature doesn't require a template to exist.
func Test_newMrBody_NoTemplate_WithTicketReference(t *testing.T) {
	mr := NewGitLabMergeRequest(
		map[string]string{
			"gitRootDir":              t.TempDir(),
			"newMrTitle":              "proj-123 something",
			"ticketBaseURLCandidates": "https://example.com/browse",
		},
		false,
		false,
	)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\n"
	actual := mr.newMrBody()

	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

// Regression: with no ticketBaseURLCandidates configured, the feature does nothing
// at all, even though the title has a ticket reference.
// Regression: with no ticketBaseURLCandidates configured, the bare reference is
// still inserted -- a base URL only upgrades it into a real link, it
// never gates whether the reference appears at all (so platforms with
// their own autolinking still get it).
func Test_newMrBody_NoTemplate_NoBaseURL_InsertsBareReference(t *testing.T) {
	mr := NewGitLabMergeRequest(
		map[string]string{
			"gitRootDir": t.TempDir(),
			"newMrTitle": "proj-123 something",
		},
		false,
		false,
	)

	expected := "### [proj-123]\n\n"
	actual := mr.newMrBody()

	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
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
		GitRootDir:      tempDir,
		InteractiveMode: true,
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
		GitRootDir:      "/path/to/repo",
		InteractiveMode: true,
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
