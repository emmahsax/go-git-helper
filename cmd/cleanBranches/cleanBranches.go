package cleanBranches

import (
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type CleanBranches struct {
	Debug    bool
	Executor executor.ExecutorInterface
}

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
			newCleanBranches(debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newCleanBranches(debug bool) *CleanBranches {
	return &CleanBranches{
		Debug:    debug,
		Executor: executor.NewExecutor(debug),
	}
}

func (cb *CleanBranches) execute() {
	g := git.NewGit(cb.Debug, cb.Executor)
	branch := g.DefaultBranch()
	g.Checkout(branch)
	g.Pull()
	g.Fetch()
	g.CleanDeletedBranches()
}
