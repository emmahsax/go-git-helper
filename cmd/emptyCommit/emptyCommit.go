package emptyCommit

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type EmptyCommit struct{}

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
			emptyCommit().execute(debug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func emptyCommit() *EmptyCommit {
	return &EmptyCommit{}
}

func (ec *EmptyCommit) execute(debug bool) {
	git.NewGitClient(debug).CreateEmptyCommit()
}
