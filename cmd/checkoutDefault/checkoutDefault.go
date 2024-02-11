package checkoutDefault

import (
	"github.com/emmahsax/go-git-helper/internal/executor"
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type CheckoutDefault struct {
	Debug    bool
	Executor executor.ExecutorInterface
}

func NewCommand() *cobra.Command {
	var (
		debug bool
	)

	cmd := &cobra.Command{
		Use:                   "checkout-default",
		Short:                 "Switches to the default branch",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			newCheckoutDefault(debug).execute()
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func newCheckoutDefault(debug bool) *CheckoutDefault {
	return &CheckoutDefault{
		Debug:    debug,
		Executor: executor.NewExecutor(debug),
	}
}

func (cd *CheckoutDefault) execute() {
	g := git.NewGit(cd.Debug, cd.Executor)
	branch := g.DefaultBranch()
	g.Checkout(branch)
}
