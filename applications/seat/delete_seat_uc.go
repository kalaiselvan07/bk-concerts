package seat

import (
	"fmt"

	"bk-concerts/db" // ⬅️ Import your database package

	"github.com/google/uuid"
)

// DeleteSeat removes a seat record from the database by its ID.
// It returns the number of rows affected or an error.
func DeleteSeat(seatID string) (int64, error) {
	// 1. Validate and convert the seatID string to uuid.UUID
	id, err := uuid.Parse(seatID)
	if err != nil {
		return 0, fmt.Errorf("invalid seat ID format: %w", err)
	}

	// 2. Define the SQL DELETE statement
	const deleteSQL = `
		DELETE FROM seat
		WHERE seat_id = $1`

	// 3. Execute the command
	result, err := db.DB.Exec(deleteSQL, id)
	if err != nil {
		return 0, fmt.Errorf("database deletion error: %w", err)
	}

	// 4. Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("could not get rows affected: %w", err)
	}

	// If 0 rows were affected, the seat wasn't found.
	if rowsAffected == 0 {
		return 0, fmt.Errorf("seat with ID %s not found", seatID)
	}

	return rowsAffected, nil
}
