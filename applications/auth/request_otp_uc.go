package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"bk-concerts/applications/user"
	"bk-concerts/db"
	"bk-concerts/logger" // ⬅️ Assuming this import path
)

const otpExpiry = 5 * time.Minute // OTP valid for 5 minutes

// generateOTP creates a secure 6-digit random code (100,000 to 999,999).
func generateOTP() (string, error) {
	// Define the range: 100,000 to 999,999 (i.e., a range of 900,000)
	const min int64 = 100000
	const max int64 = 900000 // Upper bound is max + min = 999,999

	// Generate a secure random number between 0 and 900,000
	// rand.Int() returns a random BigInt, which is always non-negative.
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", fmt.Errorf("crypto rand failed: %w", err)
	}

	// Convert to int64 and add the minimum value to ensure 6 digits
	otpValue := n.Int64() + min

	// Format to ensure 6 digits (though guaranteed by the math above)
	return fmt.Sprintf("%06d", otpValue), nil
}

// RequestUserOTP finds the user, creates them if necessary, generates an OTP, and sends it.
// Note: This returns an empty token because the user must verify the OTP first.
func RequestUserOTP(email string) (token string, role string, err error) {
	logger.Log.Info(fmt.Sprintf("[auth] Starting OTP process for email: %s", email))

	// 1. Find or Create User
	u, err := user.GetUserByEmail(email)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("[auth] User %s not found. Attempting creation.", email))
		if u, err = user.CreateUser(email); err != nil {
			logger.Log.Error(fmt.Sprintf("[auth] Failed to create user %s: %v", email, err))
			return "", "", fmt.Errorf("failed to create user: %w", err)
		}
		logger.Log.Info(fmt.Sprintf("[auth] User %s created successfully with ID: %s", email, u.UserID))
	} else {
		logger.Log.Info(fmt.Sprintf("[auth] User %s found. Role: %s. Reissuing OTP.", email, u.Role))
	}

	// 2. Generate and Store OTP
	code, err := generateOTP()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[auth] Generated OTP code: %s", code)) // WARNING: Do not log OTPs in production!

	expiresAt := time.Now().Add(otpExpiry)

	// Use ON CONFLICT DO UPDATE to handle resend requests gracefully
	const insertOTP = `
		INSERT INTO otp_codes (code, user_email, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_email) DO UPDATE 
		SET code = EXCLUDED.code, expires_at = EXCLUDED.expires_at;`

	_, err = db.DB.ExecContext(context.Background(), insertOTP, code, email, expiresAt)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to save/update OTP for %s: %v", email, err))
		return "", "", fmt.Errorf("failed to save OTP: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[auth] OTP saved/updated successfully. Expires at: %s", expiresAt.Format(time.RFC3339)))

	// 3. Send Email (now real)
	if err := SendOTP(email, code); err != nil { // ⬅️ CALL THE REAL SENDER
		logger.Log.Error(fmt.Sprintf("[auth] Failed to dispatch email for %s: %v", email, err))
		return "", "", fmt.Errorf("failed to send OTP email: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[auth] OTP dispatch triggered for %s.", email))

	return "", u.Role, nil
}
