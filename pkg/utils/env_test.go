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
				DB_PATH:    "testdb.sql",
				JWT_SECRET: "test-secret",
			},
			expected: []string{},
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
		os.Setenv("DB_PATH", "testdb.sql")
		os.Setenv("JWT_SECRET", "test-secret")
		defer os.Unsetenv("DB_PATH")
		defer os.Unsetenv("JWT_SECRET")

		config, err := LoadAndValidateEnv()
		if err != nil {
			t.Fatalf("LoadAndValidateEnv() returned an error: %v", err)
		}
		if config.DB_PATH != "testdb.sql" {
			t.Errorf("expected DB_PATH to be 'testdb.sql', got '%s'", config.DB_PATH)
		}
		if config.JWT_SECRET != "test-secret" {
			t.Errorf("expected JWT_SECRET to be 'test-secret', got '%s'", config.JWT_SECRET)
		}
	})
}

func TestGetEnvVars(t *testing.T) {
	expectedConfig := &EnvVar{
		DB_PATH:    "testdb.sql",
		JWT_SECRET: "test-secret",
	}
	Config = expectedConfig

	retrievedConfig := GetEnvVars()
	if !reflect.DeepEqual(retrievedConfig, expectedConfig) {
		t.Errorf("GetEnvVars() = %v, want %v", retrievedConfig, expectedConfig)
	}
}
