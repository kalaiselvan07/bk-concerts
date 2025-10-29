package user

import (
	"database/sql"
	"fmt"

	"bk-concerts/db"     // Assumes db is in bk-concerts/db
	"bk-concerts/logger" // ⬅️ Assuming this import path
)

// NOTE: User struct fields (UserID, Email, Password, Role, CreatedAt) are assumed to be defined in this package.

// GetUserByEmail retrieves a single user record by their email address.
func GetUserByEmail(email string) (*User, error) {
	logger.Log.Info(fmt.Sprintf("[user] Attempting to retrieve user by email: %s", email))

	const selectSQL = `
		SELECT user_id, email, password_hash, role, created_at
		FROM users
		WHERE email = $1`

	row := db.DB.QueryRow(selectSQL, email)
	u := &User{}

	var passwordHash []byte

	err := row.Scan(
		&u.UserID,
		&u.Email,
		&passwordHash, // Scan the stored hash
		&u.Role,
		&u.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[user] Retrieval failed for %s: User not found.", email))
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		logger.Log.Error(fmt.Sprintf("[user] Database query failed for %s: %v", email, err))
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// Set the stored hash to the User struct for verification
	u.Password = string(passwordHash)

	logger.Log.Info(fmt.Sprintf("[user] User %s found successfully. Role: %s.", email, u.Role))

	return u, nil
}
