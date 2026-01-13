package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage project configuration",
	Long:  `Manage project configuration at ./yar.yaml.`,
}

var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive guided setup",
	Long:  `Create a new yar.yaml with interactive guided setup.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("project init: starting interactive setup")
		fmt.Println("  [stub] would prompt for project name, environments, services")
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Display project configuration",
	Long:  `Display the current project configuration from ./yar.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("project get: displaying project config from ./yar.yaml")
		fmt.Println("  [stub] would load and display yar.yaml")
	},
}

var projectEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open project config in $EDITOR",
	Long:  `Open the project configuration file in your default editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("project edit: opening ./yar.yaml in $EDITOR")
		fmt.Println("  [stub] would open editor")
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectInitCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectEditCmd)
}
