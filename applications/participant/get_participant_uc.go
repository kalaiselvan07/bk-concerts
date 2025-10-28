package participant

import (
	"database/sql"
	"fmt"

	"bk-concerts/db" // Assumes db is in bk-concerts/db

	"github.com/google/uuid"
)

// GetParticipant retrieves a single participant record by its ID.
func GetParticipant(userID string) (*Participant, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	const selectSQL = `
		SELECT user_id, name, wa_num, email, attended
		FROM participant
		WHERE user_id = $1`

	row := db.DB.QueryRow(selectSQL, id)
	p := &Participant{}

	// Note: If UserID in Participant struct is a string, adjust the scan target.
	// Assuming UserID is string for now, but scanning UUID for DB:
	var userIDUUID uuid.UUID

	err = row.Scan(
		&userIDUUID,
		&p.Name,
		&p.WaNum,
		&p.Email,
		&p.Attended,
	)

	// Set the string ID on the struct
	p.UserID = userIDUUID.String()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("participant with ID %s not found", userID)
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}
	return p, nil
}
