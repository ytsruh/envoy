package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	cli "github.com/pressly/cli"
	"ytsruh.com/envoy/cli/controllers"
	"ytsruh.com/envoy/cli/prompts"
	"ytsruh.com/envoy/cli/utils"
	shared "ytsruh.com/envoy/shared"
)

var environmentVariablesCmd = &cli.Command{
	Name:      "variables",
	ShortHelp: "Manage environment variables",
	SubCommands: []*cli.Command{
		importVariablesCmd,
		exportVariablesCmd,
		createVariableCmd,
		listVariablesCmd,
		getVariableCmd,
		updateVariableCmd,
		deleteVariableCmd,
	},
}

func sanitizeFilename(name string) string {
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '_' || r == '-' || r == '.' {
			result = append(result, r)
		} else {
			result = append(result, '_')
		}
	}
	return strings.ToLower(string(result))
}

var importVariablesCmd = &cli.Command{
	Name:      "import",
	ShortHelp: "Import variables from .env file",
	Flags: cli.FlagsFunc(func(f *flag.FlagSet) {
		f.String("file", ".env", "Path to the .env file to import")
	}),
	FlagOptions: []cli.FlagOption{
		{Name: "file", Short: "f"},
	},
	Exec: func(ctx context.Context, s *cli.State) error {
		importFile := cli.GetFlag[string](s, "file")

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
			fmt.Fprintln(s.Stdout, "No projects found. Please create a project first with 'envoy projects create'")
			os.Exit(1)
		}

		projectOptions := make([]prompts.SelectOption, len(projects))
		for i, p := range projects {
			label := p.Name
			if p.Description != nil && *p.Description != "" {
				label += fmt.Sprintf(" - %s", *p.Description)
			}
			projectOptions[i] = prompts.SelectOption{
				Label: label,
				Value: string(p.ID),
			}
		}

		projectID, err := prompts.PromptSelect("Select a project", projectOptions, true)
		if err != nil {
			fmt.Fprintln(s.Stdout, "Import cancelled")
			return nil
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
			fmt.Fprintln(s.Stdout, "No environments found. Please create an environment first with 'envoy environments create <project_id>'")
			os.Exit(1)
		}

		environmentOptions := make([]prompts.SelectOption, len(environments))
		for i, e := range environments {
			label := e.Name
			if e.Description != nil && *e.Description != "" {
				label += fmt.Sprintf(" - %s", *e.Description)
			}
			environmentOptions[i] = prompts.SelectOption{
				Label: label,
				Value: string(e.ID),
			}
		}

		environmentID, err := prompts.PromptSelect("Select an environment", environmentOptions, true)
		if err != nil {
			fmt.Fprintln(s.Stdout, "Import cancelled")
			return nil
		}

		if _, err := os.Stat(importFile); os.IsNotExist(err) {
			fmt.Fprintf(s.Stderr, "Warning: File '%s' not found\n", importFile)
			os.Exit(1)
		}

		variables, err := utils.ParseEnvFile(importFile)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Failed to parse file '%s': %v\n", importFile, err)
			os.Exit(1)
		}

		if len(variables) == 0 {
			fmt.Fprintf(s.Stdout, "No variables found in %s\n", importFile)
			return nil
		}

		fmt.Fprintf(s.Stdout, "Found %d variable(s) in %s:\n\n", len(variables), importFile)
		for key, value := range variables {
			fmt.Fprintf(s.Stdout, "  %s=%s\n", key, value)
		}
		fmt.Fprintln(s.Stdout, "")

		confirmed, err := prompts.Confirm("Import these variables?")
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if !confirmed {
			fmt.Fprintln(s.Stdout, "Import cancelled")
			return nil
		}

		created := 0
		updated := 0
		for key, value := range variables {
			_, err := client.CreateEnvironmentVariable(projectID, environmentID, key, value)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to import variable %s: %v\n", key, err)
				continue
			}
			created++
		}

		fmt.Fprintf(s.Stdout, "Successfully imported %d variable(s)\n", created)
		if updated > 0 {
			fmt.Fprintf(s.Stdout, "Updated %d variable(s)\n", updated)
		}
		return nil
	},
}

var exportVariablesCmd = &cli.Command{
	Name:      "export",
	ShortHelp: "Export variables to .env file",
	Flags: cli.FlagsFunc(func(f *flag.FlagSet) {
		f.String("file", "", "Path to the export file (default: .env.<environment_name>)")
	}),
	FlagOptions: []cli.FlagOption{
		{Name: "file", Short: "f"},
	},
	Exec: func(ctx context.Context, s *cli.State) error {
		exportFile := cli.GetFlag[string](s, "file")

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
			fmt.Fprintln(s.Stdout, "No projects found. Please create a project first with 'envoy projects create'")
			os.Exit(1)
		}

		projectOptions := make([]prompts.SelectOption, len(projects))
		for i, p := range projects {
			label := p.Name
			if p.Description != nil && *p.Description != "" {
				label += fmt.Sprintf(" - %s", *p.Description)
			}
			projectOptions[i] = prompts.SelectOption{
				Label: label,
				Value: string(p.ID),
			}
		}

		projectID, err := prompts.PromptSelect("Select a project", projectOptions, true)
		if err != nil {
			fmt.Fprintln(s.Stdout, "Export cancelled")
			return nil
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
			fmt.Fprintln(s.Stdout, "No environments found. Please create an environment first with 'envoy environments create <project_id>'")
			os.Exit(1)
		}

		environmentOptions := make([]prompts.SelectOption, len(environments))
		for i, e := range environments {
			label := e.Name
			if e.Description != nil && *e.Description != "" {
				label += fmt.Sprintf(" - %s", *e.Description)
			}
			environmentOptions[i] = prompts.SelectOption{
				Label: label,
				Value: string(e.ID),
			}
		}

		environmentID, err := prompts.PromptSelect("Select an environment", environmentOptions, true)
		if err != nil {
			fmt.Fprintln(s.Stdout, "Export cancelled")
			return nil
		}

		var outputFilename string
		if exportFile == "" {
			environment, err := client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Warning: Failed to get environment name: %v\n", err)
				fmt.Fprintln(s.Stdout, "Using default filename .env")
				outputFilename = ".env"
			} else {
				outputFilename = ".env." + sanitizeFilename(environment.Name)
			}
		} else {
			outputFilename = exportFile
		}

		variables, err := client.ListEnvironmentVariables(projectID, environmentID)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Failed to list variables: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
			}
			os.Exit(1)
		}

		if len(variables) == 0 {
			fmt.Fprintln(s.Stdout, "No variables to export")
			return nil
		}

		if _, err := os.Stat(outputFilename); err == nil {
			fmt.Fprintf(s.Stderr, "Warning: File '%s' already exists in current directory\n", outputFilename)
			confirmed, err := prompts.Confirm("Overwrite existing file?")
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if !confirmed {
				fmt.Fprintln(s.Stdout, "Export cancelled")
				return nil
			}
		}

		variablesMap := make(map[string]string)
		for _, v := range variables {
			variablesMap[v.Key] = v.Value
		}

		if err := utils.WriteEnvFile(outputFilename, variablesMap); err != nil {
			fmt.Fprintf(s.Stderr, "Failed to write file '%s': %v\n", outputFilename, err)
			os.Exit(1)
		}

		fmt.Fprintf(s.Stdout, "Exported %d variable(s) to %s\n", len(variables), outputFilename)
		return nil
	},
}

var createVariableCmd = &cli.Command{
	Name:      "create",
	ShortHelp: "Create a new variable",
	Usage:     "envoy variables create [project_id] [environment_id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var projectID, environmentID string

		if len(s.Args) == 2 {
			projectID = s.Args[0]
			environmentID = s.Args[1]

			project, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintf(s.Stdout, "Project: %s (ID: %s)\n", project.Name, project.ID)

			environment, err := client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintf(s.Stdout, "Environment: %s (ID: %s)\n", environment.Name, environment.ID)

			key, err := prompts.PromptString("Variable key", true)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			value, err := prompts.PromptString("Variable value", true)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			variable, err := client.CreateEnvironmentVariable(projectID, environmentID, key, value)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to create variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Variable created successfully!")
			fmt.Fprintf(s.Stdout, "  ID: %s\n", variable.ID)
			fmt.Fprintf(s.Stdout, "  Key: %s\n", variable.Key)
			fmt.Fprintf(s.Stdout, "  Value: %s\n", variable.Value)
		} else if len(s.Args) == 1 {
			fmt.Fprintln(s.Stderr, "Error: Both project_id and environment_id are required")
			fmt.Fprintln(s.Stderr, "Usage: envoy variables create <project_id> <environment_id>")
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

			key, err := prompts.PromptString("Variable key", true)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			value, err := prompts.PromptString("Variable value", true)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			variable, err := client.CreateEnvironmentVariable(projectID, environmentID, key, value)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to create variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Variable created successfully!")
			fmt.Fprintf(s.Stdout, "  ID: %s\n", variable.ID)
			fmt.Fprintf(s.Stdout, "  Key: %s\n", variable.Key)
			fmt.Fprintf(s.Stdout, "  Value: %s\n", variable.Value)
		}
		return nil
	},
}

var listVariablesCmd = &cli.Command{
	Name:      "list",
	ShortHelp: "List variables",
	Usage:     "envoy variables list [project_id] [environment_id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var projectID, environmentID string

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

			_, err = client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			variables, err := client.ListEnvironmentVariables(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to list variables: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			if len(variables) == 0 {
				fmt.Fprintln(s.Stdout, "No variables found")
				return nil
			}

			fmt.Fprintf(s.Stdout, "Found %d variable(s):\n\n", len(variables))
			for _, v := range variables {
				fmt.Fprintf(s.Stdout, "  ID: %s\n", v.ID)
				fmt.Fprintf(s.Stdout, "  Key: %s\n", v.Key)
				fmt.Fprintf(s.Stdout, "  Value: %s\n", v.Value)
				fmt.Fprintf(s.Stdout, "  Updated: %s\n", v.UpdatedAt)
				fmt.Fprintln(s.Stdout, "")
			}
		} else if len(s.Args) == 1 {
			fmt.Fprintln(s.Stderr, "Error: Both project_id and environment_id are required")
			fmt.Fprintln(s.Stderr, "Usage: envoy variables list <project_id> <environment_id>")
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

			variables, err := client.ListEnvironmentVariables(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to list variables: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			if len(variables) == 0 {
				fmt.Fprintln(s.Stdout, "No variables found")
				return nil
			}

			fmt.Fprintf(s.Stdout, "Found %d variable(s):\n\n", len(variables))
			for _, v := range variables {
				fmt.Fprintf(s.Stdout, "  ID: %s\n", v.ID)
				fmt.Fprintf(s.Stdout, "  Key: %s\n", v.Key)
				fmt.Fprintf(s.Stdout, "  Value: %s\n", v.Value)
				fmt.Fprintf(s.Stdout, "  Updated: %s\n", v.UpdatedAt)
				fmt.Fprintln(s.Stdout, "")
			}
		}
		return nil
	},
}

var getVariableCmd = &cli.Command{
	Name:      "get",
	ShortHelp: "Get variable details",
	Usage:     "envoy variables get [variable_id] [project_id] [environment_id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var variableID, projectID, environmentID string

		if len(s.Args) == 3 {
			variableID = s.Args[0]
			projectID = s.Args[1]
			environmentID = s.Args[2]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			_, err = client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Variable Details:")
			fmt.Fprintf(s.Stdout, "  ID: %s\n", variable.ID)
			fmt.Fprintf(s.Stdout, "  Key: %s\n", variable.Key)
			fmt.Fprintf(s.Stdout, "  Value: %s\n", variable.Value)
			fmt.Fprintf(s.Stdout, "  Environment ID: %s\n", variable.EnvironmentID)
			fmt.Fprintf(s.Stdout, "  Created: %s\n", variable.CreatedAt)
			fmt.Fprintf(s.Stdout, "  Updated: %s\n", variable.UpdatedAt)
		} else if len(s.Args) >= 1 && len(s.Args) < 3 {
			fmt.Fprintln(s.Stderr, "Error: All three arguments are required: variable_id, project_id, and environment_id")
			fmt.Fprintln(s.Stderr, "Usage: envoy variables get <variable_id> <project_id> <environment_id>")
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

			variableID, err = prompts.PromptForVariable(client, projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Variable Details:")
			fmt.Fprintf(s.Stdout, "  ID: %s\n", variable.ID)
			fmt.Fprintf(s.Stdout, "  Key: %s\n", variable.Key)
			fmt.Fprintf(s.Stdout, "  Value: %s\n", variable.Value)
			fmt.Fprintf(s.Stdout, "  Environment ID: %s\n", variable.EnvironmentID)
			fmt.Fprintf(s.Stdout, "  Created: %s\n", variable.CreatedAt)
			fmt.Fprintf(s.Stdout, "  Updated: %s\n", variable.UpdatedAt)
		}
		return nil
	},
}

var updateVariableCmd = &cli.Command{
	Name:      "update",
	ShortHelp: "Update a variable",
	Usage:     "envoy variables update [variable_id] [project_id] [environment_id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var variableID, projectID, environmentID string

		if len(s.Args) == 3 {
			variableID = s.Args[0]
			projectID = s.Args[1]
			environmentID = s.Args[2]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			_, err = client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			key, err := prompts.PromptStringWithDefault("Variable key", variable.Key)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			value, err := prompts.PromptString("Variable value", true)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			updatedVariable, err := client.UpdateEnvironmentVariable(projectID, environmentID, variableID, key, value)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to update variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Variable updated successfully!")
			fmt.Fprintf(s.Stdout, "  Key: %s\n", updatedVariable.Key)
			fmt.Fprintf(s.Stdout, "  Value: %s\n", updatedVariable.Value)
		} else if len(s.Args) >= 1 && len(s.Args) < 3 {
			fmt.Fprintln(s.Stderr, "Error: All three arguments are required: variable_id, project_id, and environment_id")
			fmt.Fprintln(s.Stderr, "Usage: envoy variables update <variable_id> <project_id> <environment_id>")
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

			variableID, err = prompts.PromptForVariable(client, projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			key, err := prompts.PromptStringWithDefault("Variable key", variable.Key)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			value, err := prompts.PromptString("Variable value", true)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			updatedVariable, err := client.UpdateEnvironmentVariable(projectID, environmentID, variableID, key, value)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to update variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Variable updated successfully!")
			fmt.Fprintf(s.Stdout, "  Key: %s\n", updatedVariable.Key)
			fmt.Fprintf(s.Stdout, "  Value: %s\n", updatedVariable.Value)
		}
		return nil
	},
}

var deleteVariableCmd = &cli.Command{
	Name:      "delete",
	ShortHelp: "Delete a variable",
	Usage:     "envoy variables delete [variable_id] [project_id] [environment_id] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy login'")
			}
			os.Exit(1)
		}

		var variableID, projectID, environmentID string

		if len(s.Args) == 3 {
			variableID = s.Args[0]
			projectID = s.Args[1]
			environmentID = s.Args[2]

			_, err := client.GetProject(projectID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get project: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			_, err = client.GetEnvironment(projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get environment: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintf(s.Stdout, "Are you sure you want to delete variable '%s' (ID: %s)?\n", variable.Key, variable.ID)
			confirmed, err := prompts.Confirm("This action cannot be undone")
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if !confirmed {
				fmt.Fprintln(s.Stdout, "Operation cancelled")
				return nil
			}

			if err := client.DeleteEnvironmentVariable(projectID, environmentID, variableID); err != nil {
				fmt.Fprintf(s.Stderr, "Failed to delete variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Variable deleted successfully")
		} else if len(s.Args) >= 1 && len(s.Args) < 3 {
			fmt.Fprintln(s.Stderr, "Error: All three arguments are required: variable_id, project_id, and environment_id")
			fmt.Fprintln(s.Stderr, "Usage: envoy variables delete <variable_id> <project_id> <environment_id>")
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

			variableID, err = prompts.PromptForVariable(client, projectID, environmentID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			variable, err := client.GetEnvironmentVariable(projectID, environmentID, variableID)
			if err != nil {
				fmt.Fprintf(s.Stderr, "Failed to get variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintf(s.Stdout, "Are you sure you want to delete variable '%s' (ID: %s)?\n", variable.Key, variable.ID)
			confirmed, err := prompts.Confirm("This action cannot be undone")
			if err != nil {
				fmt.Fprintf(s.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if !confirmed {
				fmt.Fprintln(s.Stdout, "Operation cancelled")
				return nil
			}

			if err := client.DeleteEnvironmentVariable(projectID, environmentID, variableID); err != nil {
				fmt.Fprintf(s.Stderr, "Failed to delete variable: %v\n", err)
				if err == shared.ErrExpiredToken {
					fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy login'")
				}
				os.Exit(1)
			}

			fmt.Fprintln(s.Stdout, "Variable deleted successfully")
		}
		return nil
	},
}
