package utils

import (
	"os"
	"reflect"
	"testing"
)

func TestValidateEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		env      EnvVar
		expected []string
	}{
		{
			name: "All variables present",
			env: EnvVar{
				TURSO_DATABASE_URL: "url",
				TURSO_AUTH_TOKEN:   "token",
			},
			expected: []string{},
		},
		{
			name: "One variable missing",
			env: EnvVar{
				TURSO_DATABASE_URL: "url",
			},
			expected: []string{"TURSO_AUTH_TOKEN"},
		},
		{
			name:     "All variables missing",
			env:      EnvVar{},
			expected: []string{"TURSO_DATABASE_URL", "TURSO_AUTH_TOKEN"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing := ValidateEnvVars(tt.env)
			if len(missing) == 0 && len(tt.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(missing, tt.expected) {
				t.Errorf("ValidateEnvVars() = %v, want %v", missing, tt.expected)
			}
		})
	}
}

func TestLoadAndValidateEnv(t *testing.T) {
	t.Run("All variables present", func(t *testing.T) {
		os.Setenv("TURSO_DATABASE_URL", "test_url")
		os.Setenv("TURSO_AUTH_TOKEN", "test_token")
		defer os.Unsetenv("TURSO_DATABASE_URL")
		defer os.Unsetenv("TURSO_AUTH_TOKEN")

		config, err := LoadAndValidateEnv()
		if err != nil {
			t.Fatalf("LoadAndValidateEnv() returned an error: %v", err)
		}
		if config.TURSO_DATABASE_URL != "test_url" {
			t.Errorf("expected TURSO_DATABASE_URL to be 'test_url', got '%s'", config.TURSO_DATABASE_URL)
		}
		if config.TURSO_AUTH_TOKEN != "test_token" {
			t.Errorf("expected TURSO_AUTH_TOKEN to be 'test_token', got '%s'", config.TURSO_AUTH_TOKEN)
		}
	})

	t.Run("Missing one variable", func(t *testing.T) {
		os.Setenv("TURSO_DATABASE_URL", "test_url")
		defer os.Unsetenv("TURSO_DATABASE_URL")

		_, err := LoadAndValidateEnv()
		if err == nil {
			t.Fatal("LoadAndValidateEnv() should have returned an error but didn't")
		}
		expectedError := "missing environment variables: [TURSO_AUTH_TOKEN]"
		if err.Error() != expectedError {
			t.Errorf("LoadAndValidateEnv() error = %v, want %v", err, expectedError)
		}
	})
}

func TestGetEnvVars(t *testing.T) {
	expectedConfig := &EnvVar{
		TURSO_DATABASE_URL: "global_url",
		TURSO_AUTH_TOKEN:   "global_token",
	}
	Config = expectedConfig

	retrievedConfig := GetEnvVars()
	if !reflect.DeepEqual(retrievedConfig, expectedConfig) {
		t.Errorf("GetEnvVars() = %v, want %v", retrievedConfig, expectedConfig)
	}
}
