package newBranch

import (
	"fmt"
	"regexp"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type NewBranch struct {
	Branch string
	Debug  bool
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

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

			newNewBranch(branch, debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newNewBranch(branch string, debug bool) *NewBranch {
	return &NewBranch{
		Branch: branch,
		Debug:  debug,
	}
}

func isValidBranch(branch string) bool {
	validPattern := "^[a-zA-Z0-9-_]+$"
	return regexp.MustCompile(validPattern).MatchString(branch)
}

func getValidBranch() string {
	branch := commandline.AskOpenEndedQuestion("New branch name", false)

	if !isValidBranch(branch) {
		fmt.Println("--- Invalid branch ---")
		return getValidBranch()
	}

	return branch
}

func (nb *NewBranch) execute() {
	fmt.Println("Attempting to create a new branch:", nb.Branch)
	g := git.NewGit(nb.Debug, nil)
	g.Pull()
	g.CreateBranch(nb.Branch)
	g.Checkout(nb.Branch)
	g.PushBranch(nb.Branch)
}
