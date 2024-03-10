package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type MockExecutor struct {
	Args    []string
	Command string
	Debug   bool
	Output  []byte
}

func (me *MockExecutor) Exec(execType string, command string, args ...string) ([]byte, error) {
	me.Command = command
	me.Args = args
	return me.Output, nil
}

func Test_fetchReleaseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	executor := &MockExecutor{Debug: true}
	u := newUpdate("owner", "repo", true, executor)
	body := u.fetchReleaseBody(server.URL)

	if string(body) != "test response" {
		t.Errorf("expected 'test response', got '%s'", body)
	}
}

func Test_getDownloadURL(t *testing.T) {
	body := []byte(`{
		"assets": [
			{
				"name": "` + asset + `",
				"browser_download_url": "https://example.com/download"
			}
		]
	}`)

	executor := &MockExecutor{Debug: true}
	u := newUpdate("owner", "repo", true, executor)
	downloadURL := u.getDownloadURL(body)

	if downloadURL != "https://example.com/download" {
		t.Errorf("expected 'https://example.com/download', got '%s'", downloadURL)
	}
}

func Test_downloadAndSaveBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	executor := &MockExecutor{Debug: true}
	u := newUpdate("owner", "repo", true, executor)

	binaryName := "test_binary"

	u.downloadAndSaveBinary(server.URL, binaryName)

	content, err := os.ReadFile(binaryName)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "test response" {
		t.Errorf("expected 'test response', got '%s'", content)
	}

	os.Remove(binaryName)
}

func Test_moveGitHelper(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"mv", "./" + asset, "/usr/local/bin/git-helper"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}
		u := newUpdate("owner", "repo", true, executor)
		u.moveGitHelper()

		if executor.Command != "sudo" {
			t.Errorf("unexpected command received: expected %s, but got %s", "git", executor.Command)
		}

		if len(executor.Args) != len(test.expectedArgs) {
			t.Errorf("unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
		}

		for i, v := range executor.Args {
			if v != test.expectedArgs[i] {
				t.Errorf("unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
			}
		}
	}
}

func Test_setPermissions(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"chmod", "+x", "/usr/local/bin/git-helper"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}
		u := newUpdate("owner", "repo", true, executor)
		u.setPermissions()

		if executor.Command != "sudo" {
			t.Errorf("unexpected command received: expected %s, but got %s", "git", executor.Command)
		}

		if len(executor.Args) != len(test.expectedArgs) {
			t.Errorf("unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
		}

		for i, v := range executor.Args {
			if v != test.expectedArgs[i] {
				t.Errorf("unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
			}
		}
	}
}

func Test_outputNewVersion(t *testing.T) {
	tests := []struct {
		executorOutput string
		expectedArgs   []string
	}{
		{
			executorOutput: "Installed git-helper version",
			expectedArgs:   []string{"version"},
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: []byte(test.executorOutput),
		}

		u := newUpdate("owner", "repo", true, executor)
		u.outputNewVersion()

		if executor.Command != "git-helper" {
			t.Errorf("unexpected command received: expected %s, but got %s", "git", executor.Command)
		}

		if len(executor.Args) != len(test.expectedArgs) {
			t.Errorf("unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
		}

		for i, v := range executor.Args {
			if v != test.expectedArgs[i] {
				t.Errorf("unexpected args received: expected %v, but got %v", test.expectedArgs, executor.Args)
			}
		}
	}
}
