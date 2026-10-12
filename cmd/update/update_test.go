package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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

func Test_assetName(t *testing.T) {
	tests := []struct {
		goos, goarch, expected string
	}{
		{"darwin", "arm64", "git-helper_darwin_arm64"},
		{"darwin", "amd64", "git-helper_darwin_amd64"},
		{"linux", "amd64", "git-helper_linux_amd64"},
		{"linux", "arm64", "git-helper_linux_arm64"},
		{"windows", "amd64", "git-helper_windows_amd64.exe"},
		{"windows", "arm64", "git-helper_windows_arm64.exe"},
	}

	for _, test := range tests {
		t.Run(test.goos+"/"+test.goarch, func(t *testing.T) {
			if got := assetName(test.goos, test.goarch); got != test.expected {
				t.Errorf("expected %s, got %s", test.expected, got)
			}
		})
	}
}

func Test_installPath_override(t *testing.T) {
	u := newUpdate("owner", "repo", true, &MockExecutor{})
	u.InstallPath = filepath.Join(t.TempDir(), "git-helper")

	got, err := u.installPath()
	if err != nil {
		t.Fatal(err)
	}

	if got != u.InstallPath {
		t.Errorf("expected %s, got %s", u.InstallPath, got)
	}
}

func Test_isGoRunBinary(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/var/folders/xy/T/go-build123456/b001/exe/main", true},
		{"/tmp/go-build987/b001/exe/git-helper", true},
		{`C:\Users\me\AppData\Local\Temp\go-build42\b001\exe\main.exe`, true},
		{"/Users/me/Library/Caches/go-build/96/968ede4d83a5675950c1ca6b91d089c7bae6bcb4131de946f0bed85e52cd36e6-d/main", true},
		{"/home/me/.cache/go-build/ab/abcdef-d/main", true},
		{`C:\Users\me\AppData\Local\go-build\ab\abcdef-d\main.exe`, true},
		{"/usr/local/bin/git-helper", false},
		{"/home/me/go-builder/bin/git-helper", false},
	}

	for _, test := range tests {
		if got := isGoRunBinary(test.path); got != test.expected {
			t.Errorf("isGoRunBinary(%q) = %v, expected %v", test.path, got, test.expected)
		}
	}
}

func Test_installPath_goRunUsesBinaryOnPath(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "git-helper")
	writeFile(t, installed, "installed")

	u := newUpdate("owner", "repo", true, &MockExecutor{})
	u.Executable = func() (string, error) { return "/tmp/go-build123/b001/exe/main", nil }
	u.LookPath = func(name string) (string, error) {
		if name != "git-helper" {
			t.Errorf("expected a lookup of git-helper, got %s", name)
		}
		return installed, nil
	}

	got, err := u.installPath()
	if err != nil {
		t.Fatal(err)
	}

	want, _ := filepath.EvalSymlinks(installed)
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func Test_installPath_goRunWithoutInstalledBinary(t *testing.T) {
	u := newUpdate("owner", "repo", true, &MockExecutor{})
	u.Executable = func() (string, error) { return "/tmp/go-build123/b001/exe/main", nil }
	u.LookPath = func(string) (string, error) { return "", os.ErrNotExist }

	if _, err := u.installPath(); err == nil {
		t.Error("expected an error when go run is used and no git-helper is on PATH")
	}
}

func Test_installPath_runningBinary(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "git-helper")
	writeFile(t, installed, "installed")

	u := newUpdate("owner", "repo", true, &MockExecutor{})
	u.Executable = func() (string, error) { return installed, nil }
	u.LookPath = func(string) (string, error) {
		t.Error("did not expect a PATH lookup for a normally installed binary")
		return "", os.ErrNotExist
	}

	got, err := u.installPath()
	if err != nil {
		t.Fatal(err)
	}

	want, _ := filepath.EvalSymlinks(installed)
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func Test_fetchReleaseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	u := newUpdate("owner", "repo", true, &MockExecutor{})
	body, err := u.fetchReleaseBody(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "test response" {
		t.Errorf("expected 'test response', got '%s'", body)
	}
}

func Test_fetchReleaseBody_badStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	u := newUpdate("owner", "repo", true, &MockExecutor{})
	if _, err := u.fetchReleaseBody(server.URL); err == nil {
		t.Error("expected an error for a non-200 response")
	}
}

func Test_getDownloadURL(t *testing.T) {
	body := []byte(`{
		"assets": [
			{"name": "git-helper_linux_amd64", "browser_download_url": "https://example.com/linux"},
			{"name": "git-helper_windows_amd64.exe", "browser_download_url": "https://example.com/windows"}
		]
	}`)

	tests := []struct {
		goos, goarch, expected string
		expectErr              bool
	}{
		{"linux", "amd64", "https://example.com/linux", false},
		{"windows", "amd64", "https://example.com/windows", false},
		{"darwin", "amd64", "", true},
		{"linux", "arm64", "", true},
	}

	for _, test := range tests {
		t.Run(test.goos+"/"+test.goarch, func(t *testing.T) {
			u := newUpdate("owner", "repo", true, &MockExecutor{})
			u.GOOS, u.GOARCH = test.goos, test.goarch

			got, err := u.getDownloadURL(body)
			if test.expectErr != (err != nil) {
				t.Fatalf("expected error: %v, got: %v", test.expectErr, err)
			}

			if got != test.expected {
				t.Errorf("expected %q, got %q", test.expected, got)
			}
		})
	}
}

func Test_downloadBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	u := newUpdate("owner", "repo", true, &MockExecutor{})
	path, err := u.downloadBinary(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "test response" {
		t.Errorf("expected 'test response', got '%s'", content)
	}
}

func Test_downloadBinary_badStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	u := newUpdate("owner", "repo", true, &MockExecutor{})
	if _, err := u.downloadBinary(server.URL); err == nil {
		t.Error("expected an error for a non-200 response")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
}

func Test_install(t *testing.T) {
	for _, goos := range []string{"linux", "darwin", "windows"} {
		t.Run(goos, func(t *testing.T) {
			dir := t.TempDir()
			downloaded := filepath.Join(dir, "downloaded")
			target := filepath.Join(dir, "git-helper")
			writeFile(t, downloaded, "new")
			writeFile(t, target, "old")

			executor := &MockExecutor{}
			u := newUpdate("owner", "repo", true, executor)
			u.GOOS = goos

			if err := u.install(downloaded, target); err != nil {
				t.Fatal(err)
			}

			content, _ := os.ReadFile(target)
			if string(content) != "new" {
				t.Errorf("expected target to contain 'new', got '%s'", content)
			}

			if executor.Command != "" {
				t.Errorf("expected no external commands, got %s %v", executor.Command, executor.Args)
			}

			if goos == "windows" {
				old, _ := os.ReadFile(target + ".old")
				if string(old) != "old" {
					t.Errorf("expected the previous binary at %s.old, got '%s'", target, old)
				}

				u.removeStaleBinary(target)
				if _, err := os.Stat(target + ".old"); !os.IsNotExist(err) {
					t.Errorf("expected the .old binary to be removed")
				}
			}
		})
	}
}

func Test_removeStaleBinary_onlyOnWindows(t *testing.T) {
	for _, goos := range []string{"linux", "darwin", "windows"} {
		t.Run(goos, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "git-helper")
			writeFile(t, target+".old", "unrelated backup")

			u := newUpdate("owner", "repo", true, &MockExecutor{})
			u.GOOS = goos
			u.removeStaleBinary(target)

			_, err := os.Stat(target + ".old")
			if goos == "windows" && !os.IsNotExist(err) {
				t.Errorf("expected the stale binary to be removed on windows")
			}

			if goos != "windows" && err != nil {
				t.Errorf("expected %s.old to be left alone on %s: %v", target, goos, err)
			}
		})
	}
}

func Test_install_missingTarget(t *testing.T) {
	dir := t.TempDir()
	downloaded := filepath.Join(dir, "downloaded")
	target := filepath.Join(dir, "git-helper")
	writeFile(t, downloaded, "new")

	u := newUpdate("owner", "repo", true, &MockExecutor{})
	if err := u.install(downloaded, target); err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(target)
	if string(content) != "new" {
		t.Errorf("expected target to contain 'new', got '%s'", content)
	}
}

// readOnlyDir returns a directory the current user cannot create files in, or
// skips the test when the platform or user ignores permission bits.
func readOnlyDir(t *testing.T) string {
	t.Helper()

	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("requires a non-root Unix user")
	}

	dir := t.TempDir()
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) })

	if f, err := os.CreateTemp(dir, "probe-*"); err == nil {
		f.Close()
		t.Skip("directory permissions are not enforced here")
	}

	return dir
}

func Test_install_fallsBackToSudo(t *testing.T) {
	dir := readOnlyDir(t)

	downloaded := filepath.Join(t.TempDir(), "downloaded")
	target := filepath.Join(dir, "git-helper")
	writeFile(t, downloaded, "new")

	executor := &MockExecutor{}
	u := newUpdate("owner", "repo", true, executor)
	u.GOOS = "linux"

	if err := u.install(downloaded, target); err != nil {
		t.Fatal(err)
	}

	expected := []string{"install", "-m", "0755", downloaded, target}
	if executor.Command != "sudo" {
		t.Errorf("expected sudo, got %s", executor.Command)
	}

	if len(executor.Args) != len(expected) {
		t.Fatalf("expected args %v, got %v", expected, executor.Args)
	}

	for i, v := range expected {
		if executor.Args[i] != v {
			t.Errorf("expected args %v, got %v", expected, executor.Args)
		}
	}
}

func Test_install_windowsNeverUsesSudo(t *testing.T) {
	dir := readOnlyDir(t)

	downloaded := filepath.Join(t.TempDir(), "downloaded")
	writeFile(t, downloaded, "new")

	executor := &MockExecutor{}
	u := newUpdate("owner", "repo", true, executor)
	u.GOOS = "windows"

	if err := u.install(downloaded, filepath.Join(dir, "git-helper.exe")); err == nil {
		t.Error("expected an error when the directory is not writable")
	}

	if executor.Command != "" {
		t.Errorf("expected no external commands, got %s", executor.Command)
	}
}

func Test_outputNewVersion(t *testing.T) {
	executor := &MockExecutor{
		Debug:  true,
		Output: []byte("git-helper version 1.0.0"),
	}

	u := newUpdate("owner", "repo", true, executor)
	u.outputNewVersion("/some/dir/git-helper")

	if executor.Command != "/some/dir/git-helper" {
		t.Errorf("expected the installed binary to be run by path, got %s", executor.Command)
	}

	if len(executor.Args) != 1 || executor.Args[0] != "version" {
		t.Errorf("unexpected args received: %v", executor.Args)
	}
}
