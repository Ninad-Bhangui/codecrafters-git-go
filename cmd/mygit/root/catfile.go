package cmd

import (
	"github.com/codecrafters-io/git-starter-go/internal/catfile"
	"github.com/spf13/cobra"
)

var prettyPrint bool

var catfileCmd = &cobra.Command{
	Use:   "cat-file",
	Short: "cat git objects",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Flags().PrintDefaults()
		}
		catfile.CatFilePrettyPrint(args[0])
	},
}

func init() {
	rootCmd.AddCommand(catfileCmd)
	catfileCmd.Flags().BoolVarP(&prettyPrint, "pretty-print", "p", false, "pretty print")
}
