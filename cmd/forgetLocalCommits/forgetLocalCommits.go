package forgetLocalCommits

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

type ForgetLocalCommits struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "forget-local-commits",
		Short:                 "Forget all commits that aren't pushed to remote",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flc := &ForgetLocalCommits{}
			flc.execute()
			return nil
		},
	}

	return cmd
}

func (flc *ForgetLocalCommits) execute() {
	gitPull()
	gitReset()
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

func gitReset() {
	cmd := exec.Command("git", "reset", "--hard", "origin/HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}
