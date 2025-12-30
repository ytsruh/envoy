package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"ytsruh.com/envoy/cli/config"
	"ytsruh.com/envoy/cli/controllers"
	shared "ytsruh.com/envoy/shared"
)

var environmentsCmd = &cobra.Command{
	Use:   "environments",
	Short: "Manage environments",
	Long:  "Create, list, update, and delete environments within projects",
}

var createEnvironmentCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new environment",
	Long:  "Create a new environment within the current project",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
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

		project, err := client.GetProject(projectID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to get project: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Printf("Creating environment for project: %s (ID: %d)\n", project.Name, project.ID)
		confirmed, err := Confirm("Is this correct?")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if !confirmed {
			fmt.Println("Operation cancelled")
			return
		}

		name, err := PromptString("Environment name", true)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		description, err := PromptString("Description (optional)", false)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		environment, err := client.CreateEnvironment(projectID, name, description)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to create environment: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Environment created successfully!")
		fmt.Printf("  ID: %d\n", shared.ProjectIDToInt64(environment.ID))
		fmt.Printf("  Name: %s\n", environment.Name)
		if environment.Description != nil {
			fmt.Printf("  Description: %s\n", *environment.Description)
		}
		fmt.Printf("  Project ID: %d\n", shared.ProjectIDToInt64(environment.ProjectID))
	},
}

var listEnvironmentsCmd = &cobra.Command{
	Use:   "list [project_id]",
	Short: "List environments",
	Long:  "List all environments for a project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		currentEnvironmentID, _ := config.GetEnvironmentID()

		var projectID int64
		if len(args) > 0 {
			projectID, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid project ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			projectID, err = config.GetProjectID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if projectID == 0 {
				fmt.Println("No current project set.")
				fmt.Println("Please set a project with:")
				fmt.Println("  envoy projects use <id>")
				fmt.Println("Or provide project_id as an argument:")
				fmt.Println("  envoy environments list <project_id>")
				os.Exit(1)
			}
		}

		environments, err := client.ListEnvironments(projectID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to list environments: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		if len(environments) == 0 {
			fmt.Println("No environments found")
			return
		}

		fmt.Printf("Found %d environment(s):\n\n", len(environments))
		for _, env := range environments {
			if shared.ProjectIDToInt64(env.ID) == currentEnvironmentID {
				fmt.Printf("* ID: %d\n", shared.ProjectIDToInt64(env.ID))
			} else {
				fmt.Printf("  ID: %d\n", shared.ProjectIDToInt64(env.ID))
			}
			fmt.Printf("  Name: %s\n", env.Name)
			if env.Description != nil && *env.Description != "" {
				fmt.Printf("  Description: %s\n", *env.Description)
			}
			fmt.Printf("  Created: %s\n", env.CreatedAt)
			fmt.Println()
		}
	},
}

var getEnvironmentCmd = &cobra.Command{
	Use:   "get <environment_id> [project_id]",
	Short: "Get environment details",
	Long:  "Get detailed information about a specific environment",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		environmentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
			os.Exit(1)
		}

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var projectID int64
		if len(args) > 1 {
			projectID, err = strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid project ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			projectID, err = config.GetProjectID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if projectID == 0 {
				fmt.Println("No current project set.")
				fmt.Println("Please set a project with:")
				fmt.Println("  envoy projects use <id>")
				fmt.Println("Or provide project_id as an argument:")
				fmt.Println("  envoy environments get <environment_id> <project_id>")
				os.Exit(1)
			}
		}

		environment, err := client.GetEnvironment(projectID, environmentID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to get environment: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Environment Details:")
		fmt.Printf("  ID: %d\n", shared.ProjectIDToInt64(environment.ID))
		fmt.Printf("  Name: %s\n", environment.Name)
		if environment.Description != nil {
			fmt.Printf("  Description: %s\n", *environment.Description)
		}
		fmt.Printf("  Project ID: %d\n", shared.ProjectIDToInt64(environment.ProjectID))
		fmt.Printf("  Created: %s\n", environment.CreatedAt)
		fmt.Printf("  Updated: %s\n", environment.UpdatedAt)
	},
}

var updateEnvironmentCmd = &cobra.Command{
	Use:   "update <environment_id> [project_id]",
	Short: "Update an environment",
	Long:  "Update environment name and description",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		environmentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
			os.Exit(1)
		}

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var projectID int64
		if len(args) > 1 {
			projectID, err = strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid project ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			projectID, err = config.GetProjectID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if projectID == 0 {
				fmt.Println("No current project set.")
				fmt.Println("Please set a project with:")
				fmt.Println("  envoy projects use <id>")
				fmt.Println("Or provide project_id as an argument:")
				fmt.Println("  envoy environments update <environment_id> <project_id>")
				os.Exit(1)
			}
		}

		environment, err := client.GetEnvironment(projectID, environmentID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to get environment: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		name, err := PromptStringWithDefault("Environment name", environment.Name)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		description, err := PromptString("Description (leave empty to keep current)", false)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}
		if description == "" && environment.Description != nil {
			description = *environment.Description
		}

		updatedEnvironment, err := client.UpdateEnvironment(projectID, environmentID, name, description)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to update environment: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Environment updated successfully!")
		fmt.Printf("  Name: %s\n", updatedEnvironment.Name)
		if updatedEnvironment.Description != nil {
			fmt.Printf("  Description: %s\n", *updatedEnvironment.Description)
		}
	},
}

var deleteEnvironmentCmd = &cobra.Command{
	Use:   "delete <environment_id> [project_id]",
	Short: "Delete an environment",
	Long:  "Delete an environment permanently",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		environmentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
			os.Exit(1)
		}

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var projectID int64
		if len(args) > 1 {
			projectID, err = strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Invalid project ID: %v\n", err)
				os.Exit(1)
			}
		} else {
			projectID, err = config.GetProjectID()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if projectID == 0 {
				fmt.Println("No current project set.")
				fmt.Println("Please set a project with:")
				fmt.Println("  envoy projects use <id>")
				fmt.Println("Or provide project_id as an argument:")
				fmt.Println("  envoy environments delete <environment_id> <project_id>")
				os.Exit(1)
			}
		}

		environment, err := client.GetEnvironment(projectID, environmentID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to get environment: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Printf("Are you sure you want to delete environment '%s' (ID: %d)?\n", environment.Name, environment.ID)
		confirmed, err := Confirm("This action cannot be undone")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if !confirmed {
			fmt.Println("Operation cancelled")
			return
		}

		if err := client.DeleteEnvironment(projectID, environmentID); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to delete environment: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Environment deleted successfully")
	},
}

var useEnvironmentCmd = &cobra.Command{
	Use:   "use <id>",
	Short: "Set current environment",
	Long:  "Set the current environment for subsequent commands",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		environmentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Invalid environment ID: %v\n", err)
			os.Exit(1)
		}

		projectID, err := config.GetProjectID()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if projectID == 0 {
			fmt.Println("No current project set.")
			fmt.Println("Please set a project first with:")
			fmt.Println("  envoy projects use <id>")
			os.Exit(1)
		}

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		environment, err := client.GetEnvironment(projectID, environmentID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to get environment: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		if err := config.SetEnvironmentID(environmentID); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to set environment: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Now using environment: %s (ID: %d)\n", environment.Name, environment.ID)
	},
}

var unsetEnvironmentCmd = &cobra.Command{
	Use:   "unset",
	Short: "Unset current environment",
	Long:  "Clear the current environment context",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.ClearEnvironmentID(); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Current environment cleared")
	},
}

func init() {
	environmentsCmd.AddCommand(createEnvironmentCmd)
	environmentsCmd.AddCommand(listEnvironmentsCmd)
	environmentsCmd.AddCommand(getEnvironmentCmd)
	environmentsCmd.AddCommand(updateEnvironmentCmd)
	environmentsCmd.AddCommand(deleteEnvironmentCmd)
	environmentsCmd.AddCommand(useEnvironmentCmd)
	environmentsCmd.AddCommand(unsetEnvironmentCmd)
}
