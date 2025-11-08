package participant

import (
	"fmt"
	// Added for rows.Err()
	"supra/db"     // Using the correct module path
	"supra/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The Participant struct is assumed to be defined elsewhere in this package.

// GetAllParticipants retrieves a slice of all participant records.
func GetAllParticipants() ([]*Participant, error) {
	logger.Log.Info("[get-all-participants-uc] Starting retrieval of all participant records.")

	const selectAllSQL = `
		SELECT user_id, name, wa_num, email, attended
		FROM participant
		ORDER BY name`

	logger.Log.Info("[get-all-participants-uc] Executing SELECT all query.")
	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-participants-uc] Database query failed: %v", err))
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	participants := make([]*Participant, 0)
	recordCount := 0

	for rows.Next() {
		p := &Participant{}
		var userIDUUID uuid.UUID

		err := rows.Scan(
			&userIDUUID,
			&p.Name,
			&p.WaNum,
			&p.Email,
			&p.Attended,
		)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[get-all-participants-uc] Error scanning participant row: %v", err))
			return nil, fmt.Errorf("error scanning participant row: %w", err)
		}
		p.UserID = userIDUUID.String()
		participants = append(participants, p)
		recordCount++
	}

	if err = rows.Err(); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-participants-uc] Error during row iteration: %v", err))
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-all-participants-uc] Successfully retrieved %d participant records.", recordCount))
	return participants, nil
}
