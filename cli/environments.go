package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ytsruh.com/envoy/cli/controllers"
	shared "ytsruh.com/envoy/shared"
)

var environmentsCmd = &cobra.Command{
	Use:   "environments",
	Short: "Manage environments",
	Long:  "Create, list, update, and delete environments within projects",
}

var createEnvironmentCmd = &cobra.Command{
	Use:   "create [project_id]",
	Short: "Create a new environment",
	Long:  "Create a new environment within a project",
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

		var projectID string

		if len(args) == 1 {
			projectID = args[0]

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Printf("Project: %s (ID: %s)\n", project.Name, project.ID)
		} else {
			projectID, err = promptForProject(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
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
		fmt.Printf("  ID: %s\n", environment.ID)
		fmt.Printf("  Name: %s\n", environment.Name)
		if environment.Description != nil {
			fmt.Printf("  Description: %s\n", *environment.Description)
		}
		fmt.Printf("  Project ID: %s\n", environment.ProjectID)
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

		var projectID string

		if len(args) == 1 {
			projectID = args[0]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}
		} else {
			projectID, err = promptForProject(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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
			fmt.Printf("  ID: %s\n", env.ID)
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
	Use:   "get [environment_id] [project_id]",
	Short: "Get environment details",
	Long:  "Get detailed information about a specific environment",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var environmentID, projectID string

		if len(args) == 2 {
			projectID = args[0]
			environmentID = args[1]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
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

			fmt.Println("Environment Details:")
			fmt.Printf("  ID: %s\n", environment.ID)
			fmt.Printf("  Name: %s\n", environment.Name)
			if environment.Description != nil {
				fmt.Printf("  Description: %s\n", *environment.Description)
			}
			fmt.Printf("  Project ID: %s\n", environment.ProjectID)
			fmt.Printf("  Created: %s\n", environment.CreatedAt)
			fmt.Printf("  Updated: %s\n", environment.UpdatedAt)
		} else if len(args) == 1 {
			fmt.Fprintln(cmd.OutOrStderr(), "Error: Both environment_id and project_id are required")
			fmt.Fprintln(cmd.OutOrStderr(), "Usage: envoy environments get <environment_id> <project_id>")
			os.Exit(1)
		} else {
			projectID, err = promptForProject(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}

			environmentID, err = promptForEnvironment(client, projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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

			fmt.Println("Environment Details:")
			fmt.Printf("  ID: %s\n", environment.ID)
			fmt.Printf("  Name: %s\n", environment.Name)
			if environment.Description != nil {
				fmt.Printf("  Description: %s\n", *environment.Description)
			}
			fmt.Printf("  Project ID: %s\n", environment.ProjectID)
			fmt.Printf("  Created: %s\n", environment.CreatedAt)
			fmt.Printf("  Updated: %s\n", environment.UpdatedAt)
		}
	},
}

var updateEnvironmentCmd = &cobra.Command{
	Use:   "update [environment_id] [project_id]",
	Short: "Update an environment",
	Long:  "Update environment name and description",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var environmentID, projectID string

		if len(args) == 2 {
			projectID = args[0]
			environmentID = args[1]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
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
		} else if len(args) == 1 {
			fmt.Fprintln(cmd.OutOrStderr(), "Error: Both environment_id and project_id are required")
			fmt.Fprintln(cmd.OutOrStderr(), "Usage: envoy environments update <environment_id> <project_id>")
			os.Exit(1)
		} else {
			projectID, err = promptForProject(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}

			environmentID, err = promptForEnvironment(client, projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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
		}
	},
}

var deleteEnvironmentCmd = &cobra.Command{
	Use:   "delete [environment_id] [project_id]",
	Short: "Delete an environment",
	Long:  "Delete an environment permanently",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var environmentID, projectID string

		if len(args) == 2 {
			projectID = args[0]
			environmentID = args[1]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
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

			fmt.Printf("Are you sure you want to delete environment '%s' (ID: %s)?\n", environment.Name, environment.ID)
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
		} else if len(args) == 1 {
			fmt.Fprintln(cmd.OutOrStderr(), "Error: Both environment_id and project_id are required")
			fmt.Fprintln(cmd.OutOrStderr(), "Usage: envoy environments delete <environment_id> <project_id>")
			os.Exit(1)
		} else {
			projectID, err = promptForProject(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}

			environmentID, err = promptForEnvironment(client, projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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

			fmt.Printf("Are you sure you want to delete environment '%s' (ID: %s)?\n", environment.Name, environment.ID)
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
		}
	},
}

func init() {
	environmentsCmd.AddCommand(createEnvironmentCmd)
	environmentsCmd.AddCommand(listEnvironmentsCmd)
	environmentsCmd.AddCommand(getEnvironmentCmd)
	environmentsCmd.AddCommand(updateEnvironmentCmd)
	environmentsCmd.AddCommand(deleteEnvironmentCmd)
}
