package cmd

import (
	"github.com/spf13/cobra"

	"github.com/codecrafters-io/git-starter-go/internal/writetree"
)

var writetreeCmd = &cobra.Command{
	Use:   "write-tree",
	Short: "Write tree",
	Run: func(cmd *cobra.Command, args []string) {
		writetree.WriteTree()
	},
}

func init() {
	rootCmd.AddCommand(writetreeCmd)
}
