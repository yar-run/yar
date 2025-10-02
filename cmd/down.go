package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Lower the sails (stop local fleet)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("yar down: stopping local fleet…")
		return nil
	},
}

func init() { rootCmd.AddCommand(downCmd) }
