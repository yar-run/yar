package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Template flags
var (
	templateEnv        string
	templateFormat     string
	templateOutputDir  string
	templatePackage    bool
	templatePush       string
	templateValuesOnly bool
	templateLock       bool
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Generate deployment artifacts",
	Long:  `Generate deployment assets from packs (Helm charts, Compose files, K8s manifests).`,
}

var templateBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Generate Helm charts, Compose files, or K8s manifests",
	Long:  `Generate deployment artifacts from the project configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("template build: generating %s artifacts for environment '%s'\n", templateFormat, templateEnv)
		fmt.Printf("  output directory: %s\n", templateOutputDir)
		if templatePackage {
			fmt.Println("  [stub] would package Helm chart as .tgz")
		}
		if templatePush != "" {
			fmt.Printf("  [stub] would push to %s\n", templatePush)
		}
		fmt.Println("  [stub] would generate deployment artifacts")
	},
}

var templateRenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render templates to stdout (dry-run)",
	Long:  `Render templates to stdout without writing files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("template render: rendering templates for environment '%s'\n", templateEnv)
		fmt.Println("  [stub] would render and print templates")
	},
}

var templatePublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish charts to artifact repository",
	Long:  `Publish generated charts to an artifact repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("template publish: publishing charts")
		fmt.Println("  [stub] would publish to configured repository")
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)

	// template build flags
	templateBuildCmd.Flags().StringVar(&templateEnv, "env", "local", "Target environment")
	templateBuildCmd.Flags().StringVar(&templateFormat, "format", "helm", "Output format: helm, compose, manifest")
	templateBuildCmd.Flags().StringVar(&templateOutputDir, "output-dir", "./dist", "Output directory")
	templateBuildCmd.Flags().BoolVar(&templatePackage, "package", false, "Package Helm chart as .tgz")
	templateBuildCmd.Flags().StringVar(&templatePush, "push", "", "Push to OCI registry URL")
	templateBuildCmd.Flags().BoolVar(&templateValuesOnly, "values-only", false, "Only update values, no template changes")
	templateBuildCmd.Flags().BoolVar(&templateLock, "lock", false, "Lock dependency versions")
	templateCmd.AddCommand(templateBuildCmd)

	// template render flags
	templateRenderCmd.Flags().StringVar(&templateEnv, "env", "local", "Target environment")
	templateCmd.AddCommand(templateRenderCmd)

	// template publish
	templateCmd.AddCommand(templatePublishCmd)
}
