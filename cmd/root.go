package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev" // set via -ldflags

// Global flags
var (
	verbose      bool
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "yar",
	Short: "Yar — local <-> Kubernetes fleet bootstrapper",
	Long: `Yar bridges local development with production Kubernetes clusters.

Run services locally with Docker Compose or deploy to Kubernetes with the
same configuration. Secrets are never stored in files - they're resolved
at runtime from secure providers.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: yaml, json, table")
}
