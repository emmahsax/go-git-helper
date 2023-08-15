package emptyCommit

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type EmptyCommit struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "empty-commit",
		Short:                 "Creates an empty commit",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			emptyCommit().execute()
			return nil
		},
	}

	return cmd
}

func emptyCommit() *EmptyCommit {
	return &EmptyCommit{}
}

func (ec *EmptyCommit) execute() {
	git.CreateEmptyCommit()
}
