package git

import (
	"testing"
)

type MockExecutor struct {
	Args      []string
	Command   string
	Debug     bool
	Output    []byte
	ExecCount int
}

func (me *MockExecutor) Exec(execType string, command string, args ...string) ([]byte, error) {
	me.Command = command
	me.Args = args
	me.ExecCount++
	return me.Output, nil
}

func Test_Checkout(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"checkout", "branch"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.Checkout("branch")

		if executor.Command != "git" {
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

func Test_CleanDeletedBranches(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"branch", "-vv"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.CleanDeletedBranches()

		if executor.Command != "git" {
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

func Test_CreateBranch(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"branch", "--no-track", "branch"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.CreateBranch("branch")

		if executor.Command != "git" {
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

func Test_CreateEmptyCommit(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"commit", "--allow-empty", "-m", "Empty commit"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.CreateEmptyCommit()

		if executor.Command != "git" {
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

func Test_CurrentBranch(t *testing.T) {
	tests := []struct {
		expectedArgs   []string
		expectedOutput string
		finalOutput    string
	}{
		{
			expectedArgs:   []string{"branch"},
			expectedOutput: "* master",
			finalOutput:    "master",
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: []byte(test.expectedOutput),
		}

		g := NewGit(true, executor)
		o := g.CurrentBranch()

		if executor.Command != "git" {
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

		if string(executor.Output) != test.expectedOutput {
			t.Errorf("unexpected output received: expected %s, but got %s", test.expectedOutput, executor.Output)
		}

		if o != test.finalOutput {
			t.Errorf("unexpected output received: expected %s, but got %s", test.finalOutput, o)
		}
	}
}

func Test_DefaultBranch(t *testing.T) {
	tests := []struct {
		expectedArgs   []string
		expectedOutput string
		finalOutput    string
	}{
		{
			expectedArgs:   []string{"symbolic-ref", "refs/remotes/origin/HEAD"},
			expectedOutput: "refs/remotes/origin/master",
			finalOutput:    "master",
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: []byte(test.expectedOutput),
		}

		g := NewGit(true, executor)
		o := g.DefaultBranch()

		if executor.Command != "git" {
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

		if string(executor.Output) != test.expectedOutput {
			t.Errorf("unexpected output received: expected %s, but got %s", test.expectedOutput, executor.Output)
		}

		if o != test.finalOutput {
			t.Errorf("unexpected output received: expected %s, but got %s", test.finalOutput, o)
		}
	}
}

func Test_Fetch(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"fetch", "-p"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.Fetch()

		if executor.Command != "git" {
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

func Test_GetGitRootDir(t *testing.T) {
	tests := []struct {
		expectedArgs   []string
		expectedOutput string
	}{
		{
			expectedArgs:   []string{"rev-parse", "--show-toplevel"},
			expectedOutput: "go-git-helper",
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: []byte(test.expectedOutput),
		}

		g := NewGit(true, executor)
		o := g.GetGitRootDir()

		if executor.Command != "git" {
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

		if string(executor.Output) != test.expectedOutput {
			t.Errorf("unexpected output received: expected %s, but got %s", test.expectedOutput, executor.Output)
		}

		if o != test.expectedOutput {
			t.Errorf("unexpected output received: expected %s, but got %s", test.expectedOutput, o)

		}
	}
}

func Test_Pull(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"pull"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.Pull()

		if executor.Command != "git" {
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

func Test_PushBranch(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"push", "--set-upstream", "origin", "branch"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.PushBranch("branch")

		if executor.Command != "git" {
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

func Test_RepoName(t *testing.T) {
	tests := []struct {
		name           string
		expectedArgs   []string
		expectedOutput string
		repoName       string
	}{
		{
			name:         "SSH remote without .git",
			expectedArgs: []string{"remote", "-v"},
			expectedOutput: `origin  git@github.com:emmahsax/go-git-helper (fetch)
origin  git@github.com:emmahsax/go-git-helper (push)`,
			repoName: "emmahsax/go-git-helper",
		},
		{
			name:         "SSH remote with .git",
			expectedArgs: []string{"remote", "-v"},
			expectedOutput: `origin  git@github.com:emmahsax/go-git-helper.git (fetch)
origin  git@github.com:emmahsax/go-git-helper.git (push)`,
			repoName: "emmahsax/go-git-helper",
		},
		{
			name:         "HTTP remote without .git",
			expectedArgs: []string{"remote", "-v"},
			expectedOutput: `origin  https://github.com/emmahsax/go-git-helper (fetch)
origin  https://github.com/emmahsax/go-git-helper (push)`,
			repoName: "emmahsax/go-git-helper",
		},
		{
			name:         "HTTP remote with .git",
			expectedArgs: []string{"remote", "-v"},
			expectedOutput: `origin  https://github.com/emmahsax/go-git-helper.git (fetch)
origin  https://github.com/emmahsax/go-git-helper.git (push)`,
			repoName: "emmahsax/go-git-helper",
		},
		{
			name:         "self-hosted SSH remote on a non-.com domain",
			expectedArgs: []string{"remote", "-v"},
			expectedOutput: `origin  git@gitlab.example.ai:emmahsax/go-git-helper.git (fetch)
origin  git@gitlab.example.ai:emmahsax/go-git-helper.git (push)`,
			repoName: "emmahsax/go-git-helper",
		},
		{
			name:         "ssh:// remote with a custom port",
			expectedArgs: []string{"remote", "-v"},
			expectedOutput: `origin  ssh://git@gitlab.example.com:2222/emmahsax/go-git-helper.git (fetch)
origin  ssh://git@gitlab.example.com:2222/emmahsax/go-git-helper.git (push)`,
			repoName: "emmahsax/go-git-helper",
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: []byte(test.expectedOutput),
		}

		g := NewGit(true, executor)
		r := g.RepoName()

		if executor.Command != "git" {
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

		if string(executor.Output) != test.expectedOutput {
			t.Errorf("unexpected output received: expected %s, but got %s", test.expectedOutput, executor.Output)
		}

		if r != test.repoName {
			t.Errorf("%s unexpected output received: expected %s, but got %s", test.name, test.repoName, r)
		}
	}
}

func Test_Remotes(t *testing.T) {
	tests := []struct {
		expectedArgs   []string
		expectedOutput string
		remotes        []string
	}{
		{
			expectedArgs: []string{"remote", "-v"},
			expectedOutput: `origin  git@github.com:emmahsax/go-git-helper.git (fetch)
origin  git@github.com:emmahsax/go-git-helper.git (push)`,
			remotes: []string{"origin  git@github.com:emmahsax/go-git-helper.git (fetch)", "origin  git@github.com:emmahsax/go-git-helper.git (push)"},
		},
	}

	for _, test := range tests {
		executor := &MockExecutor{
			Debug:  true,
			Output: []byte(test.expectedOutput),
		}

		g := NewGit(true, executor)
		r := g.Remotes()

		if executor.Command != "git" {
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

		if string(executor.Output) != test.expectedOutput {
			t.Errorf("unexpected output received: expected %s, but got %s", test.expectedOutput, executor.Output)
		}

		for i, v := range r {
			if v != test.remotes[i] {
				t.Errorf("unexpected args received: expected %v, but got %v", test.remotes, r)
			}
		}
	}
}

func Test_RemoteHosts(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput string
		hosts          []string
	}{
		{
			name: "single SSH remote",
			expectedOutput: `origin  git@github.com:emmahsax/go-git-helper.git (fetch)
origin  git@github.com:emmahsax/go-git-helper.git (push)`,
			hosts: []string{"github.com"},
		},
		{
			name: "deduplicates fetch and push",
			expectedOutput: `origin  https://gitlab.com/emmahsax/go-git-helper.git (fetch)
origin  https://gitlab.com/emmahsax/go-git-helper.git (push)`,
			hosts: []string{"gitlab.com"},
		},
		{
			// The whole point of RemoteHosts: a naive substring check would
			// treat this as a match for "gitlab.com".
			name: "self-hosted domain that contains the public domain as a substring",
			expectedOutput: `origin  git@gitlab.company.com:emmahsax/go-git-helper.git (fetch)
origin  git@gitlab.company.com:emmahsax/go-git-helper.git (push)`,
			hosts: []string{"gitlab.company.com"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &MockExecutor{Debug: true, Output: []byte(test.expectedOutput)}
			g := NewGit(true, executor)
			hosts := g.RemoteHosts()

			if len(hosts) != len(test.hosts) {
				t.Fatalf("expected hosts %v, got %v", test.hosts, hosts)
			}

			for i := range test.hosts {
				if hosts[i] != test.hosts[i] {
					t.Errorf("expected hosts %v, got %v", test.hosts, hosts)
				}
			}
		})
	}
}

// Regression: RemoteHosts/RepoName must only consider the "origin" remote.
// A repo with origin on one host and a different remote (e.g. upstream) on
// another must resolve to origin's host/path, not whichever remote happens
// to match first.
func Test_RemoteHosts_IgnoresNonOriginRemotes(t *testing.T) {
	executor := &MockExecutor{Debug: true, Output: []byte(`origin  git@github.company.com:emmahsax/go-git-helper.git (fetch)
origin  git@github.company.com:emmahsax/go-git-helper.git (push)
upstream  git@github.com:upstream-org/go-git-helper.git (fetch)
upstream  git@github.com:upstream-org/go-git-helper.git (push)`)}
	g := NewGit(true, executor)

	hosts := g.RemoteHosts()
	if len(hosts) != 1 || hosts[0] != "github.company.com" {
		t.Errorf("expected only origin's host ['github.company.com'], got %v", hosts)
	}
}

func Test_RepoName_IgnoresNonOriginRemotes(t *testing.T) {
	executor := &MockExecutor{Debug: true, Output: []byte(`origin  git@github.company.com:emmahsax/go-git-helper.git (fetch)
origin  git@github.company.com:emmahsax/go-git-helper.git (push)
upstream  git@github.com:upstream-org/go-git-helper.git (fetch)
upstream  git@github.com:upstream-org/go-git-helper.git (push)`)}
	g := NewGit(true, executor)

	if repo := g.RepoName(); repo != "emmahsax/go-git-helper" {
		t.Errorf("expected origin's repo path 'emmahsax/go-git-helper', got %q", repo)
	}
}

// Remotes() is called internally by both RemoteHosts() and RepoName(); it
// must only shell out to `git remote -v` once per *Git, not once per call.
func Test_Remotes_CachesAcrossCalls(t *testing.T) {
	executor := &MockExecutor{Debug: true, Output: []byte(`origin  git@github.com:emmahsax/go-git-helper.git (fetch)
origin  git@github.com:emmahsax/go-git-helper.git (push)`)}
	g := NewGit(true, executor)

	g.RemoteHosts()
	g.RepoName()
	g.Remotes()

	if executor.ExecCount != 1 {
		t.Errorf("expected exactly 1 `git remote -v` call across repeated use, got %d", executor.ExecCount)
	}
}

func Test_Reset(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"reset", "--hard", "origin/HEAD"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.Reset()

		if executor.Command != "git" {
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

func Test_SetHeadRef(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/branch"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.SetHeadRef("branch")

		if executor.Command != "git" {
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

func Test_Stash(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"stash"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.Stash()

		if executor.Command != "git" {
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

func Test_StashDrop(t *testing.T) {
	tests := []struct {
		expectedArgs []string
	}{
		{expectedArgs: []string{"stash", "drop"}},
	}

	for _, test := range tests {
		executor := &MockExecutor{Debug: true}

		g := NewGit(true, executor)
		g.StashDrop()

		if executor.Command != "git" {
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
