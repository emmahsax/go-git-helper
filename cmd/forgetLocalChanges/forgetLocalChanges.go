package forgetLocalChanges

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type ForgetLocalChanges struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "forget-local-changes",
		Short:                 "Forget all changes that aren't committed",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			forgetLocalChanges().execute()
			return nil
		},
	}

	return cmd
}

func forgetLocalChanges() *ForgetLocalChanges {
	return &ForgetLocalChanges{}
}

func (flc *ForgetLocalChanges) execute() {
	git.Stash()
	git.StashDrop()
}
