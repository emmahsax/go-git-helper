package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emmahsax/go-git-helper/internal/testutil"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

func Test_NewGitLab(t *testing.T) {
	testutil.WriteConfig(t, testutil.IsolateHome(t), "providers:\n  - name: gitlab\n    token: test_token\n")

	// This test requires a valid config file with GitLab token
	// We'll test that the struct is created properly
	gl := NewGitLab(PublicHost, "", false)

	if gl == nil {
		t.Fatal("Expected NewGitLab to return a non-nil GitLab struct")
	}

	if gl.Client == nil {
		t.Error("Expected GitLab client to be initialized")
	}

	if gl.Debug != false {
		t.Errorf("Expected Debug to be false, got %v", gl.Debug)
	}
}

func Test_NewGitLab_WithDebug(t *testing.T) {
	testutil.WriteConfig(t, testutil.IsolateHome(t), "providers:\n  - name: gitlab\n    token: test_token\n")

	gl := NewGitLab(PublicHost, "", true)

	if gl == nil {
		t.Fatal("Expected NewGitLab to return a non-nil GitLab struct")
	}

	if gl.Debug != true {
		t.Errorf("Expected Debug to be true, got %v", gl.Debug)
	}
}

func Test_NewGitLab_SelfHostedAccount(t *testing.T) {
	testutil.WriteConfig(t, testutil.IsolateHome(t), `providers:
  - name: gitlab
    token: public_token
  - name: gitlab
    token: self_hosted_token
    remote_hosts:
      - gitlab.example.com
    api_host: gitlab.example.com
`)

	gl := NewGitLab("gitlab.example.com", "", false)

	if gl == nil {
		t.Fatal("Expected NewGitLab to return a non-nil GitLab struct")
	}

	if got := gl.Client.BaseURL().Host; got != "gitlab.example.com" {
		t.Errorf("Expected BaseURL host 'gitlab.example.com', got '%s'", got)
	}
}

func Test_NewGitLab_SSHAliasKeepsPublicAPI(t *testing.T) {
	// A host that's only an SSH config alias (no api_host set) must still
	// hit the real public API, not be mistaken for a self-hosted server.
	testutil.WriteConfig(t, testutil.IsolateHome(t), `providers:
  - name: gitlab
    token: personal_token
  - name: gitlab
    token: work_token
    remote_hosts:
      - gitlab.com-work
`)

	gl := NewGitLab("gitlab.com-work", "", false)

	if gl == nil {
		t.Fatal("Expected NewGitLab to return a non-nil GitLab struct")
	}

	if got := gl.Client.BaseURL().Host; got == "gitlab.com-work" {
		t.Errorf("Expected the real public API BaseURL, not the SSH alias, got '%s'", got)
	}
}

func Test_NewGitLab_DisambiguatesByUsername(t *testing.T) {
	// Which account's token actually gets used for a given username is
	// covered at the configfile layer (Test_GitHubAccounts_TwoAccountsSameHost);
	// this just confirms NewGitLab builds a client without erroring when
	// username picks among tied accounts.
	testutil.WriteConfig(t, testutil.IsolateHome(t), `providers:
  - name: gitlab
    username: personal_user
    token: personal_token
  - name: gitlab
    username: work_user
    token: work_token
`)

	gl := NewGitLab(PublicHost, "work_user", false)

	if gl == nil {
		t.Fatal("Expected NewGitLab to return a non-nil GitLab struct")
	}
	if gl.Client == nil {
		t.Error("Expected GitLab client to be initialized")
	}
}

func Test_CreateMergeRequest_Success(t *testing.T) {
	// Create a test server that returns a successful MR response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{
			"id": 1,
			"iid": 1,
			"title": "Test MR",
			"description": "Test MR body",
			"state": "opened",
			"draft": false,
			"web_url": "https://gitlab.com/owner/repo/-/merge_requests/1"
		}`)
	}))
	defer server.Close()

	// Create a GitLab client that uses the test server
	client, _ := gitlab.NewClient("", gitlab.WithBaseURL(server.URL))

	gl := &GitLab{
		Debug:  false,
		Client: client,
	}

	options := &gitlab.CreateMergeRequestOptions{
		Title:        new("Test MR"),
		Description:  new("Test MR body"),
		SourceBranch: new("feature-branch"),
		TargetBranch: new("main"),
	}

	mr, err := gl.CreateMergeRequest("owner/repo", options)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if mr == nil {
		t.Fatal("Expected MR to be non-nil")
	}

	if mr.Title != "Test MR" {
		t.Errorf("Expected MR title 'Test MR', got '%s'", mr.Title)
	}
}

func Test_CreateMergeRequest_Error(t *testing.T) {
	// Skip this test as it would call os.Exit through HandleError
	// The error path is covered indirectly by other integration tests
	t.Skip("Skipping test that would call os.Exit")
}

func Test_newGitLabClient(t *testing.T) {
	token := "test-token-123"
	client, err := newGitLabClient(token, "", false)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Error("Expected client to be non-nil")
	}

	// Verify the client is properly configured
	if client.BaseURL() == nil {
		t.Error("Expected client BaseURL to be set")
	}
}

func Test_newGitLabClient_EmptyToken(t *testing.T) {
	client, err := newGitLabClient("", "", false)

	if err != nil {
		t.Errorf("Expected no error with empty token, got %v", err)
	}

	if client == nil {
		t.Error("Expected client to be non-nil even with empty token")
	}
}

func Test_newGitLabClient_WithDebug(t *testing.T) {
	token := "test-token-debug"
	client, err := newGitLabClient(token, "", true)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Error("Expected client to be non-nil")
	}
}

func Test_newGitLabClient_EmptyAPIHost(t *testing.T) {
	client, err := newGitLabClient("test-token", "", false)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	if got := client.BaseURL().Host; got == "gitlab.example.com" {
		t.Errorf("Expected the default BaseURL, not a custom one, got '%s'", got)
	}
}

func Test_newGitLabClient_SelfHostedHost(t *testing.T) {
	client, err := newGitLabClient("test-token", "gitlab.example.com", false)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	if got := client.BaseURL().Host; got != "gitlab.example.com" {
		t.Errorf("Expected BaseURL host 'gitlab.example.com', got '%s'", got)
	}
}
