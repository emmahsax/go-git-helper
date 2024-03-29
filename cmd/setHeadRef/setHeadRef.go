package setHeadRef

import (
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type SetHeadRef struct {
	Debug         bool
	DefaultBranch string
	Executor      executor.ExecutorInterface
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "set-head-ref [defaultBranch]",
		Short:                 "Sets the HEAD ref as a symbolic ref",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newSetHeadRef(args[0], debug, executor.NewExecutor(debug)).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newSetHeadRef(defaultBranch string, debug bool, executor executor.ExecutorInterface) *SetHeadRef {
	return &SetHeadRef{
		Debug:         debug,
		DefaultBranch: defaultBranch,
		Executor:      executor,
	}
}

func (shr *SetHeadRef) execute() {
	g := git.NewGit(shr.Debug, shr.Executor)
	g.SetHeadRef(shr.DefaultBranch)
}
