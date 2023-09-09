package checkoutDefault

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type CheckoutDefault struct{}

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
			checkoutDefault().execute(debug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enables debug mode")

	return cmd
}

func checkoutDefault() *CheckoutDefault {
	return &CheckoutDefault{}
}

func (cd *CheckoutDefault) execute(debug bool) {
	g := git.NewGitClient(debug)
	branch := g.DefaultBranch()
	g.Checkout(branch)
}
