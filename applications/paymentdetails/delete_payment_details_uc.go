package paymentdetails

import (
	"fmt"

	"bk-concerts/db" // Assumes this is the package path for your database connection

	"github.com/google/uuid"
)

// DeletePayment removes a payment record from the database by its ID.
// It returns the number of rows affected or an error.
func DeletePayment(paymentID string) (int64, error) {
	// 1. Validate and convert the ID string to uuid.UUID
	id, err := uuid.Parse(paymentID)
	if err != nil {
		return 0, fmt.Errorf("invalid payment ID format: %w", err)
	}

	// 2. Define the SQL DELETE statement
	const deleteSQL = `
		DELETE FROM payment
		WHERE payment_id = $1`

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

	// 5. If 0 rows were affected, the payment record wasn't found.
	if rowsAffected == 0 {
		return 0, fmt.Errorf("payment with ID %s not found", paymentID)
	}

	return rowsAffected, nil
}
