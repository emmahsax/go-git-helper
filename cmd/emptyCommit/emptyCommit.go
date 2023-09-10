package emptyCommit

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type EmptyCommit struct {
	Debug bool
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
			newEmptyCommitClient(debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newEmptyCommitClient(debug bool) *EmptyCommit {
	return &EmptyCommit{
		Debug: debug,
	}
}

func (ec *EmptyCommit) execute() {
	git.NewGitClient(ec.Debug).CreateEmptyCommit()
}
