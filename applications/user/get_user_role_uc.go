// user/get_user_role_uc.go
package user

import (
	"database/sql"
	"fmt"

	"supra/db" // Assumes db is in supra/db
	"supra/logger"
)

// GetUserRoleByEmail returns the user's role if they exist, or an error.
func GetUserRoleByEmail(email string) (string, error) {
	logger.Log.Info(fmt.Sprintf("[user] Checking role for existing user: %s", email))

	// We only need the role column here
	const selectSQL = `SELECT role FROM users WHERE email = $1`

	var role string
	row := db.DB.QueryRow(selectSQL, email)

	err := row.Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			// This is not necessarily an error, but signals the user doesn't exist yet
			logger.Log.Info(fmt.Sprintf("[user] User %s does not exist yet. Defaulting to regular flow.", email))
			return RoleUser, fmt.Errorf("user not found")
		}
		logger.Log.Error(fmt.Sprintf("[user] DB error checking role for %s: %v", email, err))
		return "", fmt.Errorf("database error: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[user] User %s found. Current role: %s", email, role))
	return role, nil
}
