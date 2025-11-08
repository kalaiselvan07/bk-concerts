package paymentdetails

import (
	"fmt"
	// Added for rows.Err()
	"supra/db"     // Assumes this is the package path for your database connection
	"supra/logger" // ⬅️ Assuming this import path
	// NOTE: PaymentDetails struct is assumed here
)

// GetAllPayments retrieves a slice of all payment records from the database.
func GetAllPayments() ([]*PaymentDetails, error) {
	logger.Log.Info("[get-all-payment-details-uc] Starting retrieval of all payment records.")

	// 1. Define the SQL query
	const selectAllSQL = `
		SELECT payment_id, payment_type, details, notes
		FROM payment
		ORDER BY payment_id`

	// 2. Execute the query
	logger.Log.Info("[get-all-payment-details-uc] Executing SELECT all query.")
	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-payment-details-uc] Database query failed: %v", err))
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close() // Ensure the result set is closed

	// 3. Initialize the slice
	payments := make([]*PaymentDetails, 0)
	recordCount := 0

	// 4. Iterate through the results
	for rows.Next() {
		p := &PaymentDetails{}

		// Scan the row data
		err := rows.Scan(
			&p.PaymentID,
			&p.PaymentType,
			&p.Details,
			&p.Notes,
		)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[get-all-payment-details-uc] Error scanning payment row: %v", err))
			return nil, fmt.Errorf("error scanning payment row: %w", err)
		}

		// Add the payment to the slice
		payments = append(payments, p)
		recordCount++
	}

	// 5. Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-payment-details-uc] Error during row iteration: %v", err))
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-all-payment-details-uc] Successfully retrieved %d payment records.", recordCount))
	// 6. Return the slice of payments
	return payments, nil
}
