package git

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

func Checkout(branch string) {
	cmd := exec.Command("git", "checkout", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func CreateEmptyCommit() {
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "Empty commit")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func Pull() {
	cmd := exec.Command("git", "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func PushBranch(branch string) {
	cmd := exec.Command("git", "push", "--set-upstream", "origin", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func Reset() {
	cmd := exec.Command("git", "reset", "--hard", "origin/HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func Stash() {
	cmd := exec.Command("git", "stash")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func StashDrop() {
	cmd := exec.Command("git", "stash", "drop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}
