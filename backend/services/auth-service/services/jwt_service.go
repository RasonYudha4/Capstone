package services

import (
	"fmt"
	"time"

	"auth-service/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles creation and (future) validation of JSON Web Tokens.
type JWTService struct {
	secretKey []byte
}

// NewJWTService creates a JWTService with the configured signing secret.
func NewJWTService() *JWTService {
	return &JWTService{
		secretKey: []byte(config.JWTSecret),
	}
}

// Claims defines the JWT payload.
// Embedding jwt.RegisteredClaims gives us standard fields (exp, iat, iss, etc.).
type Claims struct {
	UserId string `json:"user_id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the given email.
// The token includes the user's email, their role, and an expiration timestamp.
func (s *JWTService) GenerateToken(email string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(config.JWTExpiration)

	claims := &Claims{
		UserId: "bc6adda6-bfe6-4a6c-a637-35dd4ba4e562",
		Email: email,
		Role:  config.DefaultRole,
		RegisteredClaims: jwt.RegisteredClaims{
			// Subject identifies the principal (the user).
			Subject: email,
			// IssuedAt records when the token was created.
			IssuedAt: jwt.NewNumericDate(now),
			// ExpiresAt marks when the token becomes invalid.
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			// Issuer identifies which service minted this token.
			Issuer: "auth-service",
		},
	}

	// Create the token with HMAC-SHA256 signing method.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key.
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}
