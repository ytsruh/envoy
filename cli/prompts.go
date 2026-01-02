package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"ytsruh.com/envoy/cli/controllers"

	"golang.org/x/term"
)

var stdinReader = bufio.NewReader(os.Stdin)

type PromptConfig struct {
	Prompt    string
	Default   string
	Required  bool
	Password  bool
	Validator func(string) error
}

func Prompt(cfg PromptConfig) (string, error) {
	for {
		if cfg.Default != "" {
			fmt.Printf("%s [%s]: ", cfg.Prompt, cfg.Default)
		} else {
			fmt.Printf("%s: ", cfg.Prompt)
		}

		var input string
		var err error

		if cfg.Password {
			if term.IsTerminal(int(os.Stdin.Fd())) {
				bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return "", fmt.Errorf("failed to read password: %w", err)
				}
				input = string(bytePassword)
				fmt.Println()
			} else {
				input, err = stdinReader.ReadString('\n')
				if err != nil {
					return "", fmt.Errorf("failed to read password: %w", err)
				}
				input = strings.TrimSpace(input)
			}
		} else {
			input, err = stdinReader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("failed to read input: %w", err)
			}
			input = strings.TrimSpace(input)
		}

		if input == "" {
			if cfg.Default != "" {
				return cfg.Default, nil
			}
			if cfg.Required {
				fmt.Println("This field is required")
				continue
			}
			return "", nil
		}

		if cfg.Validator != nil {
			if err := cfg.Validator(input); err != nil {
				fmt.Printf("Invalid input: %v\n", err)
				continue
			}
		}

		return input, nil
	}
}

func PromptString(prompt string, required bool) (string, error) {
	return Prompt(PromptConfig{
		Prompt:   prompt,
		Required: required,
	})
}

func PromptStringWithDefault(prompt, defaultValue string) (string, error) {
	return Prompt(PromptConfig{
		Prompt:  prompt,
		Default: defaultValue,
	})
}

func PromptPassword(prompt string) (string, error) {
	return Prompt(PromptConfig{
		Prompt:   prompt,
		Required: true,
		Password: true,
	})
}

func PromptEmail(prompt string) (string, error) {
	return Prompt(PromptConfig{
		Prompt:   prompt,
		Required: true,
		Validator: func(s string) error {
			if !strings.Contains(s, "@") || !strings.Contains(s, ".") {
				return fmt.Errorf("invalid email format")
			}
			return nil
		},
	})
}

func Confirm(prompt string) (bool, error) {
	for {
		fmt.Printf("%s [y/N]: ", prompt)
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no", "":
			return false, nil
		default:
			fmt.Println("Please enter 'y' or 'n'")
		}
	}
}

type SelectOption struct {
	Label string
	Value string
}

func PromptSelect(prompt string, options []SelectOption, allowCancel bool) (string, error) {
	for {
		fmt.Printf("\n%s:\n", prompt)
		for i, opt := range options {
			fmt.Printf("  %d. %s\n", i+1, opt.Label)
		}
		if allowCancel {
			fmt.Printf("  0. Cancel\n")
		}
		fmt.Printf("\nSelect an option [1-%d]: ", len(options))

		input, err := stdinReader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		if allowCancel && input == "0" {
			return "", fmt.Errorf("cancelled by user")
		}

		var selection int
		_, err = fmt.Sscanf(input, "%d", &selection)
		if err != nil {
			fmt.Printf("Please enter a number")
			if allowCancel {
				fmt.Printf(" between 0 and %d\n", len(options))
			} else {
				fmt.Printf(" between 1 and %d\n", len(options))
			}
			continue
		}

		if selection < 1 || selection > len(options) {
			fmt.Printf("Invalid selection. Please choose")
			if allowCancel {
				fmt.Printf(" 0-%d\n", len(options))
			} else {
				fmt.Printf(" 1-%d\n", len(options))
			}
			continue
		}

		return options[selection-1].Value, nil
	}
}

func promptForProject(client *controllers.Client) (string, error) {
	projects, err := client.ListProjects()
	if err != nil {
		return "", err
	}

	if len(projects) == 0 {
		return "", fmt.Errorf("no projects found. Please create a project first with 'envoy projects create'")
	}

	options := make([]SelectOption, len(projects))
	for i, p := range projects {
		label := p.Name
		if p.Description != nil && *p.Description != "" {
			label += fmt.Sprintf(" - %s", *p.Description)
		}
		options[i] = SelectOption{Label: label, Value: string(p.ID)}
	}

	return PromptSelect("Select a project", options, true)
}

func promptForEnvironment(client *controllers.Client, projectID string) (string, error) {
	environments, err := client.ListEnvironments(projectID)
	if err != nil {
		return "", err
	}

	if len(environments) == 0 {
		return "", fmt.Errorf("no environments found. Please create an environment first with 'envoy environments create %s'", projectID)
	}

	options := make([]SelectOption, len(environments))
	for i, e := range environments {
		label := e.Name
		if e.Description != nil && *e.Description != "" {
			label += fmt.Sprintf(" - %s", *e.Description)
		}
		options[i] = SelectOption{Label: label, Value: string(e.ID)}
	}

	return PromptSelect("Select an environment", options, true)
}

func promptForVariable(client *controllers.Client, projectID, environmentID string) (string, error) {
	variables, err := client.ListEnvironmentVariables(projectID, environmentID)
	if err != nil {
		return "", err
	}

	if len(variables) == 0 {
		return "", fmt.Errorf("no variables found in this environment")
	}

	options := make([]SelectOption, len(variables))
	for i, v := range variables {
		options[i] = SelectOption{
			Label: fmt.Sprintf("%s = %s", v.Key, v.Value),
			Value: string(v.ID),
		}
	}

	return PromptSelect("Select a variable", options, true)
}
