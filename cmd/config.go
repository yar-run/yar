package cmd

import (
	"fmt"

	"github.com/yar-run/yar/internal/platform"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage global configuration",
	Long:  `Manage machine-wide configuration at ~/.config/yar/config.yaml.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Display global configuration",
	Long:  `Display the current global configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		configDir, err := platform.ConfigDir()
		if err != nil {
			fmt.Printf("config get: error getting config dir: %v\n", err)
			return
		}
		fmt.Printf("config get: displaying global config from %s/config.yaml\n", configDir)
		fmt.Println("  [stub] would load and display config.yaml")
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open global config in $EDITOR",
	Long:  `Open the global configuration file in your default editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		configDir, err := platform.ConfigDir()
		if err != nil {
			fmt.Printf("config edit: error getting config dir: %v\n", err)
			return
		}
		fmt.Printf("config edit: opening %s/config.yaml in $EDITOR\n", configDir)
		fmt.Println("  [stub] would open editor")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configEditCmd)
}
