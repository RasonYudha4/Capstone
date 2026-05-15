package services

import (
	"fmt"
	"time"

	"auth-service/config"

	"github.com/golang-jwt/jwt/v5"
)

// handles creation and validation of JSON Web Tokens.
type JWTService struct {
	secretKey []byte
}

// creates a JWTService with the configured signing secret.
func NewJWTService() *JWTService {
	return &JWTService{
		secretKey: []byte(config.JWTSecret),
	}
}

// defines the JWT payload.
// embedding jwt.RegisteredClaims gives standard fields (exp, iat, iss, etc.).

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

// creates a signed JWT for the given user.
// token includes the user's ID, email, role, and an expiration timestamp.
func (s *JWTService) GenerateToken(userID, email, role string) (string, error) {
    now := time.Now()
    expiresAt := now.Add(config.AccessTokenExpiry)

    claims := &Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            // ... rest of your registered claims
        },
    }
    // ... rest of function
}
		RegisteredClaims: jwt.RegisteredClaims{
			// the principal (the user).
			Subject: userID,
			// records when the token was created.
			IssuedAt: jwt.NewNumericDate(now),
			// marks when the token becomes invalid.
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			// identifies which service minted this token.
			Issuer: "auth-service",
		},
	}

	// HMAC-SHA256 signing method for creating token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// signthe token with our secret key.
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}

// parses a raw JWT string, verifies its signature and expiration,
// and returns the embedded claims if everything checks out.
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	// parse the token and validate the signature using our secret key.
	// verifies that the signing method is HMAC,
	// preventing algorithm-switching attacks (e.g. "none" or RSA → HMAC).
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Ensure the signing method is what we expect.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token validation failed")
	}

	return claims, nil
}
