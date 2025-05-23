package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "envman",
	Short: "A simple environment management tool",
	Long:  `EnvMan is a command-line tool designed to simplify the management of your project environment. It provides tools to store & manage environment variables & project documentation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to EnvMan. The following commands are available:")
		for _, cmd := range cmd.Commands() {
			if cmd.Hidden {
				continue
			}
			fmt.Printf("%s"+" - %s\n", cmd.Name(), cmd.Short)
		}
	},
}

// Custom completion command to override the auto generated one
var completionCmd = &cobra.Command{
	Use:    "completion",
	Hidden: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
