package forgetLocalCommits

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type ForgetLocalCommits struct{}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "forget-local-commits",
		Short:                 "Forget all commits that aren't pushed to remote",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			forgetLocalCommits().execute(debug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func forgetLocalCommits() *ForgetLocalCommits {
	return &ForgetLocalCommits{}
}

func (flc *ForgetLocalCommits) execute(debug bool) {
	g := git.NewGitClient(debug)
	g.Pull()
	g.Reset()
}
