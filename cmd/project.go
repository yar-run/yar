package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yar-run/yar/internal/config"
	"github.com/yar-run/yar/internal/editor"
	"github.com/yar-run/yar/internal/errors"
	"gopkg.in/yaml.v3"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		loader := config.NewLoader()

		proj, err := loader.LoadProject()
		if err != nil {
			// Check if it's a NotFoundError for better UX
			if nfErr, ok := err.(*errors.NotFoundError); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", nfErr.Message)
				fmt.Fprintf(os.Stderr, "\nNo yar.yaml found in current directory or any parent.\n")
				fmt.Fprintf(os.Stderr, "Run 'yar project init' to create one.\n")
				os.Exit(2) // Configuration error exit code
			}
			return fmt.Errorf("failed to load project: %w", err)
		}

		// Get project path for display
		path, _ := loader.ProjectPath()

		switch outputFormat {
		case "json":
			data, err := json.MarshalIndent(proj, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal project: %w", err)
			}
			fmt.Println(string(data))

		case "yaml":
			data, err := yaml.Marshal(proj)
			if err != nil {
				return fmt.Errorf("failed to marshal project: %w", err)
			}
			fmt.Print(string(data))

		case "table":
			// For table format, show a summary
			fmt.Printf("# Project configuration from %s\n\n", path)
			fmt.Printf("project: %s\n", proj.Project)
			fmt.Printf("\nenvironments:\n")
			for name, env := range proj.Environments {
				fmt.Printf("  %s: cluster=%s, secrets=%s\n", name, env.Cluster, env.Secrets)
			}
			fmt.Printf("\nservices: (%d total)\n", len(proj.Services))
			for _, svc := range proj.Services {
				deps := ""
				if len(svc.Requires) > 0 {
					deps = fmt.Sprintf(" (requires: %v)", svc.Requires)
				}
				fmt.Printf("  - %s [%s]%s\n", svc.Name, svc.Pack, deps)
			}

		default:
			// Default to yaml
			data, err := yaml.Marshal(proj)
			if err != nil {
				return fmt.Errorf("failed to marshal project: %w", err)
			}
			fmt.Print(string(data))
		}

		return nil
	},
}

var projectEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open project config in $EDITOR",
	Long:  `Open the project configuration file in your default editor.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		loader := config.NewLoader()
		path, err := loader.ProjectPath()
		if err != nil {
			if _, ok := err.(*errors.NotFoundError); ok {
				fmt.Fprintf(os.Stderr, "Error: no yar.yaml found\n")
				fmt.Fprintf(os.Stderr, "Run 'yar project init' to create one.\n")
				os.Exit(2)
			}
			return fmt.Errorf("failed to find project config: %w", err)
		}

		// Open in editor
		if err := editor.OpenInEditor(path); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectInitCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectEditCmd)
}
