package cleanBranches

import (
	"github.com/emmahsax/go-git-helper/internal/git"
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
			cleanBranches().execute()
			return nil
		},
	}

	return cmd
}

func cleanBranches() *CleanBranches {
	return &CleanBranches{}
}

func (cb *CleanBranches) execute() {
	branch := git.DefaultBranch()
	git.Checkout(branch)
	git.Pull()
	git.Fetch()
	git.CleanDeletedBranches()
}
