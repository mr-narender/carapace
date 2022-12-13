package cmd

import (
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

var flagCmd = &cobra.Command{
	Use:     "flag",
	Short:   "flag example",
	GroupID: "main",
	Run:     func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(flagCmd)

	flagCmd.Flags().CountP("count", "c", "count flag")
	flagCmd.Flags().StringP("optarg", "o", "", "optional argument")

	flagCmd.Flag("optarg").NoOptDefVal = " "

	carapace.Gen(flagCmd).FlagCompletion(carapace.ActionMap{
		"optarg": carapace.ActionValues("optarg1", "optarg2", "optarg3"),
	})

	carapace.Gen(flagCmd).PositionalCompletion(
		carapace.ActionValues("positional1", "p1", "positional1 with space"),
		carapace.ActionValues("positional2", "p2", "positional2 with space"),
	)
}