package cleanBranches

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type CleanBranches struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "clean-branches",
		Short:                 "Switches to the default branch, git pulls, git fetches, and removes remote-deleted branches from your machine",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cb := &CleanBranches{}
			cb.execute()
			return nil
		},
	}

	return cmd
}

func (cb *CleanBranches) execute() {
	branch := getDefaultBranch()
	gitCheckout(branch)
	gitPull()
	gitFetch()
	cleanBranches()
}

func getDefaultBranch() string {
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

func gitCheckout(branch string) {
	cmd := exec.Command("git", "checkout", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}


func gitPull() {
	cmd := exec.Command("git", "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func gitFetch() {
	cmd := exec.Command("git", "fetch", "-p")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func cleanBranches() {
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
