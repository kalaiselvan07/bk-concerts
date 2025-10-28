package paymentdetails

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db" // Using the correct module path

	"github.com/google/uuid"
)

// CreatePaymentParams is used for the creation payload.
type CreatePaymentParams struct {
	PaymentType string `json:"paymentType" validate:"required"`
	Details     string `json:"details" validate:"required"`
	Notes       string `json:"notes,omitempty"`
}

// AddPayment handles the creation of a new payment record in the database.
func AddPayment(payload []byte) (*PaymentDetails, error) {
	var p CreatePaymentParams
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	pt := &PaymentDetails{
		PaymentID:   uuid.New().String(),
		PaymentType: p.PaymentType,
		Details:     p.Details,
		Notes:       p.Notes,
	}

	const insertSQL = `
		INSERT INTO payment (payment_id, payment_type, details, notes) 
		VALUES ($1, $2, $3, $4)`

	_, err := db.DB.Exec(
		insertSQL,
		pt.PaymentID,
		pt.PaymentType,
		pt.Details,
		pt.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert payment into database: %w", err)
	}

	// Return the created PaymentDetails object
	return pt, nil
}
