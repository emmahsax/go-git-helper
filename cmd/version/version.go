package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand(packageVersion string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "version",
		Short:                 "Print the version number",
		Args:                  cobra.ExactArgs(0),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("git-helper version %s\n", packageVersion)
			return nil
		},
	}

	return cmd
}
