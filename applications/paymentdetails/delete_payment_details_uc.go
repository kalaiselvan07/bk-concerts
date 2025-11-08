package paymentdetails

import (
	"fmt"

	"supra/db"     // Assumes this is the package path for your database connection
	"supra/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// DeletePayment removes a payment record from the database by its ID.
// It returns the number of rows affected or an error.
func DeletePayment(paymentID string) (int64, error) {
	logger.Log.Info(fmt.Sprintf("[delete-paymentdetails-uc] Deletion initiated for PaymentID: %s", paymentID))

	// 1. Validate and convert the ID string to uuid.UUID
	id, err := uuid.Parse(paymentID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[delete-paymentdetails-uc] Deletion failed for %s: Invalid UUID format.", paymentID))
		return 0, fmt.Errorf("invalid payment ID format: %w", err)
	}

	// 2. Define the SQL DELETE statement
	const deleteSQL = `
		DELETE FROM payment
		WHERE payment_id = $1`

	// 3. Execute the command
	logger.Log.Info(fmt.Sprintf("[delete-paymentdetails-uc] Executing DELETE statement for ID: %s", paymentID))
	result, err := db.DB.Exec(deleteSQL, id)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-paymentdetails-uc] Database deletion error for %s: %v", paymentID, err))
		return 0, fmt.Errorf("database deletion error: %w", err)
	}

	// 4. Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[delete-paymentdetails-uc] Failed to retrieve rows affected after delete for %s: %v", paymentID, err))
		return 0, fmt.Errorf("could not get rows affected: %w", err)
	}

	// 5. If 0 rows were affected, the payment record wasn't found.
	if rowsAffected == 0 {
		logger.Log.Warn(fmt.Sprintf("[delete-paymentdetails-uc] Deletion failed for %s: Payment not found (0 rows affected).", paymentID))
		return 0, fmt.Errorf("payment with ID %s not found", paymentID)
	}

	logger.Log.Info(fmt.Sprintf("[delete-paymentdetails-uc] Payment %s deleted successfully. Rows affected: %d", paymentID, rowsAffected))
	return rowsAffected, nil
}
