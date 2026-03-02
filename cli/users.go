package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ytsruh.com/envoy/cli/controllers"
	"ytsruh.com/envoy/cli/prompts"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  "Search for users and manage project membership",
}

var searchUsersCmd = &cobra.Command{
	Use:   "search [email]",
	Short: "Search for users by email",
	Long:  "Search for users by their email address to find their user ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		users, err := client.UsersController.SearchByEmail(email)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if len(users) == 0 {
			fmt.Println("No users found with that email.")
			return
		}

		fmt.Printf("Found %d user(s):\n\n", len(users))
		for _, u := range users {
			fmt.Printf("  Name:  %s\n", u.Name)
			fmt.Printf("  Email: %s\n", u.Email)
			fmt.Printf("  ID:    %s\n\n", u.UserID)
		}
	},
}

var membersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage project members",
	Long:  "Add, remove, and list members of a project",
}

var listMembersCmd = &cobra.Command{
	Use:   "list [project-name]",
	Short: "List members of a project",
	Long:  "List all members of a project and their roles",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		projects, err := client.ProjectsController.ListProjects()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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
			fmt.Fprintf(cmd.OutOrStderr(), "Error: project '%s' not found\n", projectName)
			os.Exit(1)
		}

		members, err := client.ProjectsController.ListProjectMembers(string(project.ID))
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if len(members) == 0 {
			fmt.Println("No members found in this project.")
			return
		}

		fmt.Printf("Members of '%s':\n\n", projectName)
		for _, m := range members {
			fmt.Printf("  User ID: %s\n", m.UserID)
			fmt.Printf("  Role:    %s\n\n", m.Role)
		}
	},
}

var addMemberCmd = &cobra.Command{
	Use:   "add [project-name] [email]",
	Short: "Add a member to a project",
	Long:  "Add a user to a project by their email address",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		email := args[1]

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		projects, err := client.ProjectsController.ListProjects()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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
			fmt.Fprintf(cmd.OutOrStderr(), "Error: project '%s' not found\n", projectName)
			os.Exit(1)
		}

		users, err := client.UsersController.SearchByEmail(email)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if len(users) == 0 {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: no user found with email '%s'\n", email)
			os.Exit(1)
		}

		if len(users) > 1 {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: multiple users found with similar email. Use 'envoy users search' to find the exact user.\n")
			os.Exit(1)
		}

		user := users[0]

		role, err := prompts.PromptRole("Enter role (viewer/editor)")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		err = client.ProjectsController.AddMember(string(project.ID), string(user.UserID), role)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added %s (%s) to project '%s' as %s\n", user.Name, user.Email, projectName, role)
	},
}

var removeMemberCmd = &cobra.Command{
	Use:   "remove [project-name] [email]",
	Short: "Remove a member from a project",
	Long:  "Remove a user from a project by their email address",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		email := args[1]

		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		projects, err := client.ProjectsController.ListProjects()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
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
			fmt.Fprintf(cmd.OutOrStderr(), "Error: project '%s' not found\n", projectName)
			os.Exit(1)
		}

		users, err := client.UsersController.SearchByEmail(email)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if len(users) == 0 {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: no user found with email '%s'\n", email)
			os.Exit(1)
		}

		if len(users) > 1 {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: multiple users found with similar email. Use 'envoy users search' to find the exact user.\n")
			os.Exit(1)
		}

		user := users[0]

		err = client.ProjectsController.RemoveMember(string(project.ID), user.UserID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully removed %s (%s) from project '%s'\n", user.Name, user.Email, projectName)
	},
}

func init() {
	usersCmd.AddCommand(searchUsersCmd)
	usersCmd.AddCommand(membersCmd)

	membersCmd.AddCommand(listMembersCmd)
	membersCmd.AddCommand(addMemberCmd)
	membersCmd.AddCommand(removeMemberCmd)
}
