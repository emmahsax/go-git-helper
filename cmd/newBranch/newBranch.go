package newBranch

import (
	"fmt"

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
			newNewBranch(determineBranch(args), debug, executor.NewExecutor(debug)).execute()
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

func determineBranch(args []string) string {
	if len(args) == 0 {
		return askForBranch()
	} else {
		return args[0]
	}
}

func askForBranch() string {
	return commandline.AskOpenEndedQuestion("New branch name", false)
}

func (nb *NewBranch) execute() {
	fmt.Println("Attempting to create a new branch:", nb.Branch)
	g := git.NewGit(nb.Debug, nb.Executor)
	g.Pull()

	for {
		err := g.CreateBranch(nb.Branch)
		if err == nil {
			break
		}

		fmt.Println("--- Invalid branch ---")
		nb.Branch = askForBranch()
	}

	g.Checkout(nb.Branch)
	g.PushBranch(nb.Branch)
}
