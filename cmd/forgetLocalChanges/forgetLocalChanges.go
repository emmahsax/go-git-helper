package forgetLocalChanges

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

type ForgetLocalChanges struct{}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "forget-local-changes",
		Short:                 "Forget all changes that aren't committed",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flc := &ForgetLocalChanges{}
			flc.execute()
			return nil
		},
	}

	return cmd
}

func (flc *ForgetLocalChanges) execute() {
	gitStash()
	gitStashDrop()
}

func gitStash() {
	cmd := exec.Command("git", "stash")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}

func gitStashDrop() {
	cmd := exec.Command("git", "stash", "drop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("%s", string(output))
}
