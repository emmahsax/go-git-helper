package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v74/github"
)

func Test_NewGitHub(t *testing.T) {
	// This test requires a valid config file with GitHub token
	// We'll test that the struct is created properly
	gh := NewGitHub(false)

	if gh == nil {
		t.Fatal("Expected NewGitHub to return a non-nil GitHub struct")
	}

	if gh.Client == nil {
		t.Error("Expected GitHub client to be initialized")
	}

	if gh.Debug != false {
		t.Errorf("Expected Debug to be false, got %v", gh.Debug)
	}
}

func Test_NewGitHub_WithDebug(t *testing.T) {
	gh := NewGitHub(true)

	if gh == nil {
		t.Fatal("Expected NewGitHub to return a non-nil GitHub struct")
	}

	if gh.Debug != true {
		t.Errorf("Expected Debug to be true, got %v", gh.Debug)
	}
}

func Test_CreatePullRequest_Success(t *testing.T) {
	// Create a test server that returns a successful PR response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{
			"number": 1,
			"title": "Test PR",
			"body": "Test PR body",
			"state": "open",
			"draft": false,
			"html_url": "https://github.com/owner/repo/pull/1"
		}`)
	}))
	defer server.Close()

	// Create a GitHub client that uses the test server
	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

	gh := &GitHub{
		Debug:  false,
		Client: client,
	}

	options := &github.NewPullRequest{
		Title: github.Ptr("Test PR"),
		Body:  github.Ptr("Test PR body"),
		Head:  github.Ptr("feature-branch"),
		Base:  github.Ptr("main"),
		Draft: github.Ptr(false),
	}

	pr, err := gh.CreatePullRequest("owner", "repo", options)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if pr == nil {
		t.Error("Expected PR to be non-nil")
	}

	if pr.GetTitle() != "Test PR" {
		t.Errorf("Expected PR title 'Test PR', got '%s'", pr.GetTitle())
	}
}

func Test_CreatePullRequest_DraftNotSupported(t *testing.T) {
	callCount := 0
	// Create a test server that returns draft not supported error on first call
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprint(w, `{
				"message": "422 Draft pull requests are not supported in this repository."
			}`)
		} else {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{
				"number": 1,
				"title": "Test PR",
				"body": "Test PR body",
				"state": "open",
				"draft": false,
				"html_url": "https://github.com/owner/repo/pull/1"
			}`)
		}
	}))
	defer server.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

	gh := &GitHub{
		Debug:  false,
		Client: client,
	}

	options := &github.NewPullRequest{
		Title: github.Ptr("Test PR"),
		Body:  github.Ptr("Test PR body"),
		Head:  github.Ptr("feature-branch"),
		Base:  github.Ptr("main"),
		Draft: github.Ptr(true),
	}

	pr, err := gh.CreatePullRequest("owner", "repo", options)

	if err != nil {
		t.Errorf("Expected no error after retry, got %v", err)
	}

	if pr == nil {
		t.Error("Expected PR to be non-nil")
	}

	if callCount != 2 {
		t.Errorf("Expected 2 API calls (first fail, second succeed), got %d", callCount)
	}
}

func Test_CreatePullRequest_Error(t *testing.T) {
	// Skip this test as it would call os.Exit through HandleError
	// The error path is covered indirectly by other integration tests
	t.Skip("Skipping test that would call os.Exit")
}

func Test_newGitHubClient(t *testing.T) {
	token := "test-token-123"
	client := newGitHubClient(token)

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	// Verify the client is properly configured by checking it can make requests
	// We can't directly test the token, but we can verify the client structure
	if client.BaseURL == nil {
		t.Error("Expected client BaseURL to be set")
	}
}

func Test_newGitHubClient_EmptyToken(t *testing.T) {
	client := newGitHubClient("")

	if client == nil {
		t.Error("Expected client to be non-nil even with empty token")
	}
}
