package paymentdetails

import (
	"fmt"

	"bk-concerts/db" // Assumes this is the package path for your database connection
)

// GetAllPayments retrieves a slice of all payment records from the database.
func GetAllPayments() ([]*PaymentDetails, error) {
	// 1. Define the SQL query
	const selectAllSQL = `
		SELECT payment_id, payment_type, details, notes
		FROM payment
		ORDER BY payment_id` // Ordering by ID for consistency

	// 2. Execute the query
	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close() // Ensure the result set is closed

	// 3. Initialize the slice
	payments := make([]*PaymentDetails, 0)

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
			return nil, fmt.Errorf("error scanning payment row: %w", err)
		}

		// Add the payment to the slice
		payments = append(payments, p)
	}

	// 5. Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	// 6. Return the slice of payments
	return payments, nil
}
