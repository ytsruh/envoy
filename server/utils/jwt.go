package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	shared "ytsruh.com/envoy/shared"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// JWTHeader represents the JWT header
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// GenerateJWT creates a new JWT token for a user with 7 days expiration
func GenerateJWT(userID, email, secret string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days

	header := JWTHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Iat:    now.Unix(),
		Exp:    expiresAt.Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	message := headerEncoded + "." + claimsEncoded
	signature := createSignature(message, secret)

	token := message + "." + signature

	return token, nil
}

// ValidateJWT validates a JWT token and returns the claims if valid
func ValidateJWT(token, secret string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, shared.ErrInvalidToken
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signature := parts[2]

	message := headerEncoded + "." + claimsEncoded
	expectedSignature := createSignature(message, secret)

	if signature != expectedSignature {
		return nil, shared.ErrInvalidSignature
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	if time.Now().Unix() > claims.Exp {
		return nil, shared.ErrExpiredToken
	}

	return &claims, nil
}

// createSignature creates an HMAC SHA256 signature
func createSignature(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
