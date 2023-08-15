package git

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func Checkout(branch string) {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func CleanDeletedBranches() {
	cmd := exec.Command("git", "branch", "-vv")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	branches := strings.Split(string(output), "\n")
	pattern := "origin/.*: gone"

	for _, branch := range branches {
		re := regexp.MustCompile(pattern)

		if re.MatchString(branch) {
			b := strings.Fields(branch)[0]
			cmd = exec.Command("git", "branch", "-D", b)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatal(err)
				return
			}

			fmt.Printf("%s", string(output))
		}
	}
}

func CreateBranch(branch string) {
	cmd := exec.Command("git", "branch", "--no-track", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func CreateEmptyCommit() {
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "Empty commit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func CurrentBranch() string {
	cmd := exec.Command("git", "branch")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	re := regexp.MustCompile(`\*\s(\S*)`)
	match := re.FindStringSubmatch(string(output))

	if len(match) == 2 {
		return match[1]
	}

	return ""
}

func DefaultBranch() string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	branch := strings.SplitN(strings.TrimSpace(string(output)), "/", 4)
	if len(branch) != 4 {
		log.Fatal("Invalid branch format")
		return ""
	}

	return branch[3]
}

func Fetch() {
	cmd := exec.Command("git", "fetch", "-p")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func Pull() {
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func PushBranch(branch string) {
	cmd := exec.Command("git", "push", "--set-upstream", "origin", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func RepoName() string {
	output, err := exec.Command("git", "remote", "-v").Output()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	remoteURL := string(output)
	re := regexp.MustCompile(`\S\s*\S+.com\S{1}(\S*).git`)
	match := re.FindStringSubmatch(remoteURL)
	if len(match) >= 2 {
		return match[1]
	} else {
		log.Fatal("No match found")
	}

	return ""
}

func Remotes() []string {
	cmd := exec.Command("git", "remote", "-v")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return []string{}
	}

	return strings.Split(string(output), "\n")
}

func Reset() {
	cmd := exec.Command("git", "reset", "--hard", "origin/HEAD")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func Stash() {
	cmd := exec.Command("git", "stash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func StashDrop() {
	cmd := exec.Command("git", "stash", "drop")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return
	}
}
