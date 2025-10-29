package participant

import (
	"database/sql"
	"fmt"

	"bk-concerts/db"     // Assumes db is in bk-concerts/db
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The Participant struct is assumed to be defined elsewhere in this package.

// GetParticipant retrieves a single participant record by its ID.
func GetParticipant(userID string) (*Participant, error) {
	logger.Log.Info(fmt.Sprintf("[get-participant-uc] Starting retrieval for UserID: %s", userID))

	id, err := uuid.Parse(userID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-participant-uc] Retrieval failed for %s: Invalid UUID format.", userID))
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	const selectSQL = `
		SELECT user_id, name, wa_num, email, attended
		FROM participant
		WHERE user_id = $1`

	logger.Log.Info(fmt.Sprintf("[get-participant-uc] Executing SELECT query for ID: %s", userID))
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
			logger.Log.Warn(fmt.Sprintf("[get-participant-uc] Retrieval failed for %s: Participant not found.", userID))
			return nil, fmt.Errorf("participant with ID %s not found", userID)
		}
		logger.Log.Error(fmt.Sprintf("[get-participant-uc] Database query error for %s: %v", userID, err))
		return nil, fmt.Errorf("database query error: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-participant-uc] Participant %s retrieved successfully. Name: %s", userID, p.Name))
	return p, nil
}
