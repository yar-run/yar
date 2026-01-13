package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Fleet flags
var (
	fleetDetach        bool
	fleetBuild         bool
	fleetForceRecreate bool
	fleetKeepVolumes   bool
	fleetForce         bool
)

var fleetCmd = &cobra.Command{
	Use:   "fleet",
	Short: "Manage the fleet of services",
	Long:  `Manage the fleet of services defined in yar.yaml.`,
}

var fleetUpCmd = &cobra.Command{
	Use:   "up [env]",
	Short: "Start all services for environment",
	Long: `Start all services for the specified environment (default: local).

Bootstraps Colima/VPN/DNS, validates secrets, and starts containers
in dependency order.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := "local"
		if len(args) > 0 {
			env = args[0]
		}
		fmt.Printf("fleet up: starting services for environment '%s'\n", env)
		fmt.Println("  [stub] would validate secrets, start containers in dependency order")
	},
}

var fleetDownCmd = &cobra.Command{
	Use:   "down [env]",
	Short: "Stop all services",
	Long:  `Stop all services. Containers are stopped but not removed.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := "local"
		if len(args) > 0 {
			env = args[0]
		}
		fmt.Printf("fleet down: stopping services for environment '%s'\n", env)
		fmt.Println("  [stub] would stop containers in reverse dependency order")
	},
}

var fleetDestroyCmd = &cobra.Command{
	Use:   "destroy [env]",
	Short: "Stop and remove all services, networks, and volumes",
	Long:  `Stop and remove all services, networks, and volumes.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := "local"
		if len(args) > 0 {
			env = args[0]
		}
		fmt.Printf("fleet destroy: destroying all resources for environment '%s'\n", env)
		if fleetKeepVolumes {
			fmt.Println("  [stub] would keep volumes")
		}
		fmt.Println("  [stub] would remove containers, networks, volumes")
	},
}

var fleetRestartCmd = &cobra.Command{
	Use:   "restart [env]",
	Short: "Restart all services",
	Long:  `Restart all services, applying any config changes.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := "local"
		if len(args) > 0 {
			env = args[0]
		}
		fmt.Printf("fleet restart: restarting services for environment '%s'\n", env)
		fmt.Println("  [stub] would restart containers")
	},
}

var fleetStatusCmd = &cobra.Command{
	Use:   "status [env]",
	Short: "Show status of all services",
	Long:  `Show status of all services (running, stopped, health).`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := "local"
		if len(args) > 0 {
			env = args[0]
		}
		fmt.Printf("fleet status: showing status for environment '%s'\n", env)
		fmt.Println("  SERVICE              STATUS    REPLICAS    ENDPOINTS")
		fmt.Println("  [stub] no services configured")
	},
}

var fleetUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update yar binary and pack catalog",
	Long:  `Update the yar binary and refresh the pack catalog.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("fleet update: checking for updates")
		fmt.Println("  [stub] would check for new yar version and update pack catalog")
	},
}

func init() {
	rootCmd.AddCommand(fleetCmd)

	// fleet up
	fleetUpCmd.Flags().BoolVar(&fleetDetach, "detach", true, "Run in background")
	fleetUpCmd.Flags().BoolVar(&fleetBuild, "build", false, "Build images before starting")
	fleetUpCmd.Flags().BoolVar(&fleetForceRecreate, "force-recreate", false, "Recreate containers even if unchanged")
	fleetCmd.AddCommand(fleetUpCmd)

	// fleet down
	fleetCmd.AddCommand(fleetDownCmd)

	// fleet destroy
	fleetDestroyCmd.Flags().BoolVar(&fleetKeepVolumes, "keep-volumes", false, "Don't remove volumes")
	fleetDestroyCmd.Flags().BoolVar(&fleetForce, "force", false, "Skip confirmation prompt")
	fleetCmd.AddCommand(fleetDestroyCmd)

	// fleet restart
	fleetCmd.AddCommand(fleetRestartCmd)

	// fleet status
	fleetCmd.AddCommand(fleetStatusCmd)

	// fleet update
	fleetCmd.AddCommand(fleetUpdateCmd)
}
