package forgetLocalChanges

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type ForgetLocalChanges struct{}

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
			forgetLocalChanges().execute(debug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func forgetLocalChanges() *ForgetLocalChanges {
	return &ForgetLocalChanges{}
}

func (flc *ForgetLocalChanges) execute(debug bool) {
	g := git.NewGitClient(debug)
	g.Stash()
	g.StashDrop()
}
