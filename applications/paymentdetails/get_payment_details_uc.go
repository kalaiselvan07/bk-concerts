package paymentdetails

import (
	"database/sql"
	"fmt"

	"bk-concerts/db" // Assumes this is the package path for your database connection

	"github.com/google/uuid"
)

// GetPayment retrieves a single payment record from the database by its ID.
func GetPayment(paymentID string) (*PaymentDetails, error) {
	// 1. Validate and convert the ID string to uuid.UUID
	id, err := uuid.Parse(paymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID format: %w", err)
	}

	// 2. Define the SQL query
	const selectSQL = `
		SELECT payment_id, payment_type, details, notes
		FROM payment
		WHERE payment_id = $1`

	// 3. Execute the query
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
			// No payment was found
			return nil, fmt.Errorf("payment with ID %s not found", paymentID)
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// 6. Return the retrieved payment details
	return p, nil
}
