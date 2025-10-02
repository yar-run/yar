package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var scuttleCmd = &cobra.Command{
	Use:   "scuttle",
	Short: "Nuke the environment (danger)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("yar scuttle: destroying env…")
		return nil
	},
}

func init() { rootCmd.AddCommand(scuttleCmd) }
