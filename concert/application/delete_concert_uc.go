package application

import (
	"fmt"
	"log/slog"

	"supra/db"     // Using the correct module path
	"supra/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

type DeleteConcertUC struct {
	log *slog.Logger
}

func NewDeleteConcertUC(log *slog.Logger) *DeleteConcertUC {
	return &DeleteConcertUC{
		log: log,
	}
}

// DeleteConcert removes a concert record from the database by its ID.
// It returns the number of rows affected or an error.
func (uc *DeleteConcertUC) Invoke(concertID string) (int64, error) {
	logger.Log.Info(fmt.Sprintf("[delete-concert-uc] Deletion initiated for concertID: %s", concertID))

	// 1. Validate and convert the ID string to uuid.UUID
	id, err := uuid.Parse(concertID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[delete-concert-uc] Deletion failed for %s: Invalid UUID format.", concertID))
		return 0, fmt.Errorf("invalid concert ID format: %w", err)
	}

	// 2. Define the SQL DELETE statement
	const deleteSQL = `
		DELETE FROM concert
		WHERE concert_id = $1`

	// 3. Execute the command
	logger.Log.Info(fmt.Sprintf("[delete-concert-uc] Executing DELETE statement for ID: %s", concertID))
	result, err := db.DB.Exec(deleteSQL, id)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-concert-uc] Database deletion error for %s: %v", concertID, err))
		return 0, fmt.Errorf("database deletion error: %w", err)
	}

	// 4. Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-concert-uc] Failed to retrieve rows affected after delete for %s: %v", concertID, err))
		return 0, fmt.Errorf("could not get rows affected: %w", err)
	}

	// If 0 rows were affected, the concert wasn't found.
	if rowsAffected == 0 {
		logger.Log.Warn(fmt.Sprintf("[delete-concert-uc] Deletion failed for %s: Concert not found (0 rows affected).", concertID))
		return 0, fmt.Errorf("concert with ID %s not found", concertID)
	}

	logger.Log.Info(fmt.Sprintf("[delete-concert-uc] Concert %s deleted successfully. Rows affected: %d", concertID, rowsAffected))
	return rowsAffected, nil
}
