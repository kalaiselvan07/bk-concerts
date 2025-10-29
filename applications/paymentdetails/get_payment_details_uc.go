package paymentdetails

import (
	"database/sql"
	"fmt"

	"bk-concerts/db"     // Assumes this is the package path for your database connection
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The PaymentDetails struct is assumed to be defined elsewhere in this package.

// GetPayment retrieves a single payment record from the database by its ID.
func GetPayment(paymentID string) (*PaymentDetails, error) {
	logger.Log.Info(fmt.Sprintf("[get-payment-details-uc] Starting retrieval for PaymentID: %s", paymentID))

	// 1. Validate and convert the ID string to uuid.UUID
	id, err := uuid.Parse(paymentID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-payment-details-uc] Retrieval failed for %s: Invalid UUID format.", paymentID))
		return nil, fmt.Errorf("invalid payment ID format: %w", err)
	}

	// 2. Define the SQL query
	const selectSQL = `
		SELECT payment_id, payment_type, details, notes
		FROM payment
		WHERE payment_id = $1`

	// 3. Execute the query
	logger.Log.Info(fmt.Sprintf("[get-payment-details-uc] Executing SELECT query for ID: %s", paymentID))
	row := db.DB.QueryRow(selectSQL, id)
	p := &PaymentDetails{}

	// 4. Scan the row data into the struct fields
	err = row.Scan(
		&p.PaymentID,
		&p.PaymentType,
		&p.Details,
		&p.Notes,
	)

	// 5. Check the result of the scan
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[get-payment-details-uc] Retrieval failed for %s: Payment not found.", paymentID))
			return nil, fmt.Errorf("payment with ID %s not found", paymentID)
		}
		logger.Log.Error(fmt.Sprintf("[get-payment-details-uc] Database query error for %s: %v", paymentID, err))
		return nil, fmt.Errorf("database query error: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-payment-details-uc] Payment %s retrieved successfully. Type: %s.", paymentID, p.PaymentType))
	// 6. Return the retrieved payment details
	return p, nil
}
