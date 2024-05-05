package cmd

import (
	"github.com/spf13/cobra"

	"github.com/codecrafters-io/git-starter-go/internal/hashobject"
)

var writeMode bool

var hashobjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "Hash file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Flags().PrintDefaults()
		}
		hashobject.HashObject(args[0], writeMode)
	},
}

func init() {
	rootCmd.AddCommand(hashobjectCmd)
	hashobjectCmd.Flags().BoolVarP(&writeMode, "write-mode", "w", false, "write to object file")
}
