package cmd

import (
	"github.com/spf13/cobra"

	"github.com/codecrafters-io/git-starter-go/internal/lstree"
)

var nameOnly bool

var lstreeCmd = &cobra.Command{
	Use:   "ls-tree",
	Short: "List a tree",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Flags().PrintDefaults()
		}

		lstree.LsTree(args[0], nameOnly)
	},
}

func init() {
	rootCmd.AddCommand(lstreeCmd)
	lstreeCmd.Flags().BoolVar(&nameOnly, "name-only", false, "print name only")
}
