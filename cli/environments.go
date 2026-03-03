package cli

import (
	"context"
	"fmt"
	"os"

	cli "github.com/pressly/cli"
	"ytsruh.com/envoy/cli/controllers"
	"ytsruh.com/envoy/cli/prompts"
	shared "ytsruh.com/envoy/shared"
)

var environmentsCmd = &cli.Command{
	Name:      "environments",
	ShortHelp: "Manage environments",
	SubCommands: []*cli.Command{
		createEnvironmentCmd,
		listEnvironmentsCmd,
		getEnvironmentCmd,
		updateEnvironmentCmd,
		deleteEnvironmentCmd,
	},
}

var createEnvironmentCmd = &cli.Command{
	Name:      "create",
	ShortHelp: "Create a new environment",
	Usage:     "envoy environments create [project_id] [flags]",
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

			fmt.Fprintf(s.Stdout, "Project: %s (ID: %s)\n", project.Name, project.ID)
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		name, err := prompts.PromptString("Environment name", true)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		description, err := prompts.PromptString("Description (optional)", false)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		environment, err := client.CreateEnvironment(projectID, name, description)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Failed to create environment: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		fmt.Fprintln(s.Stdout, "Environment created successfully!")
		fmt.Fprintf(s.Stdout, "  ID: %s\n", environment.ID)
		fmt.Fprintf(s.Stdout, "  Name: %s\n", environment.Name)
		if environment.Description != nil {
			fmt.Fprintf(s.Stdout, "  Description: %s\n", *environment.Description)
		}
		fmt.Fprintf(s.Stdout, "  Project ID: %s\n", environment.ProjectID)
		return nil
	},
}

var listEnvironmentsCmd = &cli.Command{
	Name:      "list",
	ShortHelp: "List environments",
	Usage:     "envoy environments list [project_id] [flags]",
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

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		environments, err := client.ListEnvironments(projectID)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Failed to list environments: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		if len(environments) == 0 {
			fmt.Fprintln(s.Stdout, "No environments found")
			return nil
		}

		fmt.Fprintf(s.Stdout, "Found %d environment(s):\n\n", len(environments))
		for _, env := range environments {
			fmt.Fprintf(s.Stdout, "  ID: %s\n", env.ID)
			fmt.Fprintf(s.Stdout, "  Name: %s\n", env.Name)
			if env.Description != nil && *env.Description != "" {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *env.Description)
			}
			fmt.Fprintf(s.Stdout, "  Created: %s\n", env.CreatedAt)
			fmt.Fprintln(s.Stdout, "")
		}
		return nil
	},
}

var getEnvironmentCmd = &cli.Command{
	Name:      "get",
	ShortHelp: "Get environment details",
	Usage:     "envoy environments get [environment_id] [project_id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var environmentID, projectID string

		if len(s.Args) == 2 {
			projectID = s.Args[0]
			environmentID = s.Args[1]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			environment, err := client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Environment Details:")
			fmt.Fprintf(s.Stdout, "  ID: %s\n", environment.ID)
			fmt.Fprintf(s.Stdout, "  Name: %s\n", environment.Name)
			if environment.Description != nil {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *environment.Description)
			}
			fmt.Fprintf(s.Stdout, "  Project ID: %s\n", environment.ProjectID)
			fmt.Fprintf(s.Stdout, "  Created: %s\n", environment.CreatedAt)
			fmt.Fprintf(s.Stdout, "  Updated: %s\n", environment.UpdatedAt)
		} else if len(s.Args) == 1 {
			fmt.Fprintln(s.Stderr, "Error: Both environment_id and project_id are required")
			fmt.Fprintln(s.Stderr, "Usage: envoy environments get <environment_id> <project_id>")
			os.Exit(1)
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			environmentID, err = prompts.PromptForEnvironment(client, projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			environment, err := client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Environment Details:")
			fmt.Fprintf(s.Stdout, "  ID: %s\n", environment.ID)
			fmt.Fprintf(s.Stdout, "  Name: %s\n", environment.Name)
			if environment.Description != nil {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *environment.Description)
			}
			fmt.Fprintf(s.Stdout, "  Project ID: %s\n", environment.ProjectID)
			fmt.Fprintf(s.Stdout, "  Created: %s\n", environment.CreatedAt)
			fmt.Fprintf(s.Stdout, "  Updated: %s\n", environment.UpdatedAt)
		}
		return nil
	},
}

var updateEnvironmentCmd = &cli.Command{
	Name:      "update",
	ShortHelp: "Update an environment",
	Usage:     "envoy environments update [environment_id] [project_id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var environmentID, projectID string

		if len(s.Args) == 2 {
			projectID = s.Args[0]
			environmentID = s.Args[1]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			environment, err := client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			name, err := prompts.PromptStringWithDefault("Environment name", environment.Name)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			description, err := prompts.PromptString("Description (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if description == "" && environment.Description != nil {
				description = *environment.Description
			}

			updatedEnvironment, err := client.UpdateEnvironment(projectID, environmentID, name, description)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to update environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Environment updated successfully!")
			fmt.Fprintf(s.Stdout, "  Name: %s\n", updatedEnvironment.Name)
			if updatedEnvironment.Description != nil {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *updatedEnvironment.Description)
			}
		} else if len(s.Args) == 1 {
			fmt.Fprintln(s.Stderr, "Error: Both environment_id and project_id are required")
			fmt.Fprintln(s.Stderr, "Usage: envoy environments update <environment_id> <project_id>")
			os.Exit(1)
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			environmentID, err = prompts.PromptForEnvironment(client, projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			environment, err := client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			name, err := prompts.PromptStringWithDefault("Environment name", environment.Name)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			description, err := prompts.PromptString("Description (leave empty to keep current)", false)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if description == "" && environment.Description != nil {
				description = *environment.Description
			}

			updatedEnvironment, err := client.UpdateEnvironment(projectID, environmentID, name, description)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to update environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Environment updated successfully!")
			fmt.Fprintf(s.Stdout, "  Name: %s\n", updatedEnvironment.Name)
			if updatedEnvironment.Description != nil {
				fmt.Fprintf(s.Stdout, "  Description: %s\n", *updatedEnvironment.Description)
			}
		}
		return nil
	},
}

var deleteEnvironmentCmd = &cli.Command{
	Name:      "delete",
	ShortHelp: "Delete an environment",
	Usage:     "envoy environments delete [environment_id] [project_id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var environmentID, projectID string

		if len(s.Args) == 2 {
			projectID = s.Args[0]
			environmentID = s.Args[1]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			environment, err := client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintf(s.Stdout, "Are you sure you want to delete environment '%s' (ID: %s)?\n", environment.Name, environment.ID)
			confirmed, err := prompts.Confirm("This action cannot be undone")
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if !confirmed {
				fmt.Fprintln(s.Stdout, "Operation cancelled")
				return nil
			}

			if err := client.DeleteEnvironment(projectID, environmentID); err != nil {
				fmt.Fprintf(s.Stderr, "Failed to delete environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Environment deleted successfully")
		} else if len(s.Args) == 1 {
			fmt.Fprintln(s.Stderr, "Error: Both environment_id and project_id are required")
			fmt.Fprintln(s.Stderr, "Usage: envoy environments delete <environment_id> <project_id>")
			os.Exit(1)
		} else {
			projectID, err = prompts.PromptForProject(client)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			environmentID, err = prompts.PromptForEnvironment(client, projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			environment, err := client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintf(s.Stdout, "Are you sure you want to delete environment '%s' (ID: %s)?\n", environment.Name, environment.ID)
			confirmed, err := prompts.Confirm("This action cannot be undone")
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if !confirmed {
				fmt.Fprintln(s.Stdout, "Operation cancelled")
				return nil
			}

			if err := client.DeleteEnvironment(projectID, environmentID); err != nil {
				fmt.Fprintf(s.Stderr, "Failed to delete environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Environment deleted successfully")
		}
		return nil
	},
}
