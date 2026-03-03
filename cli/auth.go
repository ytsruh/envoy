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

var authCmd = &cli.Command{
	Name:      "auth",
	ShortHelp: "Authentication commands",
	SubCommands: []*cli.Command{
		registerCmd,
		loginCmd,
		logoutCmd,
		profileCmd,
	},
}

var registerCmd = &cli.Command{
	Name:      "register",
	ShortHelp: "Register a new account",
	Exec: func(ctx context.Context, s *cli.State) error {
		fmt.Fprintln(s.Stdout, "Registering new account...")

		name, err := prompts.PromptString("Name", true)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		email, err := prompts.PromptEmail("Email")
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		password, err := prompts.PromptPassword("Password (min 8 characters)")
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(password) < 8 {
			fmt.Fprintln(s.Stderr, "Password must be at least 8 characters")
			os.Exit(1)
		}

		client, err := controllers.NewClient()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		authResp, err := client.Register(name, email, password)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Registration failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintln(s.Stdout, "Account registered successfully!")
		fmt.Fprintf(s.Stdout, "Welcome, %s!\n", authResp.User.Name)
		return nil
	},
}

var loginCmd = &cli.Command{
	Name:      "login",
	ShortHelp: "Login to your account",
	Exec: func(ctx context.Context, s *cli.State) error {
		fmt.Fprintln(s.Stdout, "Logging in...")

		email, err := prompts.PromptEmail("Email")
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		password, err := prompts.PromptPassword("Password")
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client, err := controllers.NewClient()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		authResp, err := client.Login(email, password)
		if err != nil {
			fmt.Fprintf(s.Stderr, "Login failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintln(s.Stdout, "Login successful!")
		fmt.Fprintf(s.Stdout, "Welcome back, %s!\n", authResp.User.Name)
		return nil
	},
}

var logoutCmd = &cli.Command{
	Name:      "logout",
	ShortHelp: "Logout from your account",
	Exec: func(ctx context.Context, s *cli.State) error {
		if err := utils.ClearToken(); err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintln(s.Stdout, "Logged out successfully")
		return nil
	},
}

var profileCmd = &cli.Command{
	Name:      "profile",
	ShortHelp: "Show your profile information",
	Exec: func(ctx context.Context, s *cli.State) error {
		client, err := controllers.RequireToken()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrNoToken {
				fmt.Fprintln(s.Stdout, "Please login first using 'envoy auth login'")
			}
			os.Exit(1)
		}

		profile, err := client.GetProfile()
		if err != nil {
			fmt.Fprintf(s.Stderr, "Error: %v\n", err)
			if err == shared.ErrExpiredToken {
				fmt.Fprintln(s.Stdout, "Your session has expired. Please login again using 'envoy auth login'")
			}
			os.Exit(1)
		}

		fmt.Fprintln(s.Stdout, "Profile Information:")
		fmt.Fprintf(s.Stdout, "  User ID: %s\n", profile.UserID)
		fmt.Fprintf(s.Stdout, "  Email: %s\n", profile.Email)
		fmt.Fprintf(s.Stdout, "  Token issued at: %d\n", profile.Iat)
		fmt.Fprintf(s.Stdout, "  Token expires at: %d\n", profile.Exp)
		return nil
	},
}
