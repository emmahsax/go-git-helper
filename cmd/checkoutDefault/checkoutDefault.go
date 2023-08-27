package checkoutDefault

import (
	"github.com/emmahsax/go-git-helper/internal/git"
	"github.com/spf13/cobra"
)

type CheckoutDefault struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "checkout-default",
		Short:                 "Switches to the default branch",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			checkoutDefault().execute()
			return nil
		},
	}

	return cmd
}

func checkoutDefault() *CheckoutDefault {
	return &CheckoutDefault{}
}

func (cd *CheckoutDefault) execute() {
	branch := git.DefaultBranch()
	git.Checkout(branch)
}
