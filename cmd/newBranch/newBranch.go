package newBranch

import (
	"fmt"
	"regexp"

	"github.com/emmahsax/go-git-helper/internal/commandline"
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type NewBranch struct {
	Branch   string
	Debug    bool
	Executor executor.ExecutorInterface
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
			newNewBranch(determineBranch(args, debug), debug, executor.NewExecutor(debug)).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newNewBranch(branch string, debug bool, executor executor.ExecutorInterface) *NewBranch {
	return &NewBranch{
		Branch:   branch,
		Debug:    debug,
		Executor: executor,
	}
}

func determineBranch(args []string, debug bool) string {
	if len(args) == 0 {
		return getValidBranch()
	} else {
		if !isValidBranch(args[0]) {
			fmt.Println("--- Invalid branch provided ---")
			return getValidBranch()
		} else {
			return args[0]
		}
	}
}

func isValidBranch(branch string) bool {
	validPattern := "^[a-zA-Z0-9-_]+$"
	return regexp.MustCompile(validPattern).MatchString(branch)
}

func getValidBranch() string {
	var branch string

	for {
		branch = commandline.AskOpenEndedQuestion("New branch name", false)

		if isValidBranch(branch) {
			break
		}

		fmt.Println("--- Invalid branch ---")
	}

	return branch
}

func (nb *NewBranch) execute() {
	fmt.Println("Attempting to create a new branch:", nb.Branch)
	g := git.NewGit(nb.Debug, nb.Executor)
	g.Pull()
	g.CreateBranch(nb.Branch)
	g.Checkout(nb.Branch)
	g.PushBranch(nb.Branch)
}
