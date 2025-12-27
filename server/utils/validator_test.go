package utils

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type TestStruct struct {
	Name     string `validate:"required,project_name"`
	Email    string `validate:"required,email"`
	Age      int    `validate:"min=0,max=150"`
	Optional string `validate:"max=50"`
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		input       TestStruct
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:  "Valid Project",
				Email: "test@example.com",
				Age:   25,
			},
			expectError: false,
		},
		{
			name: "missing required fields",
			input: TestStruct{
				Name: "",
			},
			expectError: true,
			errorMsg:    "Name is required; Email is required",
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:  "Valid Project",
				Email: "invalid-email",
				Age:   25,
			},
			expectError: true,
			errorMsg:    "Email must be a valid email address",
		},
		{
			name: "invalid project name - too short",
			input: TestStruct{
				Name:  "",
				Email: "test@example.com",
				Age:   25,
			},
			expectError: true,
			errorMsg:    "Name is required",
		},
		{
			name: "invalid project name - invalid characters",
			input: TestStruct{
				Name:  "Project@Name!",
				Email: "test@example.com",
				Age:   25,
			},
			expectError: true,
			errorMsg:    "Name must be 1-100 characters and contain only letters, numbers, spaces, hyphens, and underscores",
		},
		{
			name: "age out of range - too young",
			input: TestStruct{
				Name:  "Valid Project",
				Email: "test@example.com",
				Age:   -1,
			},
			expectError: true,
			errorMsg:    "Age must be at least 0 characters long",
		},
		{
			name: "age out of range - too old",
			input: TestStruct{
				Name:  "Valid Project",
				Email: "test@example.com",
				Age:   151,
			},
			expectError: true,
			errorMsg:    "Age must be at most 150 characters long",
		},
		{
			name: "optional field too long",
			input: TestStruct{
				Name:     "Valid Project",
				Email:    "test@example.com",
				Age:      25,
				Optional: string(make([]byte, 51)), // 51 characters
			},
			expectError: true,
			errorMsg:    "Optional must be at most 50 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid name with letters",
			input:    "ProjectName",
			expected: true,
		},
		{
			name:     "valid name with spaces",
			input:    "Project Name",
			expected: true,
		},
		{
			name:     "valid name with hyphens",
			input:    "Project-Name",
			expected: true,
		},
		{
			name:     "valid name with underscores",
			input:    "Project_Name",
			expected: true,
		},
		{
			name:     "valid name with numbers",
			input:    "Project123",
			expected: true,
		},
		{
			name:     "valid mixed characters",
			input:    "Project_123-Test",
			expected: true,
		},
		{
			name:     "valid single character",
			input:    "A",
			expected: true,
		},
		{
			name:     "valid maximum length (100 chars)",
			input:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "too long (101 chars)",
			input:    string(make([]byte, 101)),
			expected: false,
		},
		{
			name:     "invalid characters - @",
			input:    "Project@Name",
			expected: false,
		},
		{
			name:     "invalid characters - #",
			input:    "Project#Name",
			expected: false,
		},
		{
			name:     "invalid characters - $",
			input:    "Project$Name",
			expected: false,
		},
		{
			name:     "invalid characters - %",
			input:    "Project%Name",
			expected: false,
		},
		{
			name:     "invalid characters - ()",
			input:    "Project(Name)",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary validator instance to test custom validation
			v := validator.New()
			v.RegisterValidation("project_name", validateProjectName)

			err := v.Var(tt.input, "project_name")
			result := err == nil

			if result != tt.expected {
				t.Errorf("expected %v for input '%s', got %v", tt.expected, tt.input, result)
			}
		})
	}
}

func TestValidateEnvVarValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty value",
			input:    "",
			expected: true,
		},
		{
			name:     "single character",
			input:    "a",
			expected: true,
		},
		{
			name:     "valid value",
			input:    "my-secret-value",
			expected: true,
		},
		{
			name:     "maximum length (255 chars)",
			input:    string(make([]byte, 255)),
			expected: true,
		},
		{
			name:     "too long (256 chars)",
			input:    string(make([]byte, 256)),
			expected: false,
		},
		{
			name:     "special characters allowed",
			input:    "value_with-special@chars#123!",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			v.RegisterValidation("env_var_value", validateEnvVarValue)

			err := v.Var(tt.input, "env_var_value")
			result := err == nil

			if result != tt.expected {
				t.Errorf("expected %v for input length %d, got %v", tt.expected, len(tt.input), result)
			}
		})
	}
}

func TestValidateWithNilInput(t *testing.T) {
	err := Validate(nil)
	if err == nil {
		t.Error("expected error for nil input but got none")
	}
}

func TestValidateInvalidStructType(t *testing.T) {
	err := Validate("not a struct")
	if err == nil {
		t.Error("expected error for non-struct input but got none")
	}
}
