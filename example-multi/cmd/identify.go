package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var identifyCmd = &cobra.Command{
	Use:   "identify",
	Short: "identify image format",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	identifyCmd.Flags().StringP("format", "f", "", "image format")
	identifyCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.AddCommand(identifyCmd)

	carapace.Gen(identifyCmd).FlagCompletion(carapace.ActionMap{
		"format": carapace.ActionValues("png", "jpeg", "gif", "tiff", "bmp"),
	})

	carapace.Gen(identifyCmd).PositionalCompletion(
		carapace.ActionFiles(".png", ".jpg", ".jpeg", ".gif", ".tiff", ".bmp"),
	)
}
