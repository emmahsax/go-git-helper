package checkoutDefault

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type CheckoutDefault struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "checkout-default",
		Short:                 "Switches to the default branch",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cd := &CheckoutDefault{}
			cd.execute()
			return nil
		},
	}

	return cmd
}

func (cd *CheckoutDefault) execute() {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	branch := strings.SplitN(strings.TrimSpace(string(output)), "/", 4)
	if len(branch) != 4 {
		log.Fatal("Invalid branch format")
		return
	}

	checkoutCmd := exec.Command("git", "checkout", branch[3])
	output, err = checkoutCmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}
