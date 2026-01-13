package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yar-run/yar/internal/config"
	"gopkg.in/yaml.v3"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		loader := config.NewLoader()

		cfg, err := loader.LoadGlobal()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get config path for display
		path, _ := loader.GlobalPath()

		switch outputFormat {
		case "json":
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			fmt.Println(string(data))

		case "yaml":
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			fmt.Print(string(data))

		case "table":
			// For table format, show a summary
			fmt.Printf("# Global configuration from %s\n", path)
			fmt.Printf("container: %s\n", cfg.Container)
			if cfg.Hosts != nil {
				fmt.Printf("hosts.mode: %s\n", cfg.Hosts.Mode)
			}
			if cfg.Network != nil {
				fmt.Printf("network.name: %s\n", cfg.Network.Name)
				fmt.Printf("network.cidr: %s\n", cfg.Network.CIDR)
			}
			if cfg.Secrets != nil && cfg.Secrets.Local != nil {
				fmt.Printf("secrets.local.provider: %s\n", cfg.Secrets.Local.Provider)
			}
			if len(cfg.Clusters) > 0 {
				fmt.Printf("clusters: %d configured\n", len(cfg.Clusters))
			}

		default:
			// Default to yaml
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			fmt.Print(string(data))
		}

		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open global config in $EDITOR",
	Long:  `Open the global configuration file in your default editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		loader := config.NewLoader()
		path, err := loader.GlobalPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		fmt.Printf("config edit: opening %s in $EDITOR\n", path)
		fmt.Println("  [stub] would open editor")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configEditCmd)
}
