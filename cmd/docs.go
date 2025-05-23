/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "View your project documentation",
	Long:  `Navigate & view your project documentation in your terminal. Markdown files in the .docs directory are automatically parsed and displayed. Alternatively set your own documentation directory using the --dir flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Description:")
		fmt.Println(cmd.Long)
	},
}

func init() {
	docsCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.AddCommand(docsCmd)
}
