package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword() should not return the plain password")
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "mySecurePassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// bcrypt generates different hashes for the same password due to salt
	if hash1 == hash2 {
		t.Error("HashPassword() should generate unique hashes for same password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Test correct password
	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash() should return true for correct password")
	}

	// Test incorrect password
	if CheckPasswordHash("wrongPassword", hash) {
		t.Error("CheckPasswordHash() should return false for incorrect password")
	}
}

func TestCheckPasswordHashWithVariousPasswords(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		testPassword  string
		expectedMatch bool
	}{
		{
			name:          "exact match",
			password:      "correctPassword",
			testPassword:  "correctPassword",
			expectedMatch: true,
		},
		{
			name:          "wrong password",
			password:      "correctPassword",
			testPassword:  "wrongPassword",
			expectedMatch: false,
		},
		{
			name:          "case sensitive",
			password:      "Password123",
			testPassword:  "password123",
			expectedMatch: false,
		},
		{
			name:          "empty password does not match",
			password:      "somePassword",
			testPassword:  "",
			expectedMatch: false,
		},
		{
			name:          "special characters",
			password:      "p@ssw0rd!#$%",
			testPassword:  "p@ssw0rd!#$%",
			expectedMatch: true,
		},
		{
			name:          "long password",
			password:      "thisIsAVeryLongPasswordWithLotsOfCharacters123456789",
			testPassword:  "thisIsAVeryLongPasswordWithLotsOfCharacters123456789",
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword() error = %v", err)
			}

			result := CheckPasswordHash(tt.testPassword, hash)
			if result != tt.expectedMatch {
				t.Errorf("CheckPasswordHash() = %v, want %v", result, tt.expectedMatch)
			}
		})
	}
}

func TestCheckPasswordHashInvalidHash(t *testing.T) {
	password := "testPassword"
	invalidHash := "not-a-valid-bcrypt-hash"

	result := CheckPasswordHash(password, invalidHash)
	if result {
		t.Error("CheckPasswordHash() should return false for invalid hash")
	}
}

func TestPasswordHashRoundTrip(t *testing.T) {
	passwords := []string{
		"simple",
		"WithNumbers123",
		"With Spaces",
		"Special!@#$%^&*()",
		"VeryLongPasswordThatExceedsNormalLengthToTestBcryptHandling123456789",
		"🔐emoji🔑",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hash, err := HashPassword(password)
			if err != nil {
				t.Fatalf("HashPassword() error = %v for password %q", err, password)
			}

			if !CheckPasswordHash(password, hash) {
				t.Errorf("Password round trip failed for %q", password)
			}
		})
	}
}
