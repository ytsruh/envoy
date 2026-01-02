package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	shared "ytsruh.com/envoy/shared"
)

var RootCmd = &cobra.Command{
	Use:   "envoy",
	Short: "Envoy CLI client",
	Long:  "Envoy CLI client for managing projects and environments",
	CompletionOptions: cobra.CompletionOptions{
		//HiddenDefaultCmd:  true, // hides cmd
		DisableDefaultCmd: true, // removes cmd
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", shared.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(projectsCmd)
	RootCmd.AddCommand(environmentsCmd)
	RootCmd.AddCommand(environmentVariablesCmd)
}
