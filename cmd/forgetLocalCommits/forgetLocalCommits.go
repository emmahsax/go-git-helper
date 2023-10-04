package forgetLocalCommits

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type ForgetLocalCommits struct {
	Debug bool
}

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
			newForgetLocalCommits(debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newForgetLocalCommits(debug bool) *ForgetLocalCommits {
	return &ForgetLocalCommits{
		Debug: debug,
	}
}

func (flc *ForgetLocalCommits) execute() {
	g := git.NewGit(flc.Debug)
	g.Pull()
	g.Reset()
}
