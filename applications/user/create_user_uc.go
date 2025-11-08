package user

import (
	"fmt"
	"time"

	"supra/db"
	"supra/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: User struct and RoleUser constant are assumed to be defined elsewhere in the package.

// CreateUser creates a new user with the default 'user' role.
func CreateUser(email string) (*User, error) {
	newID := uuid.New()

	u := &User{
		UserID:    newID,
		Email:     email,
		Role:      RoleUser,
		CreatedAt: time.Now(),
	}

	logger.Log.Info(fmt.Sprintf("[user] Attempting to create new user with email: %s and generated ID: %s", email, newID.String()))

	const insertSQL = `
		INSERT INTO users (user_id, email, role, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id, email, role, created_at`

	// Note: We leave password_hash NULL for regular users

	err := db.DB.QueryRow(insertSQL,
		u.UserID, u.Email, u.Role, u.CreatedAt,
	).Scan(
		&u.UserID, &u.Email, &u.Role, &u.CreatedAt,
	)

	if err != nil {
		logger.Log.Error(fmt.Sprintf("[user] Failed to insert new user %s into database: %v", email, err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[user] New user %s created successfully. Role: %s", email, u.Role))
	return u, nil
}
