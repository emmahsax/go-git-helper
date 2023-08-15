package forgetLocalCommits

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type ForgetLocalCommits struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "forget-local-commits",
		Short:                 "Forget all commits that aren't pushed to remote",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			forgetLocalCommits().execute()
			return nil
		},
	}

	return cmd
}

func forgetLocalCommits() *ForgetLocalCommits {
	return &ForgetLocalCommits{}
}

func (flc *ForgetLocalCommits) execute() {
	git.Pull()
	git.Reset()
}
