package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenProvider is the concrete implementation of auth.TokenProvider
type tokenProvider struct {
	secretKey []byte
	issuer    string
}

// NewTokenProvider creates a new JWT service.
// In production, the secretKey MUST come from environment variables.
func NewTokenProvider(secret string, issuer string) *tokenProvider {
	return &tokenProvider{
		secretKey: []byte(secret),
		issuer:    issuer,
	}
}

// GenerateOTPToken creates a short-lived token just for the MPIN setup step.
func (t *tokenProvider) GenerateOTPToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iss": t.issuer,
		"aud": "mpin_setup", // SDE3 Tip: Use audience claims to restrict token usage
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secretKey)
}

// GenerateSessionTokens creates the long-lived Access and Refresh tokens for the Dashboard.
func (t *tokenProvider) GenerateSessionTokens(userID string) (string, string, error) {
	// 1. Generate Access Token (Short lived: 1 Hour)
	accessClaims := jwt.MapClaims{
		"sub": userID,
		"iss": t.issuer,
		"aud": "api_access",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(t.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Generate Refresh Token (Long lived: 30 Days)
	refreshClaims := jwt.MapClaims{
		"sub": userID,
		"iss": t.issuer,
		"aud": "token_refresh",
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(t.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
