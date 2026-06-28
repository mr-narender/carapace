package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "convert image format",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	convertCmd.Flags().StringP("output", "o", "", "output format")
	convertCmd.Flags().IntP("quality", "q", 0, "output quality")
	rootCmd.AddCommand(convertCmd)

	carapace.Gen(convertCmd).FlagCompletion(carapace.ActionMap{
		"output":  carapace.ActionValues("png", "jpeg", "gif", "tiff", "bmp"),
		"quality": carapace.ActionValues("1", "50", "75", "90", "100"),
	})

	carapace.Gen(convertCmd).PositionalCompletion(
		carapace.ActionFiles(".png", ".jpg", ".jpeg", ".gif", ".tiff", ".bmp"),
	)
}
