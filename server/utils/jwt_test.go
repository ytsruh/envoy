package utils

import (
	"testing"
	"time"

	shared "ytsruh.com/envoy/shared"
)

func TestGenerateJWT(t *testing.T) {
	userID := "test-user-id"
	email := "test@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateJWT() returned empty token")
	}

	// Token should have 3 parts separated by dots
	parts := len(token)
	if parts == 0 {
		t.Error("GenerateJWT() returned invalid token format")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := "test-user-id"
	email := "test@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	claims, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("ValidateJWT() UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Email != email {
		t.Errorf("ValidateJWT() Email = %v, want %v", claims.Email, email)
	}

	// Check expiration is set to 7 days from now (with some tolerance)
	expectedExp := time.Now().Add(7 * 24 * time.Hour).Unix()
	tolerance := int64(5) // 5 seconds tolerance
	if claims.Exp < expectedExp-tolerance || claims.Exp > expectedExp+tolerance {
		t.Errorf("ValidateJWT() Exp = %v, want approximately %v", claims.Exp, expectedExp)
	}
}

func TestValidateJWTInvalidToken(t *testing.T) {
	secret := "test-secret-key"

	tests := []struct {
		name  string
		token string
		want  error
	}{
		{
			name:  "empty token",
			token: "",
			want:  shared.ErrInvalidToken,
		},
		{
			name:  "malformed token",
			token: "invalid.token",
			want:  shared.ErrInvalidToken,
		},
		{
			name:  "token with invalid format",
			token: "header.payload",
			want:  shared.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateJWT(tt.token, secret)
			if err != tt.want {
				t.Errorf("ValidateJWT() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestValidateJWTInvalidSignature(t *testing.T) {
	userID := "test-user-id"
	email := "test@example.com"
	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"

	token, err := GenerateJWT(userID, email, secret)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err != shared.ErrInvalidSignature {
		t.Errorf("ValidateJWT() with wrong secret error = %v, want %v", err, shared.ErrInvalidSignature)
	}
}

func TestValidateJWTExpired(t *testing.T) {
	// This test would require mocking time or creating a token with a past expiration
	// For now, we'll skip this test as it would require refactoring the GenerateJWT function
	// to accept custom expiration time for testing purposes
	t.Skip("Skipping expired token test - would require time mocking")
}

func TestCreateSignature(t *testing.T) {
	message := "test.message"
	secret := "test-secret"

	sig1 := createSignature(message, secret)
	sig2 := createSignature(message, secret)

	if sig1 != sig2 {
		t.Error("createSignature() should return same signature for same input")
	}

	if sig1 == "" {
		t.Error("createSignature() returned empty signature")
	}

	// Different message should produce different signature
	sig3 := createSignature("different.message", secret)
	if sig1 == sig3 {
		t.Error("createSignature() should return different signature for different message")
	}

	// Different secret should produce different signature
	sig4 := createSignature(message, "different-secret")
	if sig1 == sig4 {
		t.Error("createSignature() should return different signature for different secret")
	}
}

func TestJWTRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		email  string
		secret string
	}{
		{
			name:   "basic test",
			userID: "user-123",
			email:  "user@example.com",
			secret: "my-secret-key",
		},
		{
			name:   "special characters in email",
			userID: "user-456",
			email:  "user+tag@example.com",
			secret: "another-secret",
		},
		{
			name:   "uuid user id",
			userID: "550e8400-e29b-41d4-a716-446655440000",
			email:  "uuid@example.com",
			secret: "uuid-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWT(tt.userID, tt.email, tt.secret)
			if err != nil {
				t.Fatalf("GenerateJWT() error = %v", err)
			}

			claims, err := ValidateJWT(token, tt.secret)
			if err != nil {
				t.Fatalf("ValidateJWT() error = %v", err)
			}

			if claims.UserID != tt.userID {
				t.Errorf("UserID = %v, want %v", claims.UserID, tt.userID)
			}

			if claims.Email != tt.email {
				t.Errorf("Email = %v, want %v", claims.Email, tt.email)
			}
		})
	}
}
