package utils

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
var Config *EnvVar

// EnvVar struct holds all environment variables used by the application
type EnvVar struct {
	JWT_SECRET string
	DB_URL     string
	DB_TOKEN   string
}

// LoadAndValidateEnv loads environment variables from .env file (in development) or from system environment (in production) and validates that all required variables are set. Returns the loaded environment variables and an error if any required variable is missing
func LoadAndValidateEnv() (*EnvVar, error) {
	// Load from .env file if it exists (typically in development). This will be silently ignored in production where env vars are set in the environment
	_ = godotenv.Load()

	env := EnvVar{
		DB_URL:     os.Getenv("DB_URL"),
		DB_TOKEN:   os.Getenv("DB_TOKEN"),
		JWT_SECRET: os.Getenv("JWT_SECRET"),
	}

	// Validate that all required environment variables are set
	missingVars := ValidateEnvVars(env)
	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing environment variables: %v", missingVars)
	}

	Config = &env
	return &env, nil
}

// ValidateEnvVars checks if all fields in the EnvVar struct are set. Returns a slice of names of missing environment variables
func ValidateEnvVars(env EnvVar) []string {
	v := reflect.ValueOf(env)
	if v.Kind() != reflect.Struct {
		log.Fatal("Invalid struct")
	}

	var missingVars []string
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.String && v.Field(i).String() == "" {
			missingVars = append(missingVars, v.Type().Field(i).Name)
		}
	}

	return missingVars
}

// GetEnvVars returns the current environment variables configuration
func GetEnvVars() *EnvVar {
	return Config
}
