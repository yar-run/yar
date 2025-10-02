package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Hoist your services (start local fleet & networking)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("yar up: starting local fleet…")
		// TODO: bootstrap VPN/DNS/hosts, bring packs up
		return nil
	},
}

func init() { rootCmd.AddCommand(upCmd) }
