package auth

import (
	"fmt"
	"time"

	"supra/logger"

	"github.com/golang-jwt/jwt/v5"
)

// Define your JWT signing key (KEEP THIS SECURE AND SECRET)
var jwtSecret = []byte("YOUR_SUPER_SECURE_JWT_SIGNING_KEY_12345")

// init runs once when the package is imported
func init() {
	// Log that the JWT configuration has been loaded upon package initialization
	logger.Log.Info("[auth] JWT configuration loaded and signing key initialized.")
}

// Claims structure to store user info in the token
type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new signed JWT for the user.
func GenerateJWT(userID, email, role string) (string, error) {
	claims := UserClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // 24-hour expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to sign JWT for user %s (%s): %v", userID, email, err))
		return "", err
	}

	logger.Log.Info(fmt.Sprintf("[auth] Successfully generated JWT for user %s (Role: %s).", userID, role))

	return tokenString, nil
}
