package concert

import (
	"fmt"

	"bk-concerts/db" // Using the correct module path

	"github.com/google/uuid"
)

// DeleteConcert removes a concert record from the database by its ID.
// It returns the number of rows affected or an error.
func DeleteConcert(concertID string) (int64, error) {
	// 1. Validate and convert the ID string to uuid.UUID
	id, err := uuid.Parse(concertID)
	if err != nil {
		return 0, fmt.Errorf("invalid concert ID format: %w", err)
	}

	// 2. Define the SQL DELETE statement
	const deleteSQL = `
		DELETE FROM concert
		WHERE concert_id = $1`

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

	// If 0 rows were affected, the concert wasn't found.
	if rowsAffected == 0 {
		return 0, fmt.Errorf("concert with ID %s not found", concertID)
	}

	return rowsAffected, nil
}
