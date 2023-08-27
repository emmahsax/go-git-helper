package newBranch

import (
	"fmt"
	"regexp"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type NewBranch struct {
	branch string
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "new-branch [optionalBranch]",
		Short:                 "Creates a new local branch and pushes to the remote",
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var branch string

			if len(args) == 0 {
				branch = getValidBranch()
			} else {
				if !isValidBranch(args[0]) {
					fmt.Println("--- Invalid branch provided ---")
					branch = getValidBranch()
				} else {
					branch = args[0]
				}
			}

			newBranch(branch).execute()
			return nil
		},
	}

	return cmd
}

func newBranch(branch string) *NewBranch {
	return &NewBranch{
		branch: branch,
	}
}

func isValidBranch(branch string) bool {
	validPattern := "^[a-zA-Z0-9-_]+$"
	return regexp.MustCompile(validPattern).MatchString(branch)
}

func getValidBranch() string {
	branch := commandline.AskOpenEndedQuestion("New branch name?", false)

	if !isValidBranch(branch) {
		fmt.Println("--- Invalid branch ---")
		return getValidBranch()
	}

	return branch
}

func (nb *NewBranch) execute() {
	fmt.Println("Attempting to create a new branch:", nb.branch)
	git.Pull()
	git.CreateBranch(nb.branch)
	git.Checkout(nb.branch)
	git.PushBranch(nb.branch)
}
