package cmd

import (
	"github.com/benluddy/depster/cmd/internal/dot"
	"github.com/benluddy/depster/cmd/internal/resolve"
	"github.com/benluddy/depster/internal/version"
	"github.com/spf13/cobra"
)

func Execute() error {
	root := &cobra.Command{
		Use:     "depster",
		Short:   "Depster is a command-line interface to operator dependency resolution.",
		Version: version.Version,
	}
	root.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	resolve.AddTo(root)
	dot.AddTo(root)
	return root.Execute()
}
