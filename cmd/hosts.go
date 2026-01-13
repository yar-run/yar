package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var hostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "Manage /etc/hosts entries",
	Long: `Manage /etc/hosts entries for container name resolution.

All entries are marked with '# yar:managed' for safe cleanup.`,
}

var hostsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List yar-managed host entries",
	Long:  `List all host entries managed by yar.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hosts list: listing yar-managed entries")
		fmt.Println("  NAME                    IP")
		fmt.Println("  [stub] would read /etc/hosts and filter yar:managed entries")
	},
}

var hostsSetCmd = &cobra.Command{
	Use:   "set <name> <ip>",
	Short: "Add or update a host entry",
	Long:  `Add or update a host entry in /etc/hosts.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name, ip := args[0], args[1]
		fmt.Printf("hosts set: setting %s -> %s\n", name, ip)
		fmt.Println("  [stub] would update /etc/hosts (may require sudo)")
	},
}

var hostsGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Show a host entry",
	Long:  `Show a specific host entry.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("hosts get: getting entry for '%s'\n", args[0])
		fmt.Println("  [stub] would look up in /etc/hosts")
	},
}

var hostsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Remove a host entry",
	Long:  `Remove a host entry from /etc/hosts.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("hosts delete: removing entry for '%s'\n", args[0])
		fmt.Println("  [stub] would remove from /etc/hosts (may require sudo)")
	},
}

func init() {
	rootCmd.AddCommand(hostsCmd)
	hostsCmd.AddCommand(hostsListCmd)
	hostsCmd.AddCommand(hostsSetCmd)
	hostsCmd.AddCommand(hostsGetCmd)
	hostsCmd.AddCommand(hostsDeleteCmd)
}
