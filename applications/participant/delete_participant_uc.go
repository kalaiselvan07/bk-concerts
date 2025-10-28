package participant

import (
	"fmt"

	"bk-concerts/db"

	"github.com/google/uuid"
)

// DeleteParticipant removes a participant record by its ID.
func DeleteParticipant(userID string) (int64, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}

	const deleteSQL = `
		DELETE FROM participant
		WHERE user_id = $1`

	result, err := db.DB.Exec(deleteSQL, id)
	if err != nil {
		return 0, fmt.Errorf("database deletion error: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("could not get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return 0, fmt.Errorf("participant with ID %s not found", userID)
	}

	return rowsAffected, nil
}
