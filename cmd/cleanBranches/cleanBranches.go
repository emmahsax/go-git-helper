package cleanBranches

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type CheckoutDefault struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "clean-branches",
		Short:                 "Switches to the default branch, git pulls, git fetches, and removes remote-deleted branches from your machine",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cd := &CheckoutDefault{}
			cd.execute()
			return nil
		},
	}

	return cmd
}

func (cd *CheckoutDefault) execute() {
	switchBranches()
	gitPull()
	gitFetch()
	cleanBranches()
}

func switchBranches() {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	branch := strings.SplitN(strings.TrimSpace(string(output)), "/", 4)
	if len(branch) != 4 {
		log.Fatal("Invalid branch format")
		return
	}

	checkoutCmd := exec.Command("git", "checkout", branch[3])
	output, err = checkoutCmd.CombinedOutput()
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
