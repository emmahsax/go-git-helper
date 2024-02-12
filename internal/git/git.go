package git

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/utils"
)

type Git struct {
	Debug    bool
	Executor executor.ExecutorInterface
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

	branches := strings.Split(string(output), "\n")
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

func (g *Git) CreateBranch(branch string) {
	_, err := g.Executor.Exec("waitAndStdout", "git", "branch", "--no-track", branch)
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) CreateEmptyCommit() {
	_, err := g.Executor.Exec("waitAndStdout", "git", "commit", "--allow-empty", "-m", "Empty commit")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}

func (g *Git) CurrentBranch() string {
	cmd := exec.Command("git", "branch")
	output, err := cmd.CombinedOutput()
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
			fmt.Printf("\nYour symbolic ref is not set up properly. Please run:\n  git-helper set-head-ref [defaultBranch]\n\nAnd then try your command again.\n\n")
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
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	output, err := cmd.CombinedOutput()
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

func (g *Git) RepoName() string {
	output, err := exec.Command("git", "remote", "-v").Output()
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return ""
	}

	remoteURL := string(output)
	re := regexp.MustCompile(`\S\s*\S+.com\S{1}(\S*).git`)
	match := re.FindStringSubmatch(remoteURL)
	if len(match) >= 2 {
		return match[1]
	} else {
		err = errors.New("no match found")
		utils.HandleError(err, g.Debug, nil)
	}

	return ""
}

func (g *Git) Remotes() []string {
	cmd := exec.Command("git", "remote", "-v")

	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return []string{}
	}

	return strings.Split(string(output), "\n")
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
	_, err := g.Executor.Exec("waitAndStdout", "git", "stash", "drop")
	if err != nil {
		utils.HandleError(err, g.Debug, nil)
		return
	}
}
