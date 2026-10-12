package version

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// Resolve returns the version to report. It prefers the version injected at
// build time, then the module version recorded by `go install`, and falls back
// to "dev" for local builds.
func Resolve(injected string) string {
	if injected != "" && injected != "dev" {
		return injected
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}

	return "dev"
}

func NewCommand(packageVersion string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "version",
		Short:                 "Print the version number",
		Args:                  cobra.ExactArgs(0),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("git-helper version %s\n", packageVersion)
			return nil
		},
	}

	return cmd
}
