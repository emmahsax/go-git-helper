package forgejo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
)

func Test_determineTitle_Draft(t *testing.T) {
	pr := &ForgejoPullRequest{
		Draft:      "true",
		NewPrTitle: "hello world",
	}

	expected := "WIP: hello world"
	actual := pr.determineTitle()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_determineTitle_NotDraft(t *testing.T) {
	pr := &ForgejoPullRequest{
		Draft:      "false",
		NewPrTitle: "hello world",
	}

	expected := "hello world"
	actual := pr.determineTitle()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_newPrBody(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tempDir, ".forgejo"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".forgejo", "pull_request_template.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	pr := &ForgejoPullRequest{
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

	expected := "content"
	actual := pr.newPrBody()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_newPrBody_NoTemplate(t *testing.T) {
	pr := &ForgejoPullRequest{
		GitRootDir:      t.TempDir(),
		InteractiveMode: true,
	}

	expected := ""
	actual := pr.newPrBody()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_templateNameToApply_Declines(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tempDir, ".forgejo"), 0755); err != nil {
		t.Fatal(err)
	}
	templatePath := filepath.Join(tempDir, ".forgejo", "pull_request_template.md")
	if err := os.WriteFile(templatePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	pr := &ForgejoPullRequest{
		GitRootDir:      tempDir,
		InteractiveMode: true,
	}

	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() {
		commandline.AskYesNoQuestion = originalAskYesNoQuestion
	})
	commandline.AskYesNoQuestion = func(question string) bool {
		return false
	}

	expected := ""
	actual := pr.templateNameToApply()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_templateNameToApply_NonInteractiveAlwaysApplies(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tempDir, ".gitea"), 0755); err != nil {
		t.Fatal(err)
	}
	templatePath := filepath.Join(tempDir, ".gitea", "PULL_REQUEST_TEMPLATE.yml")
	if err := os.WriteFile(templatePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	pr := &ForgejoPullRequest{
		GitRootDir:      tempDir,
		InteractiveMode: false,
	}

	expected := templatePath
	actual := pr.templateNameToApply()

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func Test_prTemplateOptions(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tempDir, ".forgejo"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "docs"), 0755); err != nil {
		t.Fatal(err)
	}

	files := []string{
		filepath.Join(".forgejo", "pull_request_template.md"),
		filepath.Join("docs", "PULL_REQUEST_TEMPLATE.yaml"),
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Not a recognized location/filename -- must not be picked up.
	if err := os.WriteFile(filepath.Join(tempDir, "pull_request_template.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	pr := &ForgejoPullRequest{GitRootDir: tempDir}

	// Checked as a subset, not exact equality: on a case-insensitive
	// filesystem (macOS, Windows), stat-ing both "pull_request_template.md"
	// and "PULL_REQUEST_TEMPLATE.md" can each report the one file written as
	// existing, adding a harmless duplicate to the result (prTemplateOptions
	// only ever uses the first entry, so a duplicate changes nothing
	// user-visible). What matters here is that both real files were found,
	// and the unrecognized root-level file wasn't.
	want := []string{
		filepath.Join(tempDir, ".forgejo", "pull_request_template.md"),
		filepath.Join(tempDir, "docs", "PULL_REQUEST_TEMPLATE.yaml"),
	}
	actual := pr.prTemplateOptions()

	actualSet := make(map[string]bool, len(actual))
	for _, path := range actual {
		actualSet[path] = true
	}
	for _, path := range want {
		if !actualSet[path] {
			t.Errorf("expected %s to be found, got %v", path, actual)
		}
	}
	if actualSet[filepath.Join(tempDir, "pull_request_template.md")] {
		t.Errorf("root-level pull_request_template.md is not a recognized Forgejo template location, got %v", actual)
	}
}

func Test_prTemplateOptions_None(t *testing.T) {
	pr := &ForgejoPullRequest{GitRootDir: t.TempDir()}

	actual := pr.prTemplateOptions()

	if len(actual) != 0 {
		t.Errorf("expected no templates, got %v", actual)
	}
}
