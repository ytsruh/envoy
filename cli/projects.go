package cli

import (
	"context"
	"fmt"
	"os"

	cli "github.com/pressly/cli"
	"ytsruh.com/envoy/cli/controllers"
	"ytsruh.com/envoy/cli/prompts"
	"ytsruh.com/envoy/cli/utils"
	shared "ytsruh.com/envoy/shared"
)

var projectsCmd = &cli.Command{
	Name:      "projects",
	ShortHelp: "Manage projects",
	SubCommands: []*cli.Command{
		createProjectCmd,
		listProjectsCmd,
		getProjectCmd,
		updateProjectCmd,
		deleteProjectCmd,
	},
}

var createProjectCmd = &cli.Command{
	Name:      "create",
	ShortHelp: "Create a new project",
	Exec: func(ctx context.Context, s *cli.State) error {
		name, err := prompts.PromptString("Project name", true)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		description, err := prompts.PromptString("Description (optional)", false)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		gitRepo, err := utils.GetGitRepoString()
		if err != nil {
			fmt.Fprintf(s.Stdout, "Warning: Could not detect git repository: %v\n", err)
		}

		if gitRepo == "" {
			gitRepo, err = prompts.PromptString("Git repository (owner/repo, optional)", false)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(s.Stdout, "Detected git repository: %s\n", gitRepo)
			useGit, err := prompts.Confirm("Use this git repository?")
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if !useGit {
				gitRepo, err = prompts.PromptString("Git repository (owner/repo, optional)", false)
				if err != nil {
					fmt.Fprintf(s.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			}
		}

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		project, err := client.CreateProject(name, description, gitRepo)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Failed to create project: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Fprintln(s.Stdout, "Project created successfully!")
		fmt.Fprintf(s.Stdout, "  ID: %s\n", project.ID)
		fmt.Fprintf(s.Stdout, "  Name: %s\n", project.Name)
		if project.Description != nil && *project.Description != "" {
			fmt.Fprintf(s.Stdout, "  Description: %s\n", *project.Description)
		}
		if project.GitRepo != nil && *project.GitRepo != "" {
			fmt.Fprintf(s.Stdout, "  Git Repository: %s\n", *project.GitRepo)
		}
		return nil
	},
}

var listProjectsCmd = &cli.Command{
	Name:      "list",
	ShortHelp: "List all projects",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		projects, err := client.ListProjects()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Failed to list projects: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		if len(projects) == 0 {
			fmt.Fprintln(s.Stdout, "No projects found")
			return nil
		}

		fmt.Fprintf(s.Stdout, "Found %d project(s):\n\n", len(projects))
		for _, p := range projects {
			fmt.Fprintf(s.Stdout, "  ID: %s\n", p.ID)
			fmt.Fprintf(s.Stdout, "  Name: %s\n", p.Name)
			if p.Description != nil && *p.Description != "" {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *p.Description)
			}
			if p.GitRepo != nil && *p.GitRepo != "" {
				fmt.Fprintf(s.Stdout, "  Git Repository: %s\n", *p.GitRepo)
			}
			fmt.Fprintf(s.Stdout, "  Created: %s\n", p.CreatedAt)
			fmt.Fprintln(s.Stdout, "")
		}
		return nil
	},
}

var getProjectCmd = &cli.Command{
	Name:      "get",
	ShortHelp: "Get project details",
	Usage:     "envoy projects get [id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var projectID string

		if len(s.Args) == 1 {
			projectID = s.Args[0]

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Project Details:")
			fmt.Fprintf(s.Stdout, "  ID: %s\n", project.ID)
			fmt.Fprintf(s.Stdout, "  Name: %s\n", project.Name)
			if project.Description != nil && *project.Description != "" {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *project.Description)
			}
			if project.GitRepo != nil && *project.GitRepo != "" {
				fmt.Fprintf(s.Stdout, "  Git Repository: %s\n", *project.GitRepo)
			}
			fmt.Fprintf(s.Stdout, "  Owner ID: %s\n", project.OwnerID)
			fmt.Fprintf(s.Stdout, "  Created: %s\n", project.CreatedAt)
			fmt.Fprintf(s.Stdout, "  Updated: %s\n", project.UpdatedAt)
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Project Details:")
			fmt.Fprintf(s.Stdout, "  ID: %s\n", project.ID)
			fmt.Fprintf(s.Stdout, "  Name: %s\n", project.Name)
			if project.Description != nil && *project.Description != "" {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *project.Description)
			}
			if project.GitRepo != nil && *project.GitRepo != "" {
				fmt.Fprintf(s.Stdout, "  Git Repository: %s\n", *project.GitRepo)
			}
			fmt.Fprintf(s.Stdout, "  Owner ID: %s\n", project.OwnerID)
			fmt.Fprintf(s.Stdout, "  Created: %s\n", project.CreatedAt)
			fmt.Fprintf(s.Stdout, "  Updated: %s\n", project.UpdatedAt)
		}
		return nil
	},
}

var updateProjectCmd = &cli.Command{
	Name:      "update",
	ShortHelp: "Update a project",
	Usage:     "envoy projects update [id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var projectID string

		if len(s.Args) == 1 {
			projectID = s.Args[0]

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			name, err := prompts.PromptStringWithDefault("Project name", project.Name)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			description, err := prompts.PromptString("Description (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if description == "" && project.Description != nil {
				description = *project.Description
			}

			gitRepo, err := prompts.PromptString("Git repository owner/repo (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if gitRepo == "" && project.GitRepo != nil {
				gitRepo = *project.GitRepo
			}

			updatedProject, err := client.UpdateProject(projectID, name, description, gitRepo)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to update project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Project updated successfully!")
			fmt.Fprintf(s.Stdout, "  Name: %s\n", updatedProject.Name)
			if updatedProject.Description != nil && *updatedProject.Description != "" {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *updatedProject.Description)
			}
			if updatedProject.GitRepo != nil && *updatedProject.GitRepo != "" {
				fmt.Fprintf(s.Stdout, "  Git Repository: %s\n", *updatedProject.GitRepo)
			}
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			name, err := prompts.PromptStringWithDefault("Project name", project.Name)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			description, err := prompts.PromptString("Description (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if description == "" && project.Description != nil {
				description = *project.Description
			}

			gitRepo, err := prompts.PromptString("Git repository owner/repo (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if gitRepo == "" && project.GitRepo != nil {
				gitRepo = *project.GitRepo
			}

			updatedProject, err := client.UpdateProject(projectID, name, description, gitRepo)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to update project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Project updated successfully!")
			fmt.Fprintf(s.Stdout, "  Name: %s\n", updatedProject.Name)
			if updatedProject.Description != nil && *updatedProject.Description != "" {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *updatedProject.Description)
			}
			if updatedProject.GitRepo != nil && *updatedProject.GitRepo != "" {
				fmt.Fprintf(s.Stdout, "  Git Repository: %s\n", *updatedProject.GitRepo)
			}
		}
		return nil
	},
}

var deleteProjectCmd = &cli.Command{
	Name:      "delete",
	ShortHelp: "Delete a project",
	Usage:     "envoy projects delete [id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var projectID string

		if len(s.Args) == 1 {
			projectID = s.Args[0]

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintf(s.Stdout, "Are you sure you want to delete project '%s' (ID: %s)?\n", project.Name, project.ID)
			confirmed, err := prompts.Confirm("This action cannot be undone")
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if !confirmed {
				fmt.Fprintln(s.Stdout, "Operation cancelled")
				return nil
			}

			if err := client.DeleteProject(projectID); err != nil {
				fmt.Fprintf(s.Stderr, "Failed to delete project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Project deleted successfully")
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintf(s.Stdout, "Are you sure you want to delete project '%s' (ID: %s)?\n", project.Name, project.ID)
			confirmed, err := prompts.Confirm("This action cannot be undone")
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if !confirmed {
				fmt.Fprintln(s.Stdout, "Operation cancelled")
				return nil
			}

			if err := client.DeleteProject(projectID); err != nil {
				fmt.Fprintf(s.Stderr, "Failed to delete project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Project deleted successfully")
		}
		return nil
	},
}
