package seat

import (
	"fmt"

	"bk-concerts/db"     // ⬅️ Import your database package
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// DeleteSeat removes a seat record from the database by its ID.
// It returns the number of rows affected or an error.
func DeleteSeat(seatID string) (int64, error) {
	logger.Log.Info(fmt.Sprintf("[delete-seat-uc] Deletion initiated for seatID: %s", seatID))

	// 1. Validate and convert the seatID string to uuid.UUID
	id, err := uuid.Parse(seatID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[delete-seat-uc] Deletion failed for %s: Invalid UUID format.", seatID))
		return 0, fmt.Errorf("invalid seat ID format: %w", err)
	}

	// 2. Define the SQL DELETE statement
	const deleteSQL = `
		DELETE FROM seat
		WHERE seat_id = $1`

	// 3. Execute the command
	logger.Log.Info(fmt.Sprintf("[delete-seat-uc] Executing DELETE statement for ID: %s", seatID))
	result, err := db.DB.Exec(deleteSQL, id)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-seat-uc] Database deletion error for %s: %v", seatID, err))
		return 0, fmt.Errorf("database deletion error: %w", err)
	}

	// 4. Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-seat-uc] Failed to retrieve rows affected after delete for %s: %v", seatID, err))
		return 0, fmt.Errorf("could not get rows affected: %w", err)
	}

	// If 0 rows were affected, the seat wasn't found.
	if rowsAffected == 0 {
		logger.Log.Warn(fmt.Sprintf("[delete-seat-uc] Deletion failed for %s: Seat not found (0 rows affected).", seatID))
		return 0, fmt.Errorf("seat with ID %s not found", seatID)
	}

	logger.Log.Info(fmt.Sprintf("[delete-seat-uc] Seat %s deleted successfully. Rows affected: %d", seatID, rowsAffected))
	return rowsAffected, nil
}
