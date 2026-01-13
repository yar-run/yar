package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Manage service packs",
	Long:  `Manage portable service definitions (packs).`,
}

var packListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available packs",
	Long:  `List all available packs (built-in and installed).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pack list: listing available packs")
		fmt.Println("  NAME        VERSION    DESCRIPTION")
		fmt.Println("  redis       1.0.0      Redis in-memory data store")
		fmt.Println("  postgres    1.0.0      PostgreSQL database")
		fmt.Println("  kafka       1.0.0      Apache Kafka message broker")
		fmt.Println("  [stub] would load from packs/ directory and catalog")
	},
}

var packInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a pack from catalog",
	Long:  `Install a pack from the pack catalog.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pack install: installing pack '%s'\n", args[0])
		fmt.Println("  [stub] would download and install pack")
	},
}

var packRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an installed pack",
	Long:  `Remove a previously installed pack.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pack remove: removing pack '%s'\n", args[0])
		fmt.Println("  [stub] would remove pack from local storage")
	},
}

func init() {
	rootCmd.AddCommand(packCmd)
	packCmd.AddCommand(packListCmd)
	packCmd.AddCommand(packInstallCmd)
	packCmd.AddCommand(packRemoveCmd)
}
