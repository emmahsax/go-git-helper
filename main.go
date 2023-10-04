package main

import (
	"fmt"
	"os"

	"github.com/emmahsax/go-git-helper/cmd/changeRemote"
	"github.com/emmahsax/go-git-helper/cmd/checkoutDefault"
	"github.com/emmahsax/go-git-helper/cmd/cleanBranches"
	"github.com/emmahsax/go-git-helper/cmd/codeRequest"
	"github.com/emmahsax/go-git-helper/cmd/emptyCommit"
	"github.com/emmahsax/go-git-helper/cmd/forgetLocalChanges"
	"github.com/emmahsax/go-git-helper/cmd/forgetLocalCommits"
	"github.com/emmahsax/go-git-helper/cmd/newBranch"
	"github.com/emmahsax/go-git-helper/cmd/setHeadRef"
	"github.com/emmahsax/go-git-helper/cmd/setup"
	"github.com/emmahsax/go-git-helper/cmd/update"
	"github.com/emmahsax/go-git-helper/cmd/version"
	"github.com/spf13/cobra"
)

var (
	packageVersion = "beta-1.0.1"
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
		Short: "Making it easier to work with git on the command-line",
	}

	cmd.AddCommand(changeRemote.NewCommand())
	cmd.AddCommand(checkoutDefault.NewCommand())
	cmd.AddCommand(cleanBranches.NewCommand())
	cmd.AddCommand(codeRequest.NewCommand())
	cmd.AddCommand(emptyCommit.NewCommand())
	cmd.AddCommand(forgetLocalChanges.NewCommand())
	cmd.AddCommand(forgetLocalCommits.NewCommand())
	cmd.AddCommand(newBranch.NewCommand())
	cmd.AddCommand(setup.NewCommand())
	cmd.AddCommand(setHeadRef.NewCommand())
	cmd.AddCommand(update.NewCommand())
	cmd.AddCommand(version.NewCommand(packageVersion))

	return cmd
}
