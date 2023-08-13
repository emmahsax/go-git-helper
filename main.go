package main

import (
	"fmt"
	"os"

	"github.com/emmahsax/go-git-helper/cmd/changeRemote"
	"github.com/emmahsax/go-git-helper/cmd/checkoutDefault"
	"github.com/emmahsax/go-git-helper/cmd/cleanBranches"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := newCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-helper",
		Short: "Making it easier to work with git on the command line",
	}

	cmd.DisableAutoGenTag = true
	cmd.DisableFlagParsing = true
	cmd.DisableFlagsInUseLine = true

	cmd.AddCommand(changeRemote.NewCommand())
	cmd.AddCommand(checkoutDefault.NewCommand())
	cmd.AddCommand(cleanBranches.NewCommand())

	return cmd
}
