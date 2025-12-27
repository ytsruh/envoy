package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ytsruh.com/envoy/cli/config"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new account",
	Long:  "Register a new Envoy account with email and password",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Registering new account...")

		name, err := PromptString("Name", true)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		email, err := PromptEmail("Email")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		password, err := PromptPassword("Password (min 8 characters)")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		if len(password) < 8 {
			fmt.Fprintln(cmd.OutOrStderr(), "Password must be at least 8 characters")
			os.Exit(1)
		}

		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		authResp, err := client.Register(name, email, password)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Registration failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Account registered successfully!")
		fmt.Printf("Welcome, %s!\n", authResp.User.Name)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your account",
	Long:  "Login to your Envoy account using email and password",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Logging in...")

		email, err := PromptEmail("Email")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		password, err := PromptPassword("Password")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		authResp, err := client.Login(email, password)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Login failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Login successful!")
		fmt.Printf("Welcome back, %s!\n", authResp.User.Name)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your account",
	Long:  "Logout from your Envoy account (clears stored token)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.ClearToken(); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Logged out successfully")
	},
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Show your profile information",
	Long:  "Display your current account information",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		profile, err := client.GetProfile()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Profile Information:")
		fmt.Printf("  User ID: %s\n", profile.UserID)
		fmt.Printf("  Email: %s\n", profile.Email)
		fmt.Printf("  Token issued at: %d\n", profile.Iat)
		fmt.Printf("  Token expires at: %d\n", profile.Exp)
	},
}
