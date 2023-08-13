package emptyCommit

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

type EmptyCommit struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "empty-commit",
		Short:                 "Creates an empty commit",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ec := &EmptyCommit{}
			ec.execute()
			return nil
		},
	}

	return cmd
}

func (ec *EmptyCommit) execute() {
	gitEmptyCommit()
}

func gitEmptyCommit() {
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "Empty commit")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}
