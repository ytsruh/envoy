package cli

import (
	"context"
	"fmt"
	"os"

	cli "github.com/pressly/cli"
	"ytsruh.com/envoy/cli/controllers"
	"ytsruh.com/envoy/cli/prompts"
)

var usersCmd = &cli.Command{
	Name:      "users",
	ShortHelp: "Manage users",
	SubCommands: []*cli.Command{
		searchUsersCmd,
		membersCmd,
	},
}

var searchUsersCmd = &cli.Command{
	Name:      "search",
	ShortHelp: "Search for users by email",
	Usage:     "envoy users search [email] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		if len(s.Args) != 1 {
			fmt.Fprintln(s.Stderr, "Error: email is required")
			fmt.Fprintln(s.Stderr, "Usage: envoy users search <email>")
			os.Exit(1)
		}

		email := s.Args[0]

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		users, err := client.UsersController.SearchByEmail(email)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(users) == 0 {
			fmt.Fprintln(s.Stdout, "No users found with that email.")
			return nil
		}

		fmt.Fprintf(s.Stdout, "Found %d user(s):\n\n", len(users))
		for _, u := range users {
			fmt.Fprintf(s.Stdout, "  Name:  %s\n", u.Name)
			fmt.Fprintf(s.Stdout, "  Email: %s\n", u.Email)
			fmt.Fprintf(s.Stdout, "  ID:    %s\n\n", u.UserID)
		}
		return nil
	},
}

var membersCmd = &cli.Command{
	Name:      "members",
	ShortHelp: "Manage project members",
	SubCommands: []*cli.Command{
		listMembersCmd,
		addMemberCmd,
		removeMemberCmd,
	},
}

var listMembersCmd = &cli.Command{
	Name:      "list",
	ShortHelp: "List members of a project",
	Usage:     "envoy users members list [project-name] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		if len(s.Args) != 1 {
			fmt.Fprintln(s.Stderr, "Error: project-name is required")
			fmt.Fprintln(s.Stderr, "Usage: envoy users members list <project-name>")
			os.Exit(1)
		}

		projectName := s.Args[0]

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		projects, err := client.ProjectsController.ListProjects()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var project controllers.ProjectResponse
		found := false
		for _, p := range projects {
			if p.Name == projectName {
				project = p
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(s.Stderr, "Error: project '%s' not found\n", projectName)
			os.Exit(1)
		}

		members, err := client.ProjectsController.ListProjectMembers(string(project.ID))
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(members) == 0 {
			fmt.Fprintln(s.Stdout, "No members found in this project.")
			return nil
		}

		fmt.Fprintf(s.Stdout, "Members of '%s':\n\n", projectName)
		for _, m := range members {
			fmt.Fprintf(s.Stdout, "  User ID: %s\n", m.UserID)
			fmt.Fprintf(s.Stdout, "  Role:    %s\n\n", m.Role)
		}
		return nil
	},
}

var addMemberCmd = &cli.Command{
	Name:      "add",
	ShortHelp: "Add a member to a project",
	Usage:     "envoy users members add [project-name] [email] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		if len(s.Args) != 2 {
			fmt.Fprintln(s.Stderr, "Error: project-name and email are required")
			fmt.Fprintln(s.Stderr, "Usage: envoy users members add <project-name> <email>")
			os.Exit(1)
		}

		projectName := s.Args[0]
		email := s.Args[1]

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		projects, err := client.ProjectsController.ListProjects()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var project controllers.ProjectResponse
		found := false
		for _, p := range projects {
			if p.Name == projectName {
				project = p
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(s.Stderr, "Error: project '%s' not found\n", projectName)
			os.Exit(1)
		}

		users, err := client.UsersController.SearchByEmail(email)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(users) == 0 {
			fmt.Fprintf(s.Stderr, "Error: no user found with email '%s'\n", email)
			os.Exit(1)
		}

		if len(users) > 1 {
			fmt.Fprintf(s.Stderr, "Error: multiple users found with similar email. Use 'envoy users search' to find the exact user.\n")
			os.Exit(1)
		}

		user := users[0]

		role, err := prompts.PromptRole("Enter role (viewer/editor)")
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = client.ProjectsController.AddMember(string(project.ID), string(user.UserID), role)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(s.Stdout, "Successfully added %s (%s) to project '%s' as %s\n", user.Name, user.Email, projectName, role)
		return nil
	},
}

var removeMemberCmd = &cli.Command{
	Name:      "remove",
	ShortHelp: "Remove a member from a project",
	Usage:     "envoy users members remove [project-name] [email] [flags]",
	Exec: func(ctx context.Context, s *cli.State) error {
		if len(s.Args) != 2 {
			fmt.Fprintln(s.Stderr, "Error: project-name and email are required")
			fmt.Fprintln(s.Stderr, "Usage: envoy users members remove <project-name> <email>")
			os.Exit(1)
		}

		projectName := s.Args[0]
		email := s.Args[1]

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		projects, err := client.ProjectsController.ListProjects()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var project controllers.ProjectResponse
		found := false
		for _, p := range projects {
			if p.Name == projectName {
				project = p
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(s.Stderr, "Error: project '%s' not found\n", projectName)
			os.Exit(1)
		}

		users, err := client.UsersController.SearchByEmail(email)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(users) == 0 {
			fmt.Fprintf(s.Stderr, "Error: no user found with email '%s'\n", email)
			os.Exit(1)
		}

		if len(users) > 1 {
			fmt.Fprintf(s.Stderr, "Error: multiple users found with similar email. Use 'envoy users search' to find the exact user.\n")
			os.Exit(1)
		}

		user := users[0]

		err = client.ProjectsController.RemoveMember(string(project.ID), user.UserID)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(s.Stdout, "Successfully removed %s (%s) from project '%s'\n", user.Name, user.Email, projectName)
		return nil
	},
}
