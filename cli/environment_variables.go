package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"ytsruh.com/envoy/cli/config"
)

var environmentVariablesCmd = &cobra.Command{
	Use:   "variables",
	Short: "Manage environment variables",
	Long:  "Import, export, create, list, update, and delete environment variables",
}

var importVariablesCmd = &cobra.Command{
	Use:   "import [environment_id]",
	Short: "Import variables from .env file",
	Long:  "Import environment variables from .env file in current directory",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projectID, err := config.GetProjectID()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if projectID == 0 {
			fmt.Println("No current project set.")
			fmt.Println("Please set a project with:")
			fmt.Println("  envoy projects use <id>")
			os.Exit(1)
		}

		var environmentID int64
		if len(args) > 0 {
			environmentID, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			environmentID, err = config.GetEnvironmentID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if environmentID == 0 {
				fmt.Println("No current environment set.")
				fmt.Println("Please set an environment with:")
				fmt.Println("  envoy environments use <id>")
				fmt.Println("Or provide environment_id as an argument:")
				fmt.Println("  envoy variables import <environment_id>")
				os.Exit(1)
			}
		}

		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			fmt.Println("No .env file found in current directory")
			os.Exit(1)
		}

		variables, err := ParseEnvFile(".env")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to parse .env file: %v\n", err)
			os.Exit(1)
		}

		if len(variables) == 0 {
			fmt.Println("No variables found in .env file")
			return
		}

		fmt.Printf("Found %d variable(s) in .env file:\n\n", len(variables))
		for key, value := range variables {
			fmt.Printf("  %s=%s\n", key, value)
		}
		fmt.Println()

		confirmed, err := Confirm("Import these variables?")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if !confirmed {
			fmt.Println("Import cancelled")
			return
		}

		created := 0
		updated := 0
		for key, value := range variables {
			_, err := client.CreateEnvironmentVariable(projectID, environmentID, key, value)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to import variable %s: %v\n", key, err)
				continue
			}
			created++
		}

		fmt.Printf("Successfully imported %d variable(s)\n", created)
		if updated > 0 {
			fmt.Printf("Updated %d variable(s)\n", updated)
		}
	},
}

var exportVariablesCmd = &cobra.Command{
	Use:   "export [environment_id]",
	Short: "Export variables to .env file",
	Long:  "Export environment variables to .env file in current directory",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projectID, err := config.GetProjectID()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if projectID == 0 {
			fmt.Println("No current project set.")
			fmt.Println("Please set a project with:")
			fmt.Println("  envoy projects use <id>")
			os.Exit(1)
		}

		var environmentID int64
		if len(args) > 0 {
			environmentID, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			environmentID, err = config.GetEnvironmentID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if environmentID == 0 {
				fmt.Println("No current environment set.")
				fmt.Println("Please set an environment with:")
				fmt.Println("  envoy environments use <id>")
				fmt.Println("Or provide environment_id as an argument:")
				fmt.Println("  envoy variables export <environment_id>")
				os.Exit(1)
			}
		}

		variables, err := client.ListEnvironmentVariables(projectID, environmentID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to list variables: %v\n", err)
			if err == ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		if len(variables) == 0 {
			fmt.Println("No variables to export")
			return
		}

		if _, err := os.Stat(".env"); err == nil {
			fmt.Println("Warning: .env file already exists in current directory")
			confirmed, err := Confirm("Overwrite existing .env file?")
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if !confirmed {
				fmt.Println("Export cancelled")
				return
			}
		}

		variablesMap := make(map[string]string)
		for _, v := range variables {
			variablesMap[v.Key] = v.Value
		}

		if err := WriteEnvFile(".env", variablesMap); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to write .env file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Exported %d variable(s) to .env\n", len(variables))
	},
}

var createVariableCmd = &cobra.Command{
	Use:   "create [environment_id]",
	Short: "Create a new variable",
	Long:  "Create a new environment variable",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projectID, err := config.GetProjectID()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if projectID == 0 {
			fmt.Println("No current project set.")
			fmt.Println("Please set a project with:")
			fmt.Println("  envoy projects use <id>")
			os.Exit(1)
		}

		var environmentID int64
		if len(args) > 0 {
			environmentID, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			environmentID, err = config.GetEnvironmentID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if environmentID == 0 {
				fmt.Println("No current environment set.")
				fmt.Println("Please set an environment with:")
				fmt.Println("  envoy environments use <id>")
				fmt.Println("Or provide environment_id as an argument:")
				fmt.Println("  envoy variables create <environment_id>")
				os.Exit(1)
			}
		}

		key, err := PromptString("Variable key", true)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		value, err := PromptString("Variable value", true)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		variable, err := client.CreateEnvironmentVariable(projectID, environmentID, key, value)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to create variable: %v\n", err)
			if err == ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Variable created successfully!")
		fmt.Printf("  ID: %d\n", variable.ID)
		fmt.Printf("  Key: %s\n", variable.Key)
		fmt.Printf("  Value: %s\n", variable.Value)
	},
}

var listVariablesCmd = &cobra.Command{
	Use:   "list [environment_id]",
	Short: "List variables",
	Long:  "List all variables for an environment",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projectID, err := config.GetProjectID()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if projectID == 0 {
			fmt.Println("No current project set.")
			fmt.Println("Please set a project with:")
			fmt.Println("  envoy projects use <id>")
			os.Exit(1)
		}

		var environmentID int64
		if len(args) > 0 {
			environmentID, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			environmentID, err = config.GetEnvironmentID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if environmentID == 0 {
				fmt.Println("No current environment set.")
				fmt.Println("Please set an environment with:")
				fmt.Println("  envoy environments use <id>")
				fmt.Println("Or provide environment_id as an argument:")
				fmt.Println("  envoy variables list <environment_id>")
				os.Exit(1)
			}
		}

		variables, err := client.ListEnvironmentVariables(projectID, environmentID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to list variables: %v\n", err)
			if err == ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		if len(variables) == 0 {
			fmt.Println("No variables found")
			return
		}

		fmt.Printf("Found %d variable(s):\n\n", len(variables))
		for _, v := range variables {
			fmt.Printf("  ID: %d\n", v.ID)
			fmt.Printf("  Key: %s\n", v.Key)
			fmt.Printf("  Value: %s\n", v.Value)
			fmt.Printf("  Updated: %s\n", v.UpdatedAt)
			fmt.Println()
		}
	},
}

var getVariableCmd = &cobra.Command{
	Use:   "get <variable_id> [environment_id]",
	Short: "Get variable details",
	Long:  "Get detailed information about a specific variable",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		variableID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Invalid variable ID: %v\n", err)
			os.Exit(1)
		}

		client, err := RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projectID, err := config.GetProjectID()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if projectID == 0 {
			fmt.Println("No current project set.")
			fmt.Println("Please set a project with:")
			fmt.Println("  envoy projects use <id>")
			os.Exit(1)
		}

		var environmentID int64
		if len(args) > 1 {
			environmentID, err = strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			environmentID, err = config.GetEnvironmentID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if environmentID == 0 {
				fmt.Println("No current environment set.")
				fmt.Println("Please set an environment with:")
				fmt.Println("  envoy environments use <id>")
				fmt.Println("Or provide environment_id as an argument:")
				fmt.Println("  envoy variables get <variable_id> <environment_id>")
				os.Exit(1)
			}
		}

		variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to get variable: %v\n", err)
			if err == ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Variable Details:")
		fmt.Printf("  ID: %d\n", variable.ID)
		fmt.Printf("  Key: %s\n", variable.Key)
		fmt.Printf("  Value: %s\n", variable.Value)
		fmt.Printf("  Environment ID: %d\n", variable.EnvironmentID)
		fmt.Printf("  Created: %s\n", variable.CreatedAt)
		fmt.Printf("  Updated: %s\n", variable.UpdatedAt)
	},
}

var updateVariableCmd = &cobra.Command{
	Use:   "update <variable_id> [environment_id]",
	Short: "Update a variable",
	Long:  "Update variable key and value",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		variableID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Invalid variable ID: %v\n", err)
			os.Exit(1)
		}

		client, err := RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projectID, err := config.GetProjectID()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if projectID == 0 {
			fmt.Println("No current project set.")
			fmt.Println("Please set a project with:")
			fmt.Println("  envoy projects use <id>")
			os.Exit(1)
		}

		var environmentID int64
		if len(args) > 1 {
			environmentID, err = strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			environmentID, err = config.GetEnvironmentID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if environmentID == 0 {
				fmt.Println("No current environment set.")
				fmt.Println("Please set an environment with:")
				fmt.Println("  envoy environments use <id>")
				fmt.Println("Or provide environment_id as an argument:")
				fmt.Println("  envoy variables update <variable_id> <environment_id>")
				os.Exit(1)
			}
		}

		variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to get variable: %v\n", err)
			if err == ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		key, err := PromptStringWithDefault("Variable key", variable.Key)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		value, err := PromptString("Variable value", true)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		updatedVariable, err := client.UpdateEnvironmentVariable(projectID, environmentID, variableID, key, value)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to update variable: %v\n", err)
			if err == ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Variable updated successfully!")
		fmt.Printf("  Key: %s\n", updatedVariable.Key)
		fmt.Printf("  Value: %s\n", updatedVariable.Value)
	},
}

var deleteVariableCmd = &cobra.Command{
	Use:   "delete <variable_id> [environment_id]",
	Short: "Delete a variable",
	Long:  "Delete a variable permanently",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		variableID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Invalid variable ID: %v\n", err)
			os.Exit(1)
		}

		client, err := RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projectID, err := config.GetProjectID()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if projectID == 0 {
			fmt.Println("No current project set.")
			fmt.Println("Please set a project with:")
			fmt.Println("  envoy projects use <id>")
			os.Exit(1)
		}

		var environmentID int64
		if len(args) > 1 {
			environmentID, err = strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			environmentID, err = config.GetEnvironmentID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if environmentID == 0 {
				fmt.Println("No current environment set.")
				fmt.Println("Please set an environment with:")
				fmt.Println("  envoy environments use <id>")
				fmt.Println("Or provide environment_id as an argument:")
				fmt.Println("  envoy variables delete <variable_id> <environment_id>")
				os.Exit(1)
			}
		}

		variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to get variable: %v\n", err)
			if err == ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Printf("Are you sure you want to delete variable '%s' (ID: %d)?\n", variable.Key, variable.ID)
		confirmed, err := Confirm("This action cannot be undone")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if !confirmed {
			fmt.Println("Operation cancelled")
			return
		}

		if err := client.DeleteEnvironmentVariable(projectID, environmentID, variableID); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to delete variable: %v\n", err)
			if err == ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Variable deleted successfully")
	},
}

func init() {
	environmentVariablesCmd.AddCommand(importVariablesCmd)
	environmentVariablesCmd.AddCommand(exportVariablesCmd)
	environmentVariablesCmd.AddCommand(createVariableCmd)
	environmentVariablesCmd.AddCommand(listVariablesCmd)
	environmentVariablesCmd.AddCommand(getVariableCmd)
	environmentVariablesCmd.AddCommand(updateVariableCmd)
	environmentVariablesCmd.AddCommand(deleteVariableCmd)
}
