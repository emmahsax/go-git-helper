package forgetLocalChanges

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type ForgetLocalChanges struct {
	Debug bool
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "forget-local-changes",
		Short:                 "Forget all changes that aren't committed",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newForgetLocalChanges(debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newForgetLocalChanges(debug bool) *ForgetLocalChanges {
	return &ForgetLocalChanges{
		Debug: debug,
	}
}

func (flc *ForgetLocalChanges) execute() {
	g := git.NewGit(flc.Debug, nil)
	g.Stash()
	g.StashDrop()
}
