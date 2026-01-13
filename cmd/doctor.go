package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Doctor flags
var (
	doctorFix      bool
	doctorFixCache bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose and repair environment issues",
	Long:  `Run health checks and optionally repair environment issues.`,
}

var doctorRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all health checks",
	Long: `Run all health checks (VPN, DNS, hosts, clusters, secrets).

Use --fix to attempt auto-repair of issues.
Use --fix-cache to clear caches and regenerate state.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("doctor run: running health checks")
		fmt.Println("")
		fmt.Println("  CHECK           STATUS    MESSAGE")
		fmt.Println("  docker          [stub]    checking Docker daemon...")
		fmt.Println("  colima          [stub]    checking Colima status...")
		fmt.Println("  vpn             [stub]    checking VPN connection...")
		fmt.Println("  hosts           [stub]    checking /etc/hosts entries...")
		fmt.Println("  secrets         [stub]    checking secret store access...")
		fmt.Println("")
		if doctorFix {
			fmt.Println("  --fix: would attempt to repair issues")
		}
		if doctorFixCache {
			fmt.Println("  --fix-cache: would clear caches and regenerate state")
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	doctorRunCmd.Flags().BoolVar(&doctorFix, "fix", false, "Attempt to auto-repair issues")
	doctorRunCmd.Flags().BoolVar(&doctorFixCache, "fix-cache", false, "Clear caches and regenerate state")
	doctorCmd.AddCommand(doctorRunCmd)
}
