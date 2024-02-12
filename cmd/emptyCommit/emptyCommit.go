package emptyCommit

import (
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type EmptyCommit struct {
	Debug    bool
	Executor executor.ExecutorInterface
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "empty-commit",
		Short:                 "Creates an empty commit",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newEmptyCommit(debug, executor.NewExecutor(debug)).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newEmptyCommit(debug bool, executor executor.ExecutorInterface) *EmptyCommit {
	return &EmptyCommit{
		Debug:    debug,
		Executor: executor,
	}
}

func (ec *EmptyCommit) execute() {
	git.NewGit(ec.Debug, ec.Executor).CreateEmptyCommit()
}
