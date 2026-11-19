package forgejo

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/commandline"
)

// captureStdout redirects os.Stdout for the duration of fn and returns
// whatever it wrote, restoring the original afterward.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	original := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = original })

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}

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

	pr := NewForgejoPullRequest(
		map[string]string{
			"baseBranch":  "main",
			"localRepo":   "repo",
			"localBranch": "feature",
			"gitRootDir":  tempDir,
			"newPrTitle":  "hello world",
			"draft":       "false",
		},
		false,
		true,
	)

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

func Test_relativeToGitRoot(t *testing.T) {
	pr := &ForgejoPullRequest{GitRootDir: filepath.Join("repo", "root")}

	path := filepath.Join(pr.GitRootDir, ".forgejo", "pull_request_template.md")
	expected := filepath.Join(".forgejo", "pull_request_template.md")
	actual := pr.relativeToGitRoot(path)

	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

// Regression: a path that isn't actually under GitRootDir must fall back to
// itself rather than produce a nonsensical "../" relative path.
func Test_relativeToGitRoot_NotUnderGitRoot(t *testing.T) {
	pr := &ForgejoPullRequest{GitRootDir: filepath.Join("repo", "root")}

	path := filepath.Join("somewhere", "else", "pull_request_template.md")
	actual := pr.relativeToGitRoot(path)

	if actual != path {
		t.Errorf("expected the path unchanged as a fallback, got '%s'", actual)
	}
}

// Regression: when the server doesn't honor the "WIP:" convention (e.g. a
// self-hosted instance reconfigured WORK_IN_PROGRESS_PREFIXES), the user
// must be warned that a requested draft wasn't actually created as one.
func Test_warnIfDraftMismatched_WarnsWhenServerIgnoredDraftRequest(t *testing.T) {
	pr := &ForgejoPullRequest{Draft: "true"}

	output := captureStdout(t, func() {
		pr.warnIfDraftMismatched(false)
	})

	if !strings.Contains(output, "Warning") {
		t.Errorf("expected a warning to be printed, got %q", output)
	}
}

func Test_warnIfDraftMismatched_NoWarningWhenMatched(t *testing.T) {
	pr := &ForgejoPullRequest{Draft: "true"}

	output := captureStdout(t, func() {
		pr.warnIfDraftMismatched(true)
	})

	if output != "" {
		t.Errorf("expected no warning, got %q", output)
	}
}

func Test_warnIfDraftMismatched_NoWarningWhenDraftNotRequested(t *testing.T) {
	pr := &ForgejoPullRequest{Draft: "false"}

	output := captureStdout(t, func() {
		pr.warnIfDraftMismatched(false)
	})

	if output != "" {
		t.Errorf("expected no warning, got %q", output)
	}
}
