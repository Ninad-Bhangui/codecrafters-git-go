package cmd

import (
	"github.com/spf13/cobra"

	"github.com/codecrafters-io/git-starter-go/internal/committree"
)

var (
	parentHashes []string
	commitMsg    string
)

var committreeCmd = &cobra.Command{
	Use:   "commit-tree",
	Short: "Commit a tree",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Flags().PrintDefaults()
		}

		committree.InvokeCommitTree(args[0], commitMsg, parentHashes)
	},
}

func init() {
	rootCmd.AddCommand(committreeCmd)

	committreeCmd.Flags().StringVarP(&commitMsg, "message", "m", "", "commit message")
	committreeCmd.Flags().
		StringArrayVarP(&parentHashes, "parent", "p", []string{}, "parent hash")
}
