package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev" // set via -ldflags

var rootCmd = &cobra.Command{
	Use:   "yar",
	Short: "Yar — local ↔ Kubernetes fleet bootstrapper",
	Long:  "Yar bridges local dev with a prod-like cluster network: services, DNS, secrets, and packs.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("yar: it sails 🏴‍☠️  (try `yar up`)")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
}
