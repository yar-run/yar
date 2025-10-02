package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var swabCmd = &cobra.Command{
	Use:   "swab",
	Short: "Clean caches/volumes/temp (housekeeping)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("yar swab: cleaning workspace…")
		return nil
	},
}

func init() { rootCmd.AddCommand(swabCmd) }
