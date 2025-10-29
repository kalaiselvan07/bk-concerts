package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID    uuid.UUID `json:"userID"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`    // Store hashed password, but never return it
	Role      string    `json:"role"` // "admin" or "user"
	CreatedAt time.Time `json:"createdAt"`
}

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// LoginParams for incoming credentials
type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}
