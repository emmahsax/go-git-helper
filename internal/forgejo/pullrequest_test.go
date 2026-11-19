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

func Test_splitOwnerRepo(t *testing.T) {
	tests := []struct {
		name          string
		localRepo     string
		expectedOwner string
		expectedRepo  string
		expectedOK    bool
	}{
		{"owner/repo", "emmahsax/go-git-helper", "emmahsax", "go-git-helper", true},
		{
			// Regression: an scp-like remote with no owner segment in its
			// path (e.g. "git@forgejo.example.com:reponame.git") parses to
			// a path with no "/" at all -- this must not panic.
			name:       "no slash",
			localRepo:  "reponame",
			expectedOK: false,
		},
		{"too many slashes", "a/b/c", "", "", false},
		{"empty", "", "", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner, repoName, ok := splitOwnerRepo(test.localRepo)

			if ok != test.expectedOK {
				t.Fatalf("expected ok=%v, got %v", test.expectedOK, ok)
			}
			if ok && (owner != test.expectedOwner || repoName != test.expectedRepo) {
				t.Errorf("expected (%q, %q), got (%q, %q)", test.expectedOwner, test.expectedRepo, owner, repoName)
			}
		})
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

// Regression: a ticket reference in the title must be linked even with no
// template at all -- the feature doesn't require a template to exist.
func Test_newPrBody_NoTemplate_WithTicketReference(t *testing.T) {
	pr := NewForgejoPullRequest(
		map[string]string{
			"gitRootDir":    t.TempDir(),
			"newPrTitle":    "proj-123 something",
			"ticketBaseURL": "https://example.com/browse",
		},
		false,
		false,
	)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\n"
	actual := pr.newPrBody()

	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

// Regression: with no ticketBaseURL configured, the prompt still appears
// and the bare reference still gets inserted -- a base URL only upgrades
// it into a real link, it never gates whether the reference appears at all
// (so platforms with their own autolinking, e.g. GitHub, still get it).
func Test_newPrBody_NoTemplate_NoBaseURL_InsertsBareReference(t *testing.T) {
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })

	asked := false
	commandline.AskYesNoQuestion = func(question string) bool {
		asked = true
		return true
	}

	pr := NewForgejoPullRequest(
		map[string]string{
			"gitRootDir": t.TempDir(),
			"newPrTitle": "proj-123 something",
		},
		false,
		true,
	)

	expected := "### [proj-123]\n\n"
	actual := pr.newPrBody()

	if !asked {
		t.Error("expected the prompt even with no ticketBaseURL configured")
	}
	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

// Regression: interactively, the ticket-link prompt must still appear with
// no template configured -- it's not gated on a template existing, only on
// a base URL being configured.
func Test_newPrBody_NoTemplate_InteractivelyAsksAboutTicketReference(t *testing.T) {
	originalAskYesNoQuestion := commandline.AskYesNoQuestion
	t.Cleanup(func() { commandline.AskYesNoQuestion = originalAskYesNoQuestion })

	var asked bool
	commandline.AskYesNoQuestion = func(question string) bool {
		asked = true
		return true
	}

	pr := NewForgejoPullRequest(
		map[string]string{
			"gitRootDir":    t.TempDir(),
			"newPrTitle":    "proj-123 something",
			"ticketBaseURL": "https://example.com/browse",
		},
		false,
		true,
	)

	expected := "### [proj-123](https://example.com/browse/proj-123)\n\n"
	actual := pr.newPrBody()

	if !asked {
		t.Error("expected the ticket-link prompt to be asked even with no template")
	}
	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
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

func Test_templateName_FindsFirstMatchInDirOrder(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tempDir, ".forgejo"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "docs"), 0755); err != nil {
		t.Fatal(err)
	}

	// Written with the exact casing templateName checks first
	// ("PULL_REQUEST_TEMPLATE.md" before any lowercase variant) so the
	// expected path below is unambiguous even on a case-insensitive
	// filesystem (macOS, Windows), where a lowercase-named file would also
	// satisfy an os.Stat of the uppercase variant.
	files := []string{
		filepath.Join(".forgejo", "PULL_REQUEST_TEMPLATE.md"),
		filepath.Join("docs", "PULL_REQUEST_TEMPLATE.yaml"),
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Not a recognized location/filename -- must not be picked up, even
	// though it would be checked first if it were.
	if err := os.WriteFile(filepath.Join(tempDir, "pull_request_template.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	pr := &ForgejoPullRequest{GitRootDir: tempDir}

	// .forgejo is checked before docs, so that's the one that must win even
	// though both files exist.
	expected := filepath.Join(tempDir, ".forgejo", "PULL_REQUEST_TEMPLATE.md")
	actual := pr.templateName()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func Test_templateName_None(t *testing.T) {
	pr := &ForgejoPullRequest{GitRootDir: t.TempDir()}

	if actual := pr.templateName(); actual != "" {
		t.Errorf("expected no template, got %s", actual)
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
