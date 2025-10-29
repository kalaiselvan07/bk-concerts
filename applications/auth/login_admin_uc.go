package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"bk-concerts/applications/user" // Import user use cases
	"bk-concerts/logger"            // ⬅️ Assuming this import path
)

// LoginAdmin handles the secure login flow for the Administrator.
func LoginAdmin(email, password string) (token string, role string, err error) {
	logger.Log.Info(fmt.Sprintf("[auth] Admin login attempt started for email: %s", email))

	// 1. Retrieve the user record by email
	u, err := user.GetUserByEmail(email)
	if err != nil {
		// Log the failure to find the user
		logger.Log.Warn(fmt.Sprintf("[auth] Admin login failed for %s: User not found or DB error: %v", email, err))
		return "", "", errors.New("invalid credentials")
	}

	// 2. Check if the user has the necessary permissions
	if u.Role != user.RoleAdmin {
		logger.Log.Warn(fmt.Sprintf("[auth] Admin login blocked for %s: Role is '%s', not 'admin'.", email, u.Role))
		return "", "", errors.New("access denied: account is not an administrator")
	}

	logger.Log.Info(fmt.Sprintf("[auth] User %s found with role 'admin'. Proceeding to password comparison.", email))

	// 3. Compare the provided password against the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// This handles bcrypt.ErrMismatchedHashAndPassword
		logger.Log.Warn(fmt.Sprintf("[auth] Admin login failed for %s: Password mismatch.", email))
		return "", "", errors.New("invalid credentials")
	}

	logger.Log.Info(fmt.Sprintf("[auth] Password matched for Admin %s. Generating JWT.", email))

	// 4. If credentials match, generate the JWT
	token, err = GenerateJWT(u.UserID.String(), u.Email, u.Role)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to generate JWT for %s: %v", email, err))
		return "", "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[auth] Admin login successful for %s. JWT issued.", email))
	return token, u.Role, nil
}
