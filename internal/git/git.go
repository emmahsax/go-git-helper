package git

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/utils"
)

type Git struct {
	Debug    bool
	Executor executor.ExecutorInterface

	// remotes caches Remotes()'s result for this *Git's lifetime: `git
	// remote -v` doesn't change within a single invocation, and several
	// callers (RepoName, RemoteHosts) ask for it more than once.
	remotes       []string
	remotesCached bool
}

func NewGit(debug bool, executor executor.ExecutorInterface) *Git {
	return &Git{
		Debug:    debug,
		Executor: executor,
	}
}

func (g *Git) Checkout(branch string) {
	_, err := g.Executor.Exec("waitAndStdout", "git", "checkout", branch)
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) CleanDeletedBranches() {
	output, err := g.Executor.Exec("actionAndOutput", "git", "branch", "-vv")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}

	branches := utils.SplitLines(string(output))
	pattern := "origin/.*: gone"

	for _, branch := range branches {
		re := regexp.MustCompile(pattern)

		if re.MatchString(branch) {
			b := strings.Fields(branch)[0]
			output, err = g.Executor.Exec("actionAndOutput", "git", "branch", "-D", b)
			if err != nil {
				utils.HandleError(err, g.Debug, nil)
				return
			}

			fmt.Printf("%s", string(output))
		}
	}
}

func (g *Git) CreateBranch(branch string) error {
	_, err := g.Executor.Exec("waitAndStdout", "git", "branch", "--no-track", branch)
	if err != nil {
		return err
	}

	return nil
}

func (g *Git) CreateEmptyCommit() {
	_, err := g.Executor.Exec("waitAndStdout", "git", "commit", "--allow-empty", "-m", "Empty commit")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) CurrentBranch() string {
	output, err := g.Executor.Exec("actionAndOutput", "git", "branch")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return ""
	}

	re := regexp.MustCompile(`\*\s(\S*)`)
	match := re.FindStringSubmatch(string(output))

	if len(match) == 2 {
		return match[1]
	}

	return ""
}

func (g *Git) DefaultBranch() string {
	output, err := g.Executor.Exec("actionAndOutput", "git", "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		if strings.Contains(err.Error(), "fatal: ") {
			fmt.Printf("\nYour symbolic ref is not set up properly. Please run:\n  git-helper set-head-ref [defaultBranch]\n\nAnd then try your command again\n\n")
		}
		utils.HandleError(err, g.Debug, nil)
		return ""
	}

	branch := strings.SplitN(strings.TrimSpace(string(output)), "/", 4)
	if len(branch) != 4 {
		err = errors.New("invalid branch format")
		utils.HandleError(err, g.Debug, nil)
		return ""
	}

	return branch[3]
}

func (g *Git) Fetch() {
	_, err := g.Executor.Exec("waitAndStdout", "git", "fetch", "-p")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) GetGitRootDir() string {
	output, err := g.Executor.Exec("actionAndOutput", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return ""
	}

	return strings.TrimSpace(string(output))
}

func (g *Git) Pull() {
	_, err := g.Executor.Exec("waitAndStdout", "git", "pull")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) PushBranch(branch string) {
	_, err := g.Executor.Exec("waitAndStdout", "git", "push", "--set-upstream", "origin", branch)
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

// RepoName is the owner/repo path (e.g. "emmahsax/go-git-helper") of the
// origin remote's push URL, with any ".git" suffix stripped.
func (g *Git) RepoName() string {
	for _, line := range g.originRemotes() {
		if !strings.HasSuffix(strings.TrimSpace(line), "(push)") {
			continue
		}

		if remote, ok := parseRemoteURL(remoteURLFromLine(line)); ok {
			return remote.Path
		}
	}

	err := errors.New("no match found")
	utils.HandleError(err, g.Debug, nil)
	return ""
}

func (g *Git) Remotes() []string {
	if g.remotesCached {
		return g.remotes
	}

	output, err := g.Executor.Exec("actionAndOutput", "git", "remote", "-v")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return []string{}
	}

	g.remotes = utils.SplitLines(string(output))
	g.remotesCached = true

	return g.remotes
}

// originRemotes returns Remotes() filtered to the "origin" remote
// specifically, not any other configured remote (e.g. upstream). A
// repo-request command always targets origin, so detection, account
// resolution, and the repo path it uses must all agree on that same
// remote -- otherwise, e.g., a repo with origin on a self-hosted instance
// and upstream on the public host could resolve the wrong account for the
// repo actually being worked on.
func (g *Git) originRemotes() []string {
	var lines []string

	for _, line := range g.Remotes() {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "origin" {
			lines = append(lines, line)
		}
	}

	return lines
}

// RemoteHosts returns the distinct, exact hosts (e.g. "github.com",
// "gitlab.example.com") parsed from the origin remote's URLs (fetch and
// push, in case they ever differ) -- not any other configured remote.
// Callers matching against these should compare hosts exactly, not with
// substring containment: a naive substring check would treat
// "gitlab.company.com" as a match for "gitlab.com", or "git.example.com" as
// a match for "sub.git.example.com".
func (g *Git) RemoteHosts() []string {
	var hosts []string

	for _, line := range g.originRemotes() {
		remote, ok := parseRemoteURL(remoteURLFromLine(line))
		if !ok || slices.Contains(hosts, remote.Host) {
			continue
		}

		hosts = append(hosts, remote.Host)
	}

	return hosts
}

// remoteURLFromLine extracts the URL field from a `git remote -v` line, e.g.
// "origin  git@github.com:owner/repo.git (push)" -> "git@github.com:owner/repo.git".
func remoteURLFromLine(line string) string {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ""
	}

	return fields[1]
}

// parsedRemoteURL holds the pieces of a git remote URL relevant to Git
// Helper: the exact host, and the repo path with any ".git" suffix removed.
type parsedRemoteURL struct {
	Host string
	Path string
}

// scpLikeRemoteURL matches the SCP-like shorthand git uses for SSH remotes,
// e.g. "git@github.com:owner/repo.git" -- as opposed to a standard URL like
// "ssh://git@github.com/owner/repo.git", which net/url already understands.
var scpLikeRemoteURL = regexp.MustCompile(`^(?:[^@\s]+@)?([^:/\s]+):(.+)$`)

// parseRemoteURL extracts the host and path from a git remote URL, in either
// the SCP-like shorthand (git@host:owner/repo.git) or a standard URL
// (https://host/owner/repo.git, ssh://git@host/owner/repo.git, etc.). ok is
// false if rawURL doesn't look like either.
func parseRemoteURL(rawURL string) (parsedRemoteURL, bool) {
	rawURL = strings.TrimSpace(rawURL)

	if !strings.Contains(rawURL, "://") {
		if m := scpLikeRemoteURL.FindStringSubmatch(rawURL); m != nil {
			return parsedRemoteURL{Host: m[1], Path: strings.TrimSuffix(m[2], ".git")}, true
		}

		return parsedRemoteURL{}, false
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Hostname() == "" {
		return parsedRemoteURL{}, false
	}

	path := strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), ".git")
	return parsedRemoteURL{Host: u.Hostname(), Path: path}, true
}

func (g *Git) Reset() {
	_, err := g.Executor.Exec("waitAndStdout", "git", "reset", "--hard", "origin/HEAD")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) SetHeadRef(defaultBranch string) {
	_, err := g.Executor.Exec("waitAndStdout", "git", "branch", "--set-upstream-to=origin/"+defaultBranch, defaultBranch)
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}

	_, err = g.Executor.Exec("waitAndStdout", "git", "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/"+defaultBranch)
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) Stash() {
	_, err := g.Executor.Exec("waitAndStdout", "git", "stash")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) StashDrop() {
	_, _ = g.Executor.Exec("waitAndStdout", "git", "stash", "drop")
}
