package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type PromptConfig struct {
	Prompt    string
	Default   string
	Required  bool
	Password  bool
	Validator func(string) error
}

func Prompt(cfg PromptConfig) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		if cfg.Default != "" {
			fmt.Printf("%s [%s]: ", cfg.Prompt, cfg.Default)
		} else {
			fmt.Printf("%s: ", cfg.Prompt)
		}

		var input string
		var err error

		if cfg.Password {
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return "", fmt.Errorf("failed to read password: %w", err)
			}
			input = string(bytePassword)
			fmt.Println()
		} else {
			input, err = reader.ReadString('\n')
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
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/N]: ", prompt)
		input, err := reader.ReadString('\n')
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
