package participant

import (
	"fmt"

	"supra/db"     // Using the correct module path
	"supra/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// DeleteParticipant removes a participant record by its ID.
func DeleteParticipant(userID string) (int64, error) {
	logger.Log.Info(fmt.Sprintf("[delete-participant-uc] Deletion initiated for UserID: %s", userID))

	// 1. Validate and convert the ID string to uuid.UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[delete-participant-uc] Deletion failed for %s: Invalid UUID format.", userID))
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}

	// 2. Define the SQL DELETE statement
	const deleteSQL = `
		DELETE FROM participant
		WHERE user_id = $1`

	// 3. Execute the command
	logger.Log.Info(fmt.Sprintf("[delete-participant-uc] Executing DELETE statement for ID: %s", userID))
	result, err := db.DB.Exec(deleteSQL, id)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-participant-uc] Database deletion error for %s: %v", userID, err))
		return 0, fmt.Errorf("database deletion error: %w", err)
	}

	// 4. Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-participant-uc] Failed to retrieve rows affected after delete for %s: %v", userID, err))
		return 0, fmt.Errorf("could not get rows affected: %w", err)
	}

	// If 0 rows were affected, the participant wasn't found.
	if rowsAffected == 0 {
		logger.Log.Warn(fmt.Sprintf("[delete-participant-uc] Deletion failed for %s: Participant not found (0 rows affected).", userID))
		return 0, fmt.Errorf("participant with ID %s not found", userID)
	}

	logger.Log.Info(fmt.Sprintf("[delete-participant-uc] Participant %s deleted successfully. Rows affected: %d", userID, rowsAffected))
	return rowsAffected, nil
}
