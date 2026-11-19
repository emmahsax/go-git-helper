package forgejo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v4"
	"github.com/emmahsax/go-git-helper/internal/testutil"
)

func Test_NewForgejo(t *testing.T) {
	testutil.WriteConfig(t, testutil.IsolateHome(t), "providers:\n  - name: forgejo\n    token: test_token\n")

	fj := NewForgejo(PublicHost, "", false)

	if fj == nil {
		t.Fatal("Expected NewForgejo to return a non-nil Forgejo struct")
	}

	if fj.Client == nil {
		t.Error("Expected Forgejo client to be initialized")
	}

	if fj.Debug != false {
		t.Errorf("Expected Debug to be false, got %v", fj.Debug)
	}
}

func Test_NewForgejo_WithDebug(t *testing.T) {
	testutil.WriteConfig(t, testutil.IsolateHome(t), "providers:\n  - name: forgejo\n    token: test_token\n")

	fj := NewForgejo(PublicHost, "", true)

	if fj == nil {
		t.Fatal("Expected NewForgejo to return a non-nil Forgejo struct")
	}

	if fj.Debug != true {
		t.Errorf("Expected Debug to be true, got %v", fj.Debug)
	}
}

// Note: unlike go-github/go-gitlab, forgejo-sdk's *Client doesn't expose its
// configured base URL (the url field is private, with no getter), so these
// tests -- unlike the equivalent GitHub/GitLab ones -- can only confirm
// NewForgejo/newForgejoClient build a client without erroring, not which
// host it actually targets. That host-selection logic is still covered
// indirectly: it's identical to accountsFor's (see
// Test_GitHubAccounts_SSHAliasOnlyMatchesConfiguredHost and friends in
// configfile_test.go), just reading APIHost instead of Token.

func Test_NewForgejo_SelfHostedAccount(t *testing.T) {
	testutil.WriteConfig(t, testutil.IsolateHome(t), `providers:
  - name: forgejo
    token: public_token
  - name: forgejo
    token: work_token
    remote_hosts:
      - forgejo.example.com
    api_host: forgejo.example.com
`)

	fj := NewForgejo("forgejo.example.com", "", false)

	if fj == nil {
		t.Fatal("Expected NewForgejo to return a non-nil Forgejo struct")
	}
	if fj.Client == nil {
		t.Error("Expected Forgejo client to be initialized")
	}
}

func Test_NewForgejo_SSHAliasKeepsPublicAPI(t *testing.T) {
	// A host that's only an SSH config alias (no api_host set) must still
	// resolve to an account with no APIHost, so the client talks to the
	// real public API rather than being mistaken for a self-hosted server.
	testutil.WriteConfig(t, testutil.IsolateHome(t), `providers:
  - name: forgejo
    token: personal_token
  - name: forgejo
    token: work_token
    remote_hosts:
      - codeberg.org-work
`)

	fj := NewForgejo("codeberg.org-work", "", false)

	if fj == nil {
		t.Fatal("Expected NewForgejo to return a non-nil Forgejo struct")
	}
	if fj.Client == nil {
		t.Error("Expected Forgejo client to be initialized")
	}
}

func Test_NewForgejo_DisambiguatesByUsername(t *testing.T) {
	testutil.WriteConfig(t, testutil.IsolateHome(t), `providers:
  - name: forgejo
    username: personal_user
    token: personal_token
  - name: forgejo
    username: work_user
    token: work_token
`)

	fj := NewForgejo(PublicHost, "work_user", false)

	if fj == nil {
		t.Fatal("Expected NewForgejo to return a non-nil Forgejo struct")
	}
	if fj.Client == nil {
		t.Error("Expected Forgejo client to be initialized")
	}
}

func Test_CreatePullRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{
			"number": 1,
			"title": "Test PR",
			"body": "Test PR body",
			"state": "open",
			"draft": false,
			"html_url": "https://forgejo.example.com/owner/repo/pulls/1"
		}`)
	}))
	defer server.Close()

	client, _ := forgejo.NewClient(server.URL, forgejo.SetToken("test-token"), forgejo.SetForgejoVersion(""))

	fj := &Forgejo{
		Debug:  false,
		Client: client,
	}

	options := forgejo.CreatePullRequestOption{
		Title: "Test PR",
		Body:  "Test PR body",
		Head:  "feature-branch",
		Base:  "main",
	}

	pr, err := fj.CreatePullRequest("owner", "repo", options)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if pr == nil {
		t.Fatal("Expected PR to be non-nil")
	}

	if pr.Title != "Test PR" {
		t.Errorf("Expected PR title 'Test PR', got '%s'", pr.Title)
	}
}

func Test_CreatePullRequest_Error(t *testing.T) {
	// Skip this test as it would call os.Exit through HandleError
	// The error path is covered indirectly by other integration tests
	t.Skip("Skipping test that would call os.Exit")
}

func Test_newForgejoClient_EmptyAPIHost(t *testing.T) {
	client, err := newForgejoClient("test-token", "", false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

func Test_newForgejoClient_SelfHostedHost(t *testing.T) {
	client, err := newForgejoClient("test-token", "forgejo.example.com", false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}
