package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ytsruh.com/envoy/cli/controllers"
	"ytsruh.com/envoy/cli/prompts"
	"ytsruh.com/envoy/cli/utils"
	shared "ytsruh.com/envoy/shared"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Long:  "Create, list, update, and delete projects",
}

var createProjectCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long:  "Create a new Envoy project with optional git repository tracking",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := prompts.PromptString("Project name", true)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		description, err := prompts.PromptString("Description (optional)", false)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		gitRepo, err := utils.GetGitRepoString()
		if err != nil {
			fmt.Printf("Warning: Could not detect git repository: %v\n", err)
		}

		if gitRepo == "" {
			gitRepo, err = prompts.PromptString("Git repository (owner/repo, optional)", false)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Detected git repository: %s\n", gitRepo)
			useGit, err := prompts.Confirm("Use this git repository?")
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if !useGit {
				gitRepo, err = prompts.PromptString("Git repository (owner/repo, optional)", false)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
					os.Exit(1)
				}
			}
		}

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		project, err := client.CreateProject(name, description, gitRepo)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to create project: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Println("Project created successfully!")
		fmt.Printf("  ID: %s\n", project.ID)
		fmt.Printf("  Name: %s\n", project.Name)
		if project.Description != nil && *project.Description != "" {
			fmt.Printf("  Description: %s\n", *project.Description)
		}
		if project.GitRepo != nil && *project.GitRepo != "" {
			fmt.Printf("  Git Repository: %s\n", *project.GitRepo)
		}
	},
}

var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  "List all projects you have access to",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Println("Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projects, err := client.ListProjects()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to list projects: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Println("Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		if len(projects) == 0 {
			fmt.Println("No projects found")
			return
		}

		fmt.Printf("Found %d project(s):\n\n", len(projects))
		for _, p := range projects {
			fmt.Printf("  ID: %s\n", p.ID)
			fmt.Printf("  Name: %s\n", p.Name)
			if p.Description != nil && *p.Description != "" {
				fmt.Printf("  Description: %s\n", *p.Description)
			}
			if p.GitRepo != nil && *p.GitRepo != "" {
				fmt.Printf("  Git Repository: %s\n", *p.GitRepo)
			}
			fmt.Printf("  Created: %s\n", p.CreatedAt)
			fmt.Println()
		}
	},
}

var getProjectCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get project details",
	Long:  "Get detailed information about a specific project",
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
				os.Exit(1)
			}

			fmt.Println("Project Details:")
			fmt.Printf("  ID: %s\n", project.ID)
			fmt.Printf("  Name: %s\n", project.Name)
			if project.Description != nil && *project.Description != "" {
				fmt.Printf("  Description: %s\n", *project.Description)
			}
			if project.GitRepo != nil && *project.GitRepo != "" {
				fmt.Printf("  Git Repository: %s\n", *project.GitRepo)
			}
			fmt.Printf("  Owner ID: %s\n", project.OwnerID)
			fmt.Printf("  Created: %s\n", project.CreatedAt)
			fmt.Printf("  Updated: %s\n", project.UpdatedAt)
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to get project: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Project Details:")
			fmt.Printf("  ID: %s\n", project.ID)
			fmt.Printf("  Name: %s\n", project.Name)
			if project.Description != nil && *project.Description != "" {
				fmt.Printf("  Description: %s\n", *project.Description)
			}
			if project.GitRepo != nil && *project.GitRepo != "" {
				fmt.Printf("  Git Repository: %s\n", *project.GitRepo)
			}
			fmt.Printf("  Owner ID: %s\n", project.OwnerID)
			fmt.Printf("  Created: %s\n", project.CreatedAt)
			fmt.Printf("  Updated: %s\n", project.UpdatedAt)
		}
	},
}

var updateProjectCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update a project",
	Long:  "Update project name, description, or git repository",
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

			name, err := prompts.PromptStringWithDefault("Project name", project.Name)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}

			description, err := prompts.PromptString("Description (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if description == "" && project.Description != nil {
				description = *project.Description
			}

			gitRepo, err := prompts.PromptString("Git repository owner/repo (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if gitRepo == "" && project.GitRepo != nil {
				gitRepo = *project.GitRepo
			}

			updatedProject, err := client.UpdateProject(projectID, name, description, gitRepo)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to update project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Println("Project updated successfully!")
			fmt.Printf("  Name: %s\n", updatedProject.Name)
			if updatedProject.Description != nil && *updatedProject.Description != "" {
				fmt.Printf("  Description: %s\n", *updatedProject.Description)
			}
			if updatedProject.GitRepo != nil && *updatedProject.GitRepo != "" {
				fmt.Printf("  Git Repository: %s\n", *updatedProject.GitRepo)
			}
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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

			name, err := prompts.PromptStringWithDefault("Project name", project.Name)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}

			description, err := prompts.PromptString("Description (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if description == "" && project.Description != nil {
				description = *project.Description
			}

			gitRepo, err := prompts.PromptString("Git repository owner/repo (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
			if gitRepo == "" && project.GitRepo != nil {
				gitRepo = *project.GitRepo
			}

			updatedProject, err := client.UpdateProject(projectID, name, description, gitRepo)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to update project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Println("Project updated successfully!")
			fmt.Printf("  Name: %s\n", updatedProject.Name)
			if updatedProject.Description != nil && *updatedProject.Description != "" {
				fmt.Printf("  Description: %s\n", *updatedProject.Description)
			}
			if updatedProject.GitRepo != nil && *updatedProject.GitRepo != "" {
				fmt.Printf("  Git Repository: %s\n", *updatedProject.GitRepo)
			}
		}
	},
}

var deleteProjectCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a project",
	Long:  "Delete a project permanently",
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

			fmt.Printf("Are you sure you want to delete project '%s' (ID: %s)?\n", project.Name, project.ID)
			confirmed, err := prompts.Confirm("This action cannot be undone")
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}

			if !confirmed {
				fmt.Println("Operation cancelled")
				return
			}

			if err := client.DeleteProject(projectID); err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to delete project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Println("Project deleted successfully")
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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

			fmt.Printf("Are you sure you want to delete project '%s' (ID: %s)?\n", project.Name, project.ID)
			confirmed, err := prompts.Confirm("This action cannot be undone")
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}

			if !confirmed {
				fmt.Println("Operation cancelled")
				return
			}

			if err := client.DeleteProject(projectID); err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to delete project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Println("Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Println("Project deleted successfully")
		}
	},
}

func init() {
	projectsCmd.AddCommand(createProjectCmd)
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(getProjectCmd)
	projectsCmd.AddCommand(updateProjectCmd)
	projectsCmd.AddCommand(deleteProjectCmd)
}
