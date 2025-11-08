package auth

import (
	// "database/sql"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"supra/applications/user"
	"supra/db"
	"supra/logger" // ⬅️ Assuming this import path
)

// VerifyOTP checks the submitted code, logs the user in, and deletes the code.
func VerifyOTP(email, code string) (token string, role string, err error) {
	logger.Log.Info(fmt.Sprintf("[auth] Verification attempt for email: %s", email))

	// 1. Retrieve the user record first
	u, err := user.GetUserByEmail(email)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[auth] Verification failed for %s: User not found.", email))
		return "", "", errors.New("invalid code or user not found")
	}

	logger.Log.Info(fmt.Sprintf("[auth] User %s found. Proceeding to OTP validation.", email))

	// 2. Check OTP in the database
	const selectOTP = `
		SELECT expires_at, code 
		FROM otp_codes 
		WHERE user_email = $1 AND code = $2`

	var expiresAt time.Time
	var storedCode string

	row := db.DB.QueryRow(selectOTP, email, code)
	err = row.Scan(&expiresAt, &storedCode)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[auth] Verification failed for %s: OTP not found in DB or code mismatch.", email))
			return "", "", errors.New("invalid OTP code")
		}
		logger.Log.Error(fmt.Sprintf("[auth] DB error retrieving OTP for %s: %v", email, err))
		return "", "", errors.New("database error during verification")
	}

	// Log successful code match before checking expiry
	logger.Log.Info(fmt.Sprintf("[auth] OTP code matched for %s. Checking expiry...", email))

	// 3. Check for Expiration
	if time.Now().After(expiresAt) {
		logger.Log.Warn(fmt.Sprintf("[auth] Verification failed for %s: OTP expired at %s.", email, expiresAt.Format(time.RFC3339)))
		return "", "", errors.New("OTP expired. Please request a new one")
	}

	// 4. Clean up the OTP code (prevent reuse)
	const deleteOTP = `DELETE FROM otp_codes WHERE user_email = $1`
	if _, err := db.DB.Exec(deleteOTP, email); err != nil {
		// Log the failure to cleanup, but do not stop the login process
		logger.Log.Error(fmt.Sprintf("[auth] Cleanup failed: Failed to delete used OTP for %s: %v", email, err))
	} else {
		logger.Log.Info(fmt.Sprintf("[auth] OTP cleaned up successfully for %s.", email))
	}

	// 5. Generate and return the JWT
	token, err = GenerateJWT(u.UserID.String(), u.Email, u.Role)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to generate final JWT for %s: %v", email, err))
		return "", "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[auth] Verification successful for %s. JWT issued. Role: %s.", email, u.Role))
	return token, u.Role, nil
}
