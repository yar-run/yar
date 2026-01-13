package cmd

import (
	"github.com/spf13/cobra"
)

// Aliases provide ergonomic shortcuts for common operations

func init() {
	// yar up [env] -> yar fleet up [env]
	upCmd := &cobra.Command{
		Use:   "up [env]",
		Short: "Alias for 'fleet up'",
		Long:  `Start all services for environment (alias for 'fleet up').`,
		Args:  cobra.MaximumNArgs(1),
		Run:   fleetUpCmd.Run,
	}
	upCmd.Flags().AddFlagSet(fleetUpCmd.Flags())
	rootCmd.AddCommand(upCmd)

	// yar down [env] -> yar fleet down [env]
	downCmd := &cobra.Command{
		Use:   "down [env]",
		Short: "Alias for 'fleet down'",
		Long:  `Stop all services (alias for 'fleet down').`,
		Args:  cobra.MaximumNArgs(1),
		Run:   fleetDownCmd.Run,
	}
	rootCmd.AddCommand(downCmd)

	// yar hoist [env] -> yar fleet up [env]
	hoistCmd := &cobra.Command{
		Use:   "hoist [env]",
		Short: "Alias for 'fleet up' (nautical)",
		Long:  `Start all services for environment (alias for 'fleet up').`,
		Args:  cobra.MaximumNArgs(1),
		Run:   fleetUpCmd.Run,
	}
	hoistCmd.Flags().AddFlagSet(fleetUpCmd.Flags())
	rootCmd.AddCommand(hoistCmd)

	// yar dock [env] -> yar fleet down [env]
	dockCmd := &cobra.Command{
		Use:   "dock [env]",
		Short: "Alias for 'fleet down' (nautical)",
		Long:  `Stop all services (alias for 'fleet down').`,
		Args:  cobra.MaximumNArgs(1),
		Run:   fleetDownCmd.Run,
	}
	rootCmd.AddCommand(dockCmd)

	// yar scuttle [env] -> yar fleet destroy [env]
	scuttleCmd := &cobra.Command{
		Use:   "scuttle [env]",
		Short: "Alias for 'fleet destroy' (nautical)",
		Long:  `Stop and remove all services, networks, and volumes (alias for 'fleet destroy').`,
		Args:  cobra.MaximumNArgs(1),
		Run:   fleetDestroyCmd.Run,
	}
	scuttleCmd.Flags().AddFlagSet(fleetDestroyCmd.Flags())
	rootCmd.AddCommand(scuttleCmd)

	// yar swab -> yar doctor run --fix-cache
	swabCmd := &cobra.Command{
		Use:   "swab",
		Short: "Alias for 'doctor run --fix-cache' (nautical)",
		Long:  `Clear caches and regenerate state (alias for 'doctor run --fix-cache').`,
		Run: func(cmd *cobra.Command, args []string) {
			doctorFixCache = true
			doctorRunCmd.Run(cmd, args)
		},
	}
	rootCmd.AddCommand(swabCmd)
}
