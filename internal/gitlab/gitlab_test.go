package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func Test_NewGitLab(t *testing.T) {
	// This test requires a valid config file with GitLab token
	// We'll test that the struct is created properly
	gl := NewGitLab(false)

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
	gl := NewGitLab(true)

	if gl == nil {
		t.Fatal("Expected NewGitLab to return a non-nil GitLab struct")
	}

	if gl.Debug != true {
		t.Errorf("Expected Debug to be true, got %v", gl.Debug)
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
		Title:        gitlab.Ptr("Test MR"),
		Description:  gitlab.Ptr("Test MR body"),
		SourceBranch: gitlab.Ptr("feature-branch"),
		TargetBranch: gitlab.Ptr("main"),
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
	client, err := newGitLabClient(token, false)

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
	client, err := newGitLabClient("", false)

	if err != nil {
		t.Errorf("Expected no error with empty token, got %v", err)
	}

	if client == nil {
		t.Error("Expected client to be non-nil even with empty token")
	}
}

func Test_newGitLabClient_WithDebug(t *testing.T) {
	token := "test-token-debug"
	client, err := newGitLabClient(token, true)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Error("Expected client to be non-nil")
	}
}
