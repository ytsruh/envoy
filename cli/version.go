package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"ytsruh.com/envoy/shared"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", shared.Version)
	},
}
