package setHeadRef

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type SetHeadRef struct {
	defaultBranch string
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
			setHeadRef(args[0]).execute(debug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func setHeadRef(defaultBranch string) *SetHeadRef {
	return &SetHeadRef{
		defaultBranch: defaultBranch,
	}
}

func (shr *SetHeadRef) execute(debug bool) {
	g := git.NewGitClient(debug)
	g.SetHeadRef(shr.defaultBranch)
}
