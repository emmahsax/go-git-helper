package cleanBranches

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type CleanBranches struct{}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "clean-branches",
		Short:                 "Switches to the default branch, git pulls, git fetches, and removes remote-deleted branches from your machine",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cleanBranches().execute(debug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func cleanBranches() *CleanBranches {
	return &CleanBranches{}
}

func (cb *CleanBranches) execute(debug bool) {
	g := git.NewGitClient(debug)
	branch := g.DefaultBranch()
	g.Checkout(branch)
	g.Pull()
	g.Fetch()
	g.CleanDeletedBranches()
}
