package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "example-multi",
	Short: "multi-completer example",
}

func init() {
	carapace.Gen(rootCmd, carapace.WithSubcommands(identifyCmd, convertCmd))
}

// Execute executes cmd.
func Execute() error {
	return carapace.Gen(rootCmd).Execute()
}
